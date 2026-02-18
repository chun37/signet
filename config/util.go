package config

import (
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

// openFile はファイルを開くヘルパー関数
func openFile(path string) (*os.File, error) {
	return os.Open(path)
}
