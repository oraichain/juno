package types

import (
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

// this file contains code related to custom governance proposals
const (
	ProposalTypeUpdateParams = "GlobalFeeUpdateParams"
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
