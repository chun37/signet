package core

import (
	"crypto/sha256"
	"encoding/hex"
)

// CalcSHA256 は与えられた文字列のSHA-256ハッシュを計算し、hexエンコードして返す
func CalcSHA256(data string) string {
	h := sha256.New()
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}
