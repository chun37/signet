package storage

import (
	"encoding/json"
	"path/filepath"
	"signet/core"
	"testing"
	"time"
)

func TestNewPendingStore(t *testing.T) {
	store := NewPendingStore("/test/path")
	if store == nil {
		t.Fatal("NewPendingStore() returned nil")
	}
	if store.path != "/test/path" {
		t.Errorf("store.path = %v, want /test/path", store.path)
	}
}

func TestPendingStoreLoad(t *testing.T) {
	t.Run("nonexistent file returns empty slice", func(t *testing.T) {
		tmpDir := t.TempDir()
		store := NewPendingStore(filepath.Join(tmpDir, "nonexistent.json"))

		items, err := store.Load()
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		if len(items) != 0 {
			t.Errorf("Load() returned %d items, want 0", len(items))
		}
	})

	t.Run("load pending transactions", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "pending.json")

		// テストデータを作成 - 新しい構造に合わせる
		tx1 := core.TransactionData{From: "node1", To: "node2", Amount: 1000, Title: "Test"}
		tx1Data, _ := json.Marshal(tx1)
		payload1 := core.BlockPayload{
			Type: "transaction",
			Data: json.RawMessage(tx1Data),
		}

		tx2 := core.TransactionData{From: "node2", To: "node3", Amount: 2000, Title: "Test2"}
		tx2Data, _ := json.Marshal(tx2)
		payload2 := core.BlockPayload{
			Type: "transaction",
			Data: json.RawMessage(tx2Data),
		}

		pending1 := core.NewPendingTransaction("id1", payload1)
		pending2 := core.NewPendingTransaction("id2", payload2)

		items := []*core.PendingTransaction{pending1, pending2}
		store := NewPendingStore(filePath)
		if err := store.Save(items); err != nil {
			t.Fatalf("failed to save test data: %v", err)
		}

		// 読み込んで確認
		loaded, err := store.Load()
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		if len(loaded) != 2 {
			t.Fatalf("Load() returned %d items, want 2", len(loaded))
		}

		if loaded[0].ID != "id1" {
			t.Errorf("loaded[0].ID = %v, want id1", loaded[0].ID)
		}
		if loaded[1].ID != "id2" {
			t.Errorf("loaded[1].ID = %v, want id2", loaded[1].ID)
		}
	})
}

func TestPendingStoreSave(t *testing.T) {
	t.Run("save pending transactions", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "pending.json")
		store := NewPendingStore(filePath)

		// テストデータを作成
		tx1 := core.TransactionData{From: "node1", To: "node2", Amount: 1000, Title: "Test"}
		tx1Data, _ := json.Marshal(tx1)
		payload1 := core.BlockPayload{
			Type: "transaction",
			Data: json.RawMessage(tx1Data),
		}

		tx2 := core.TransactionData{From: "node2", To: "node3", Amount: 2000, Title: "Test2"}
		tx2Data, _ := json.Marshal(tx2)
		payload2 := core.BlockPayload{
			Type: "transaction",
			Data: json.RawMessage(tx2Data),
		}

		pending1 := core.NewPendingTransaction("id1", payload1)
		pending2 := core.NewPendingTransaction("id2", payload2)

		items := []*core.PendingTransaction{pending1, pending2}
		if err := store.Save(items); err != nil {
			t.Fatalf("Save() error = %v", err)
		}

		// 読み込んで確認
		loaded, err := store.Load()
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		if len(loaded) != 2 {
			t.Errorf("Load() returned %d items, want 2", len(loaded))
		}
		if loaded[0].ID != "id1" {
			t.Errorf("loaded[0].ID = %v, want id1", loaded[0].ID)
		}
		if loaded[1].ID != "id2" {
			t.Errorf("loaded[1].ID = %v, want id2", loaded[1].ID)
		}
	})

	t.Run("save empty slice", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "pending.json")
		store := NewPendingStore(filePath)

		if err := store.Save([]*core.PendingTransaction{}); err != nil {
			t.Fatalf("Save() error = %v", err)
		}

		loaded, err := store.Load()
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		if len(loaded) != 0 {
			t.Errorf("Load() returned %d items, want 0", len(loaded))
		}
	})
}

