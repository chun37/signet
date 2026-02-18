package crypto

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"signet/core"
)

// Sign はデータにEd25519で署名し、Base64エンコードされた署名を返す
func Sign(privKey ed25519.PrivateKey, data []byte) string {
	signature := ed25519.Sign(privKey, data)
	return base64.StdEncoding.EncodeToString(signature)
}

// Verify は署名を検証する
func Verify(pubKey ed25519.PublicKey, data []byte, signatureBase64 string) bool {
	signature, err := base64.StdEncoding.DecodeString(signatureBase64)
	if err != nil {
		return false
	}

	return ed25519.Verify(pubKey, data, signature)
}

// MakeSigningPayload は署名対象のペイロードバイト列を作成する
// Type + Data をJSON直列化して連結
func MakeSigningPayload(payload *core.BlockPayload) ([]byte, error) {
	typeData := struct {
		Type string          `json:"type"`
		Data json.RawMessage `json:"data"`
	}{
		Type: payload.Type,
		Data: payload.Data,
	}

	jsonData, err := json.Marshal(typeData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal signing payload: %w", err)
	}

	return jsonData, nil
}

// SignPayload はBlockPayloadに署名する
func SignPayload(privKey ed25519.PrivateKey, payload *core.BlockPayload) (string, error) {
	signingData, err := MakeSigningPayload(payload)
	if err != nil {
		return "", err
	}

	return Sign(privKey, signingData), nil
}

// VerifyPayloadSignature はペイロードの署名を検証する
func VerifyPayloadSignature(pubKey ed25519.PublicKey, payload *core.BlockPayload, signatureBase64 string) bool {
	signingData, err := MakeSigningPayload(payload)
	if err != nil {
		return false
	}

	return Verify(pubKey, signingData, signatureBase64)
}

// SignTransaction はトランザクションデータに署名する
func SignTransaction(privKey ed25519.PrivateKey, tx *core.TransactionData) (string, error) {
	data, err := json.Marshal(tx)
	if err != nil {
		return "", err
	}

	return Sign(privKey, data), nil
}

// VerifyTransactionSignature はトランザクションの署名を検証する
func VerifyTransactionSignature(pubKey ed25519.PublicKey, tx *core.TransactionData, signatureBase64 string) bool {
	data, err := json.Marshal(tx)
	if err != nil {
		return false
	}

	return Verify(pubKey, data, signatureBase64)
}

// SignData は生データに署名するヘルパー関数
func SignData(privKey ed25519.PrivateKey, data string) string {
	return Sign(privKey, []byte(data))
}

// VerifyDataSignature は生データの署名を検証するヘルパー関数
func VerifyDataSignature(pubKey ed25519.PublicKey, data string, signatureBase64 string) bool {
	return Verify(pubKey, []byte(data), signatureBase64)
}
