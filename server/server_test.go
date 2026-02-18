package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// mockNodeService はテスト用のモック実装
type mockNodeService struct {
	chain       []*Block
	pending     []*PendingTransaction
	peers       map[string]*NodeInfo
	nodeName    string
	proposeErr  error
	approveErr  error
	receiveErr  error
	registerErr error

	proposeCalled  bool
	approveCalled  bool
	rejectCalled   bool
	registerCalled bool
	receiveCalled  bool
	rejectErr      error
	broadcastBlock *Block
}

func (m *mockNodeService) GetChain() []*Block {
	return m.chain
}

func (m *mockNodeService) GetChainLen() int {
	return len(m.chain)
}

func (m *mockNodeService) ReceiveBlock(b *Block) error {
	m.receiveCalled = true
	if m.receiveErr != nil {
		return m.receiveErr
	}
	m.chain = append(m.chain, b)
	return nil
}

func (m *mockNodeService) ProposeTransaction(data *TransactionData, fromSignature string) error {
	m.proposeCalled = true
	return m.proposeErr
}

func (m *mockNodeService) ApproveTransaction(id string) (*Block, error) {
	m.approveCalled = true
	if m.approveErr != nil {
		return nil, m.approveErr
	}
	block := &Block{
		Header: BlockHeader{
			Index:     1,
			CreatedAt: time.Now().Unix(),
			PrevHash:  "prev-hash",
			Hash:      "test-block-hash",
		},
		Payload: BlockPayload{
			Type: "transaction",
			Transaction: &TransactionData{
				From:   "alice",
				To:     "bob",
				Amount: 1000,
				Title:  "Test transaction",
			},
		},
	}
	return block, nil
}

func (m *mockNodeService) ListPending() []*PendingTransaction {
	return m.pending
}

func (m *mockNodeService) GetPending(id string) *PendingTransaction {
	for _, p := range m.pending {
		if p.ID == id {
			return p
		}
	}
	return nil
}

func (m *mockNodeService) RejectTransaction(id string) error {
	m.rejectCalled = true
	return m.rejectErr
}

func (m *mockNodeService) RegisterNode(nodeName, nickName, address, publicKey string) (*Block, error) {
	m.registerCalled = true
	if m.registerErr != nil {
		return nil, m.registerErr
	}
	block := &Block{
		Header: BlockHeader{
			Index:     1,
			CreatedAt: time.Now().Unix(),
			PrevHash:  "prev-hash",
			Hash:      "register-block-hash",
		},
		Payload: BlockPayload{
			Type: "add_node",
			AddNode: &AddNodeData{
				PublicKey: publicKey,
				NodeName:  nodeName,
				NickName:  nickName,
				Address:   address,
			},
		},
	}
	return block, nil
}

func (m *mockNodeService) GetPeers() map[string]*NodeInfo {
	return m.peers
}

func (m *mockNodeService) GetNodeName() string {
	return m.nodeName
}

func (m *mockNodeService) BroadcastBlock(b *Block) {
	m.broadcastBlock = b
}

func TestNewServer(t *testing.T) {
	mock := &mockNodeService{
		chain:    []*Block{},
		pending:  []*PendingTransaction{},
		peers:    make(map[string]*NodeInfo),
		nodeName: "test-node",
	}

	server := NewServer(":8080", mock)
	if server == nil {
		t.Fatal("NewServer returned nil")
	}
	if server.addr != ":8080" {
		t.Errorf("Expected addr :8080, got %s", server.addr)
	}
	if server.node != mock {
		t.Error("Expected node to be set to mock")
	}
}

