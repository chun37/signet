package core

import (
	"fmt"
	"sync"
)

// Chain はブロックチェーンを表す
type Chain struct {
	mu      sync.RWMutex
	blocks  []*Block
	hashSet map[string]struct{} // 重複検知用
}

// NewChain は新しいブロックチェーンを作成する
func NewChain() *Chain {
	genesis := NewGenesisBlock()
	hashSet := make(map[string]struct{})
	hashSet[genesis.Header.Hash] = struct{}{}

	return &Chain{
		blocks:  []*Block{genesis},
		hashSet: hashSet,
	}
}

// NewChainFromBlocks はストレージから読んだブロックでチェーンを構築する
// ジェネシスブロックの二重生成を防ぐ
func NewChainFromBlocks(blocks []*Block) (*Chain, error) {
	if len(blocks) == 0 {
		return nil, fmt.Errorf("blocks is empty")
	}

	if !blocks[0].IsGenesisBlock() {
		return nil, fmt.Errorf("first block is not a genesis block")
	}

	hashSet := make(map[string]struct{})
	for _, b := range blocks {
		hashSet[b.Header.Hash] = struct{}{}
	}

	chain := &Chain{
		blocks:  make([]*Block, len(blocks)),
		hashSet: hashSet,
	}
	copy(chain.blocks, blocks)

	return chain, nil
}

// AddBlock はブロックをチェーンに追加する
func (c *Chain) AddBlock(b *Block) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// ブロックの検証
	if err := ValidateBlock(b); err != nil {
		return fmt.Errorf("block validation failed: %w", err)
	}

	// 前のブロックのハッシュをチェック
	if len(c.blocks) > 0 {
		lastBlock := c.blocks[len(c.blocks)-1]
		if b.Header.PrevHash != lastBlock.Header.Hash {
			return fmt.Errorf("prev_hash mismatch: expected %s, got %s", lastBlock.Header.Hash, b.Header.PrevHash)
		}

		// インデックスが連続しているかチェック
		if b.Header.Index != lastBlock.Header.Index+1 {
			return fmt.Errorf("index mismatch: expected %d, got %d", lastBlock.Header.Index+1, b.Header.Index)
		}
	}

	// 重複チェック
	if _, exists := c.hashSet[b.Header.Hash]; exists {
		return fmt.Errorf("duplicate block: %s", b.Header.Hash)
	}

	c.blocks = append(c.blocks, b)
	c.hashSet[b.Header.Hash] = struct{}{}

	return nil
}

// GetBlocks は全ブロックのコピーを返す
func (c *Chain) GetBlocks() []*Block {
	c.mu.RLock()
	defer c.mu.RUnlock()

	blocks := make([]*Block, len(c.blocks))
	copy(blocks, c.blocks)
	return blocks
}

// LastBlock は最後のブロックを返す
func (c *Chain) LastBlock() *Block {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.blocks) == 0 {
		return nil
	}
	return c.blocks[len(c.blocks)-1]
}

// Len はチェーンの長さを返す
func (c *Chain) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.blocks)
}

// ValidateChain はチェーン全体の整合性を検証する
func (c *Chain) ValidateChain() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.blocks) == 0 {
		return fmt.Errorf("empty chain")
	}

	// ジェネシスブロックのチェック
	genesis := c.blocks[0]
	if !genesis.IsGenesisBlock() {
		return fmt.Errorf("first block is not a valid genesis block")
	}

	// 各ブロックの検証
	for i := 1; i < len(c.blocks); i++ {
		current := c.blocks[i]
		prev := c.blocks[i-1]

		// ブロック自体のハッシュ検証
		if err := ValidateBlock(current); err != nil {
			return fmt.Errorf("block at index %d validation failed: %w", i, err)
		}

		// 前のブロックとの連結検証
		if current.Header.PrevHash != prev.Header.Hash {
			return fmt.Errorf("block at index %d has invalid prev_hash: expected %s, got %s",
				i, prev.Header.Hash, current.Header.PrevHash)
		}

		// インデックスの連続性
		if current.Header.Index != prev.Header.Index+1 {
			return fmt.Errorf("block at index %d has invalid index: expected %d, got %d",
				i, prev.Header.Index+1, current.Header.Index)
		}
	}

	return nil
}

