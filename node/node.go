package node

import (
	"bytes"
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"signet/config"
	"signet/core"
	"signet/crypto"
	"signet/p2p"
	"signet/server"
	"signet/storage"
	"sync"
	"time"
)

// httpClient はタイムアウト付きHTTPクライアント
var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

// Node は全コンポーネントを統合するノード構造体
type Node struct {
	Config       *config.Config
	Chain        *core.Chain
	PendingPool  *core.PendingPool
	BlockStore   *storage.BlockStore
	NodeStore    *storage.NodeStore
	PendingStore *storage.PendingStore
	PrivKey      ed25519.PrivateKey
	PubKey       ed25519.PublicKey
	broadcastLock sync.Mutex
}

// NewNode は新しいノードを作成・初期化する
func NewNode(cfg *config.Config) (*Node, error) {
	// 秘密鍵読み込み
	privKey, err := crypto.LoadPrivateKey(cfg.PrivKeyPath())
	if err != nil {
		return nil, fmt.Errorf("failed to load private key: %w", err)
	}

	// 公開鍵を取得
	pubKey := privKey.Public().(ed25519.PublicKey)

	// ストレージ初期化
	blockStore := storage.NewBlockStore(cfg.BlockFilePath())
	nodeStore := storage.NewNodeStore(cfg.NodesDir())
	pendingStore := storage.NewPendingStore(cfg.PendingFilePath())

	// ブロックチェーン読み込み
	blocks, err := blockStore.LoadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to load blocks: %w", err)
	}

	var chain *core.Chain
	if len(blocks) == 0 {
		// ブロックがなければジェネシスブロックで初期化（フォールバック）
		chain = core.NewChain()
	} else {
		// ストレージのブロックからチェーンを直接構築（ジェネシス二重生成を防止）
		var chainErr error
		chain, chainErr = core.NewChainFromBlocks(blocks)
		if chainErr != nil {
			return nil, fmt.Errorf("failed to build chain from blocks: %w", chainErr)
		}
	}

	// 承認待ちトランザクション読み込み
	pendingItems, err := pendingStore.Load()
	if err != nil {
		log.Printf("Warning: failed to load pending transactions: %v", err)
		pendingItems = []*core.PendingTransaction{}
	}

	pendingPool := core.NewPendingPool()
	for _, item := range pendingItems {
		pendingPool.Add(item)
	}

	return &Node{
		Config:       cfg,
		Chain:        chain,
		PendingPool:  pendingPool,
		BlockStore:   blockStore,
		NodeStore:    nodeStore,
		PendingStore: pendingStore,
		PrivKey:      privKey,
		PubKey:       pubKey,
	}, nil
}

// GetChain はチェーンを返す（server.NodeServiceインターフェース実装）
func (n *Node) GetChain() []*server.Block {
	blocks := n.Chain.GetBlocks()
	result := make([]*server.Block, len(blocks))
	for i, b := range blocks {
		result[i] = convertBlockToServer(b)
	}
	return result
}

// GetChainLen はチェーンの長さを返す
func (n *Node) GetChainLen() int {
	return n.Chain.Len()
}

