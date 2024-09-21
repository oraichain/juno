package keeper

import (
	"context"
	"fmt"

	"github.com/CosmosContracts/juno/v15/x/globalfee/types"
)

var _ types.MsgServer = msgServer{}

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the x/mint MsgServer interface.
func NewMsgServerImpl(k Keeper) types.MsgServer {
	return &msgServer{
		Keeper: k,
	}
}

func (ms msgServer) UpdateParams(goCtx context.Context, req *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {

	return nil, fmt.Errorf("UpdateParams for globalfee is not currently supported!. Please use it via gov proposal")
	// ctx := sdk.UnwrapSDKContext(goCtx)
	// if err := ms.SetParams(ctx, req.Params); err != nil {
	// 	return nil, err
	// }

	// return &types.MsgUpdateParamsResponse{}, nil
}
