package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Digital-MOB-Filecoin/grafana-dashboards/model"
)

// ---- block_headers ----

type blockHeaderRepo struct{ pool *pgxpool.Pool }

func (r *blockHeaderRepo) Insert(ctx context.Context, bh *model.BlockHeader) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO block_headers
			(height, cid, miner_id, parent_weight, parent_base_fee, parent_state_root, timestamp, win_count, fork_signaling)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		bh.Height, string(bh.CID), bh.MinerID, bh.ParentWeight, bh.ParentBaseFee,
		string(bh.ParentStateRoot), bh.Timestamp, bh.WinCount, bh.ForkSignaling,
	)
	return err
}

func (r *blockHeaderRepo) GetAtHeight(ctx context.Context, height int64) ([]*model.BlockHeader, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT height, cid, miner_id, parent_weight, parent_base_fee, parent_state_root, timestamp, win_count, fork_signaling
		FROM block_headers WHERE height = $1`, height)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*model.BlockHeader
	for rows.Next() {
		var bh model.BlockHeader
		var cid, stateRoot string
		if err := rows.Scan(&bh.Height, &cid, &bh.MinerID, &bh.ParentWeight, &bh.ParentBaseFee,
			&stateRoot, &bh.Timestamp, &bh.WinCount, &bh.ForkSignaling); err != nil {
			return nil, err
		}
		bh.CID = model.CID(cid)
		bh.ParentStateRoot = model.CID(stateRoot)
		out = append(out, &bh)
	}
	return out, rows.Err()
}

func (r *blockHeaderRepo) DeleteAtHeight(ctx context.Context, height int64) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM block_headers WHERE height = $1`, height)
	return err
}

// ---- chain_economics ----

type chainEconomicsRepo struct{ pool *pgxpool.Pool }

func (r *chainEconomicsRepo) Insert(ctx context.Context, ce *model.ChainEconomics) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO chain_economics (
			height, state_root,
			circulating_fil, vested_fil, mined_fil, burnt_fil, locked_fil, fil_reserve_disbursed,
			total_raw_bytes_power, total_raw_bytes_committed, total_qa_bytes_power, total_qa_bytes_committed,
			total_pledge_collateral, qa_smoothed_position_estimate, qa_smoothed_velocity_estimate,
			miner_count, participating_miner_count
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)`,
		ce.Height, string(ce.StateRoot),
		ce.CirculatingFIL, ce.VestedFIL, ce.MinedFIL, ce.BurntFIL, ce.LockedFIL, ce.FilReserveDisbursed,
		ce.TotalRawBytesPower, ce.TotalRawBytesCommitted, ce.TotalQABytesPower, ce.TotalQABytesCommitted,
		ce.TotalPledgeCollateral, ce.QASmoothedPositionEstimate, ce.QASmoothedVelocityEstimate,
		ce.MinerCount, ce.ParticipatingMinerCount,
	)
	return err
}

func (r *chainEconomicsRepo) GetAtHeight(ctx context.Context, height int64) ([]*model.ChainEconomics, error) {
	rows, err := r.pool.Query(ctx, `SELECT height FROM chain_economics WHERE height = $1`, height)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.ChainEconomics
	for rows.Next() {
		var ce model.ChainEconomics
		if err := rows.Scan(&ce.Height); err != nil {
			return nil, err
		}
		out = append(out, &ce)
	}
	return out, rows.Err()
}

func (r *chainEconomicsRepo) DeleteAtHeight(ctx context.Context, height int64) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM chain_economics WHERE height = $1`, height)
	return err
}

// ---- derived_gas_outputs ----

type derivedGasOutputRepo struct{ pool *pgxpool.Pool }

