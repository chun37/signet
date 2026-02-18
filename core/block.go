package core

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// BlockHeader はブロックのヘッダーを表す
type BlockHeader struct {
	Index     int       `json:"index"`
	CreatedAt time.Time `json:"created_at"`
	PrevHash  string    `json:"prev_hash"`
	Hash      string    `json:"hash"`
}

// BlockPayload はブロックのペイロードを表す
type BlockPayload struct {
	Type          string          `json:"type"`
	Data          json.RawMessage `json:"data"`
	FromSignature string          `json:"from_signature"`
	ToSignature   string          `json:"to_signature"`
}

// Block はブロックチェーンの1つのブロックを表す
type Block struct {
	Header  BlockHeader  `json:"header"`
	Payload BlockPayload `json:"payload"`
}

// CalcBlockHash はブロックのハッシュを計算する
// Index + CreatedAt(RFC3339) + PrevHash + Payload(JSON) を連結してSHA-256
func CalcBlockHash(b *Block) string {
	payloadJSON, err := json.Marshal(b.Payload)
	if err != nil {
		return ""
	}

	data := fmt.Sprintf("%d%s%s%s", b.Header.Index, b.Header.CreatedAt.Format(time.RFC3339), b.Header.PrevHash, string(payloadJSON))
	h := sha256.New()
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// NewBlock は新しいブロックを生成する
func NewBlock(index int, prevHash string, payload BlockPayload) *Block {
	now := time.Now().UTC()
	block := &Block{
		Header: BlockHeader{
			Index:     index,
			CreatedAt: now,
			PrevHash:  prevHash,
		},
		Payload: payload,
	}
	block.Header.Hash = CalcBlockHash(block)
	return block
}

// NewGenesisBlock はジェネシスブロックを生成する
// addNode に初期化ノードの情報を渡す
func NewGenesisBlock(addNode *AddNodeData) *Block {
	data, _ := json.Marshal(addNode)
	payload := BlockPayload{
		Type:          "add_node",
		Data:          json.RawMessage(data),
		FromSignature: "",
		ToSignature:   "",
	}

	block := &Block{
		Header: BlockHeader{
			Index:     0,
			CreatedAt: time.Time{}.UTC(), // ゼロ値
			PrevHash:  "0",
		},
		Payload: payload,
	}
	block.Header.Hash = CalcBlockHash(block)
	return block
}

// ValidateBlock はブロックのハッシュが正しいか検証する
func ValidateBlock(b *Block) error {
	calculatedHash := CalcBlockHash(b)
	if calculatedHash != b.Header.Hash {
		return fmt.Errorf("invalid block hash: expected %s, got %s", calculatedHash, b.Header.Hash)
	}

	// PayloadのTypeが有効かチェック
	validTypes := map[string]bool{
		"transaction": true,
		"add_node":    true,
	}
	if !validTypes[b.Payload.Type] {
		return fmt.Errorf("invalid payload type: %s", b.Payload.Type)
	}

	return nil
}

// IsValidBlockType はブロックタイプが有効かを返す
func IsValidBlockType(blockType string) bool {
	validTypes := map[string]bool{
		"transaction": true,
		"add_node":    true,
	}
	return validTypes[blockType]
}

// GetTransactionData はペイロードからTransactionDataを取り出す
func (b *Block) GetTransactionData() (*TransactionData, error) {
	if b.Payload.Type != "transaction" {
		return nil, fmt.Errorf("payload type is not transaction: %s", b.Payload.Type)
	}

	var data TransactionData
	if err := json.Unmarshal(b.Payload.Data, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal transaction data: %w", err)
	}
	return &data, nil
}

// GetAddNodeData はペイロードからAddNodeDataを取り出す
func (b *Block) GetAddNodeData() (*AddNodeData, error) {
	if b.Payload.Type != "add_node" {
		return nil, fmt.Errorf("payload type is not add_node: %s", b.Payload.Type)
	}

	var data AddNodeData
	if err := json.Unmarshal(b.Payload.Data, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal add_node data: %w", err)
	}
	return &data, nil
}

// SetTransactionData はペイロードにTransactionDataを設定する
func SetTransactionData(tx *TransactionData) (json.RawMessage, error) {
	data, err := json.Marshal(tx)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal transaction data: %w", err)
	}
	return json.RawMessage(data), nil
}

