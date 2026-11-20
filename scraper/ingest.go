package scraper

import (
	"context"
	"fmt"
	"log"

	"github.com/Digital-MOB-Filecoin/grafana-dashboards/model"
	"github.com/Digital-MOB-Filecoin/grafana-dashboards/rpc"
	"github.com/Digital-MOB-Filecoin/grafana-dashboards/store"
)

// IngestJob fetches and stores chain data by walking backwards from the chain head.
type IngestJob struct {
	api   *rpc.LotusAPI
	store store.Store
}

func NewIngestJob(api *rpc.LotusAPI, st store.Store) *IngestJob {
	return &IngestJob{api: api, store: st}
}

// RunFrom walks backwards from start, ingesting each tipset, until it reaches
// endHeight (inclusive). The caller is responsible for validating start.
func (j *IngestJob) RunFrom(ctx context.Context, start *model.TipSet, endHeight int64) error {
	ts := start

	for ts.Height >= endHeight {
		if err := j.runTipSet(ctx, ts); err != nil {
			log.Printf("ingest: error at height %d: %v", ts.Height, err)
		}

		if ts.Height == 0 {
			break
		}

		var err error
		ts, err = j.api.ChainGetTipSet(ctx, parentKey(ts))
		if err != nil {
			return fmt.Errorf("ChainGetTipSet at height %d: %w", ts.Height, err)
		}
	}
	return nil
}

// runTipSet ingests all data for one tipset, skipping it if already stored.
func (j *IngestJob) runTipSet(ctx context.Context, ts *model.TipSet) error {
	existing, err := j.store.BlockHeaders().GetAtHeight(ctx, ts.Height)
	if err != nil {
		return fmt.Errorf("check existing headers at %d: %w", ts.Height, err)
	}
	if len(existing) > 0 {
		return nil
	}

	log.Printf("ingest: processing height %d (%d blocks)", ts.Height, len(ts.Blocks))

	if err := j.ingestBlockHeaders(ctx, ts); err != nil {
		return err
	}
	if err := j.ingestChainEconomics(ctx, ts); err != nil {
		return err
	}
	if err := j.ingestMessages(ctx, ts); err != nil {
		return err
	}

	return nil
}

// ingestBlockHeaders stores one row per block in the tipset.
// ts.Cids[i] is the CID of ts.Blocks[i].
func (j *IngestJob) ingestBlockHeaders(ctx context.Context, ts *model.TipSet) error {
	for i, block := range ts.Blocks {
		if i >= len(ts.Cids) {
			break
		}
		_ = block
		_ = i
		// TODO: parse block.Miner address → miner ID integer, then:
		// bh := &model.BlockHeader{
		//     Height:          ts.Height,
		//     CID:             ts.Cids[i],
		//     MinerID:         parsedMinerID,
		//     ParentWeight:    block.ParentWeight,
		//     ParentBaseFee:   block.ParentBaseFee,
		//     ParentStateRoot: block.ParentStateRoot,
		//     Timestamp:       int64(block.Timestamp),
		//     ForkSignaling:   int64(block.ForkSignaling),
		// }
		// if block.ElectionProof != nil { bh.WinCount = &block.ElectionProof.WinCount }
		// if err := j.store.BlockHeaders().Insert(ctx, bh); err != nil { return err }
	}
	return nil
}

// ingestChainEconomics fetches circulating supply and power actor state for the tipset.
func (j *IngestJob) ingestChainEconomics(ctx context.Context, ts *model.TipSet) error {
	cs, err := j.api.StateVMCirculatingSupplyInternal(ctx, ts.Cids)
	if err != nil {
		return fmt.Errorf("StateVMCirculatingSupplyInternal: %w", err)
	}
	_ = cs
	// TODO: fetch power actor state, then:
	// ce := &model.ChainEconomics{
	//     Height:              ts.Height,
	//     StateRoot:           ts.Blocks[0].ParentStateRoot,
	//     CirculatingFIL:      cs.FilCirculating,
	//     ...
	// }
	// return j.store.ChainEconomics().Insert(ctx, ce)
	return nil
}

// ingestMessages fetches parent messages + receipts for each block and stores
// derived_gas_outputs rows.
func (j *IngestJob) ingestMessages(ctx context.Context, ts *model.TipSet) error {
	for _, blockCID := range ts.Cids {
		msgs, err := j.api.ChainGetParentMessages(ctx, blockCID)
		if err != nil {
			return fmt.Errorf("ChainGetParentMessages for %s: %w", blockCID, err)
		}
		receipts, err := j.api.ChainGetParentReceipts(ctx, blockCID)
		if err != nil {
			return fmt.Errorf("ChainGetParentReceipts for %s: %w", blockCID, err)
		}

		for i, msg := range msgs {
			if i >= len(receipts) {
				continue
			}
			_ = msg
			_ = receipts[i]
			// TODO: compute gas fields and store:
			// dgo := &model.DerivedGasOutput{ ... }
			// if err := j.store.DerivedGasOutputs().Insert(ctx, dgo); err != nil { return err }
		}
	}
	return nil
}
