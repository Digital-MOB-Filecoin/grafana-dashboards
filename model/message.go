package model

// Message is the RPC response from ChainGetParentMessages.
type Message struct {
	CID        CID    `json:"CID"`
	Message    *RawMessage `json:"Message"`
}

// RawMessage holds the fields of a Filecoin message.
type RawMessage struct {
	Version    uint64 `json:"Version"`
	To         string `json:"To"`
	From       string `json:"From"`
	Nonce      uint64 `json:"Nonce"`
	Value      string `json:"Value"`
	GasLimit   int64  `json:"GasLimit"`
	GasFeeCap  string `json:"GasFeeCap"`
	GasPremium string `json:"GasPremium"`
	Method     uint64 `json:"Method"`
	Params     string `json:"Params"` // base64-encoded
}

// MessageReceipt is the RPC response from ChainGetParentReceipts.
type MessageReceipt struct {
	ExitCode int64  `json:"ExitCode"`
	Return   string `json:"Return"` // base64-encoded
	GasUsed  int64  `json:"GasUsed"`
}