func TestHandleGetChain(t *testing.T) {
	mockChain := []*Block{
		{
			Header: BlockHeader{
				Index:     0,
				CreatedAt: 1234567890,
				PrevHash:  "",
				Hash:      "genesis-hash",
			},
			Payload: BlockPayload{
				Type: "add_node",
				AddNode: &AddNodeData{
					PublicKey: "pub-key",
					NodeName:  "genesis",
					NickName:  "Genesis Node",
					Address:   "localhost",
				},
			},
		},
	}

	mock := &mockNodeService{
		chain:    mockChain,
		pending:  []*PendingTransaction{},
		peers:    make(map[string]*NodeInfo),
		nodeName: "test-node",
	}

	server := NewServer(":8080", mock)

	req := httptest.NewRequest("GET", "/chain", nil)
	w := httptest.NewRecorder()
	server.handleGetChain(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var chain []*Block
	if err := json.NewDecoder(w.Body).Decode(&chain); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(chain) != 1 {
		t.Fatalf("Expected 1 block, got %d", len(chain))
	}

	if chain[0].Header.Hash != "genesis-hash" {
		t.Errorf("Expected hash genesis-hash, got %s", chain[0].Header.Hash)
	}
}

func TestHandleReceiveBlock(t *testing.T) {
	mock := &mockNodeService{
		chain:    []*Block{},
		pending:  []*PendingTransaction{},
		peers:    make(map[string]*NodeInfo),
		nodeName: "test-node",
	}

	server := NewServer(":8080", mock)

	block := Block{
		Header: BlockHeader{
			Index:     1,
			CreatedAt: time.Now().Unix(),
			PrevHash:  "prev-hash",
			Hash:      "test-hash",
		},
		Payload: BlockPayload{
			Type: "transaction",
			Transaction: &TransactionData{
				From:   "alice",
				To:     "bob",
				Amount: 1000,
				Title:  "Test",
			},
		},
	}

	blockJSON, _ := json.Marshal(block)
	req := httptest.NewRequest("POST", "/block", bytes.NewBuffer(blockJSON))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	server.handleReceiveBlock(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if !mock.receiveCalled {
		t.Error("Expected ReceiveBlock to be called")
	}

	if len(mock.chain) != 1 {
		t.Errorf("Expected 1 block in chain, got %d", len(mock.chain))
	}
}

func TestHandleReceiveBlockInvalidJSON(t *testing.T) {
	mock := &mockNodeService{
		chain:    []*Block{},
		pending:  []*PendingTransaction{},
		peers:    make(map[string]*NodeInfo),
		nodeName: "test-node",
	}

	server := NewServer(":8080", mock)

	req := httptest.NewRequest("POST", "/block", nil)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	server.handleReceiveBlock(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHandlePropose(t *testing.T) {
	mock := &mockNodeService{
		chain:    []*Block{},
		pending:  []*PendingTransaction{},
		peers:    make(map[string]*NodeInfo),
		nodeName: "test-node",
	}

	server := NewServer(":8080", mock)

	reqBody := map[string]any{
		"from":           "alice",
		"to":             "bob",
		"amount":         1000,
		"title":          "飲み会代",
		"from_signature": "signature123",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/transaction/propose", nil)
	// Fix request body
	buf := bytes.NewBuffer(body)
	req = httptest.NewRequest("POST", "/transaction/propose", buf)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	server.handlePropose(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if !mock.proposeCalled {
		t.Error("Expected ProposeTransaction to be called")
	}

	var resp struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Status != "proposed" {
		t.Errorf("Expected status 'proposed', got '%s'", resp.Status)
	}
}

func TestHandleProposeInvalidJSON(t *testing.T) {
	mock := &mockNodeService{
		chain:    []*Block{},
		pending:  []*PendingTransaction{},
		peers:    make(map[string]*NodeInfo),
		nodeName: "test-node",
	}

	server := NewServer(":8080", mock)

	req := httptest.NewRequest("POST", "/transaction/propose", nil)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	server.handlePropose(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHandleApprove(t *testing.T) {
	mock := &mockNodeService{
		chain:    []*Block{},
		pending:  []*PendingTransaction{},
		peers:    make(map[string]*NodeInfo),
		nodeName: "test-node",
	}

	server := NewServer(":8080", mock)

	reqBody := map[string]string{
		"id": "uuid-xxx",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/transaction/approve", nil)
	buf := bytes.NewBuffer(body)
	req = httptest.NewRequest("POST", "/transaction/approve", buf)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	server.handleApprove(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if !mock.approveCalled {
		t.Error("Expected ApproveTransaction to be called")
	}

	if mock.broadcastBlock == nil {
		t.Error("Expected block to be broadcasted")
	}

	var resp struct {
		Status string `json:"status"`
		Block  *Block `json:"block"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Status != "approved" {
		t.Errorf("Expected status 'approved', got '%s'", resp.Status)
	}

	if resp.Block == nil {
		t.Error("Expected block in response")
	}
}

func TestHandleGetPending(t *testing.T) {
	pending := []*PendingTransaction{
		{
			ID: "uuid-1",
			Transaction: &TransactionData{
				From:   "alice",
				To:     "bob",
				Amount: 1000,
				Title:  "Test",
			},
			FromSig: "sig123",
		},
	}

	mock := &mockNodeService{
		chain:    []*Block{},
		pending:  pending,
		peers:    make(map[string]*NodeInfo),
		nodeName: "test-node",
	}

	server := NewServer(":8080", mock)

	req := httptest.NewRequest("GET", "/transaction/pending", nil)
	w := httptest.NewRecorder()
	server.handleGetPending(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result []*PendingTransaction
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("Expected 1 pending transaction, got %d", len(result))
	}

	if result[0].ID != "uuid-1" {
		t.Errorf("Expected ID 'uuid-1', got '%s'", result[0].ID)
	}
}

func TestHandleRegister(t *testing.T) {
	mock := &mockNodeService{
		chain:    []*Block{},
		pending:  []*PendingTransaction{},
		peers:    make(map[string]*NodeInfo),
		nodeName: "test-node",
	}

	server := NewServer(":8080", mock)

	reqBody := map[string]string{
		"node_name":  "alice",
		"nick_name":  "アリス",
		"address":    "10.0.0.1",
		"public_key": "pub-key-123",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/register", nil)
	buf := bytes.NewBuffer(body)
	req = httptest.NewRequest("POST", "/register", buf)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	server.handleRegister(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if !mock.registerCalled {
		t.Error("Expected RegisterNode to be called")
	}

	if mock.broadcastBlock == nil {
		t.Error("Expected block to be broadcasted")
	}

	var resp struct {
		Status string `json:"status"`
		Block  *Block `json:"block"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Status != "registered" {
		t.Errorf("Expected status 'registered', got '%s'", resp.Status)
	}
}

func TestHandleRegisterInvalidJSON(t *testing.T) {
	mock := &mockNodeService{
		chain:    []*Block{},
		pending:  []*PendingTransaction{},
		peers:    make(map[string]*NodeInfo),
		nodeName: "test-node",
	}

	server := NewServer(":8080", mock)

	req := httptest.NewRequest("POST", "/register", nil)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	server.handleRegister(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHandleGetPeers(t *testing.T) {
	peers := map[string]*NodeInfo{
		"alice": {
			Name:      "alice",
			NickName:  "アリス",
			Address:   "10.0.0.1",
			PublicKey: "pub-key-alice",
		},
		"bob": {
			Name:      "bob",
			NickName:  "ボブ",
			Address:   "10.0.0.2",
			PublicKey: "pub-key-bob",
		},
	}

	mock := &mockNodeService{
		chain:    []*Block{},
		pending:  []*PendingTransaction{},
		peers:    peers,
		nodeName: "test-node",
	}

	server := NewServer(":8080", mock)

	req := httptest.NewRequest("GET", "/peers", nil)
	w := httptest.NewRecorder()
	server.handleGetPeers(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result map[string]*NodeInfo
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("Expected 2 peers, got %d", len(result))
	}

	if result["alice"].NickName != "アリス" {
		t.Errorf("Expected nick name 'アリス', got '%s'", result["alice"].NickName)
	}
}

func TestServerStartAndStop(t *testing.T) {
	mock := &mockNodeService{
		chain:    []*Block{},
		pending:  []*PendingTransaction{},
		peers:    make(map[string]*NodeInfo),
		nodeName: "test-node",
	}

	server := NewServer(":0", mock) // Use :0 to get a random port

	// Start server in background
	go func() {
		_ = server.Start()
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Stop server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Stop(ctx); err != nil {
		t.Errorf("Failed to stop server: %v", err)
	}
}
