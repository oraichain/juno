package cli

import (
	"fmt"
	"strconv"

	errorsmod "cosmossdk.io/errors"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

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
		CmdRemoveContractProposal(),
	}...)

	return clockTxCmd
}

// CmdAddContractProposal enables users to create a proposal to add new contract to clock module
func CmdAddContractProposal() *cobra.Command {
	// nolint: exhaustruct
	cmd := &cobra.Command{
		Use:   "add-contract [contract-address] [contract-gas-limit] [title] [initial-deposit] [description]",
		Short: "Creates a governance proposal to add contract address to clock module",
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
			clockContracts, err := queryClient.ClockContracts(cmd.Context(), &types.QueryClockContracts{})
			if err != nil {
				return err
			}

			cosmosAddr := cliCtx.GetFromAddress()

			initialDeposit, err := sdk.ParseCoinsNormalized(args[3])
			if err != nil {
				return errorsmod.Wrap(err, "bad initial deposit amount")
			}

			if len(initialDeposit) != 1 {
				return fmt.Errorf("unexpected coin amounts, expecting just 1 coin amount for initialDeposit")
			}

			contractAddress := args[0]
			// Valid address check
			if _, err := sdk.AccAddressFromBech32(contractAddress); err != nil {
				return errorsmod.Wrapf(
					sdkerrors.ErrInvalidAddress,
					"invalid contract address: %s", err.Error(),
				)
			}

			for _, addr := range clockContracts.ContractAddresses {
				if contractAddress == addr {
					return errorsmod.Wrapf(
						sdkerrors.ErrInvalidAddress,
						"duplicate contract address: %s", contractAddress,
					)
				}
			}
			newContractAddresses := append(clockContracts.ContractAddresses, contractAddress)
			contractGasLimit, err := strconv.ParseUint(string(args[1]), 10, 64)
			if err != nil {
				return fmt.Errorf("ContractGasLimit should be an unsigned integer")
			}

			proposal := &types.UpdateParamsProposal{
				Params: types.Params{
					ContractAddresses: newContractAddresses,
					ContractGasLimit:  contractGasLimit,
				},
				Title:       args[2],
				Description: args[4],
			}
			if err := proposal.ValidateBasic(); err != nil {
				return err
			}
			proposalAny, err := codectypes.NewAnyWithValue(proposal)
			if err != nil {
				return errorsmod.Wrap(err, "invalid metadata or proposal details!")
			}

			// Make the message
			msg := govtypes.MsgSubmitProposal{
				Proposer:       cosmosAddr.String(),
				InitialDeposit: initialDeposit,
				Messages:       []*codectypes.Any{proposalAny},
			}

			// Send it
			return tx.GenerateOrBroadcastTxCLI(cliCtx, cmd.Flags(), &msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdRemoveContractProposal enables users to create a proposal to remove existed contract from clock module
func CmdRemoveContractProposal() *cobra.Command {
	// nolint: exhaustruct
	cmd := &cobra.Command{
		Use:   "remove-contract [contract-address] [title] [initial-deposit] [description]",
		Short: "Creates a governance proposal to remove existed contract address from clock module",
		Args:  cobra.ExactArgs(4),
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
			clockParams, err := queryClient.Params(cmd.Context(), &types.QueryParamsRequest{})
			if err != nil {
				return err
			}

			cosmosAddr := cliCtx.GetFromAddress()
			initialDeposit, err := sdk.ParseCoinsNormalized(args[2])
			if err != nil {
				return errorsmod.Wrap(err, "bad initial deposit amount")
			}

			if len(initialDeposit) != 1 {
				return fmt.Errorf("unexpected coin amounts, expecting just 1 coin amount for initialDeposit")
			}

			contractAddress := args[0]
			newContractAddresses := clockParams.Params.ContractAddresses

			// Valid address check
			if _, err := sdk.AccAddressFromBech32(contractAddress); err != nil {
				return errorsmod.Wrapf(
					sdkerrors.ErrInvalidAddress,
					"invalid contract address: %s", contractAddress,
				)
			}

			index := 0
			count := 0
			for idx, addr := range newContractAddresses {
				if contractAddress == addr {
					index = idx
					count++
				}
			}

			if count > 0 {
				newContractAddresses = append(newContractAddresses[:index], newContractAddresses[:index+1]...)
			} else {
				return errorsmod.Wrapf(
					sdkerrors.ErrInvalidAddress,
					"Cannot remove un-existed contract address: %s", contractAddress,
				)
			}

			proposal := &types.UpdateParamsProposal{
				Params: types.Params{
					ContractAddresses: newContractAddresses,
					ContractGasLimit:  clockParams.Params.ContractGasLimit,
				},
				Title:       args[1],
				Description: args[3],
			}
			if err := proposal.ValidateBasic(); err != nil {
				return err
			}
			proposalAny, err := codectypes.NewAnyWithValue(proposal)
			if err != nil {
				return errorsmod.Wrap(err, "invalid metadata or proposal details!")
			}

			// Make the message
			msg := govtypes.MsgSubmitProposal{
				Proposer:       cosmosAddr.String(),
				InitialDeposit: initialDeposit,
				Messages:       []*codectypes.Any{proposalAny},
			}

			// Send it
			return tx.GenerateOrBroadcastTxCLI(cliCtx, cmd.Flags(), &msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
