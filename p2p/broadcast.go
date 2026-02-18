package p2p

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"signet/storage"
	"sync"
)

// BroadcastBlock は全ピア（自分以外）にブロックを送信する
// block は server.Block 型に変換済みのものを渡すこと
func BroadcastBlock(block any, peers map[string]*storage.NodeInfo, selfName string) {
	var wg sync.WaitGroup

	for name, peer := range peers {
		if name == selfName {
			continue // 自分には送信しない
		}

		wg.Add(1)
		go func(nodeName string, addr string) {
			defer wg.Done()

			if err := sendBlock(addr, block); err != nil {
				// エラーはログに出力するだけ（送信失敗しても続行）
				fmt.Printf("Warning: failed to send block to %s (%s): %v\n", nodeName, addr, err)
			}
		}(name, peer.Address)
	}

	wg.Wait()
}

// sendBlock は指定したアドレスにブロックをPOSTする
func sendBlock(addr string, block any) error {
	// JSONエンコード
	data, err := json.Marshal(block)
	if err != nil {
		return fmt.Errorf("failed to marshal block: %w", err)
	}

	// POSTリクエスト
	url := fmt.Sprintf("http://%s/block", addr)
	resp, err := http.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// ステータスコードチェック
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}
