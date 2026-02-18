package core

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNewPendingPool(t *testing.T) {
	pool := NewPendingPool()

	if pool.Len() != 0 {
		t.Errorf("New pool length = %d, want 0", pool.Len())
	}
}

func TestPendingPool_AddAndGet(t *testing.T) {
	pool := NewPendingPool()

	txData := &TransactionData{
		From:   "node1",
		To:     "node2",
		Amount: 1000,
		Title:  "test",
	}

	data, _ := json.Marshal(txData)
	payload := BlockPayload{
		Type:          "transaction",
		Data:          json.RawMessage(data),
		FromSignature: "sig1",
		ToSignature:   "",
	}

	pt := NewPendingTransaction("id1", payload)
	pool.Add(pt)

	retrieved := pool.Get("id1")
	if retrieved == nil {
		t.Fatal("Get returned nil for existing id")
	}

	if retrieved.ID != "id1" {
		t.Errorf("ID = %s, want id1", retrieved.ID)
	}

	// 存在しないID
	if pool.Get("nonexistent") != nil {
		t.Error("Get should return nil for non-existent id")
	}
}

func TestPendingPool_Remove(t *testing.T) {
	pool := NewPendingPool()

	txData := &TransactionData{From: "a", To: "b", Amount: 100, Title: "test"}
	data, _ := json.Marshal(txData)
	payload := BlockPayload{
		Type:          "transaction",
		Data:          json.RawMessage(data),
		FromSignature: "sig1",
	}

	pt := NewPendingTransaction("id1", payload)
	pool.Add(pt)

	if pool.Len() != 1 {
		t.Errorf("Pool length = %d, want 1", pool.Len())
	}

	pool.Remove("id1")

	if pool.Len() != 0 {
		t.Errorf("Pool length after remove = %d, want 0", pool.Len())
	}

	if pool.Get("id1") != nil {
		t.Error("Get should return nil after remove")
	}
}

func TestPendingPool_List(t *testing.T) {
	pool := NewPendingPool()

	// 複数追加
	for i := 0; i < 3; i++ {
		txData := &TransactionData{From: "a", To: "b", Amount: 100, Title: "test"}
		data, _ := json.Marshal(txData)
		payload := BlockPayload{
			Type:          "transaction",
			Data:          json.RawMessage(data),
			FromSignature: "sig1",
		}
		pt := NewPendingTransaction(string(rune('a'+i)), payload)
		pool.Add(pt)
	}

	list := pool.List()
	if len(list) != 3 {
		t.Errorf("List length = %d, want 3", len(list))
	}

	// 返されたリストを修改しても元に影響しない
	list[0] = nil

	if pool.Len() != 3 {
		t.Error("Modifying returned slice affected original pool")
	}
}

func TestPendingPool_GetAll(t *testing.T) {
	pool := NewPendingPool()

	txData := &TransactionData{From: "a", To: "b", Amount: 100, Title: "test"}
	data, _ := json.Marshal(txData)
	payload := BlockPayload{
		Type:          "transaction",
		Data:          json.RawMessage(data),
		FromSignature: "sig1",
	}
	pt := NewPendingTransaction("id1", payload)
	pool.Add(pt)

	all := pool.GetAll()
	if len(all) != 1 {
		t.Errorf("GetAll length = %d, want 1", len(all))
	}

	if _, exists := all["id1"]; !exists {
		t.Error("GetAll should contain id1")
	}

	// 返されたマップを修改しても元に影響しない
	all["id2"] = nil

	if pool.Len() != 1 {
		t.Error("Modifying returned map affected original pool")
	}
}

func TestPendingPool_Has(t *testing.T) {
	pool := NewPendingPool()

	txData := &TransactionData{From: "a", To: "b", Amount: 100, Title: "test"}
	data, _ := json.Marshal(txData)
	payload := BlockPayload{
		Type:          "transaction",
		Data:          json.RawMessage(data),
		FromSignature: "sig1",
	}
	pt := NewPendingTransaction("id1", payload)
	pool.Add(pt)

	if !pool.Has("id1") {
		t.Error("Has should return true for existing id")
	}

	if pool.Has("nonexistent") {
		t.Error("Has should return false for non-existent id")
	}
}

