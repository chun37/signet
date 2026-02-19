package core

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// PendingTransaction は承認待ちのトランザクションを表す
type PendingTransaction struct {
	ID        string       `json:"id"`
	CreatedAt time.Time    `json:"created_at"`
	Payload   BlockPayload `json:"payload"`
}

// PendingPool は承認待ちトランザクションのプールを表す
type PendingPool struct {
	mu    sync.RWMutex
	items map[string]*PendingTransaction
}

// NewPendingPool は新しい承認待ちプールを作成する
func NewPendingPool() *PendingPool {
	return &PendingPool{
		items: make(map[string]*PendingTransaction),
	}
}

// Add は承認待ちトランザクションを追加する
func (p *PendingPool) Add(pt *PendingTransaction) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.items[pt.ID] = pt
}

// Remove は指定したIDの承認待ちトランザクションを削除する
func (p *PendingPool) Remove(id string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	delete(p.items, id)
}

// Get は指定したIDの承認待ちトランザクションを返す
func (p *PendingPool) Get(id string) *PendingTransaction {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.items[id]
}

// List は全承認待ちトランザクションのリストを返す
func (p *PendingPool) List() []*PendingTransaction {
	p.mu.RLock()
	defer p.mu.RUnlock()

	result := make([]*PendingTransaction, 0, len(p.items))
	for _, pt := range p.items {
		result = append(result, pt)
	}

	return result
}

// GetAll は全承認待ちトランザクションのマップを返す
func (p *PendingPool) GetAll() map[string]*PendingTransaction {
	p.mu.RLock()
	defer p.mu.RUnlock()

	result := make(map[string]*PendingTransaction, len(p.items))
	for k, v := range p.items {
		result[k] = v
	}

	return result
}

// Len はプール内のトランザクション数を返す
func (p *PendingPool) Len() int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return len(p.items)
}

// Has は指定したIDのトランザクションが存在するかを返す
func (p *PendingPool) Has(id string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	_, exists := p.items[id]
	return exists
}

// Clear はプールをクリアする
func (p *PendingPool) Clear() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.items = make(map[string]*PendingTransaction)
}

// GetByToNode は指定したノード宛のトランザクションを返す
func (p *PendingPool) GetByToNode(nodeName string) []*PendingTransaction {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var result []*PendingTransaction
	for _, pt := range p.items {
		if pt.Payload.Type == "transaction" {
			var txData TransactionData
			if err := json.Unmarshal(pt.Payload.Data, &txData); err == nil && txData.To == nodeName {
				result = append(result, pt)
			}
		}
	}

	return result
}

// GetByFromNode は指定したノードが提案したトランザクションを返す
func (p *PendingPool) GetByFromNode(nodeName string) []*PendingTransaction {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var result []*PendingTransaction
	for _, pt := range p.items {
		if pt.Payload.Type == "transaction" {
			var txData TransactionData
			if err := json.Unmarshal(pt.Payload.Data, &txData); err == nil && txData.From == nodeName {
				result = append(result, pt)
			}
		}
	}

	return result
}

// NewPendingTransaction は新しい承認待ちトランザクションを作成する
func NewPendingTransaction(id string, payload BlockPayload) *PendingTransaction {
	return &PendingTransaction{
		ID:        id,
		CreatedAt: time.Now().UTC(),
		Payload:   payload,
	}
}

// GenerateID は一意なIDを生成する（ハッシュベース）
func GenerateID(payload BlockPayload, t time.Time) string {
	data := fmt.Sprintf("%d%s%s", t.UnixNano(), payload.Type, string(payload.Data))
	return CalcSHA256(data)
}

// GetTransactionData はPendingTransactionのペイロードからTransactionDataを取得する
func (pt *PendingTransaction) GetTransactionData() (*TransactionData, error) {
	if pt.Payload.Type != "transaction" {
		return nil, fmt.Errorf("payload type is not transaction: %s", pt.Payload.Type)
	}

	var data TransactionData
	if err := json.Unmarshal(pt.Payload.Data, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal transaction data: %w", err)
	}

	return &data, nil
}
