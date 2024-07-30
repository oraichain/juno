package clock

import (
	"encoding/base64"
	"fmt"
	"log"
	"time"

	"github.com/CosmosContracts/juno/v18/x/clock/keeper"
	"github.com/CosmosContracts/juno/v18/x/clock/types"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// EndBlocker executes on contracts at the end of the block.
func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)

	message := []byte(fmt.Sprintf(types.EndBlockSudoMessage, base64.StdEncoding.EncodeToString(ctx.HeaderHash())))

	p := k.GetParams(ctx)

	errorExecs := []string{}

	for _, addr := range p.ContractAddresses {
		contract, err := sdk.AccAddressFromBech32(addr)
		if err != nil {
			errorExecs = append(errorExecs, addr)
			continue
		}

		childCtx := ctx.WithGasMeter(sdk.NewGasMeter(p.ContractGasLimit))
		_, err = k.GetContractKeeper().Sudo(childCtx, contract, message)
		if err != nil {
			errorExecs = append(errorExecs, addr)
			continue
		}
	}

	if len(errorExecs) > 0 {
		log.Printf("[x/clock] Execute Errors: %v", errorExecs)
	}
}