// verifyBlockSignatures はトランザクションブロックの署名を暗号学的に検証する
func (n *Node) verifyBlockSignatures(block *core.Block) error {
	if block.Payload.Type != "transaction" {
		return nil // add_node ブロックには署名不要
	}

	txData, err := block.GetTransactionData()
	if err != nil {
		return fmt.Errorf("failed to get transaction data: %w", err)
	}

	txDataBytes, err := json.Marshal(txData)
	if err != nil {
		return fmt.Errorf("failed to marshal transaction data: %w", err)
	}

	peers, err := n.NodeStore.LoadAll()
	if err != nil {
		return fmt.Errorf("failed to load peers for signature verification: %w", err)
	}

	// From 署名検証
	if block.Payload.FromSignature == "" {
		return fmt.Errorf("missing from signature")
	}
	fromPeer, ok := peers[txData.From]
	if !ok {
		return fmt.Errorf("unknown from node: %s", txData.From)
	}
	fromPubKey, err := crypto.HexToPublicKey(fromPeer.PublicKey)
	if err != nil {
		return fmt.Errorf("failed to decode from node's public key: %w", err)
	}
	if !crypto.Verify(fromPubKey, txDataBytes, block.Payload.FromSignature) {
		return fmt.Errorf("invalid from signature")
	}

	// To 署名検証
	if block.Payload.ToSignature == "" {
		return fmt.Errorf("missing to signature")
	}
	toPeer, ok := peers[txData.To]
	if !ok {
		return fmt.Errorf("unknown to node: %s", txData.To)
	}
	toPubKey, err := crypto.HexToPublicKey(toPeer.PublicKey)
	if err != nil {
		return fmt.Errorf("failed to decode to node's public key: %w", err)
	}
	if !crypto.Verify(toPubKey, txDataBytes, block.Payload.ToSignature) {
		return fmt.Errorf("invalid to signature")
	}

	return nil
}

// ReceiveBlock はブロックを受信してチェーンに追加する
func (n *Node) ReceiveBlock(b *server.Block) error {
	coreBlock := convertServerToBlock(b)

	// ハッシュ再計算チェック
	if err := core.ValidateBlock(coreBlock); err != nil {
		return fmt.Errorf("block validation failed: %w", err)
	}

	// 署名検証
	if err := n.verifyBlockSignatures(coreBlock); err != nil {
		return fmt.Errorf("signature verification failed: %w", err)
	}

	lastHash := n.Chain.GetLastHash()
	lastIndex := n.Chain.GetLastIndex()

	// PrevHash整合性チェック
	if coreBlock.Header.PrevHash == lastHash {
		// 自分の末尾と一致→追加
		if err := n.Chain.AddBlock(coreBlock); err != nil {
			return fmt.Errorf("failed to add block: %w", err)
		}
		// 永続化
		if err := n.BlockStore.Append(coreBlock); err != nil {
			return fmt.Errorf("failed to persist block: %w", err)
		}
		// ブロードキャスト
		go n.BroadcastBlock(b)
		return nil
	}

	// Indexが大きい→同期
	if coreBlock.Header.Index > lastIndex {
		return fmt.Errorf("block index %d is ahead of our chain %d, sync needed", coreBlock.Header.Index, lastIndex)
	}

	// Index以下→無視（既に持っているか、競合）
	if n.Chain.HasBlock(coreBlock.Header.Hash) {
		return nil // 重複ブロックは無視
	}

	return fmt.Errorf("block index %d is behind or equal to our chain %d", coreBlock.Header.Index, lastIndex)
}

// ProposeTransaction はトランザクションを提案する
// fromSignature が空の場合は自ノードの秘密鍵で自動署名する（ローカル提案）
// fromSignature が指定されている場合はそのまま使用する（他ノードからの転送）
func (n *Node) ProposeTransaction(data *server.TransactionData, fromSignature string) error {
	// 署名用ペイロード作成
	txData := &core.TransactionData{
		From:   data.From,
		To:     data.To,
		Amount: data.Amount,
		Title:  data.Title,
	}

	// TransactionDataをJSONに変換
	txDataBytes, err := json.Marshal(txData)
	if err != nil {
		return fmt.Errorf("failed to marshal transaction data: %w", err)
	}

	// From側の署名（未指定の場合は自動生成）
	if fromSignature == "" {
		fromSignature = crypto.Sign(n.PrivKey, txDataBytes)
	}

	// BlockPayload作成
	payload := core.BlockPayload{
		Type:          "transaction",
		Data:          txDataBytes,
		FromSignature: fromSignature,
		ToSignature:   "",
	}

	// ID生成
	id := core.GenerateID(payload, time.Now().UTC())

	// PendingTransaction作成
	pendingTx := core.NewPendingTransaction(id, payload)

	// プールに追加
	n.PendingPool.Add(pendingTx)

	// 永続化
	items := n.PendingPool.List()
	if err := n.PendingStore.Save(items); err != nil {
		log.Printf("Warning: failed to save pending transaction: %v", err)
	}

	// Toノードが別ノードの場合は送信
	if data.To != n.Config.NodeName {
		peers, err := n.NodeStore.LoadAll()
		if err == nil {
			if peer, exists := peers[data.To]; exists {
				go n.sendProposeTransaction(peer.Address, pendingTx)
			}
		}
	}

	return nil
}

