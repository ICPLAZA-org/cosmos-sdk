package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
)

// BurnTokens handles burning the collected fees
func (k Keeper) BurnTokens(ctx sdk.Context) {
	// burnt
	burntFeeCollector := k.authKeeper.GetModuleAccount(ctx, k.burntFeeCollectorName)
	burntFeesCollectedInt := k.bankKeeper.GetAllBalances(ctx, burntFeeCollector.GetAddress())

	if burntFeesCollectedInt.IsZero() {
		return
	}

	// transfer collected fees to the gov module account
	if err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, k.burntFeeCollectorName, types.ModuleName, burntFeesCollectedInt); err != nil {
		panic(err)
	}

	if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, burntFeesCollectedInt); err != nil {
		panic(err)
	}
}
