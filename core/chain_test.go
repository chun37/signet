package core

import (
	"fmt"
	"testing"
)

func TestNewChain(t *testing.T) {
	chain := NewChain(&AddNodeData{})

	if chain.Len() != 1 {
		t.Errorf("NewChain length = %d, want 1", chain.Len())
	}

	genesis := chain.LastBlock()
	if genesis == nil {
		t.Fatal("LastBlock returned nil")
	}

	if !genesis.IsGenesisBlock() {
		t.Error("First block is not genesis block")
	}
}

func TestAddBlock(t *testing.T) {
	chain := NewChain(&AddNodeData{})

	tx := &TransactionData{
		From:   "node1",
		To:     "node2",
		Amount: 1000,
		Title:  "test",
	}

	block, err := CreateBlockWithTransaction(1, chain.GetLastHash(), tx, "sig1", "sig2")
	if err != nil {
		t.Fatalf("CreateBlockWithTransaction failed: %v", err)
	}

	if err := chain.AddBlock(block); err != nil {
		t.Fatalf("AddBlock failed: %v", err)
	}

	if chain.Len() != 2 {
		t.Errorf("Chain length = %d, want 2", chain.Len())
	}

	lastBlock := chain.LastBlock()
	if lastBlock.Header.Index != 1 {
		t.Errorf("Last block index = %d, want 1", lastBlock.Header.Index)
	}
}

func TestAddBlock_InvalidPrevHash(t *testing.T) {
	chain := NewChain(&AddNodeData{})

	tx := &TransactionData{From: "a", To: "b", Amount: 100, Title: "test"}
	block, _ := CreateBlockWithTransaction(1, "wronghash", tx, "sig1", "sig2")

	err := chain.AddBlock(block)
	if err == nil {
		t.Error("Expected error for invalid prev_hash, got nil")
	}
}

func TestAddBlock_Duplicate(t *testing.T) {
	chain := NewChain(&AddNodeData{})

	tx := &TransactionData{From: "a", To: "b", Amount: 100, Title: "test"}
	block, _ := CreateBlockWithTransaction(1, chain.GetLastHash(), tx, "sig1", "sig2")

	chain.AddBlock(block)

	// 同じブロックを再度追加
	err := chain.AddBlock(block)
	if err == nil {
		t.Error("Expected error for duplicate block, got nil")
	}
}

func TestAddBlock_InvalidHash(t *testing.T) {
	chain := NewChain(&AddNodeData{})

	tx := &TransactionData{From: "a", To: "b", Amount: 100, Title: "test"}
	data, _ := SetTransactionData(tx)

	block := &Block{
		Header: BlockHeader{
			Index:     1,
			CreatedAt: chain.LastBlock().Header.CreatedAt,
			PrevHash:  chain.GetLastHash(),
			Hash:      "invalidhash",
		},
		Payload: BlockPayload{
			Type:          "transaction",
			Data:          data,
			FromSignature: "sig1",
			ToSignature:   "sig2",
		},
	}

	err := chain.AddBlock(block)
	if err == nil {
		t.Error("Expected error for invalid hash, got nil")
	}
}

func TestGetBlocks(t *testing.T) {
	chain := NewChain(&AddNodeData{})

	blocks := chain.GetBlocks()
	if len(blocks) != 1 {
		t.Errorf("GetBlocks length = %d, want 1", len(blocks))
	}

	// 返されたスライスを修改しても元に影響しないことを確認
	blocks[0] = nil

	blocks2 := chain.GetBlocks()
	if blocks2[0] == nil {
		t.Error("Modifying returned slice affected original chain")
	}
}

func TestLastBlock(t *testing.T) {
	chain := NewChain(&AddNodeData{})

	last := chain.LastBlock()
	if last == nil {
		t.Fatal("LastBlock returned nil for non-empty chain")
	}

	// ブロックを追加
	tx := &TransactionData{From: "a", To: "b", Amount: 100, Title: "test"}
	block, _ := CreateBlockWithTransaction(1, chain.GetLastHash(), tx, "sig1", "sig2")
	chain.AddBlock(block)

	last2 := chain.LastBlock()
	if last2.Header.Index != 1 {
		t.Errorf("LastBlock index = %d, want 1", last2.Header.Index)
	}
}

func TestLen(t *testing.T) {
	chain := NewChain(&AddNodeData{})

	if chain.Len() != 1 {
		t.Errorf("Len = %d, want 1", chain.Len())
	}

	// 複数ブロックを追加
	for i := 0; i < 5; i++ {
		tx := &TransactionData{From: "a", To: "b", Amount: 100, Title: "test"}
		block, _ := CreateBlockWithTransaction(i+1, chain.GetLastHash(), tx, "sig1", "sig2")
		chain.AddBlock(block)
	}

	if chain.Len() != 6 {
		t.Errorf("Len = %d, want 6", chain.Len())
	}
}

