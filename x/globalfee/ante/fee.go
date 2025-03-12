package ante

import (
	"errors"

	tmstrings "github.com/cometbft/cometbft/libs/strings"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"

	"cosmossdk.io/math"
	globalfeekeeper "github.com/CosmosContracts/juno/v18/x/globalfee/keeper"
)

// FeeWithBypassDecorator checks if the transaction's fee is at least as large
// as the local validator's minimum gasFee (defined in validator config) and global fee, and the fee denom should be in the global fees' denoms.
//
// If fee is too low, decorator returns error and tx is rejected from mempool.
// Note this only applies when ctx.CheckTx = true. If fee is high enough or not
// CheckTx, then call next AnteHandler.
//
// CONTRACT: Tx must implement FeeTx to use FeeDecorator
// If the tx msg type is one of the bypass msg types, the tx is valid even if the min fee is lower than normally required.
// If the bypass tx still carries fees, the fee denom should be the same as global fee required.

var _ sdk.AnteDecorator = FeeDecorator{}

type FeeDecorator struct {
	BypassMinFeeMsgTypes            []string
	GlobalFeeKeeper                 globalfeekeeper.Keeper
	StakingKeeper                   stakingkeeper.Keeper
	MaxTotalBypassMinFeeMsgGasUsage uint64
}

func NewFeeDecorator(bypassMsgTypes []string, gfk globalfeekeeper.Keeper, sk stakingkeeper.Keeper, maxTotalBypassMinFeeMsgGasUsage uint64) FeeDecorator {
	return FeeDecorator{
		BypassMinFeeMsgTypes:            bypassMsgTypes,
		GlobalFeeKeeper:                 gfk,
		StakingKeeper:                   sk,
		MaxTotalBypassMinFeeMsgGasUsage: maxTotalBypassMinFeeMsgGasUsage,
	}
}

// AnteHandle implements the AnteDecorator interface
func (mfd FeeDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return ctx, errorsmod.Wrap(sdkerrors.ErrTxDecode, "Tx must implement the sdk.FeeTx interface")
	}

	// Only check for minimum fees and global fee if the execution mode is CheckTx
	if !ctx.IsCheckTx() || simulate {
		return next(ctx, tx, simulate)
	}

	// if msg contains only bypass msgs then we just bypass it
	if mfd.ContainsOnlyBypassMinFeeMsgs(feeTx.GetMsgs()) {
		return next(ctx, tx, simulate)
	}

	// Get global gas prices
	requiredGlobalGasPrices, err := mfd.GetGlobalGasPrices(ctx)
	if err != nil {
		return ctx, err
	}

	// Get local minimum-gas-prices
	localMinGasPrices := ctx.MinGasPrices()

	// CombinedGasPrices should never be empty since
	// global fee is set to its default value, i.e. 0uatom, if empty
	combinedMinGasPrices := CombinedGasPrices(requiredGlobalGasPrices, localMinGasPrices)
	if len(combinedMinGasPrices) == 0 {
		return ctx, errorsmod.Wrapf(sdkerrors.ErrNotFound, "required fees are not setup.")
	}

	return next(ctx.WithMinGasPrices(combinedMinGasPrices), tx, simulate)
}

// GetGlobalGasPrices returns the global min gas prices
// (might also return 0denom if globalMinGasPrice is 0)
// sorted in ascending order.
// Note that ParamStoreKeyMinGasPrices type requires coins sorted.
func (mfd FeeDecorator) GetGlobalGasPrices(ctx sdk.Context) (sdk.DecCoins, error) {
	var (
		globalMinGasPrices sdk.DecCoins
		err                error
	)

	globalMinGasPrices = mfd.GlobalFeeKeeper.GetParams(ctx).MinimumGasPrices

	// global fee is empty set, set global fee to 0uatom
	if len(globalMinGasPrices) == 0 {
		globalMinGasPrices, err = mfd.DefaultZeroGlobalGasPrices(ctx)
		if err != nil {
			return sdk.DecCoins{}, err
		}
	}

	return globalMinGasPrices, nil
}

func (mfd FeeDecorator) DefaultZeroGlobalGasPrices(ctx sdk.Context) ([]sdk.DecCoin, error) {
	bondDenom, err := mfd.getBondDenom(ctx)
	if err != nil {
		return nil, err
	}
	if bondDenom == "" {
		return nil, errors.New("empty staking bond denomination")
	}

	return []sdk.DecCoin{sdk.NewDecCoinFromDec(bondDenom, math.LegacyNewDec(0))}, nil
}

func (mfd FeeDecorator) getBondDenom(ctx sdk.Context) (string, error) {
	return mfd.StakingKeeper.BondDenom(ctx)
}

// ContainsOnlyBypassMinFeeMsgs returns true if all the given msgs type are listed
// in the BypassMinFeeMsgTypes of the FeeDecorator.
func (mfd FeeDecorator) ContainsOnlyBypassMinFeeMsgs(msgs []sdk.Msg) bool {
	for _, msg := range msgs {
		if tmstrings.StringInSlice(sdk.MsgTypeURL(msg), mfd.BypassMinFeeMsgTypes) {
			continue
		}
		return false
	}

	return true
}
