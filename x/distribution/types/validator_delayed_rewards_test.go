package types_test

import (
	"testing"
	"time"
	"fmt"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

func TestValidateGenesis(t *testing.T) {
	vdr := types.NewValidatorDelayedReward(int64(1), time.Now(), 
		// sdk.NewDecCoins(),
		sdk.NewDecCoinsFromCoins(sdk.NewCoin("pptoken", sdk.NewInt(10))),
		sdk.NewDecCoinsFromCoins(sdk.NewCoin("mytoken", sdk.NewInt(10))),
		)

	fmt.Printf("%s\n", vdr.String())
	require.Nil(t, nil)
}
