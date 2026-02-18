package crypto

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"testing"

	"signet/core"
)

func TestSignAndVerify(t *testing.T) {
	pub, priv, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair failed: %v", err)
	}

	data := []byte("test message")

	signature := Sign(priv, data)
	if signature == "" {
		t.Fatal("Sign returned empty string")
	}

	if !Verify(pub, data, signature) {
		t.Error("Verify failed for valid signature")
	}

	// データを変更すると検証が失敗することを確認
	wrongData := []byte("wrong message")
	if Verify(pub, wrongData, signature) {
		t.Error("Verify should fail for modified data")
	}
}

func TestSign_VerifyWithWrongKey(t *testing.T) {
	_, priv1, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair failed: %v", err)
	}

	pub2, _, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair failed: %v", err)
	}

	data := []byte("test message")
	signature := Sign(priv1, data)

	// 違う公開鍵で検証
	if Verify(pub2, data, signature) {
		t.Error("Verify should fail with different public key")
	}
}

func TestVerify_InvalidSignature(t *testing.T) {
	pub, _, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair failed: %v", err)
	}

	data := []byte("test message")

	// 不正なBase64
	if Verify(pub, data, "!!!invalid!!!") {
		t.Error("Verify should fail for invalid base64")
	}

	// 空文字列
	if Verify(pub, data, "") {
		t.Error("Verify should fail for empty signature")
	}

	// 短すぎる署名
	if Verify(pub, data, "aGVsbG8=") { // "hello" in base64
		t.Error("Verify should fail for short signature")
	}
}

func TestMakeSigningPayload(t *testing.T) {
	txData := &core.TransactionData{
		From:   "node1",
		To:     "node2",
		Amount: 1000,
		Title:  "test",
	}

	data, _ := json.Marshal(txData)
	payload := &core.BlockPayload{
		Type:          "transaction",
		Data:          json.RawMessage(data),
		FromSignature: "sig1",
		ToSignature:   "sig2",
	}

	signingData, err := MakeSigningPayload(payload)
	if err != nil {
		t.Fatalf("MakeSigningPayload failed: %v", err)
	}

	// JSONとして有効か確認
	var result map[string]interface{}
	if err := json.Unmarshal(signingData, &result); err != nil {
		t.Errorf("Signing payload is not valid JSON: %v", err)
	}

	if result["type"] != "transaction" {
		t.Errorf("type = %v, want transaction", result["type"])
	}

	// 署名フィールドは含まれない
	if _, exists := result["from_signature"]; exists {
		t.Error("from_signature should not be in signing payload")
	}
}

func TestSignPayload_VerifyPayloadSignature(t *testing.T) {
	pub, priv, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair failed: %v", err)
	}

	txData := &core.TransactionData{
		From:   "node1",
		To:     "node2",
		Amount: 1000,
		Title:  "test",
	}

	data, _ := json.Marshal(txData)
	payload := &core.BlockPayload{
		Type:          "transaction",
		Data:          json.RawMessage(data),
		FromSignature: "",
		ToSignature:   "",
	}

	// 署名
	signature, err := SignPayload(priv, payload)
	if err != nil {
		t.Fatalf("SignPayload failed: %v", err)
	}

	// 検証
	if !VerifyPayloadSignature(pub, payload, signature) {
		t.Error("VerifyPayloadSignature failed for valid signature")
	}

	// ペイロードを変更すると検証が失敗することを確認
	wrongTxData := &core.TransactionData{
		From:   "node1",
		To:     "node2",
		Amount: 9999, // 異なる金額
		Title:  "test",
	}
	wrongData, _ := json.Marshal(wrongTxData)
	wrongPayload := &core.BlockPayload{
		Type:          "transaction",
		Data:          json.RawMessage(wrongData),
		FromSignature: "",
		ToSignature:   "",
	}

	if VerifyPayloadSignature(pub, wrongPayload, signature) {
		t.Error("VerifyPayloadSignature should fail for modified payload")
	}
}