func TestValidateChain_ValidChain(t *testing.T) {
	chain := NewChain(&AddNodeData{})

	// 複数ブロックを追加
	for i := 0; i < 5; i++ {
		tx := &TransactionData{From: "a", To: "b", Amount: 100, Title: "test"}
		block, _ := CreateBlockWithTransaction(i+1, chain.GetLastHash(), tx, "sig1", "sig2")
		chain.AddBlock(block)
	}

	if err := chain.ValidateChain(); err != nil {
		t.Errorf("ValidateChain failed: %v", err)
	}
}

func TestValidateChain_EmptyChain(t *testing.T) {
	chain := &Chain{
		blocks:  []*Block{},
		hashSet: map[string]struct{}{},
	}

	err := chain.ValidateChain()
	if err == nil {
		t.Error("Expected error for empty chain, got nil")
	}
}

func TestReplaceChain_LongerChain(t *testing.T) {
	chain1 := NewChain(&AddNodeData{})

	// chain1に3ブロック追加
	for i := 0; i < 3; i++ {
		tx := &TransactionData{From: "a", To: "b", Amount: 100, Title: "test"}
		block, _ := CreateBlockWithTransaction(i+1, chain1.GetLastHash(), tx, "sig1", "sig2")
		chain1.AddBlock(block)
	}

	// より長いチェーンを作成
	chain2 := NewChain(&AddNodeData{})
	for i := 0; i < 5; i++ {
		tx := &TransactionData{From: "a", To: "b", Amount: 100, Title: "test"}
		block, _ := CreateBlockWithTransaction(i+1, chain2.GetLastHash(), tx, "sig1", "sig2")
		chain2.AddBlock(block)
	}

	// chain1をchain2で置換
	err := chain1.ReplaceChain(chain2.GetBlocks())
	if err != nil {
		t.Fatalf("ReplaceChain failed: %v", err)
	}

	if chain1.Len() != 6 {
		t.Errorf("Chain length = %d, want 6", chain1.Len())
	}
}

func TestReplaceChain_ShorterChain(t *testing.T) {
	chain1 := NewChain(&AddNodeData{})

	// chain1に5ブロック追加
	for i := 0; i < 5; i++ {
		tx := &TransactionData{From: "a", To: "b", Amount: 100, Title: "test"}
		block, _ := CreateBlockWithTransaction(i+1, chain1.GetLastHash(), tx, "sig1", "sig2")
		chain1.AddBlock(block)
	}

	// 短いチェーンを作成
	chain2 := NewChain(&AddNodeData{})
	tx := &TransactionData{From: "a", To: "b", Amount: 100, Title: "test"}
	block, _ := CreateBlockWithTransaction(1, chain2.GetLastHash(), tx, "sig1", "sig2")
	chain2.AddBlock(block)

	// 短いチェーンで置換しようとする
	err := chain1.ReplaceChain(chain2.GetBlocks())
	if err == nil {
		t.Error("Expected error for shorter chain, got nil")
	}
}

func TestReplaceChain_BrokenChain(t *testing.T) {
	chain := NewChain(&AddNodeData{})

	// 不正なチェーンを作成（連結が壊れている）
	genesis := NewGenesisBlock(&AddNodeData{})
	tx := &TransactionData{From: "a", To: "b", Amount: 100, Title: "test"}
	brokenBlock, _ := CreateBlockWithTransaction(2, "wronghash", tx, "sig1", "sig2")

	brokenBlocks := []*Block{genesis, brokenBlock}

	err := chain.ReplaceChain(brokenBlocks)
	if err == nil {
		t.Error("Expected error for broken chain, got nil")
	}
}

func TestHasBlock(t *testing.T) {
	chain := NewChain(&AddNodeData{})

	genesisHash := chain.LastBlock().Header.Hash
	if !chain.HasBlock(genesisHash) {
		t.Error("HasBlock should return true for genesis block")
	}

	if chain.HasBlock("nonexistent") {
		t.Error("HasBlock should return false for non-existent block")
	}

	// ブロック追加
	tx := &TransactionData{From: "a", To: "b", Amount: 100, Title: "test"}
	block, _ := CreateBlockWithTransaction(1, chain.GetLastHash(), tx, "sig1", "sig2")
	chain.AddBlock(block)

	if !chain.HasBlock(block.Header.Hash) {
		t.Error("HasBlock should return true for added block")
	}
}

