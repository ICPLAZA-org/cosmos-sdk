package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// initialize starting info for a new delegation
func (k Keeper) initializeDelegation(ctx sdk.Context, val sdk.ValAddress, del sdk.AccAddress) {
	// period has already been incremented - we want to store the period ended by this delegation action
	previousPeriod := k.GetValidatorCurrentRewards(ctx, val).Period - 1

	// increment reference count for the period we're going to track
	k.incrementReferenceCount(ctx, val, previousPeriod)

	validator := k.stakingKeeper.Validator(ctx, val)
	delegation := k.stakingKeeper.Delegation(ctx, del, val)

	// calculate delegation stake in tokens
	// we don't store directly, so multiply delegation shares * (tokens per share)
	// note: necessary to truncate so we don't allow withdrawing more rewards than owed
	stake := validator.TokensFromSharesTruncated(delegation.GetShares())
	k.SetDelegatorStartingInfo(ctx, val, del, types.NewDelegatorStartingInfo(previousPeriod, stake, uint64(ctx.BlockHeight())))
}

func (k Keeper) calculateDelegationOnlyRewardsBetween(ctx sdk.Context, val stakingtypes.ValidatorI,
	startingPeriod, endingPeriod uint64, stake sdk.Dec) (rewards sdk.DecCoins) {
	rewards, _ = k.calculateDelegationRewardsBetween(ctx, val, startingPeriod, endingPeriod, stake, false)
	return rewards
}

func (k Keeper) calculateDelegationAllRewardsBetween(ctx sdk.Context, val stakingtypes.ValidatorI,
	startingPeriod, endingPeriod uint64, stake sdk.Dec) (rewards sdk.DecCoins, recommandersRewards sdk.DecCoins) {
	rewards, recommandersRewards = k.calculateDelegationRewardsBetween(ctx, val, startingPeriod, endingPeriod, stake, true)
	return rewards, recommandersRewards
}

// calculate the rewards accrued by a delegation between two periods
func (k Keeper) calculateDelegationRewardsBetween(ctx sdk.Context, val stakingtypes.ValidatorI,
	startingPeriod, endingPeriod uint64, stake sdk.Dec, isCalcRecommandersRewards bool) (rewards sdk.DecCoins, recommandersRewards sdk.DecCoins) {
	// sanity check
	if startingPeriod > endingPeriod {
		panic("startingPeriod cannot be greater than endingPeriod")
	}

	// sanity check
	if stake.IsNegative() {
		panic("stake should not be negative")
	}

	// return staking * (ending - starting)
	starting := k.GetValidatorHistoricalRewards(ctx, val.GetOperator(), startingPeriod)
	ending := k.GetValidatorHistoricalRewards(ctx, val.GetOperator(), endingPeriod)
	difference := ending.CumulativeRewardRatio.Sub(starting.CumulativeRewardRatio)
	if difference.IsAnyNegative() {
		panic("negative rewards should not be possible")
	}
	// note: necessary to truncate so we don't allow withdrawing more rewards than owed
	rewards = difference.MulDecTruncate(stake)

	if isCalcRecommandersRewards {
		difference = ending.CumulativeRecommandersRewardRatio.Sub(starting.CumulativeRecommandersRewardRatio)
		if difference.IsAnyNegative() {
			panic("negative recommanders rewards should not be possible")
		}

		// note: necessary to truncate so we don't allow withdrawing more rewards than owed
		recommandersRewards = difference.MulDecTruncate(stake)
	}

	return rewards, recommandersRewards
}