func TestPendingStoreAdd(t *testing.T) {
	t.Run("add pending transaction", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "pending.json")
		store := NewPendingStore(filePath)

		tx1 := core.TransactionData{From: "node1", To: "node2", Amount: 1000, Title: "Test"}
		tx1Data, _ := json.Marshal(tx1)
		payload1 := core.BlockPayload{
			Type: "transaction",
			Data: json.RawMessage(tx1Data),
		}
		pending1 := core.NewPendingTransaction("id1", payload1)

		if err := store.Add(pending1); err != nil {
			t.Fatalf("Add() error = %v", err)
		}

		// もう1つ追加
		tx2 := core.TransactionData{From: "node2", To: "node3", Amount: 2000, Title: "Test2"}
		tx2Data, _ := json.Marshal(tx2)
		payload2 := core.BlockPayload{
			Type: "transaction",
			Data: json.RawMessage(tx2Data),
		}
		pending2 := core.NewPendingTransaction("id2", payload2)
		if err := store.Add(pending2); err != nil {
			t.Fatalf("Add() error = %v", err)
		}

		// 読み込んで確認
		loaded, err := store.Load()
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		if len(loaded) != 2 {
			t.Errorf("Load() returned %d items, want 2", len(loaded))
		}
	})
}

func TestPendingStoreRemove(t *testing.T) {
	t.Run("remove pending transaction", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "pending.json")
		store := NewPendingStore(filePath)

		// 3つのアイテムを追加
		for i := 0; i < 3; i++ {
			tx := core.TransactionData{From: "node1", To: "node2", Amount: int64(i * 1000), Title: "Test"}
			txData, _ := json.Marshal(tx)
			payload := core.BlockPayload{
				Type: "transaction",
				Data: json.RawMessage(txData),
			}
			pending := core.NewPendingTransaction(time.Now().Format(time.RFC3339Nano), payload)
			store.Add(pending)
		}

		// 真ん中を削除
		if err := store.Remove(1); err != nil {
			t.Fatalf("Remove() error = %v", err)
		}

		// 読み込んで確認
		loaded, err := store.Load()
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		if len(loaded) != 2 {
			t.Errorf("Load() returned %d items, want 2", len(loaded))
		}
	})

	t.Run("remove with invalid index", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "pending.json")
		store := NewPendingStore(filePath)

		tx := core.TransactionData{From: "node1", To: "node2", Amount: 1000, Title: "Test"}
		txData, _ := json.Marshal(tx)
		payload := core.BlockPayload{
			Type: "transaction",
			Data: json.RawMessage(txData),
		}
		pending := core.NewPendingTransaction("id1", payload)
		store.Add(pending)

		// 範囲外のインデックスで削除
		if err := store.Remove(10); err == nil {
			t.Error("Remove() should return error for out of range index")
		}

		// データが変わっていないことを確認
		loaded, _ := store.Load()
		if len(loaded) != 1 {
			t.Errorf("Load() returned %d items, want 1 (data should be unchanged)", len(loaded))
		}
	})
}

func TestPendingStoreClear(t *testing.T) {
	t.Run("clear all pending transactions", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "pending.json")
		store := NewPendingStore(filePath)

		// アイテムを追加
		tx := core.TransactionData{From: "node1", To: "node2", Amount: 1000, Title: "Test"}
		txData, _ := json.Marshal(tx)
		payload := core.BlockPayload{
			Type: "transaction",
			Data: json.RawMessage(txData),
		}
		pending := core.NewPendingTransaction("id1", payload)
		store.Add(pending)
		store.Add(pending)

		// クリア
		if err := store.Clear(); err != nil {
			t.Fatalf("Clear() error = %v", err)
		}

		// 読み込んで確認
		loaded, err := store.Load()
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		if len(loaded) != 0 {
			t.Errorf("Load() returned %d items, want 0", len(loaded))
		}
	})
}