// SetAddNodeData はペイロードにAddNodeDataを設定する
func SetAddNodeData(addNode *AddNodeData) (json.RawMessage, error) {
	data, err := json.Marshal(addNode)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal add_node data: %w", err)
	}
	return json.RawMessage(data), nil
}

// CreateBlockWithTransaction はトランザクションデータを含むブロックを作成する
func CreateBlockWithTransaction(index int, prevHash string, tx *TransactionData, fromSig, toSig string) (*Block, error) {
	data, err := SetTransactionData(tx)
	if err != nil {
		return nil, err
	}

	payload := BlockPayload{
		Type:          "transaction",
		Data:          data,
		FromSignature: fromSig,
		ToSignature:   toSig,
	}

	return NewBlock(index, prevHash, payload), nil
}

// CreateBlockWithAddNode はノード追加データを含むブロックを作成する
func CreateBlockWithAddNode(index int, prevHash string, addNode *AddNodeData) (*Block, error) {
	data, err := SetAddNodeData(addNode)
	if err != nil {
		return nil, err
	}

	payload := BlockPayload{
		Type:          "add_node",
		Data:          data,
		FromSignature: "",
		ToSignature:   "",
	}

	return NewBlock(index, prevHash, payload), nil
}

// MakeSigningPayload は署名対象のペイロードバイト列を作成する
// Type + Data をJSON直列化して連結
func MakeSigningPayload(payload *BlockPayload) ([]byte, error) {
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

// HashWithoutSignature は署名用ハッシュを計算する（署名を除いたペイロード）
func (b *Block) HashWithoutSignature() (string, error) {
	signingData, err := MakeSigningPayload(&b.Payload)
	if err != nil {
		return "", err
	}

	h := sha256.New()
	h.Write(signingData)
	return hex.EncodeToString(h.Sum(nil)), nil
}

// VerifySignatures はブロックの署名を検証する（公開鍵検証は呼び出し側で行う）
func (b *Block) VerifySignatures() (fromSigValid, toSigValid bool, err error) {
	signingData, err := MakeSigningPayload(&b.Payload)
	if err != nil {
		return false, false, err
	}

	// 署名の有無をチェック（空文字列は署名なしとして扱う）
	fromSigValid = b.Payload.FromSignature != ""
	toSigValid = b.Payload.ToSignature != ""

	_ = signingData // 実際の署名検証はcryptoパッケージで行う
	return fromSigValid, toSigValid, nil
}

// BlockType はブロックの種類を表す
type BlockType string

const (
	BlockTypeTransaction BlockType = "transaction"
	BlockTypeAddNode     BlockType = "add_node"
)

// String はBlockTypeの文字列表現を返す
func (bt BlockType) String() string {
	return string(bt)
}

// ParseBlockType は文字列からBlockTypeをパースする
func ParseBlockType(s string) (BlockType, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	switch s {
	case "transaction":
		return BlockTypeTransaction, nil
	case "add_node":
		return BlockTypeAddNode, nil
	default:
		return "", fmt.Errorf("unknown block type: %s", s)
	}
}

// IsGenesisBlock はジェネシスブロックかどうかを判定する
func (b *Block) IsGenesisBlock() bool {
	return b.Header.Index == 0 && b.Header.PrevHash == "0"
}

// JSONRawMessage はjson.RawMessageの型エイリアス（cryptoパッケージから使用）
type JSONRawMessage = json.RawMessage

// MarshalJSON は構造体をJSONにマーシャルする
func MarshalJSON(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// MarshalTransactionData はTransactionDataをJSONにマーシャルする
func MarshalTransactionData(tx *TransactionData) ([]byte, error) {
	return json.Marshal(tx)
}