// sendProposeTransaction は指定したアドレスにトランザクション提案を送信する
func (n *Node) sendProposeTransaction(addr string, tx *core.PendingTransaction) error {
	txData, err := tx.GetTransactionData()
	if err != nil {
		return fmt.Errorf("failed to get transaction data: %w", err)
	}

	reqBody := struct {
		From          string `json:"from"`
		To            string `json:"to"`
		Amount        int64  `json:"amount"`
		Title         string `json:"title"`
		FromSignature string `json:"from_signature"`
	}{
		From:          txData.From,
		To:            txData.To,
		Amount:        txData.Amount,
		Title:         txData.Title,
		FromSignature: tx.Payload.FromSignature,
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("http://%s/transaction/propose", addr)
	resp, err := httpClient.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	log.Printf("Proposed transaction sent to %s", addr)
	return nil
}

// ApproveTransaction はトランザクションを承認する
func (n *Node) ApproveTransaction(id string) (*server.Block, error) {
	// プールから取得
	pendingTx := n.PendingPool.Get(id)
	if pendingTx == nil {
		return nil, fmt.Errorf("pending transaction not found: %s", id)
	}

	// TransactionDataを取得
	txData, err := pendingTx.GetTransactionData()
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction data: %w", err)
	}

	// 自分（To）の署名を追加（From署名と同じ形式: トランザクションデータに対して署名）
	txDataBytes, err := json.Marshal(txData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal transaction data for signing: %w", err)
	}
	toSignature := crypto.Sign(n.PrivKey, txDataBytes)

	// ブロック生成
	lastBlock := n.Chain.LastBlock()
	prevHash := lastBlock.Header.Hash
	index := lastBlock.Header.Index + 1

	block, err := core.CreateBlockWithTransaction(index, prevHash, txData, pendingTx.Payload.FromSignature, toSignature)
	if err != nil {
		return nil, fmt.Errorf("failed to create block: %w", err)
	}

	// チェーンに追加
	if err := n.Chain.AddBlock(block); err != nil {
		return nil, fmt.Errorf("failed to add block to chain: %w", err)
	}

	// 永続化
	if err := n.BlockStore.Append(block); err != nil {
		return nil, fmt.Errorf("failed to persist block: %w", err)
	}

	// プールから削除
	n.PendingPool.Remove(id)
	items := n.PendingPool.List()
	if err := n.PendingStore.Save(items); err != nil {
		log.Printf("Warning: failed to save pending transactions: %v", err)
	}

	return convertBlockToServer(block), nil
}

// RejectTransaction はトランザクションを拒否する
func (n *Node) RejectTransaction(id string) error {
	// プールから取得
	pendingTx := n.PendingPool.Get(id)
	if pendingTx == nil {
		return fmt.Errorf("pending transaction not found: %s", id)
	}

	// プールから削除
	n.PendingPool.Remove(id)

	// 永続化
	items := n.PendingPool.List()
	if err := n.PendingStore.Save(items); err != nil {
		log.Printf("Warning: failed to save pending transactions: %v", err)
	}

	return nil
}

// ListPending は全承認待ちトランザクションを返す
func (n *Node) ListPending() []*server.PendingTransaction {
	items := n.PendingPool.List()
	result := make([]*server.PendingTransaction, 0, len(items))
	for _, item := range items {
		txData, err := item.GetTransactionData()
		if err != nil {
			continue
		}
		result = append(result, &server.PendingTransaction{
			Transaction: &server.TransactionData{
				From:   txData.From,
				To:     txData.To,
				Amount: txData.Amount,
				Title:  txData.Title,
			},
			FromSig: item.Payload.FromSignature,
			ID:      item.ID,
		})
	}
	return result
}

