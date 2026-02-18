package p2p

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"signet/core"
	"signet/storage"
)

// SyncChain は全ピアからチェーンを取得し、最長チェーンで同期する
func SyncChain(chain *core.Chain, peers map[string]*storage.NodeInfo) error {
	var longestBlocks []*core.Block
	maxLen := chain.Len()

	// 現在のチェーンを初期値として設定
	longestBlocks = chain.GetBlocks()

	for name, peer := range peers {
		blocks, err := fetchChain(peer.Address)
		if err != nil {
			// エラーはログに出力して続行
			fmt.Printf("Warning: failed to fetch chain from %s (%s): %v\n", name, peer.Address, err)
			continue
		}

		// より長いチェーンが見つかったら更新
		if len(blocks) > maxLen {
			maxLen = len(blocks)
			longestBlocks = blocks
		}
	}

	// 自分より長いチェーンが見つかった場合は置換
	if len(longestBlocks) > chain.Len() {
		if err := chain.ReplaceChain(longestBlocks); err != nil {
			return fmt.Errorf("failed to replace chain: %w", err)
		}
		fmt.Printf("Chain synced: %d blocks\n", len(longestBlocks))
	}

	return nil
}

// fetchChain は指定したアドレスからチェーンを取得する
func fetchChain(addr string) ([]*core.Block, error) {
	url := fmt.Sprintf("http://%s/chain", addr)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var blocks []*core.Block
	if err := json.NewDecoder(resp.Body).Decode(&blocks); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return blocks, nil
}
