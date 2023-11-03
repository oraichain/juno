package cli

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/CosmosContracts/juno/v18/x/clock/keeper"
	"github.com/CosmosContracts/juno/v18/x/clock/types"
)

// GetTxCmd bundles all the subcmds together so they appear under `clock tx`
func GetTxCmd(storeKey string) *cobra.Command {
	// needed for governance proposal txs in cli case
	// internal check prevents double registration in node case
	keeper.RegisterProposalTypes()

	// nolint: exhaustruct
	clockTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Clock subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	clockTxCmd.AddCommand([]*cobra.Command{
		CmdAddContractProposal(),
	}...)

	return clockTxCmd
}

// CmdAddContractProposal enables users to create a proposal to add new contract to clock module
func CmdAddContractProposal() *cobra.Command {
	// nolint: exhaustruct
	cmd := &cobra.Command{
		Use:   "add-contract [contract_addresses] [contract_gas_limit] [title] [initial-deposit] [description]",
		Short: "Creates a governance proposal to add contract addresses to clock module",
		Args:  cobra.ExactArgs(8),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientTxContext(cmd)
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

			contractAddress := args[0]
			contractGasLimit, err := strconv.ParseUint(string(args[1]), 10, 64)
			if err != nil {
				return fmt.Errorf("ContractGasLimit should be an unsigned integer")
			}

			proposal := &types.UpdateParamsProposal{
				Params: types.Params{
					ContractAddresses: []string{contractAddress},
					ContractGasLimit:  contractGasLimit, // 1 billion
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
