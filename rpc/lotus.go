package rpc

import (
	"context"

	"github.com/Digital-MOB-Filecoin/grafana-dashboards/model"
)

// LotusAPI provides typed access to the Lotus JSON-RPC API.
// Add new methods here as new endpoints are needed.
type LotusAPI struct {
	client Client
}

func NewLotusAPI(client Client) *LotusAPI {
	return &LotusAPI{client: client}
}

// ChainGetHead returns the current heaviest tipset.
func (l *LotusAPI) ChainGetHead(ctx context.Context) (*model.TipSet, error) {
	var ts model.TipSet
	if err := l.client.Call(ctx, "ChainGetHead", nil, &ts); err != nil {
		return nil, err
	}
	return &ts, nil
}

// ChainGetTipSetByHeight returns the tipset at the given height, anchored to the
// heaviest chain when key is nil.
func (l *LotusAPI) ChainGetTipSetByHeight(ctx context.Context, height int64, key model.TipSetKey) (*model.TipSet, error) {
	var ts model.TipSet
	if err := l.client.Call(ctx, "ChainGetTipSetByHeight", []any{height, key}, &ts); err != nil {
		return nil, err
	}
	return &ts, nil
}

// ChainGetTipSet returns the tipset identified by the given key.
func (l *LotusAPI) ChainGetTipSet(ctx context.Context, key model.TipSetKey) (*model.TipSet, error) {
	var ts model.TipSet
	if err := l.client.Call(ctx, "ChainGetTipSet", []any{key}, &ts); err != nil {
		return nil, err
	}
	return &ts, nil
}

// StateVMCirculatingSupplyInternal returns the circulating supply breakdown at a tipset.
func (l *LotusAPI) StateVMCirculatingSupplyInternal(ctx context.Context, key model.TipSetKey) (*model.CirculatingSupply, error) {
	var cs model.CirculatingSupply
	if err := l.client.Call(ctx, "StateVMCirculatingSupplyInternal", []any{key}, &cs); err != nil {
		return nil, err
	}
	return &cs, nil
}

// ChainGetParentMessages returns the messages included in the parent of the given block CID.
func (l *LotusAPI) ChainGetParentMessages(ctx context.Context, blockCID model.CID) ([]*model.Message, error) {
	var msgs []*model.Message
	if err := l.client.Call(ctx, "ChainGetParentMessages", []any{blockCID}, &msgs); err != nil {
		return nil, err
	}
	return msgs, nil
}

// ChainGetParentReceipts returns the receipts for the parent messages of the given block CID.
func (l *LotusAPI) ChainGetParentReceipts(ctx context.Context, blockCID model.CID) ([]*model.MessageReceipt, error) {
	var receipts []*model.MessageReceipt
	if err := l.client.Call(ctx, "ChainGetParentReceipts", []any{blockCID}, &receipts); err != nil {
		return nil, err
	}
	return receipts, nil
}

// StateDecodeParams decodes actor method params returned from the chain.
func (l *LotusAPI) StateDecodeParams(ctx context.Context, to string, method uint64, params []byte, key model.TipSetKey) (any, error) {
	var result any
	if err := l.client.Call(ctx, "StateDecodeParams", []any{to, method, params, key}, &result); err != nil {
		return nil, err
	}
	return result, nil
}