// calculate the total rewards accrued by a delegation
func (k Keeper) CalculateDelegationRewards(ctx sdk.Context, val stakingtypes.ValidatorI, del stakingtypes.DelegationI, 
	endingPeriod uint64) (rewards sdk.DecCoins, recommandersRewards sdk.DecCoins) {
	// fetch starting info for delegation
	startingInfo := k.GetDelegatorStartingInfo(ctx, del.GetValidatorAddr(), del.GetDelegatorAddr())

	if startingInfo.Height == uint64(ctx.BlockHeight()) {
		// started this height, no rewards yet
		return
	}

	startingPeriod := startingInfo.PreviousPeriod
	stake := startingInfo.Stake

	// Iterate through slashes and withdraw with calculated staking for
	// distribution periods. These period offsets are dependent on *when* slashes
	// happen - namely, in BeginBlock, after rewards are allocated...
	// Slashes which happened in the first block would have been before this
	// delegation existed, UNLESS they were slashes of a redelegation to this
	// validator which was itself slashed (from a fault committed by the
	// redelegation source validator) earlier in the same BeginBlock.
	startingHeight := startingInfo.Height
	// Slashes this block happened after reward allocation, but we have to account
	// for them for the stake sanity check below.
	endingHeight := uint64(ctx.BlockHeight())
	if endingHeight > startingHeight {
		k.IterateValidatorSlashEventsBetween(ctx, del.GetValidatorAddr(), startingHeight, endingHeight,
			func(height uint64, event types.ValidatorSlashEvent) (stop bool) {
				endingPeriod := event.ValidatorPeriod
				if endingPeriod > startingPeriod {
					// rewards = rewards.Add(k.calculateDelegationRewardsBetween(ctx, val, startingPeriod, endingPeriod, stake)...)
					rewards = rewards.Add(k.calculateDelegationOnlyRewardsBetween(ctx, val, startingPeriod, endingPeriod, stake)...)

					// Note: It is necessary to truncate so we don't allow withdrawing
					// more rewards than owed.
					stake = stake.MulTruncate(sdk.OneDec().Sub(event.Fraction))
					startingPeriod = endingPeriod
				}
				return false
			},
		)
	}

	// A total stake sanity check; Recalculated final stake should be less than or
	// equal to current stake here. We cannot use Equals because stake is truncated
	// when multiplied by slash fractions (see above). We could only use equals if
	// we had arbitrary-precision rationals.
	currentStake := val.TokensFromShares(del.GetShares())

	if stake.GT(currentStake) {
		// AccountI for rounding inconsistencies between:
		//
		//     currentStake: calculated as in staking with a single computation
		//     stake:        calculated as an accumulation of stake
		//                   calculations across validator's distribution periods
		//
		// These inconsistencies are due to differing order of operations which
		// will inevitably have different accumulated rounding and may lead to
		// the smallest decimal place being one greater in stake than
		// currentStake. When we calculated slashing by period, even if we
		// round down for each slash fraction, it's possible due to how much is
		// being rounded that we slash less when slashing by period instead of
		// for when we slash without periods. In other words, the single slash,
		// and the slashing by period could both be rounding down but the
		// slashing by period is simply rounding down less, thus making stake >
		// currentStake
		//
		// A small amount of this error is tolerated and corrected for,
		// however any greater amount should be considered a breach in expected
		// behaviour.
		marginOfErr := sdk.SmallestDec().MulInt64(3)
		if stake.LTE(currentStake.Add(marginOfErr)) {
			stake = currentStake
		} else {
			panic(fmt.Sprintf("calculated final stake for delegator %s greater than current stake"+
				"\n\tfinal stake:\t%s"+
				"\n\tcurrent stake:\t%s",
				del.GetDelegatorAddr(), stake, currentStake))
		}
	}

	// calculate rewards for final period
	rewardsL, recommandersRewards := k.calculateDelegationAllRewardsBetween(ctx, val, startingPeriod, endingPeriod, stake)
	// rewards = rewards.Add(k.calculateDelegationRewardsBetween(ctx, val, startingPeriod, endingPeriod, stake)...)
	rewards = rewards.Add(rewardsL...)
	return rewards, recommandersRewards
}

