package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

// biz
func (k Keeper) NewAllocateTokensToValidator(ctx sdk.Context, val stakingtypes.ValidatorI, tokens sdk.DecCoins, validatorVariance sdk.Dec) {
	delayedReward := tokens.MulDec(k.GetDelayedRewardProportion(ctx))
	reward := tokens.Sub(delayedReward)

	if !reward.IsZero() {
		k.AllocateTokensToValidator(ctx, val, reward)		
	}

	if !delayedReward.IsZero() {
		k.equeuTokensToValidatorQueue(ctx, val, delayedReward, validatorVariance)
	}
}

func (k Keeper) equeuTokensToValidatorQueue(ctx sdk.Context, val stakingtypes.ValidatorI, tokens sdk.DecCoins, validatorVariance sdk.Dec) {
	startTime := k.GetDelayedRewardTime(ctx, ctx.BlockHeader().Time)

	delayedReward, ok := k.GetValidatorDelayedReward(ctx, val.GetOperator(), startTime)
	if ok {
		delayedReward.Reward = delayedReward.Reward.Add(tokens...)
	} else {
		period := k.calcValidatorDelayedRewardPeriod(ctx, val, validatorVariance)
		delayedReward = types.NewValidatorDelayedReward(period, startTime, sdk.NewDecCoins(), tokens)
	}

	k.SetValidatorDelayedReward(ctx, val.GetOperator(), startTime, delayedReward)
}

func (k Keeper) calcValidatorDelayedRewardPeriod(ctx sdk.Context, val stakingtypes.ValidatorI, validatorVariance sdk.Dec) int64 {
	factor := sdk.OneDec()

	// |((t/g)/(tt/tg)) - 1| * interval + min_start_time
	if val.GetDelegators().IsPositive() {
		validatorTG := val.GetTokens().ToDec().Quo(val.GetDelegators().ToDec())
		factor = validatorTG.Mul(validatorVariance).Sub(sdk.OneDec()).Abs()
	}
	interval := sdk.NewDec(k.GetMaxDelayedRewardInterval(ctx)).MulTruncate(factor).TruncateInt64()

	return (k.GetMinDelayedRewardPeriod(ctx) + interval)
}


func (k Keeper) WithdrawValidatorDelayedRewardsOf(ctx sdk.Context, val sdk.ValAddress) {
	lastDoneTime := k.GetValidatorDelayedRewardInfo(ctx, val)
	blockTime := k.GetDelayedRewardTime(ctx, ctx.BlockHeader().Time)
	
	if (AfterOfDelayedRewardTime(blockTime, lastDoneTime.LastTime)){
		k.withdrawValidatorDelayedRewardsOf(ctx, val, blockTime)

		lastDoneTime.LastTime = blockTime
		k.SetValidatorDelayedRewardInfo(ctx, val, lastDoneTime)
	}
}

// both startTime and currTime must be DelayedRewardTime
func (k Keeper) withdrawValidatorDelayedRewardsOf(ctx sdk.Context, valAdr sdk.ValAddress, currTime time.Time) {
	tokens := sdk.NewDecCoins()
	k.IterateValidatorDelayedRewardsOf(ctx, valAdr, func(startTime time.Time, v types.ValidatorDelayedReward) (stop bool) {
		tokens = tokens.Add(k.withdrawValidatorDelayedRewardOf(ctx, valAdr, startTime, currTime, v)...)
		return false
	})
	if !tokens.IsZero() {
		if val := k.stakingKeeper.Validator(ctx, valAdr); val != nil {
			k.AllocateTokensToValidator(ctx, val, tokens)
		}
	}
}

// both startTime and currTime must be DelayedRewardTime
func (k Keeper) withdrawValidatorDelayedRewardOf(ctx sdk.Context, valAdr sdk.ValAddress, startTime, currTime time.Time, v types.ValidatorDelayedReward) sdk.DecCoins {
	tokens := sdk.NewDecCoins()
	// n is time.duration
	n := currTime.Sub(v.UpdateTime) / k.GetDelayedRewardUnit(ctx)

	if int64(n) == 0 {
		return tokens
	}

	for index, decCoin := range v.Reward {
		// unit is decimal
		unit := v.Unit.AmountOf(decCoin.Denom)
		if unit.IsZero() {
			// unit = amount / period
			unit = decCoin.Amount.QuoInt64(v.Period)
			v.Unit = v.Unit.Add(sdk.NewDecCoinFromDec(decCoin.Denom, unit))
		}

		// amount is decimal
		amount := unit.MulInt64(int64(n))
        	if decCoin.Amount.LT(amount) {
			amount = decCoin.Amount
		}

		v.Reward[index].Amount = v.Reward[index].Amount.Sub(amount)
		tokens = tokens.Add(sdk.NewDecCoinFromDec(decCoin.Denom, amount))
	}

	if v.Reward.IsZero() {
		k.DeleteValidatorDelayedReward(ctx, valAdr, startTime)
	} else {
		v.UpdateTime = currTime
		k.SetValidatorDelayedReward(ctx, valAdr, startTime, v)
	}

	return tokens
}

