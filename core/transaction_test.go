package core

import (
	"encoding/json"
	"testing"
)

func TestTransactionData_JSON(t *testing.T) {
	data := TransactionData{
		From:   "node1",
		To:     "node2",
		Amount: 1000,
		Title:  "lunch",
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var decoded TransactionData
	if err := json.Unmarshal(bytes, &decoded); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if decoded.From != data.From {
		t.Errorf("From = %q, want %q", decoded.From, data.From)
	}
	if decoded.To != data.To {
		t.Errorf("To = %q, want %q", decoded.To, data.To)
	}
	if decoded.Amount != data.Amount {
		t.Errorf("Amount = %d, want %d", decoded.Amount, data.Amount)
	}
	if decoded.Title != data.Title {
		t.Errorf("Title = %q, want %q", decoded.Title, data.Title)
	}
}

func TestAddNodeData_JSON(t *testing.T) {
	data := AddNodeData{
		PublicKey: "abcd1234",
		NodeName:  "node1",
		NickName:  "Tanaka",
		Address:   "10.0.0.1",
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var decoded AddNodeData
	if err := json.Unmarshal(bytes, &decoded); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if decoded.PublicKey != data.PublicKey {
		t.Errorf("PublicKey = %q, want %q", decoded.PublicKey, data.PublicKey)
	}
	if decoded.NodeName != data.NodeName {
		t.Errorf("NodeName = %q, want %q", decoded.NodeName, data.NodeName)
	}
	if decoded.NickName != data.NickName {
		t.Errorf("NickName = %q, want %q", decoded.NickName, data.NickName)
	}
	if decoded.Address != data.Address {
		t.Errorf("Address = %q, want %q", decoded.Address, data.Address)
	}
}
