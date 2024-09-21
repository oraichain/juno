package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/CosmosContracts/juno/v15/x/globalfee/keeper"
	"github.com/CosmosContracts/juno/v15/x/globalfee/types"
)

// GetTxCmd bundles all the subcmds together so they appear under `clock tx`
func GetTxCmd() *cobra.Command {
	// needed for governance proposal txs in cli case
	// internal check prevents double registration in node case
	keeper.RegisterProposalTypes()

	// nolint: exhaustruct
	clockTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Globalfee subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	clockTxCmd.AddCommand([]*cobra.Command{
		CmdAddMinimumGlobalFee(),
	}...)

	return clockTxCmd
}

// CmdAddMinimumGlobalFee enables users to create a proposal to add new contract to clock module
func CmdAddMinimumGlobalFee() *cobra.Command {
	// nolint: exhaustruct
	cmd := &cobra.Command{
		Use:   "add-globalfee [denom] [amount] [title] [initial-deposit] [description]",
		Short: "Creates a governance proposal to add or update global fees",
		Args:  cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			minimumGasPrices, err := queryClient.MinimumGasPrices(cmd.Context(), &types.QueryMinimumGasPricesRequest{})
			if err != nil {
				return err
			}

			cosmosAddr := cliCtx.GetFromAddress()

			initialDeposit, err := sdk.ParseCoinsNormalized(args[3])
			if err != nil {
				return sdkerrors.Wrap(err, "bad initial deposit amount")
			}

			if len(initialDeposit) != 1 {
				return fmt.Errorf("unexpected coin amounts, expecting just 1 coin amount for initialDeposit")
			}

			denom := args[0]
			amount := args[1]
			isAddNew := true
			minGasPrice := sdk.NewDecCoinFromDec(denom, sdk.MustNewDecFromStr(amount))

			for i, gasPrice := range minimumGasPrices.MinimumGasPrices {
				if denom == gasPrice.Denom {
					minimumGasPrices.MinimumGasPrices[i] = minGasPrice
					isAddNew = false
					break
				}
			}
			if isAddNew {
				minimumGasPrices.MinimumGasPrices = minimumGasPrices.MinimumGasPrices.Add(minGasPrice)
			}
			proposal := &types.UpdateParamsProposal{
				Params: types.Params{
					MinimumGasPrices: minimumGasPrices.MinimumGasPrices,
				},
				Title:       args[2],
				Description: args[4],
			}
			proposalAny, err := codectypes.NewAnyWithValue(proposal)
			if err != nil {
				return sdkerrors.Wrap(err, "invalid metadata or proposal details!")
			}

			// Make the message
			msg := govtypes.MsgSubmitProposal{
				Proposer:       cosmosAddr.String(),
				InitialDeposit: initialDeposit,
				Content:        proposalAny,
			}
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			// Send it
			return tx.GenerateOrBroadcastTxCLI(cliCtx, cmd.Flags(), &msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
