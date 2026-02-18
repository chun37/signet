package crypto

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"os"
)

// GenerateKeyPair はEd25519の鍵ペアを生成する
func GenerateKeyPair() (ed25519.PublicKey, ed25519.PrivateKey, error) {
	return ed25519.GenerateKey(nil)
}

// SavePrivateKey は秘密鍵をBase64エンコードしてファイルに保存する
func SavePrivateKey(path string, key ed25519.PrivateKey) error {
	if len(key) != ed25519.PrivateKeySize {
		return fmt.Errorf("invalid private key size: %d", len(key))
	}

	// Base64エンコード
	encoded := base64.StdEncoding.EncodeToString(key)

	// PEM形式で保存（人間が識別しやすくするため）
	block := &pem.Block{
		Type:  "ED25519 PRIVATE KEY",
		Bytes: []byte(encoded),
	}

	// ディレクトリが存在しない場合は作成を試みる
	// (呼び出し側でディレクトリを作成することを推奨)

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create private key file: %w", err)
	}
	defer file.Close()

	if err := pem.Encode(file, block); err != nil {
		return fmt.Errorf("failed to encode PEM: %w", err)
	}

	return nil
}

// SavePrivateKeyRaw は秘密鍵を生のBase64文字列としてファイルに保存する
func SavePrivateKeyRaw(path string, key ed25519.PrivateKey) error {
	if len(key) != ed25519.PrivateKeySize {
		return fmt.Errorf("invalid private key size: %d", len(key))
	}

	encoded := base64.StdEncoding.EncodeToString(key)

	if err := os.WriteFile(path, []byte(encoded), 0600); err != nil {
		return fmt.Errorf("failed to write private key file: %w", err)
	}

	return nil
}

// LoadPrivateKey はファイルから秘密鍵を読み込む
func LoadPrivateKey(path string) (ed25519.PrivateKey, error) {
	// まずPEM形式を試みる
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key file: %w", err)
	}

	// PEMデコードを試みる
	block, _ := pem.Decode(data)
	if block != nil && block.Type == "ED25519 PRIVATE KEY" {
		// PEM形式
		encoded := string(block.Bytes)
		key, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			return nil, fmt.Errorf("failed to decode base64 private key: %w", err)
		}

		if len(key) != ed25519.PrivateKeySize {
			return nil, fmt.Errorf("invalid private key size: %d", len(key))
		}

		return ed25519.PrivateKey(key), nil
	}

	// 生のBase64形式として試みる
	key, err := base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 private key: %w", err)
	}

	if len(key) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("invalid private key size: %d", len(key))
	}

	return ed25519.PrivateKey(key), nil
}

// PublicKeyToBase64 は公開鍵をBase64エンコードして文字列にする
func PublicKeyToBase64(pub ed25519.PublicKey) string {
	return base64.StdEncoding.EncodeToString(pub)
}

// Base64ToPublicKey はBase64エンコードされた文字列から公開鍵を復元する
func Base64ToPublicKey(s string) (ed25519.PublicKey, error) {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 public key: %w", err)
	}

	if len(data) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("invalid public key size: %d", len(data))
	}

	return ed25519.PublicKey(data), nil
}

// PrivateKeyToBase64 は秘密鍵をBase64エンコードして文字列にする
func PrivateKeyToBase64(priv ed25519.PrivateKey) string {
	return base64.StdEncoding.EncodeToString(priv)
}

// Base64ToPrivateKey はBase64エンコードされた文字列から秘密鍵を復元する
func Base64ToPrivateKey(s string) (ed25519.PrivateKey, error) {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 private key: %w", err)
	}

	if len(data) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("invalid private key size: %d", len(data))
	}

	return ed25519.PrivateKey(data), nil
}

// GetPublicKeyFromPrivateKey は秘密鍵から対応する公開鍵を取得する
func GetPublicKeyFromPrivateKey(priv ed25519.PrivateKey) ed25519.PublicKey {
	return priv.Public().(ed25519.PublicKey)
}

// HexToPublicKey はhexエンコードされた文字列から公開鍵を復元する
func HexToPublicKey(s string) (ed25519.PublicKey, error) {
	data, err := hex.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("failed to decode hex public key: %w", err)
	}

	if len(data) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("invalid public key size: %d", len(data))
	}

	return ed25519.PublicKey(data), nil
}
