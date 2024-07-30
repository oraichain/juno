package types

import (
	fmt "fmt"
	"strings"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

// this file contains code related to custom governance proposals
const (
	EndBlockSudoMessage      = `{"clock_end_block":{"hash":"%s"}}`
	ProposalTypeUpdateParams = "UpdateParams"
)

func (p *UpdateParamsProposal) GetTitle() string { return p.Title }

func (p *UpdateParamsProposal) GetDescription() string { return p.Description }

func (p *UpdateParamsProposal) ProposalRoute() string { return RouterKey }

func (p *UpdateParamsProposal) ProposalType() string {
	return ProposalTypeUpdateParams
}

func (p *UpdateParamsProposal) ValidateBasic() error {
	err := govtypes.ValidateAbstract(p)
	if err != nil {
		return err
	}
	return nil
}

func (p UpdateParamsProposal) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Unhalt Bridge Proposal:
  Title:              %s
  Description:        %s
  Contract Address:   %s
  Contract Gas Limit: %d
`, p.Title, p.Description, p.Params.GetContractAddresses(), p.Params.GetContractGasLimit()))
	return b.String()
}
