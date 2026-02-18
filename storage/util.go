package storage

import (
	"encoding/json"
	"os"
)

// readFile はファイルを読み込むヘルパー関数
func readFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// writeFile はファイルに書き込むヘルパー関数
func writeFile(path string, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}

// appendFile はファイルに追記するヘルパー関数
func appendFile(path string, data []byte) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(data)
	return err
}

// splitLines はバイト列を行ごとに分割するヘルパー関数
func splitLines(data []byte) [][]byte {
	var lines [][]byte
	start := 0
	for i, b := range data {
		if b == '\n' {
			lines = append(lines, data[start:i])
			start = i + 1
		}
	}
	if start < len(data) {
		lines = append(lines, data[start:])
	}
	return lines
}

// openFile はファイルを開くヘルパー関数
func openFile(path string) (*os.File, error) {
	return os.Open(path)
}

// encodeJSON はJSONエンコードのヘルパー関数
func encodeJSON(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}
