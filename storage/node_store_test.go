package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewNodeStore(t *testing.T) {
	store := NewNodeStore("/test/nodes")
	if store == nil {
		t.Fatal("NewNodeStore() returned nil")
	}
	if store.dir != "/test/nodes" {
		t.Errorf("store.dir = %v, want /test/nodes", store.dir)
	}
}

func TestNodeStoreSave(t *testing.T) {
	t.Run("save node info", func(t *testing.T) {
		tmpDir := t.TempDir()
		store := NewNodeStore(tmpDir)

		info := &NodeInfo{
			Name:      "node1",
			NickName:  "田中",
			Address:   "10.0.0.1",
			PublicKey: "test_public_key",
		}

		if err := store.Save("node1", info); err != nil {
			t.Fatalf("Save() error = %v", err)
		}

		// ファイルが存在することを確認
		filePath := filepath.Join(tmpDir, "node1")
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Error("Save() did not create file")
		}

		// 内容を確認
		content, _ := readFile(filePath)
		expectedContent := "NickName = \"田中\"\nAddress = \"10.0.0.1\"\nEd25519PublicKey = \"test_public_key\"\n"
		if string(content) != expectedContent {
			t.Errorf("File content = %q, want %q", string(content), expectedContent)
		}
	})

	t.Run("save creates directory if not exists", func(t *testing.T) {
		tmpDir := t.TempDir()
		nodesDir := filepath.Join(tmpDir, "nodes")
		store := NewNodeStore(nodesDir)

		info := &NodeInfo{
			Name:      "node1",
			NickName:  "Test",
			Address:   "10.0.0.1",
			PublicKey: "key",
		}

		if err := store.Save("node1", info); err != nil {
			t.Fatalf("Save() error = %v", err)
		}

		// ディレクトリが作成されたことを確認
		if _, err := os.Stat(nodesDir); os.IsNotExist(err) {
			t.Error("Save() did not create directory")
		}
	})
}

func TestNodeStorePathTraversal(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewNodeStore(tmpDir)

	info := &NodeInfo{Name: "evil", NickName: "Evil", Address: "1.2.3.4", PublicKey: "key"}

	tests := []struct {
		name     string
		nodeName string
	}{
		{"dot-dot-slash", "../evil"},
		{"slash", "sub/evil"},
		{"backslash", "sub\\evil"},
		{"dot-dot", ".."},
		{"dot", "."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := store.Save(tt.nodeName, info); err == nil {
				t.Errorf("Save(%q) should return error for path traversal", tt.nodeName)
			}
		})
	}
}

func TestNodeStoreLoad(t *testing.T) {
	t.Run("load existing node", func(t *testing.T) {
		tmpDir := t.TempDir()
		store := NewNodeStore(tmpDir)

		// 先にファイルを作成
		info := &NodeInfo{
			Name:      "node1",
			NickName:  "田中",
			Address:   "10.0.0.1",
			PublicKey: "test_public_key",
		}
		store.Save("node1", info)

		// 読み込む
		loaded, err := store.Load("node1")
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		if loaded.NickName != "田中" {
			t.Errorf("NickName = %v, want 田中", loaded.NickName)
		}
		if loaded.Address != "10.0.0.1" {
			t.Errorf("Address = %v, want 10.0.0.1", loaded.Address)
		}
		if loaded.PublicKey != "test_public_key" {
			t.Errorf("PublicKey = %v, want test_public_key", loaded.PublicKey)
		}
	})

	t.Run("load nonexistent node returns error", func(t *testing.T) {
		tmpDir := t.TempDir()
		store := NewNodeStore(tmpDir)

		_, err := store.Load("nonexistent")
		if err == nil {
			t.Error("Load() should return error for nonexistent node")
		}
	})
}

func TestNodeStoreLoadAll(t *testing.T) {
	t.Run("load all nodes", func(t *testing.T) {
		tmpDir := t.TempDir()
		store := NewNodeStore(tmpDir)

		// 複数のノードを保存
		node1 := &NodeInfo{Name: "node1", NickName: "田中", Address: "10.0.0.1", PublicKey: "key1"}
		node2 := &NodeInfo{Name: "node2", NickName: "佐藤", Address: "10.0.0.2", PublicKey: "key2"}
		node3 := &NodeInfo{Name: "node3", NickName: "鈴木", Address: "10.0.0.3", PublicKey: "key3"}

		store.Save("node1", node1)
		store.Save("node2", node2)
		store.Save("node3", node3)

		// 全て読み込む
		all, err := store.LoadAll()
		if err != nil {
			t.Fatalf("LoadAll() error = %v", err)
		}

		if len(all) != 3 {
			t.Errorf("LoadAll() returned %d nodes, want 3", len(all))
		}

		// 各ノードを確認
		if all["node1"].NickName != "田中" {
			t.Errorf("node1.NickName = %v, want 田中", all["node1"].NickName)
		}
		if all["node2"].NickName != "佐藤" {
			t.Errorf("node2.NickName = %v, want 佐藤", all["node2"].NickName)
		}
		if all["node3"].NickName != "鈴木" {
			t.Errorf("node3.NickName = %v, want 鈴木", all["node3"].NickName)
		}
	})

	t.Run("load from nonexistent directory returns empty map", func(t *testing.T) {
		tmpDir := t.TempDir()
		store := NewNodeStore(filepath.Join(tmpDir, "nonexistent"))

		all, err := store.LoadAll()
		if err != nil {
			t.Fatalf("LoadAll() error = %v", err)
		}

		if len(all) != 0 {
			t.Errorf("LoadAll() returned %d nodes, want 0", len(all))
		}
	})

	t.Run("load from empty directory returns empty map", func(t *testing.T) {
		tmpDir := t.TempDir()
		store := NewNodeStore(tmpDir)

		all, err := store.LoadAll()
		if err != nil {
			t.Fatalf("LoadAll() error = %v", err)
		}

		if len(all) != 0 {
			t.Errorf("LoadAll() returned %d nodes, want 0", len(all))
		}
	})
}

func TestNodeStoreDelete(t *testing.T) {
	t.Run("delete existing node", func(t *testing.T) {
		tmpDir := t.TempDir()
		store := NewNodeStore(tmpDir)

		// ノードを保存
		info := &NodeInfo{Name: "node1", NickName: "Test", Address: "10.0.0.1", PublicKey: "key"}
		store.Save("node1", info)

		// 削除
		if err := store.Delete("node1"); err != nil {
			t.Fatalf("Delete() error = %v", err)
		}

		// 存在しないことを確認
		if store.Exists("node1") {
			t.Error("Delete() did not remove the node")
		}
	})

	t.Run("delete nonexistent node does not error", func(t *testing.T) {
		tmpDir := t.TempDir()
		store := NewNodeStore(tmpDir)

		// 存在しないノードを削除してもエラーにならない
		if err := store.Delete("nonexistent"); err != nil {
			t.Errorf("Delete() error = %v", err)
		}
	})
}

func TestNodeStoreExists(t *testing.T) {
	t.Run("existing node returns true", func(t *testing.T) {
		tmpDir := t.TempDir()
		store := NewNodeStore(tmpDir)

		info := &NodeInfo{Name: "node1", NickName: "Test", Address: "10.0.0.1", PublicKey: "key"}
		store.Save("node1", info)

		if !store.Exists("node1") {
			t.Error("Exists() returned false for existing node")
		}
	})

	t.Run("nonexistent node returns false", func(t *testing.T) {
		tmpDir := t.TempDir()
		store := NewNodeStore(tmpDir)

		if store.Exists("nonexistent") {
			t.Error("Exists() returned true for nonexistent node")
		}
	})
}