// time
func (k Keeper) GetDelayedRewardTime(ctx sdk.Context, t time.Time) (newTime time.Time) {
	return t.UTC().Truncate(k.GetDelayedRewardUnit(ctx))
}

// reports whether the time instant t1 is after t2
// both t1 and t2 must be DelayedRewardTime
func AfterOfDelayedRewardTime(t1, t2 time.Time) bool {
	return t1.After(t2)
}

// store
// get the reward
func (k Keeper) GetValidatorDelayedReward(ctx sdk.Context, val sdk.ValAddress, startTime time.Time) (v types.ValidatorDelayedReward, ok bool) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(types.GetValidatorDelayedRewardKey(val, startTime))
	if b == nil {
		return v, false
	}
	k.cdc.MustUnmarshal(b, &v)
	return v, true
}

// set the reward of a validator
func (k Keeper) SetValidatorDelayedReward(ctx sdk.Context, val sdk.ValAddress, startTime time.Time, v types.ValidatorDelayedReward) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&v)
	store.Set(types.GetValidatorDelayedRewardKey(val, startTime), bz)
}

// delete a reward of a validator
func (k Keeper) DeleteValidatorDelayedReward(ctx sdk.Context, val sdk.ValAddress, startTime time.Time) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetValidatorDelayedRewardKey(val, startTime))
}

// delete all rewards of a validator
func (k Keeper) DeleteAllValidatorDelayedReward(ctx sdk.Context, val sdk.ValAddress) sdk.DecCoins {
	ret := sdk.NewDecCoins()

	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.GetValidatorDelayedRewardWithAddressPrefix(val))

	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var v types.ValidatorDelayedReward
		k.cdc.MustUnmarshal(iter.Value(), &v)

		ret = ret.Add(v.Reward...)

		store.Delete(iter.Key())
	}

	return ret
}

func (k Keeper) GetValidatorDelayedRewardInfo(ctx sdk.Context, val sdk.ValAddress) (lastTime types.ValidatorDelayedRewardInfo) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(types.GetValidatorDelayedRewardInfoKey(val))
	if b == nil {
		return types.NewValidatorDelayedRewardInfo(time.Date(1970,1,1,0,0,0,0, time.UTC))
	}
	k.cdc.MustUnmarshal(b, &lastTime)
	return lastTime
}

func (k Keeper) SetValidatorDelayedRewardInfo(ctx sdk.Context, val sdk.ValAddress, lastTime types.ValidatorDelayedRewardInfo) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&lastTime)
	store.Set(types.GetValidatorDelayedRewardInfoKey(val), bz)
}

func (k Keeper) DeleteValidatorDelayedRewardInfo(ctx sdk.Context, val sdk.ValAddress) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetValidatorDelayedRewardInfoKey(val))
}

// iterate all rewards of a validator
func (k Keeper) IterateValidatorDelayedRewardsOf(ctx sdk.Context, val sdk.ValAddress, handler func(startTime time.Time, v types.ValidatorDelayedReward) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.GetValidatorDelayedRewardWithAddressPrefix(val))

	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var v types.ValidatorDelayedReward
		k.cdc.MustUnmarshal(iter.Value(), &v)
		_, startTime := types.SplitValidatorDelayedRewardKey(iter.Key())
		if handler(startTime, v) {
			break
		}
	}
}

// iterate all rewards of all validators
func (k Keeper) IterateValidatorDelayedRewards(ctx sdk.Context, handler func(val sdk.ValAddress, startTime time.Time, v types.ValidatorDelayedReward) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.ValidatorDelayedRewardPrefix)

	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var v types.ValidatorDelayedReward
		k.cdc.MustUnmarshal(iter.Value(), &v)
		val, startTime := types.SplitValidatorDelayedRewardKey(iter.Key())
		if handler(val, startTime, v) {
			break
		}
	}
}
