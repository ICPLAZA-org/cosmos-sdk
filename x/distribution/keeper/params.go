package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

// GetParams returns the total set of distribution parameters.
func (k Keeper) GetParams(clientCtx sdk.Context) (params types.Params) {
	k.paramSpace.GetParamSet(clientCtx, &params)
	return params
}

// SetParams sets the distribution parameters to the param space.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}

// GetCommunityTax returns the current distribution community tax.
func (k Keeper) GetCommunityTax(ctx sdk.Context) (percent sdk.Dec) {
	k.paramSpace.Get(ctx, types.ParamStoreKeyCommunityTax, &percent)
	return percent
}

// GetBaseProposerReward returns the current distribution base proposer rate.
func (k Keeper) GetBaseProposerReward(ctx sdk.Context) (percent sdk.Dec) {
	k.paramSpace.Get(ctx, types.ParamStoreKeyBaseProposerReward, &percent)
	return percent
}

// GetBonusProposerReward returns the current distribution bonus proposer reward
// rate.
func (k Keeper) GetBonusProposerReward(ctx sdk.Context) (percent sdk.Dec) {
	k.paramSpace.Get(ctx, types.ParamStoreKeyBonusProposerReward, &percent)
	return percent
}

// GetWithdrawAddrEnabled returns the current distribution withdraw address
// enabled parameter.
func (k Keeper) GetWithdrawAddrEnabled(ctx sdk.Context) (enabled bool) {
	k.paramSpace.Get(ctx, types.ParamStoreKeyWithdrawAddrEnabled, &enabled)
	return enabled
}

// GetDelayedRewardProportion returns the current distribution DelayedRewardProportion
func (k Keeper) GetDelayedRewardProportion(ctx sdk.Context) (proportion sdk.Dec) {
	k.paramSpace.Get(ctx, types.ParamStoreKeyDelayedRewardProportion, &proportion)
	return proportion
}

// GetMinDelayedRewardPeriod returns the current distribution MinRelayedRewardPeriod.
func (k Keeper) GetMinDelayedRewardPeriod(ctx sdk.Context) (period int64) {
	k.paramSpace.Get(ctx, types.ParamStoreKeyMinDelayedRewardPeriod, &period)
	return period
}

// GetMaxDelayedRewardInterval returns the current distribution MaxDelayedRewardInterval.
func (k Keeper) GetMaxDelayedRewardInterval(ctx sdk.Context) (interval int64) {
	k.paramSpace.Get(ctx, types.ParamStoreKeyMaxDelayedRewardInterval, &interval)
	return interval
}

// GetDelayedRewardUnit returns the current distribution RelayedRewardUnit.
func (k Keeper) GetDelayedRewardUnit(ctx sdk.Context) (unit time.Duration) {
	k.paramSpace.Get(ctx, types.ParamStoreKeyDelayedRewardUnit, &unit)
	return unit
}
