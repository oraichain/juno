package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmosContracts/juno/v18/x/globalfee/types"
)

var _ types.QueryServer = Querier{}

// Querier defines a wrapper around the x/FeeShare keeper providing gRPC method
// handlers.
type Querier struct {
	Keeper
}

func NewQuerier(k Keeper) Querier {
	return Querier{Keeper: k}
}

// Params returns the fees module params
func (q Querier) MinimumGasPrices(
	c context.Context,
	_ *types.QueryMinimumGasPricesRequest,
) (*types.QueryMinimumGasPricesResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	params := q.GetParams(ctx)
	return &types.QueryMinimumGasPricesResponse{MinimumGasPrices: params.MinimumGasPrices}, nil
}
