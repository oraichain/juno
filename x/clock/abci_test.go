package clock_test

import (
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/CosmosContracts/juno/v18/x/clock/types"
)

func TestBeginBlock(t *testing.T) {
	hash := []byte{1, 2, 3}
	msg := []byte(fmt.Sprintf(types.EndBlockSudoMessage, base64.StdEncoding.EncodeToString(hash)))
	println(string(msg))
}