func TestPendingPool_Clear(t *testing.T) {
	pool := NewPendingPool()

	// 複数追加
	for i := 0; i < 3; i++ {
		txData := &TransactionData{From: "a", To: "b", Amount: 100, Title: "test"}
		data, _ := json.Marshal(txData)
		payload := BlockPayload{
			Type:          "transaction",
			Data:          json.RawMessage(data),
			FromSignature: "sig1",
		}
		pt := NewPendingTransaction(string(rune('a'+i)), payload)
		pool.Add(pt)
	}

	pool.Clear()

	if pool.Len() != 0 {
		t.Errorf("Pool length after clear = %d, want 0", pool.Len())
	}
}

func TestPendingPool_GetByToNode(t *testing.T) {
	pool := NewPendingPool()

	// node2宛のトランザクションを追加
	txData1 := &TransactionData{From: "node1", To: "node2", Amount: 1000, Title: "test1"}
	data1, _ := json.Marshal(txData1)
	payload1 := BlockPayload{
		Type:          "transaction",
		Data:          json.RawMessage(data1),
		FromSignature: "sig1",
	}
	pt1 := NewPendingTransaction("id1", payload1)
	pool.Add(pt1)

	// node3宛のトランザクションを追加
	txData2 := &TransactionData{From: "node1", To: "node3", Amount: 2000, Title: "test2"}
	data2, _ := json.Marshal(txData2)
	payload2 := BlockPayload{
		Type:          "transaction",
		Data:          json.RawMessage(data2),
		FromSignature: "sig2",
	}
	pt2 := NewPendingTransaction("id2", payload2)
	pool.Add(pt2)

	// node2宛を取得
	results := pool.GetByToNode("node2")
	if len(results) != 1 {
		t.Errorf("GetByToNode returned %d items, want 1", len(results))
	}

	if len(results) > 0 && results[0].ID != "id1" {
		t.Errorf("Result ID = %s, want id1", results[0].ID)
	}

	// node3宛を取得
	results = pool.GetByToNode("node3")
	if len(results) != 1 {
		t.Errorf("GetByToNode returned %d items, want 1", len(results))
	}

	// 存在しないノード
	results = pool.GetByToNode("nonexistent")
	if len(results) != 0 {
		t.Errorf("GetByToNode returned %d items, want 0", len(results))
	}
}

func TestPendingPool_ConcurrentAccess(t *testing.T) {
	pool := NewPendingPool()

	done := make(chan bool)

	// 並行書き込み
	for i := 0; i < 10; i++ {
		go func(idx int) {
			txData := &TransactionData{From: "a", To: "b", Amount: 100, Title: "test"}
			data, _ := json.Marshal(txData)
			payload := BlockPayload{
				Type:          "transaction",
				Data:          json.RawMessage(data),
				FromSignature: "sig",
			}
			pt := NewPendingTransaction(string(rune('0'+idx)), payload)
			pool.Add(pt)
			done <- true
		}(i)
	}

	// 並行読み込み
	for i := 0; i < 10; i++ {
		go func() {
			_ = pool.Len()
			_ = pool.List()
			done <- true
		}()
	}

	// 全完了待ち
	for i := 0; i < 20; i++ {
		<-done
	}
}

func TestNewPendingTransaction(t *testing.T) {
	txData := &TransactionData{From: "a", To: "b", Amount: 100, Title: "test"}
	data, _ := json.Marshal(txData)
	payload := BlockPayload{
		Type:          "transaction",
		Data:          json.RawMessage(data),
		FromSignature: "sig1",
	}

	pt := NewPendingTransaction("test-id", payload)

	if pt.ID != "test-id" {
		t.Errorf("ID = %s, want test-id", pt.ID)
	}

	if pt.Payload.Type != "transaction" {
		t.Errorf("Type = %s, want transaction", pt.Payload.Type)
	}

	if pt.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}

	// CreatedAtはUTCであるべき
	if _, offset := pt.CreatedAt.Zone(); offset != 0 {
		t.Error("CreatedAt should be in UTC")
	}
}

