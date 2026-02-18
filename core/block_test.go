package core

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNewBlock(t *testing.T) {
	txData := &TransactionData{
		From:   "node1",
		To:     "node2",
		Amount: 1000,
		Title:  "test",
	}

	data, err := SetTransactionData(txData)
	if err != nil {
		t.Fatalf("SetTransactionData failed: %v", err)
	}

	payload := BlockPayload{
		Type:          "transaction",
		Data:          data,
		FromSignature: "sig1",
		ToSignature:   "sig2",
	}

	block := NewBlock(1, "prevhash123", payload)

	if block.Header.Index != 1 {
		t.Errorf("Index = %d, want 1", block.Header.Index)
	}
	if block.Header.PrevHash != "prevhash123" {
		t.Errorf("PrevHash = %s, want prevhash123", block.Header.PrevHash)
	}
	if block.Header.Hash == "" {
		t.Error("Hash is empty")
	}
}

func TestNewGenesisBlock(t *testing.T) {
	genesis := NewGenesisBlock(&AddNodeData{})

	if genesis.Header.Index != 0 {
		t.Errorf("Genesis Index = %d, want 0", genesis.Header.Index)
	}
	if genesis.Header.PrevHash != "0" {
		t.Errorf("Genesis PrevHash = %s, want '0'", genesis.Header.PrevHash)
	}
	if genesis.Payload.Type != "add_node" {
		t.Errorf("Genesis Type = %s, want 'add_node'", genesis.Payload.Type)
	}
}

func TestValidateBlock_ValidBlock(t *testing.T) {
	txData := &TransactionData{
		From:   "node1",
		To:     "node2",
		Amount: 1000,
		Title:  "test",
	}

	data, _ := SetTransactionData(txData)
	payload := BlockPayload{
		Type:          "transaction",
		Data:          data,
		FromSignature: "sig1",
		ToSignature:   "sig2",
	}

	block := NewBlock(1, "prevhash", payload)

	err := ValidateBlock(block)
	if err != nil {
		t.Errorf("ValidateBlock failed: %v", err)
	}
}

func TestValidateBlock_InvalidHash(t *testing.T) {
	data, _ := json.Marshal(AddNodeData{})
	block := &Block{
		Header: BlockHeader{
			Index:     1,
			CreatedAt: time.Now().UTC(),
			PrevHash:  "prev",
			Hash:      "invalidhash",
		},
		Payload: BlockPayload{
			Type:          "add_node",
			Data:          json.RawMessage(data),
			FromSignature: "",
			ToSignature:   "",
		},
	}

	err := ValidateBlock(block)
	if err == nil {
		t.Error("Expected error for invalid hash, got nil")
	}
}

func TestValidateBlock_InvalidType(t *testing.T) {
	data, _ := json.Marshal(AddNodeData{})
	block := &Block{
		Header: BlockHeader{
			Index:     1,
			CreatedAt: time.Now().UTC(),
			PrevHash:  "prev",
			Hash:      "hash",
		},
		Payload: BlockPayload{
			Type:          "invalid_type",
			Data:          json.RawMessage(data),
			FromSignature: "",
			ToSignature:   "",
		},
	}

	// ハッシュを再計算して有効にする
	block.Header.Hash = CalcBlockHash(block)

	err := ValidateBlock(block)
	if err == nil {
		t.Error("Expected error for invalid type, got nil")
	}
}

func TestCalcBlockHash_Deterministic(t *testing.T) {
	data, _ := SetTransactionData(&TransactionData{
		From:   "node1",
		To:     "node2",
		Amount: 1000,
		Title:  "test",
	})

	payload := BlockPayload{
		Type:          "transaction",
		Data:          data,
		FromSignature: "sig1",
		ToSignature:   "sig2",
	}

	// 同じ時刻を使用するため固定時刻を設定
	fixedTime := time.Date(2026, 2, 18, 12, 0, 0, 0, time.UTC)
	block1 := &Block{
		Header: BlockHeader{
			Index:     1,
			CreatedAt: fixedTime,
			PrevHash:  "prev",
		},
		Payload: payload,
	}
	block1.Header.Hash = CalcBlockHash(block1)

	block2 := &Block{
		Header: BlockHeader{
			Index:     1,
			CreatedAt: fixedTime,
			PrevHash:  "prev",
		},
		Payload: payload,
	}
	block2.Header.Hash = CalcBlockHash(block2)

	if block1.Header.Hash != block2.Header.Hash {
		t.Errorf("Hash not deterministic: %s != %s", block1.Header.Hash, block2.Header.Hash)
	}
}

