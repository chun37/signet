package config

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// ParseTOML は簡易TOMLパーサー（key = value 形式のみサポート）
func ParseTOML(r io.Reader) (map[string]string, error) {
	result := make(map[string]string)
	scanner := bufio.NewScanner(r)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// 空行は無視
		if line == "" {
			continue
		}

		// コメントは無視
		if strings.HasPrefix(line, "#") {
			continue
		}

		// key = value 形式を解析
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid format at line %d: %s", lineNum, line)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// クォートがあれば除去
		if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) ||
			(strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
			value = value[1 : len(value)-1]
		}

		result[key] = value
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading: %w", err)
	}

	return result, nil
}

// ParseTOMLFile はファイルからTOMLを読み込む
func ParseTOMLFile(path string) (map[string]string, error) {
	f, err := openFile(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ParseTOML(f)
}