// ReplaceChain はチェーンを置換する（最長チェーンルール）
func (c *Chain) ReplaceChain(blocks []*Block) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 新しいチェーンが空でないこと
	if len(blocks) == 0 {
		return fmt.Errorf("new chain is empty")
	}

	// 新しいチェーンが現在より長いこと
	if len(blocks) <= len(c.blocks) {
		return fmt.Errorf("new chain is not longer: new length %d, current length %d",
			len(blocks), len(c.blocks))
	}

	// 新しいチェーンの検証
	newChain := &Chain{
		blocks:  make([]*Block, len(blocks)),
		hashSet: make(map[string]struct{}),
	}
	copy(newChain.blocks, blocks)

	for _, b := range blocks {
		// ブロックの検証
		if err := ValidateBlock(b); err != nil {
			return fmt.Errorf("new chain contains invalid block: %w", err)
		}

		// 重複チェック
		if _, exists := newChain.hashSet[b.Header.Hash]; exists {
			return fmt.Errorf("new chain contains duplicate block: %s", b.Header.Hash)
		}
		newChain.hashSet[b.Header.Hash] = struct{}{}
	}

	// 連結性の検証
	if !blocks[0].IsGenesisBlock() {
		return fmt.Errorf("new chain does not start with genesis block")
	}

	for i := 1; i < len(blocks); i++ {
		current := blocks[i]
		prev := blocks[i-1]

		if current.Header.PrevHash != prev.Header.Hash {
			return fmt.Errorf("new chain has broken link at index %d", i)
		}

		if current.Header.Index != prev.Header.Index+1 {
			return fmt.Errorf("new chain has invalid index at %d", i)
		}
	}

	// チェーンを置換
	c.blocks = newChain.blocks
	c.hashSet = newChain.hashSet

	return nil
}

// HasBlock は指定したハッシュのブロックが存在するかを返す
func (c *Chain) HasBlock(hash string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	_, exists := c.hashSet[hash]
	return exists
}

// GetBlockByIndex は指定したインデックスのブロックを返す
func (c *Chain) GetBlockByIndex(index int) (*Block, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if index < 0 || index >= len(c.blocks) {
		return nil, fmt.Errorf("index out of range: %d", index)
	}

	return c.blocks[index], nil
}

// GetBlockByHash は指定したハッシュのブロックを返す
func (c *Chain) GetBlockByHash(hash string) (*Block, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, b := range c.blocks {
		if b.Header.Hash == hash {
			return b, nil
		}
	}

	return nil, fmt.Errorf("block not found: %s", hash)
}

// ForEach はチェーン内の各ブロックに対して関数を実行する
func (c *Chain) ForEach(fn func(b *Block) error) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, b := range c.blocks {
		if err := fn(b); err != nil {
			return err
		}
	}

	return nil
}

// Clone はチェーンのディープコピーを作成する
func (c *Chain) Clone() *Chain {
	c.mu.RLock()
	defer c.mu.RUnlock()

	blocks := make([]*Block, len(c.blocks))
	copy(blocks, c.blocks)

	hashSet := make(map[string]struct{}, len(c.hashSet))
	for k := range c.hashSet {
		hashSet[k] = struct{}{}
	}

	return &Chain{
		blocks:  blocks,
		hashSet: hashSet,
	}
}

// GetLastHash は最後のブロックのハッシュを返す
func (c *Chain) GetLastHash() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.blocks) == 0 {
		return ""
	}

	return c.blocks[len(c.blocks)-1].Header.Hash
}

// GetLastIndex は最後のブロックのインデックスを返す
func (c *Chain) GetLastIndex() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.blocks) == 0 {
		return -1
	}

	return c.blocks[len(c.blocks)-1].Header.Index
}
