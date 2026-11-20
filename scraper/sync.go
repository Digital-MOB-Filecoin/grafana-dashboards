package scraper

import (
	"context"
	"fmt"
	"log"

	"github.com/Digital-MOB-Filecoin/grafana-dashboards/model"
	"github.com/Digital-MOB-Filecoin/grafana-dashboards/rpc"
	"github.com/Digital-MOB-Filecoin/grafana-dashboards/store"
)

// ChainSyncJob reconciles the DB against the canonical chain head.
// On a reorg it walks backwards, deleting all data for reorged heights
// until the DB and chain agree on a common ancestor.
type ChainSyncJob struct {
	api   *rpc.LotusAPI
	store store.Store
}

func NewChainSyncJob(api *rpc.LotusAPI, st store.Store) *ChainSyncJob {
	return &ChainSyncJob{api: api, store: st}
}

func (j *ChainSyncJob) Run(ctx context.Context) error {
	head, err := j.api.ChainGetHead(ctx)
	if err != nil {
		return fmt.Errorf("ChainGetHead: %w", err)
	}

	height := head.Height
	for {
		dbHeaders, err := j.store.BlockHeaders().GetAtHeight(ctx, height)
		if err != nil {
			return fmt.Errorf("get block headers at height %d: %w", height, err)
		}

		// No data in DB for this height — nothing to reconcile, stop.
		if len(dbHeaders) == 0 {
			break
		}

		if j.match(head.Cids, dbHeaders) {
			break
		}

		log.Printf("sync: reorg detected at height %d, rolling back", height)
		if err := j.rollback(ctx, height); err != nil {
			return fmt.Errorf("rollback height %d: %w", height, err)
		}

		// Move to parent tipset and repeat.
		height--
		head, err = j.api.ChainGetTipSet(ctx, parentKey(head))
		if err != nil {
			return fmt.Errorf("ChainGetTipSet at height %d: %w", height, err)
		}
	}
	return nil
}

// parentKey builds the parent tipset key by collecting deduplicated Parents CIDs
// across all blocks in the tipset. All blocks share the same parent set by protocol,
// but we deduplicate defensively.
func parentKey(ts *model.TipSet) model.TipSetKey {
	seen := make(map[model.CID]struct{})
	var key model.TipSetKey
	for _, block := range ts.Blocks {
		for _, cid := range block.Parents {
			if _, ok := seen[cid]; !ok {
				seen[cid] = struct{}{}
				key = append(key, cid)
			}
		}
	}
	return key
}

// match returns true when every CID in the chain tipset is present in the DB headers.
func (j *ChainSyncJob) match(tipCids model.TipSetKey, dbHeaders []*model.BlockHeader) bool {
	dbSet := make(map[model.CID]struct{}, len(dbHeaders))
	for _, h := range dbHeaders {
		dbSet[h.CID] = struct{}{}
	}
	for _, cid := range tipCids {
		if _, ok := dbSet[cid]; !ok {
			return false
		}
	}
	return true
}

// rollback deletes all data for a given height across all affected tables.
// decoded_params is cleaned up by the FK cascade on derived_gas_outputs.
func (j *ChainSyncJob) rollback(ctx context.Context, height int64) error {
	tables := []struct {
		name string
		del  func(context.Context, int64) error
	}{
		{"block_headers", j.store.BlockHeaders().DeleteAtHeight},
		{"chain_economics", j.store.ChainEconomics().DeleteAtHeight},
		{"derived_gas_outputs", j.store.DerivedGasOutputs().DeleteAtHeight},
		{"miner_sector_events", j.store.MinerSectorEvents().DeleteAtHeight},
	}
	for _, t := range tables {
		if err := t.del(ctx, height); err != nil {
			return fmt.Errorf("delete %s at %d: %w", t.name, height, err)
		}
	}
	return nil
}
