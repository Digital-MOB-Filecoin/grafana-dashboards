package model

import "encoding/json"

// DecodedParams is the DB model for the decoded_params table.
// Currently only populated for ProveCommitAggregate messages.
type DecodedParams struct {
	Height       *int64
	IndexInBlock *int
	Params       json.RawMessage
}
