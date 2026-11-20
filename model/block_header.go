package model

// BlockHeader is the DB model for the block_headers table.
// Fields are populated from ChainGetTipSet.
type BlockHeader struct {
	Height          int64
	CID             CID
	MinerID         int64
	ParentWeight    string // attoFIL as decimal string → postgres numeric
	ParentBaseFee   string
	ParentStateRoot CID
	Timestamp       int64
	WinCount        *int64 // nullable
	ForkSignaling   int64
}

// TipSet is the RPC response from ChainGetHead / ChainGetTipSet.
type TipSet struct {
	Cids   TipSetKey         `json:"Cids"`
	Blocks []TipSetBlock     `json:"Blocks"`
	Height int64             `json:"Height"`
}

// TipSetBlock is the raw block header as returned by the Lotus RPC.
type TipSetBlock struct {
	Miner           string         `json:"Miner"`
	Parents         TipSetKey      `json:"Parents"`
	ParentWeight    string         `json:"ParentWeight"`
	Height          int64          `json:"Height"`
	ParentStateRoot CID            `json:"ParentStateRoot"`
	Timestamp       uint64         `json:"Timestamp"`
	ElectionProof   *ElectionProof `json:"ElectionProof,omitempty"`
	ForkSignaling   uint64         `json:"ForkSignaling"`
	ParentBaseFee   string         `json:"ParentBaseFee"`
}

type ElectionProof struct {
	WinCount int64 `json:"WinCount"`
}
