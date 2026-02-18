package storage

// NodeInfo はピアノードの情報を表す
type NodeInfo struct {
	Name      string `json:"name"`
	NickName  string `json:"nick_name"`
	Address   string `json:"address"`
	PublicKey string `json:"public_key"`
}
