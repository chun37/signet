package crypto

import (
	"crypto/ed25519"
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateKeyPair(t *testing.T) {
	pub, priv, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair failed: %v", err)
	}

	if len(pub) != ed25519.PublicKeySize {
		t.Errorf("Public key size = %d, want %d", len(pub), ed25519.PublicKeySize)
	}

	if len(priv) != ed25519.PrivateKeySize {
		t.Errorf("Private key size = %d, want %d", len(priv), ed25519.PrivateKeySize)
	}
}

func TestPublicKeyToBase64(t *testing.T) {
	pub, _, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair failed: %v", err)
	}

	encoded := PublicKeyToBase64(pub)
	if encoded == "" {
		t.Error("PublicKeyToBase64 returned empty string")
	}

	// デコードして元の鍵と比較
	decoded, err := Base64ToPublicKey(encoded)
	if err != nil {
		t.Fatalf("Base64ToPublicKey failed: %v", err)
	}

	if string(decoded) != string(pub) {
		t.Error("Decoded public key does not match original")
	}
}

func TestBase64ToPublicKey_InvalidInput(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "invalid base64",
			input:   "!!!invalid!!!",
			wantErr: true,
		},
		{
			name:    "too short",
			input:   "YQ==", // "a" in base64
			wantErr: true,
		},
		{
			name:    "too long",
			input:   "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA==",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Base64ToPublicKey(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Base64ToPublicKey() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPrivateKeyToBase64(t *testing.T) {
	_, priv, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair failed: %v", err)
	}

	encoded := PrivateKeyToBase64(priv)
	if encoded == "" {
		t.Error("PrivateKeyToBase64 returned empty string")
	}

	// デコードして元の鍵と比較
	decoded, err := Base64ToPrivateKey(encoded)
	if err != nil {
		t.Fatalf("Base64ToPrivateKey failed: %v", err)
	}

	if string(decoded) != string(priv) {
		t.Error("Decoded private key does not match original")
	}
}

func TestBase64ToPrivateKey_InvalidInput(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "invalid base64",
			input:   "!!!invalid!!!",
			wantErr: true,
		},
		{
			name:    "too short",
			input:   "YQ==",
			wantErr: true,
		},
		{
			name:    "too long",
			input:   "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA==",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Base64ToPrivateKey(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Base64ToPrivateKey() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSavePrivateKeyRaw_LoadPrivateKey(t *testing.T) {
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "test_key.priv")

	// 鍵を生成
	_, priv, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair failed: %v", err)
	}

	// 保存
	if err := SavePrivateKeyRaw(keyPath, priv); err != nil {
		t.Fatalf("SavePrivateKeyRaw failed: %v", err)
	}

	// ファイルが存在するか確認
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		t.Fatal("Private key file was not created")
	}

	// 読み込み
	loaded, err := LoadPrivateKey(keyPath)
	if err != nil {
		t.Fatalf("LoadPrivateKey failed: %v", err)
	}

	if string(loaded) != string(priv) {
		t.Error("Loaded private key does not match original")
	}
}

func TestSavePrivateKeyPEM_LoadPrivateKey(t *testing.T) {
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "test_key.pem")

	// 鍵を生成
	_, priv, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair failed: %v", err)
	}

	// PEM形式で保存
	if err := SavePrivateKey(keyPath, priv); err != nil {
		t.Fatalf("SavePrivateKey failed: %v", err)
	}

	// ファイルが存在するか確認
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		t.Fatal("Private key file was not created")
	}

	// 読み込み
	loaded, err := LoadPrivateKey(keyPath)
	if err != nil {
		t.Fatalf("LoadPrivateKey failed: %v", err)
	}

	if string(loaded) != string(priv) {
		t.Error("Loaded private key does not match original")
	}
}

func TestLoadPrivateKey_FileNotFound(t *testing.T) {
	_, err := LoadPrivateKey("/nonexistent/path/key.priv")
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}

func TestSavePrivateKey_InvalidSize(t *testing.T) {
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "invalid_key.priv")

	invalidKey := ed25519.PrivateKey([]byte("too_short"))

	err := SavePrivateKey(keyPath, invalidKey)
	if err == nil {
		t.Error("Expected error for invalid key size, got nil")
	}
}

func TestSavePrivateKeyRaw_InvalidSize(t *testing.T) {
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "invalid_key.priv")

	invalidKey := ed25519.PrivateKey([]byte("too_short"))

	err := SavePrivateKeyRaw(keyPath, invalidKey)
	if err == nil {
		t.Error("Expected error for invalid key size, got nil")
	}
}

func TestGetPublicKeyFromPrivateKey(t *testing.T) {
	pub, priv, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair failed: %v", err)
	}

	extractedPub := GetPublicKeyFromPrivateKey(priv)

	if string(extractedPub) != string(pub) {
		t.Error("Extracted public key does not match original")
	}
}

func TestKeyPairSignVerifyCycle(t *testing.T) {
	pub, priv, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair failed: %v", err)
	}

	// 鍵を文字列化して戻す
	pubStr := PublicKeyToBase64(pub)
	privStr := PrivateKeyToBase64(priv)

	pub2, err := Base64ToPublicKey(pubStr)
	if err != nil {
		t.Fatalf("Base64ToPublicKey failed: %v", err)
	}

	priv2, err := Base64ToPrivateKey(privStr)
	if err != nil {
		t.Fatalf("Base64ToPrivateKey failed: %v", err)
	}

	// 元の鍵と一致することを確認
	if string(pub2) != string(pub) {
		t.Error("Public key round-trip failed")
	}

	if string(priv2) != string(priv) {
		t.Error("Private key round-trip failed")
	}

	// 秘密鍵から抽出した公開鍵も一致することを確認
	extractedPub := GetPublicKeyFromPrivateKey(priv2)
	if string(extractedPub) != string(pub) {
		t.Error("Extracted public key does not match original")
	}
}
