package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// combine multiple mint hooks, all hook functions are run in array sequence
type MultiMintHooks []MintHooks

func NewMultiMintHooks(hooks ...MintHooks) MultiMintHooks {
	return hooks
}

func (h MultiMintHooks) BeforeNextAnnualProvisions(blockHeight int64, blocksPerYear uint64, totalSupply sdk.Int, customProvision bool) (rc sdk.Dec) {
	rc = sdk.NewDec(-1)

	for i := range h {
		rc = h[i].BeforeNextAnnualProvisions(blockHeight, blocksPerYear, totalSupply, customProvision)
	}
	return rc
}
