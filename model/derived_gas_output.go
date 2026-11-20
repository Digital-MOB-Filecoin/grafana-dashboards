package model

// DerivedGasOutput is the DB model for the derived_gas_outputs table.
// It is assembled from ChainGetParentMessages + ChainGetParentReceipts.
// Computed columns (base_fee_burn, over_estimation_burn, miner_tip, refund)
// are GENERATED ALWAYS in Postgres 18 and are not set here.
type DerivedGasOutput struct {
	Height        int64
	IndexInBlock  *int
	CID           CID
	StateRoot     CID
	From          string
	To            string
	Value         string
	GasFeeCap     string
	GasPremium    string
	GasLimit      *int64
	SizeBytes     *int64
	Nonce         *uint64
	Method        uint64
	ExitCode      int64
	GasUsed       int64
	ParentBaseFee string
	MinerPenalty  string
	GasRefund     int64
	GasBurned     int64
	ActorTypeID   int
	Params        string
}
