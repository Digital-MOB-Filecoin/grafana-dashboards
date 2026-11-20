package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Digital-MOB-Filecoin/grafana-dashboards/model"
	"github.com/Digital-MOB-Filecoin/grafana-dashboards/store"
)

// Store is the Postgres implementation of store.Store.
type Store struct {
	pool *pgxpool.Pool
}

func New(ctx context.Context, dsn string) (*Store, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("open postgres pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	return &Store{pool: pool}, nil
}

func (s *Store) LatestBlockHeight(ctx context.Context) (int64, bool, error) {
	var height *int64
	if err := s.pool.QueryRow(ctx, `SELECT MAX(height) FROM block_headers`).Scan(&height); err != nil {
		return 0, false, err
	}
	if height == nil {
		return 0, false, nil
	}
	return *height, true, nil
}

func (s *Store) Close() error {
	s.pool.Close()
	return nil
}

func (s *Store) BlockHeaders() store.HeightRepository[model.BlockHeader] {
	return &blockHeaderRepo{pool: s.pool}
}

func (s *Store) ChainEconomics() store.HeightRepository[model.ChainEconomics] {
	return &chainEconomicsRepo{pool: s.pool}
}

func (s *Store) DerivedGasOutputs() store.HeightRepository[model.DerivedGasOutput] {
	return &derivedGasOutputRepo{pool: s.pool}
}

func (s *Store) MinerSectorEvents() store.HeightRepository[model.MinerSectorEvent] {
	return &minerSectorEventRepo{pool: s.pool}
}

func (s *Store) DecodedParams() store.Repository[model.DecodedParams] {
	return &decodedParamsRepo{pool: s.pool}
}

func (s *Store) ActorTypes() store.Repository[model.ActorType] {
	return &actorTypeRepo{pool: s.pool}
}

func (s *Store) ActorCodes() store.Repository[model.ActorCode] {
	return &actorCodeRepo{pool: s.pool}
}

func (s *Store) ActorMethods() store.Repository[model.ActorMethod] {
	return &actorMethodRepo{pool: s.pool}
}
