package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"
)

// NodeService はノードサービスのインターフェース
// nodeパッケージのNode構造体に依存するためにインターフェースを定義
type NodeService interface {
	// Chain operations
	GetChain() []*Block
	GetChainLen() int
	ReceiveBlock(b *Block) error

	// Transaction operations
	ProposeTransaction(data *TransactionData) error
	ApproveTransaction(id string) (*Block, error)
	ListPending() []*PendingTransaction
	GetPending(id string) *PendingTransaction

	// Transaction rejection
	RejectTransaction(id string) error

	// Registration
	RegisterNode(nodeName, nickName, address, publicKey string) (*Block, error)

	// Peer operations
	GetPeers() map[string]*NodeInfo

	// Node info
	GetNodeName() string

	// Broadcast
	BroadcastBlock(b *Block)
}

// Block はブロックチェーンの1つのブロックを表す（core.Blockのエイリアス）
type Block struct {
	Header  BlockHeader  `json:"header"`
	Payload BlockPayload `json:"payload"`
}

// BlockHeader はブロックのヘッダーを表す
type BlockHeader struct {
	Index     int    `json:"index"`
	CreatedAt int64  `json:"created_at"`
	PrevHash  string `json:"prev_hash"`
	Hash      string `json:"hash"`
}

// BlockPayload はブロックのペイロードを表す
type BlockPayload struct {
	Type          string           `json:"type"`
	Transaction   *TransactionData `json:"transaction,omitempty"`
	AddNode       *AddNodeData     `json:"add_node,omitempty"`
	FromSignature string           `json:"from_signature"`
	ToSignature   string           `json:"to_signature"`
}

// TransactionData は金銭的取引のデータを表す
type TransactionData struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Amount int64  `json:"amount"`
	Title  string `json:"title"`
}

// AddNodeData はノード追加のデータを表す
type AddNodeData struct {
	PublicKey string `json:"public_key"`
	NodeName  string `json:"node_name"`
	NickName  string `json:"nick_name"`
	Address   string `json:"address"`
}

// PendingTransaction は承認待ちのトランザクションを表す
type PendingTransaction struct {
	Transaction *TransactionData `json:"transaction"`
	FromSig     string           `json:"from_sig"`
	ID          string           `json:"id"`
}

// NodeInfo はピアノードの情報を表す
type NodeInfo struct {
	Name      string `json:"name"`
	NickName  string `json:"nick_name"`
	Address   string `json:"address"`
	PublicKey string `json:"public_key"`
}

// Server はHTTPサーバーを表す
type Server struct {
	node       NodeService
	httpServer *http.Server
	addr       string
	mu         sync.Mutex
}

// NewServer は新しいサーバーを作成する
func NewServer(addr string, node NodeService) *Server {
	s := &Server{
		addr: addr,
		node: node,
	}

	mux := http.NewServeMux()

	// Go 1.22+ のパターン構文を使用
	mux.HandleFunc("GET /chain", s.handleGetChain)
	mux.HandleFunc("POST /block", s.handleReceiveBlock)
	mux.HandleFunc("POST /transaction/propose", s.handlePropose)
	mux.HandleFunc("POST /transaction/approve", s.handleApprove)
	mux.HandleFunc("POST /transaction/reject", s.handleReject)
	mux.HandleFunc("GET /transaction/pending", s.handleGetPending)
	mux.HandleFunc("POST /register", s.handleRegister)
	mux.HandleFunc("GET /peers", s.handleGetPeers)

	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return s
}

// Start はサーバーを起動する
func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.httpServer.Addr)
	if err != nil {
		return err
	}
	fmt.Printf("Server starting on %s\n", ln.Addr().String())
	return s.httpServer.Serve(ln)
}

// Stop はサーバーを停止する
func (s *Server) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