func TestGetTransactionData(t *testing.T) {
	txData := &TransactionData{
		From:   "node1",
		To:     "node2",
		Amount: 1000,
		Title:  "test",
	}

	data, _ := SetTransactionData(txData)
	payload := BlockPayload{
		Type: "transaction",
		Data: data,
	}
	block := NewBlock(1, "prev", payload)

	retrieved, err := block.GetTransactionData()
	if err != nil {
		t.Fatalf("GetTransactionData failed: %v", err)
	}

	if retrieved.From != txData.From {
		t.Errorf("From = %s, want %s", retrieved.From, txData.From)
	}
	if retrieved.To != txData.To {
		t.Errorf("To = %s, want %s", retrieved.To, txData.To)
	}
	if retrieved.Amount != txData.Amount {
		t.Errorf("Amount = %d, want %d", retrieved.Amount, txData.Amount)
	}
	if retrieved.Title != txData.Title {
		t.Errorf("Title = %s, want %s", retrieved.Title, txData.Title)
	}
}

func TestGetTransactionData_WrongType(t *testing.T) {
	data, _ := json.Marshal(AddNodeData{
		PublicKey: "key",
		NodeName:  "node1",
	})

	payload := BlockPayload{
		Type: "add_node",
		Data: json.RawMessage(data),
	}
	block := NewBlock(1, "prev", payload)

	_, err := block.GetTransactionData()
	if err == nil {
		t.Error("Expected error for wrong type, got nil")
	}
}

func TestGetAddNodeData(t *testing.T) {
	addNodeData := &AddNodeData{
		PublicKey: "pubkey123",
		NodeName:  "node1",
		NickName:  "Tanaka",
		Address:   "10.0.0.1",
	}

	data, _ := SetAddNodeData(addNodeData)
	payload := BlockPayload{
		Type: "add_node",
		Data: data,
	}
	block := NewBlock(1, "prev", payload)

	retrieved, err := block.GetAddNodeData()
	if err != nil {
		t.Fatalf("GetAddNodeData failed: %v", err)
	}

	if retrieved.PublicKey != addNodeData.PublicKey {
		t.Errorf("PublicKey = %s, want %s", retrieved.PublicKey, addNodeData.PublicKey)
	}
	if retrieved.NodeName != addNodeData.NodeName {
		t.Errorf("NodeName = %s, want %s", retrieved.NodeName, addNodeData.NodeName)
	}
	if retrieved.NickName != addNodeData.NickName {
		t.Errorf("NickName = %s, want %s", retrieved.NickName, addNodeData.NickName)
	}
	if retrieved.Address != addNodeData.Address {
		t.Errorf("Address = %s, want %s", retrieved.Address, addNodeData.Address)
	}
}

func TestCreateBlockWithTransaction(t *testing.T) {
	tx := &TransactionData{
		From:   "node1",
		To:     "node2",
		Amount: 5000,
		Title:  "dinner",
	}

	block, err := CreateBlockWithTransaction(1, "prevhash", tx, "fromsig", "tosig")
	if err != nil {
		t.Fatalf("CreateBlockWithTransaction failed: %v", err)
	}

	if block.Header.Index != 1 {
		t.Errorf("Index = %d, want 1", block.Header.Index)
	}
	if block.Payload.Type != "transaction" {
		t.Errorf("Type = %s, want transaction", block.Payload.Type)
	}
	if block.Payload.FromSignature != "fromsig" {
		t.Errorf("FromSignature = %s, want fromsig", block.Payload.FromSignature)
	}
	if block.Payload.ToSignature != "tosig" {
		t.Errorf("ToSignature = %s, want tosig", block.Payload.ToSignature)
	}
}

