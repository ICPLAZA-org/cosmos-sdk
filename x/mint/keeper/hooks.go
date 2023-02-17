package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/mint/types"
)

// Implements MintHooks interface
var _ types.MintHooks = Keeper{}

// BeforeNextAnnualProvisions - call hook if NextAnnualProvisions
func (k Keeper) BeforeNextAnnualProvisions(blockHeight int64, blocksPerYear uint64, totalSupply sdk.Int, customProvision bool) sdk.Dec {
	rc := sdk.NewDec(-1)

	if k.hooks != nil {
		rc = k.hooks.BeforeNextAnnualProvisions(blockHeight, blocksPerYear, totalSupply, customProvision)
	}

	return rc
}