func TestPendingTransaction_GetTransactionData(t *testing.T) {
	txData := &TransactionData{
		From:   "node1",
		To:     "node2",
		Amount: 5000,
		Title:  "dinner",
	}

	data, _ := json.Marshal(txData)
	payload := BlockPayload{
		Type:          "transaction",
		Data:          json.RawMessage(data),
		FromSignature: "sig1",
	}

	pt := NewPendingTransaction("id1", payload)

	retrieved, err := pt.GetTransactionData()
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
}

func TestPendingTransaction_GetTransactionData_WrongType(t *testing.T) {
	addNodeData := &AddNodeData{
		PublicKey: "pubkey",
		NodeName:  "node1",
	}
	data, _ := json.Marshal(addNodeData)
	payload := BlockPayload{
		Type: "add_node",
		Data: json.RawMessage(data),
	}

	pt := NewPendingTransaction("id1", payload)

	_, err := pt.GetTransactionData()
	if err == nil {
		t.Error("Expected error for wrong payload type, got nil")
	}
}

func TestGenerateID(t *testing.T) {
	txData := &TransactionData{From: "a", To: "b", Amount: 100, Title: "test"}
	data, _ := json.Marshal(txData)
	payload := BlockPayload{
		Type:          "transaction",
		Data:          json.RawMessage(data),
		FromSignature: "sig1",
	}

	fixedTime := time.Date(2026, 2, 18, 12, 0, 0, 0, time.UTC)

	id1 := GenerateID(payload, fixedTime)
	id2 := GenerateID(payload, fixedTime)

	if id1 != id2 {
		t.Errorf("GenerateID is not deterministic: %s != %s", id1, id2)
	}

	// 時刻が違えばIDも違うはず
	id3 := GenerateID(payload, fixedTime.Add(time.Second))
	if id1 == id3 {
		t.Error("GenerateID should produce different IDs for different times")
	}

	// ペイロードが違えばIDも違うはず
	txData2 := &TransactionData{From: "c", To: "d", Amount: 200, Title: "test2"}
	data2, _ := json.Marshal(txData2)
	payload2 := BlockPayload{
		Type:          "transaction",
		Data:          json.RawMessage(data2),
		FromSignature: "sig2",
	}
	id4 := GenerateID(payload2, fixedTime)
	if id1 == id4 {
		t.Error("GenerateID should produce different IDs for different payloads")
	}
}

func TestPendingPool_ReplaceExisting(t *testing.T) {
	pool := NewPendingPool()

	txData1 := &TransactionData{From: "a", To: "b", Amount: 100, Title: "test1"}
	data1, _ := json.Marshal(txData1)
	payload1 := BlockPayload{
		Type:          "transaction",
		Data:          json.RawMessage(data1),
		FromSignature: "sig1",
	}
	pt1 := NewPendingTransaction("id1", payload1)
	pool.Add(pt1)

	if pool.Len() != 1 {
		t.Errorf("Pool length = %d, want 1", pool.Len())
	}

	// 同じIDで上書き
	txData2 := &TransactionData{From: "c", To: "d", Amount: 200, Title: "test2"}
	data2, _ := json.Marshal(txData2)
	payload2 := BlockPayload{
		Type:          "transaction",
		Data:          json.RawMessage(data2),
		FromSignature: "sig2",
	}
	pt2 := NewPendingTransaction("id1", payload2)
	pool.Add(pt2)

	if pool.Len() != 1 {
		t.Errorf("Pool length after replace = %d, want 1", pool.Len())
	}

	retrieved := pool.Get("id1")
	if retrieved.Payload.FromSignature != "sig2" {
		t.Errorf("Payload was not replaced: FromSignature = %s, want sig2", retrieved.Payload.FromSignature)
	}
}