func (k Keeper) withdrawDelegationRewards(ctx sdk.Context, val stakingtypes.ValidatorI, del stakingtypes.DelegationI) (sdk.Coins, sdk.Coins, error) {
	// check existence of delegator starting info
	if !k.HasDelegatorStartingInfo(ctx, del.GetValidatorAddr(), del.GetDelegatorAddr()) {
		return nil, nil, types.ErrEmptyDelegationDistInfo
	}

	// end current period and calculate rewards
	endingPeriod := k.IncrementValidatorPeriod(ctx, val)
	// rewardsRaw := k.CalculateDelegationRewards(ctx, val, del, endingPeriod)
	rewardsRaw, recommandersRewardsRaw := k.CalculateDelegationRewards(ctx, val, del, endingPeriod)
	outstanding := k.GetValidatorOutstandingRewardsCoins(ctx, del.GetValidatorAddr())

	// defensive edge case may happen on the very final digits
	// of the decCoins due to operation order of the distribution mechanism.
	rewards := rewardsRaw.Intersect(outstanding)
	if !rewards.IsEqual(rewardsRaw) {
		logger := k.Logger(ctx)
		logger.Info(
			"rounding error withdrawing rewards from validator",
			"delegator", del.GetDelegatorAddr().String(),
			"validator", val.GetOperator().String(),
			"got", rewards.String(),
			"expected", rewardsRaw.String(),
		)
	}

	// truncate coins, return remainder to community pool
	coins, remainder := rewards.TruncateDecimal()

	// add coins to user account
	if !coins.IsZero() {
		withdrawAddr := k.GetDelegatorWithdrawAddr(ctx, del.GetDelegatorAddr())
		err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, withdrawAddr, coins)
		if err != nil {
			return nil, nil, err
		}
	}

	// defensive edge case may happen on the very final digits
	// of the decCoins due to operation order of the distribution mechanism.
	recommandersRewards := recommandersRewardsRaw.Intersect(outstanding)
	if !recommandersRewards.IsEqual(recommandersRewardsRaw) {
		logger := k.Logger(ctx)
		logger.Info(
			"rounding error withdrawing recommanders rewards from validator",
			"delegator", del.GetDelegatorAddr().String(),
			"validator", val.GetOperator().String(),
			"got", recommandersRewards.String(),
			"expected", recommandersRewardsRaw.String(),
		)
	}
	
	// truncate coins, return remainder to community pool
	coinsL, remainderL, err := k.withdrawDelegationRecommandersRewards(ctx, val, del, recommandersRewards)
	if err != nil {
		return nil, nil, err
	}

	// summary
	remainder = remainder.Add(remainderL...)

	// update the outstanding rewards and the community pool only if the
	// transaction was successful
	// k.SetValidatorOutstandingRewards(ctx, del.GetValidatorAddr(), types.ValidatorOutstandingRewards{Rewards: outstanding.Sub(rewards)})
	k.SetValidatorOutstandingRewards(ctx, del.GetValidatorAddr(), types.ValidatorOutstandingRewards{Rewards: outstanding.Sub(rewards).Sub(recommandersRewards)})
	feePool := k.GetFeePool(ctx)
	feePool.CommunityPool = feePool.CommunityPool.Add(remainder...)
	k.SetFeePool(ctx, feePool)

	// decrement reference count of starting period
	startingInfo := k.GetDelegatorStartingInfo(ctx, del.GetValidatorAddr(), del.GetDelegatorAddr())
	startingPeriod := startingInfo.PreviousPeriod
	k.decrementReferenceCount(ctx, del.GetValidatorAddr(), startingPeriod)

	// remove delegator starting info
	k.DeleteDelegatorStartingInfo(ctx, del.GetValidatorAddr(), del.GetDelegatorAddr())

	return coins, coinsL, nil
}

func (k Keeper) withdrawDelegationRecommandersRewards(ctx sdk.Context, val stakingtypes.ValidatorI, del stakingtypes.DelegationI,
	recommandersRewards sdk.DecCoins) (sdk.Coins, sdk.DecCoins, error) {
	
	if recommandersRewards.IsZero() {
		return sdk.Coins{}, sdk.DecCoins{}, nil
	}

	totalRewards := sdk.Coins{}
	remainderRewards := recommandersRewards
	currentDel := del
	_, _, _, recommanderClassRates := val.GetReallocatedCommissionRule()

	for _, rate := range recommanderClassRates {
		if currentDel == nil { // to validator
			break
		}

		currentRewards := recommandersRewards.MulDec(rate.Rate)
		if remainderRewards.Sub(currentRewards).IsAnyNegative() {
			logger := k.Logger(ctx)
			logger.Info(
				"rounding error withdrawing a recommander rewards from validator",
				"delegator", currentDel.GetDelegatorAddr().String(),
				"validator", val.GetOperator().String(),
				"got", rate.String(),
			)
			break;
		}

		rewards, remainder := currentRewards.TruncateDecimal()
		remainderRewards = remainderRewards.Sub(currentRewards).Add(remainder...)

		recommanderAddr := k.getRealRecommander(ctx, currentDel, del.GetValidatorAddr())
		if recommanderAddr == nil {
			break;
		}
		if !rewards.IsZero() {
			if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, recommanderAddr, rewards); err != nil {
				return nil, nil, err
			}
		}
		totalRewards = totalRewards.Add(rewards...)

		currentDel = k.stakingKeeper.Delegation(ctx, recommanderAddr, del.GetValidatorAddr())
	}

	rewards, remainder := remainderRewards.TruncateDecimal()
	if !rewards.IsZero() {
		err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, sdk.AccAddress(del.GetValidatorAddr()), rewards)
		if err != nil {
			return nil, nil, err
		}
	}
	totalRewards = totalRewards.Add(rewards...)
	
	return totalRewards, remainder, nil
}

func (k Keeper) getRealRecommander(ctx sdk.Context, del stakingtypes.DelegationI, valAddr sdk.ValAddress) sdk.AccAddress {
	recommanderAddr := del.GetRecommanderAddr()

	if recommanderAddr.Equals(del.GetDelegatorAddr()) {
		recommanderAddr = nil
	}

	return recommanderAddr
}