func (r *derivedGasOutputRepo) Insert(ctx context.Context, dgo *model.DerivedGasOutput) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO derived_gas_outputs (
			height, index_in_block, cid, state_root, "from", "to", value,
			gas_fee_cap, gas_premium, gas_limit, size_bytes, nonce, method,
			exit_code, gas_used, parent_base_fee, miner_penalty,
			gas_refund, gas_burned, actor_type_id, params
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21)`,
		dgo.Height, dgo.IndexInBlock, string(dgo.CID), string(dgo.StateRoot),
		dgo.From, dgo.To, dgo.Value, dgo.GasFeeCap, dgo.GasPremium,
		dgo.GasLimit, dgo.SizeBytes, dgo.Nonce, dgo.Method,
		dgo.ExitCode, dgo.GasUsed, dgo.ParentBaseFee, dgo.MinerPenalty,
		dgo.GasRefund, dgo.GasBurned, dgo.ActorTypeID, dgo.Params,
	)
	return err
}

func (r *derivedGasOutputRepo) GetAtHeight(ctx context.Context, height int64) ([]*model.DerivedGasOutput, error) {
	rows, err := r.pool.Query(ctx, `SELECT height FROM derived_gas_outputs WHERE height = $1`, height)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.DerivedGasOutput
	for rows.Next() {
		var dgo model.DerivedGasOutput
		if err := rows.Scan(&dgo.Height); err != nil {
			return nil, err
		}
		out = append(out, &dgo)
	}
	return out, rows.Err()
}

func (r *derivedGasOutputRepo) DeleteAtHeight(ctx context.Context, height int64) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM derived_gas_outputs WHERE height = $1`, height)
	return err
}

// ---- miner_sector_events ----

type minerSectorEventRepo struct{ pool *pgxpool.Pool }

func (r *minerSectorEventRepo) Insert(ctx context.Context, e *model.MinerSectorEvent) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO miner_sector_events (height, miner_id, sector_id, state_root, event)
		VALUES ($1,$2,$3,$4,$5)`,
		e.Height, e.MinerID, e.SectorID, string(e.StateRoot), string(e.Event),
	)
	return err
}

func (r *minerSectorEventRepo) GetAtHeight(ctx context.Context, height int64) ([]*model.MinerSectorEvent, error) {
	rows, err := r.pool.Query(ctx, `SELECT height FROM miner_sector_events WHERE height = $1`, height)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.MinerSectorEvent
	for rows.Next() {
		var e model.MinerSectorEvent
		if err := rows.Scan(&e.Height); err != nil {
			return nil, err
		}
		out = append(out, &e)
	}
	return out, rows.Err()
}

func (r *minerSectorEventRepo) DeleteAtHeight(ctx context.Context, height int64) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM miner_sector_events WHERE height = $1`, height)
	return err
}

// ---- decoded_params ----

type decodedParamsRepo struct{ pool *pgxpool.Pool }

func (r *decodedParamsRepo) Insert(ctx context.Context, dp *model.DecodedParams) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO decoded_params (height, index_in_block, params)
		VALUES ($1,$2,$3)`,
		dp.Height, dp.IndexInBlock, dp.Params,
	)
	return err
}

// ---- actor_types ----

type actorTypeRepo struct{ pool *pgxpool.Pool }

func (r *actorTypeRepo) Insert(ctx context.Context, at *model.ActorType) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO actor_types (actor_type_id, actor_id, actor_code_cid)
		VALUES ($1,$2,$3)`,
		at.ActorTypeID, at.ActorID, string(at.ActorCodeCID),
	)
	return err
}

// ---- actor_codes ----

type actorCodeRepo struct{ pool *pgxpool.Pool }

func (r *actorCodeRepo) Insert(ctx context.Context, ac *model.ActorCode) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO actor_codes (actor_type_id, actor_code_cid, network_version)
		VALUES ($1,$2,$3)`,
		ac.ActorTypeID, string(ac.ActorCodeCID), ac.NetworkVersion,
	)
	return err
}

// ---- actor_methods ----

type actorMethodRepo struct{ pool *pgxpool.Pool }

func (r *actorMethodRepo) Insert(ctx context.Context, am *model.ActorMethod) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO actor_methods (actor_type_id, method_name, method)
		VALUES ($1,$2,$3)`,
		am.ActorTypeID, am.MethodName, am.Method,
	)
	return err
}