func TestGetBlockByIndex(t *testing.T) {
	chain := NewChain(&AddNodeData{})

	// ジェネシスブロック
	block, err := chain.GetBlockByIndex(0)
	if err != nil {
		t.Fatalf("GetBlockByIndex(0) failed: %v", err)
	}
	if block.Header.Index != 0 {
		t.Errorf("Block index = %d, want 0", block.Header.Index)
	}

	// 存在しないインデックス
	_, err = chain.GetBlockByIndex(100)
	if err == nil {
		t.Error("Expected error for out of range index, got nil")
	}

	// 負のインデックス
	_, err = chain.GetBlockByIndex(-1)
	if err == nil {
		t.Error("Expected error for negative index, got nil")
	}
}

func TestGetBlockByHash(t *testing.T) {
	chain := NewChain(&AddNodeData{})

	genesis := chain.LastBlock()
	block, err := chain.GetBlockByHash(genesis.Header.Hash)
	if err != nil {
		t.Fatalf("GetBlockByHash failed: %v", err)
	}

	if block.Header.Hash != genesis.Header.Hash {
		t.Error("Retrieved block hash does not match")
	}

	// 存在しないハッシュ
	_, err = chain.GetBlockByHash("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent hash, got nil")
	}
}

func TestGetLastHash(t *testing.T) {
	chain := NewChain(&AddNodeData{})

	lastHash := chain.GetLastHash()
	if lastHash == "" {
		t.Error("GetLastHash returned empty string")
	}

	genesisHash := chain.LastBlock().Header.Hash
	if lastHash != genesisHash {
		t.Errorf("GetLastHash = %s, want %s", lastHash, genesisHash)
	}
}

func TestGetLastIndex(t *testing.T) {
	chain := NewChain(&AddNodeData{})

	if chain.GetLastIndex() != 0 {
		t.Errorf("GetLastIndex = %d, want 0", chain.GetLastIndex())
	}

	// ブロック追加
	tx := &TransactionData{From: "a", To: "b", Amount: 100, Title: "test"}
	block, _ := CreateBlockWithTransaction(1, chain.GetLastHash(), tx, "sig1", "sig2")
	chain.AddBlock(block)

	if chain.GetLastIndex() != 1 {
		t.Errorf("GetLastIndex = %d, want 1", chain.GetLastIndex())
	}
}

func TestClone(t *testing.T) {
	chain := NewChain(&AddNodeData{})

	// ブロック追加
	tx := &TransactionData{From: "a", To: "b", Amount: 100, Title: "test"}
	block, _ := CreateBlockWithTransaction(1, chain.GetLastHash(), tx, "sig1", "sig2")
	chain.AddBlock(block)

	cloned := chain.Clone()

	// 同じ長さ
	if cloned.Len() != chain.Len() {
		t.Errorf("Cloned chain length = %d, want %d", cloned.Len(), chain.Len())
	}

	// 同じハッシュ
	if cloned.GetLastHash() != chain.GetLastHash() {
		t.Errorf("Cloned chain last hash = %s, want %s", cloned.GetLastHash(), chain.GetLastHash())
	}

	// 複製を変更しても元に影響しない
	tx2 := &TransactionData{From: "b", To: "c", Amount: 200, Title: "test2"}
	block2, _ := CreateBlockWithTransaction(2, cloned.GetLastHash(), tx2, "sig3", "sig4")
	cloned.AddBlock(block2)

	if chain.Len() != 2 {
		t.Errorf("Original chain length = %d, want 2", chain.Len())
	}
}

func TestForEach(t *testing.T) {
	chain := NewChain(&AddNodeData{})

	// ブロック追加
	for i := 0; i < 3; i++ {
		tx := &TransactionData{From: "a", To: "b", Amount: 100, Title: "test"}
		block, _ := CreateBlockWithTransaction(i+1, chain.GetLastHash(), tx, "sig1", "sig2")
		chain.AddBlock(block)
	}

	count := 0
	err := chain.ForEach(func(b *Block) error {
		count++
		return nil
	})

	if err != nil {
		t.Errorf("ForEach failed: %v", err)
	}

	if count != 4 {
		t.Errorf("ForEach visited %d blocks, want 4", count)
	}

	// エラーが早期リターンされるかテスト
	errorCalled := false
	chain.ForEach(func(b *Block) error {
		if b.Header.Index == 1 {
			errorCalled = true
			return fmt.Errorf("test error")
		}
		return nil
	})

	if !errorCalled {
		t.Error("ForEach did not call function for all blocks")
	}
}
