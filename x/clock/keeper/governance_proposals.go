package keeper

import (
	errorsmod "cosmossdk.io/errors"
	"github.com/CosmosContracts/juno/v18/x/clock/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

func RegisterProposalTypes() {
	// use of prefix stripping to prevent a typo between the proposal we check
	// and the one we register, any issues with the registration string will prevent
	// the proposal from working. We must check for double registration so that cli commands
	// submitting these proposals work.
	// For some reason the cli code is run during app.go startup, but of course app.go is not
	// run during operation of one off tx commands, so we need to run this 'twice'
	if !govtypes.IsValidProposalType(types.ProposalTypeUpdateParams) {
		govtypes.RegisterProposalType(types.ProposalTypeUpdateParams)
	}
}

func NewClockProposalHandler(k Keeper) govtypes.Handler {
	return func(ctx sdk.Context, content govtypes.Content) error {
		switch c := content.(type) {
		case *types.UpdateParamsProposal:
			return k.HandleUpdateParamsProposal(ctx, c)
		default:
			return errorsmod.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized Clock proposal content type: %T", c)
		}
	}
}

func (k Keeper) HandleUpdateParamsProposal(ctx sdk.Context, p *types.UpdateParamsProposal) error {
	if err := k.SetParams(ctx, p.Params); err != nil {
		return err
	}
	return nil
}