func TestCreateBlockWithAddNode(t *testing.T) {
	addNode := &AddNodeData{
		PublicKey: "pubkey",
		NodeName:  "node1",
		NickName:  "Test",
		Address:   "192.168.1.1",
	}

	block, err := CreateBlockWithAddNode(1, "prevhash", addNode)
	if err != nil {
		t.Fatalf("CreateBlockWithAddNode failed: %v", err)
	}

	if block.Header.Index != 1 {
		t.Errorf("Index = %d, want 1", block.Header.Index)
	}
	if block.Payload.Type != "add_node" {
		t.Errorf("Type = %s, want add_node", block.Payload.Type)
	}
}

func TestMakeSigningPayload(t *testing.T) {
	txData := &TransactionData{
		From:   "node1",
		To:     "node2",
		Amount: 1000,
		Title:  "test",
	}

	data, _ := SetTransactionData(txData)
	payload := &BlockPayload{
		Type:          "transaction",
		Data:          data,
		FromSignature: "sig1",
		ToSignature:   "sig2",
	}

	signingData, err := MakeSigningPayload(payload)
	if err != nil {
		t.Fatalf("MakeSigningPayload failed: %v", err)
	}

	// JSONとして有効かチェック
	var result map[string]interface{}
	if err := json.Unmarshal(signingData, &result); err != nil {
		t.Errorf("Signing payload is not valid JSON: %v", err)
	}

	if result["type"] != "transaction" {
		t.Errorf("type = %v, want transaction", result["type"])
	}
	// 署名は含まれないはず
	if _, exists := result["from_signature"]; exists {
		t.Error("from_signature should not be in signing payload")
	}
}

func TestIsValidBlockType(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"transaction", true},
		{"add_node", true},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := IsValidBlockType(tt.input)
			if result != tt.expected {
				t.Errorf("IsValidBlockType(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseBlockType(t *testing.T) {
	tests := []struct {
		input    string
		expected BlockType
		hasError bool
	}{
		{"transaction", BlockTypeTransaction, false},
		{"TRANSACTION", BlockTypeTransaction, false},
		{"  transaction  ", BlockTypeTransaction, false},
		{"add_node", BlockTypeAddNode, false},
		{"ADD_NODE", BlockTypeAddNode, false},
		{"invalid", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := ParseBlockType(tt.input)
			if tt.hasError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("ParseBlockType(%q) = %v, want %v", tt.input, result, tt.expected)
				}
			}
		})
	}
}

func TestIsGenesisBlock(t *testing.T) {
	genesis := NewGenesisBlock(&AddNodeData{})
	if !genesis.IsGenesisBlock() {
		t.Error("NewGenesisBlock should be recognized as genesis block")
	}

	txData := &TransactionData{From: "a", To: "b", Amount: 100, Title: "test"}
	data, _ := SetTransactionData(txData)
	regularBlock := NewBlock(1, "hash", BlockPayload{Type: "transaction", Data: data})

	if regularBlock.IsGenesisBlock() {
		t.Error("Regular block should not be recognized as genesis block")
	}
}

func TestBlockJSON(t *testing.T) {
	txData := &TransactionData{
		From:   "node1",
		To:     "node2",
		Amount: 1000,
		Title:  "test",
	}

	data, _ := SetTransactionData(txData)
	payload := BlockPayload{
		Type:          "transaction",
		Data:          data,
		FromSignature: "sig1",
		ToSignature:   "sig2",
	}

	block := NewBlock(1, "prevhash", payload)

	// JSONシリアライズ・デシリアライズ
	bytes, err := json.Marshal(block)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var decoded Block
	if err := json.Unmarshal(bytes, &decoded); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if decoded.Header.Index != block.Header.Index {
		t.Errorf("Index mismatch: %d != %d", decoded.Header.Index, block.Header.Index)
	}
	if decoded.Header.Hash != block.Header.Hash {
		t.Errorf("Hash mismatch: %s != %s", decoded.Header.Hash, block.Header.Hash)
	}
	if decoded.Payload.Type != block.Payload.Type {
		t.Errorf("Type mismatch: %s != %s", decoded.Payload.Type, block.Payload.Type)
	}
}
