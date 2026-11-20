package store

import (
	"context"

	"github.com/Digital-MOB-Filecoin/grafana-dashboards/model"
)

// Repository is the base interface for inserting and removing entities by height.
type Repository[T any] interface {
	Insert(ctx context.Context, entity *T) error
}

// HeightRepository extends Repository with height-scoped operations used by
// the sync job to roll back reorged data.
type HeightRepository[T any] interface {
	Repository[T]
	GetAtHeight(ctx context.Context, height int64) ([]*T, error)
	DeleteAtHeight(ctx context.Context, height int64) error
}

// Store is the top-level database abstraction.
// Each method returns the repository for a specific table.
// Add new tables here as the schema grows.
type Store interface {
	BlockHeaders() HeightRepository[model.BlockHeader]
	ChainEconomics() HeightRepository[model.ChainEconomics]
	DerivedGasOutputs() HeightRepository[model.DerivedGasOutput]
	MinerSectorEvents() HeightRepository[model.MinerSectorEvent]

	// decoded_params cascades deletes via FK on derived_gas_outputs, so only Insert is needed.
	DecodedParams() Repository[model.DecodedParams]

	// Actor tables are wiped and repopulated on network version changes.
	ActorTypes() Repository[model.ActorType]
	ActorCodes() Repository[model.ActorCode]
	ActorMethods() Repository[model.ActorMethod]

	// LatestBlockHeight returns the highest height stored in block_headers.
	// found is false when the table is empty.
	LatestBlockHeight(ctx context.Context) (height int64, found bool, err error)

	Close() error
}