// GetPending は指定したIDの承認待ちトランザクションを返す
func (n *Node) GetPending(id string) *server.PendingTransaction {
	item := n.PendingPool.Get(id)
	if item == nil {
		return nil
	}
	txData, err := item.GetTransactionData()
	if err != nil {
		return nil
	}
	return &server.PendingTransaction{
		Transaction: &server.TransactionData{
			From:   txData.From,
			To:     txData.To,
			Amount: txData.Amount,
			Title:  txData.Title,
		},
		FromSig: item.Payload.FromSignature,
		ID:      item.ID,
	}
}

// RegisterNode はノードを登録する
func (n *Node) RegisterNode(nodeName, nickName, address, publicKey string) (*server.Block, error) {
	// ブロック生成
	lastBlock := n.Chain.LastBlock()
	prevHash := lastBlock.Header.Hash
	index := lastBlock.Header.Index + 1

	addNodeData := &core.AddNodeData{
		PublicKey: publicKey,
		NodeName:  nodeName,
		NickName:  nickName,
		Address:   address,
	}

	block, err := core.CreateBlockWithAddNode(index, prevHash, addNodeData)
	if err != nil {
		return nil, fmt.Errorf("failed to create block: %w", err)
	}

	// チェーンに追加
	if err := n.Chain.AddBlock(block); err != nil {
		return nil, fmt.Errorf("failed to add block to chain: %w", err)
	}

	// 永続化
	if err := n.BlockStore.Append(block); err != nil {
		return nil, fmt.Errorf("failed to persist block: %w", err)
	}

	// ノードファイル保存
	nodeInfo := &storage.NodeInfo{
		Name:      nodeName,
		NickName:  nickName,
		Address:   address,
		PublicKey: publicKey,
	}
	if err := n.NodeStore.Save(nodeName, nodeInfo); err != nil {
		log.Printf("Warning: failed to save node file: %v", err)
	}

	return convertBlockToServer(block), nil
}

// GetPeers はピアノード情報を返す
func (n *Node) GetPeers() map[string]*server.NodeInfo {
	peers, err := n.NodeStore.LoadAll()
	if err != nil {
		log.Printf("Warning: failed to load peers: %v", err)
		return make(map[string]*server.NodeInfo)
	}

	result := make(map[string]*server.NodeInfo)
	for name, peer := range peers {
		result[name] = &server.NodeInfo{
			Name:      name,
			NickName:  peer.NickName,
			Address:   peer.Address,
			PublicKey: peer.PublicKey,
		}
	}
	return result
}

// GetNodeName は自ノード名を返す
func (n *Node) GetNodeName() string {
	return n.Config.NodeName
}

// BroadcastBlock はブロックを全ピアにブロードキャストする
func (n *Node) BroadcastBlock(b *server.Block) {
	n.broadcastLock.Lock()
	defer n.broadcastLock.Unlock()

	// ピア取得
	peers, err := n.NodeStore.LoadAll()
	if err != nil {
		log.Printf("Warning: failed to load peers for broadcast: %v", err)
		return
	}

	// server.Block をそのまま渡す（受信側も server.Block でデコードする）
	p2p.BroadcastBlock(b, peers, n.Config.NodeName)
}

