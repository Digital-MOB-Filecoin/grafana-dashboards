package model

// ChainEconomics is the DB model for the chain_economics table.
// Circulating supply fields come from StateVMCirculatingSupplyInternal;
// power actor fields come from the power actor state.
type ChainEconomics struct {
	Height    int64
	StateRoot CID

	// From StateVMCirculatingSupplyInternal
	CirculatingFIL      string
	VestedFIL           string
	MinedFIL            string
	BurntFIL            string
	LockedFIL           string
	FilReserveDisbursed string

	// From power actor state
	TotalRawBytesPower         string
	TotalRawBytesCommitted     string
	TotalQABytesPower          string
	TotalQABytesCommitted      string
	TotalPledgeCollateral      string
	QASmoothedPositionEstimate string
	QASmoothedVelocityEstimate string
	MinerCount                 *int64
	ParticipatingMinerCount    *int64
}

// CirculatingSupply is the RPC response from StateVMCirculatingSupplyInternal.
type CirculatingSupply struct {
	FilVested           string `json:"FilVested"`
	FilMined            string `json:"FilMined"`
	FilBurnt            string `json:"FilBurnt"`
	FilLocked           string `json:"FilLocked"`
	FilCirculating      string `json:"FilCirculating"`
	FilReserveDisbursed string `json:"FilReserveDisbursed"`
}
