package core

// TransactionData は金銭的取引のデータを表す
type TransactionData struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Amount int64  `json:"amount"`
	Title  string `json:"title"`
}

// AddNodeData はノード追加のデータを表す
type AddNodeData struct {
	PublicKey string `json:"public_key"`
	NodeName  string `json:"node_name"`
	NickName  string `json:"nick_name"`
	Address   string `json:"address"`
}