// SyncChain は全ピアからチェーンを取得し、最長チェーンで同期する
func (n *Node) SyncChain() error {
	peers, err := n.NodeStore.LoadAll()
	if err != nil {
		return fmt.Errorf("failed to load peers: %w", err)
	}

	var longestBlocks []*core.Block
	maxLen := n.Chain.Len()

	for name, peer := range peers {
		if name == n.Config.NodeName {
			continue
		}

		serverBlocks, err := n.fetchChain(peer.Address)
		if err != nil {
			log.Printf("Warning: failed to fetch chain from %s (%s): %v", name, peer.Address, err)
			continue
		}

		// server.Block -> core.Block に変換
		coreBlocks := make([]*core.Block, len(serverBlocks))
		for i, sb := range serverBlocks {
			coreBlocks[i] = convertServerToBlock(sb)
		}

		if len(coreBlocks) > maxLen {
			maxLen = len(coreBlocks)
			longestBlocks = coreBlocks
		}
	}

	// 自分より長いチェーンが見つかった場合は置換
	if longestBlocks != nil && len(longestBlocks) > n.Chain.Len() {
		if err := n.Chain.ReplaceChain(longestBlocks); err != nil {
			return fmt.Errorf("failed to replace chain: %w", err)
		}
		// 永続化
		if err := n.BlockStore.ReplaceAll(longestBlocks); err != nil {
			return fmt.Errorf("failed to persist replaced chain: %w", err)
		}
		log.Printf("Chain synced: %d blocks", len(longestBlocks))
	}

	return nil
}

// fetchChain は指定したアドレスからチェーンを取得する
func (n *Node) fetchChain(addr string) ([]*server.Block, error) {
	url := fmt.Sprintf("http://%s/chain", addr)
	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var blocks []*server.Block
	if err := json.NewDecoder(resp.Body).Decode(&blocks); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return blocks, nil
}

// convertBlockToServer はcore.Blockをserver.Blockに変換する
func convertBlockToServer(b *core.Block) *server.Block {
	serverBlock := &server.Block{
		Header: server.BlockHeader{
			Index:     b.Header.Index,
			CreatedAt: b.Header.CreatedAt.Unix(),
			PrevHash:  b.Header.PrevHash,
			Hash:      b.Header.Hash,
		},
		Payload: server.BlockPayload{
			Type:          b.Payload.Type,
			FromSignature: b.Payload.FromSignature,
			ToSignature:   b.Payload.ToSignature,
		},
	}

	// ペイロードデータをコピー
	if b.Payload.Type == "transaction" {
		if txData, err := b.GetTransactionData(); err == nil {
			serverBlock.Payload.Transaction = &server.TransactionData{
				From:   txData.From,
				To:     txData.To,
				Amount: txData.Amount,
				Title:  txData.Title,
			}
		}
	} else if b.Payload.Type == "add_node" {
		if addNodeData, err := b.GetAddNodeData(); err == nil {
			serverBlock.Payload.AddNode = &server.AddNodeData{
				PublicKey: addNodeData.PublicKey,
				NodeName:  addNodeData.NodeName,
				NickName:  addNodeData.NickName,
				Address:   addNodeData.Address,
			}
		}
	}

	return serverBlock
}

// convertServerToBlock はserver.Blockをcore.Blockに変換する
func convertServerToBlock(b *server.Block) *core.Block {
	coreBlock := &core.Block{
		Header: core.BlockHeader{
			Index:     b.Header.Index,
			CreatedAt: time.Unix(b.Header.CreatedAt, 0).UTC(),
			PrevHash:  b.Header.PrevHash,
			Hash:      b.Header.Hash,
		},
		Payload: core.BlockPayload{
			Type:          b.Payload.Type,
			FromSignature: b.Payload.FromSignature,
			ToSignature:   b.Payload.ToSignature,
		},
	}

	// ペイロードデータをコピー
	if b.Payload.Transaction != nil {
		txData := &core.TransactionData{
			From:   b.Payload.Transaction.From,
			To:     b.Payload.Transaction.To,
			Amount: b.Payload.Transaction.Amount,
			Title:  b.Payload.Transaction.Title,
		}
		if data, err := core.SetTransactionData(txData); err == nil {
			coreBlock.Payload.Data = data
		}
	} else if b.Payload.AddNode != nil {
		addNodeData := &core.AddNodeData{
			PublicKey: b.Payload.AddNode.PublicKey,
			NodeName:  b.Payload.AddNode.NodeName,
			NickName:  b.Payload.AddNode.NickName,
			Address:   b.Payload.AddNode.Address,
		}
		if data, err := core.SetAddNodeData(addNodeData); err == nil {
			coreBlock.Payload.Data = data
		}
	}

	return coreBlock
}