func TestSignTransaction_VerifyTransactionSignature(t *testing.T) {
	pub, priv, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair failed: %v", err)
	}

	tx := &core.TransactionData{
		From:   "node1",
		To:     "node2",
		Amount: 5000,
		Title:  "dinner",
	}

	// 署名
	signature, err := SignTransaction(priv, tx)
	if err != nil {
		t.Fatalf("SignTransaction failed: %v", err)
	}

	// 検証
	if !VerifyTransactionSignature(pub, tx, signature) {
		t.Error("VerifyTransactionSignature failed for valid signature")
	}

	// 変更されたトランザクションで検証
	modifiedTx := &core.TransactionData{
		From:   "node1",
		To:     "node2",
		Amount: 10000, // 異なる金額
		Title:  "dinner",
	}

	if VerifyTransactionSignature(pub, modifiedTx, signature) {
		t.Error("VerifyTransactionSignature should fail for modified transaction")
	}
}

func TestSignData_VerifyDataSignature(t *testing.T) {
	pub, priv, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair failed: %v", err)
	}

	data := "important message"

	// 署名
	signature := SignData(priv, data)
	if signature == "" {
		t.Fatal("SignData returned empty string")
	}

	// 検証
	if !VerifyDataSignature(pub, data, signature) {
		t.Error("VerifyDataSignature failed for valid signature")
	}

	// 異なるデータで検証
	if VerifyDataSignature(pub, "different message", signature) {
		t.Error("VerifyDataSignature should fail for different data")
	}
}

func TestSignAndVerify_MultipleMessages(t *testing.T) {
	pub, priv, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair failed: %v", err)
	}

	messages := []string{
		"hello",
		"world",
		"test message 123",
		"日本語のメッセージ",
		"",
	}

	for _, msg := range messages {
		t.Run(msg, func(t *testing.T) {
			signature := SignData(priv, msg)
			if !VerifyDataSignature(pub, msg, signature) {
				t.Errorf("Verify failed for message: %q", msg)
			}

			// 異なるメッセージでは失敗
			if msg != "" && VerifyDataSignature(pub, msg+"x", signature) {
				t.Error("Verify should fail for modified message")
			}
		})
	}
}

func TestSignAndVerify_AddNodePayload(t *testing.T) {
	pub, priv, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair failed: %v", err)
	}

	addNodeData := &core.AddNodeData{
		PublicKey: "pubkey123",
		NodeName:  "node1",
		NickName:  "Tanaka",
		Address:   "10.0.0.1",
	}

	data, _ := json.Marshal(addNodeData)
	payload := &core.BlockPayload{
		Type:          "add_node",
		Data:          json.RawMessage(data),
		FromSignature: "",
		ToSignature:   "",
	}

	// 署名
	signature, err := SignPayload(priv, payload)
	if err != nil {
		t.Fatalf("SignPayload failed: %v", err)
	}

	// 検証
	if !VerifyPayloadSignature(pub, payload, signature) {
		t.Error("VerifyPayloadSignature failed for add_node payload")
	}
}

func TestSignatureDeterminism(t *testing.T) {
	_, priv, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair failed: %v", err)
	}

	data := []byte("test")

	sig1 := Sign(priv, data)
	sig2 := Sign(priv, data)

	// Ed25519の署名は毎回異なる値になる（ランダム性がある）
	// ただし両方とも検証には成功するはず
	pub := priv.Public().(ed25519.PublicKey)

	if !Verify(pub, data, sig1) {
		t.Error("First signature is not valid")
	}

	if !Verify(pub, data, sig2) {
		t.Error("Second signature is not valid")
	}
}

func TestSignWithEmptyData(t *testing.T) {
	pub, priv, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair failed: %v", err)
	}

	data := []byte("")

	signature := Sign(priv, data)
	if !Verify(pub, data, signature) {
		t.Error("Verify failed for empty data")
	}
}

func TestSignAndVerify_SignatureFormat(t *testing.T) {
	_, priv, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair failed: %v", err)
	}

	data := []byte("test")

	signature := Sign(priv, data)

	// Base64エンコードされていることを確認
	decoded, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		t.Fatalf("Signature is not valid base64: %v", err)
	}

	// Ed25519の署名サイズは64バイト
	if len(decoded) != ed25519.SignatureSize {
		t.Errorf("Signature size = %d, want %d", len(decoded), ed25519.SignatureSize)
	}
}
