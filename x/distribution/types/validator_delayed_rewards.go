package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// create a new ValidatorDelayedReward
func NewValidatorDelayedReward(period int64, updateTime time.Time, unit, reward sdk.DecCoins) ValidatorDelayedReward {
	return ValidatorDelayedReward{
		Period: period,
		UpdateTime: updateTime,
		Unit: unit,
		Reward: reward,
	}
}

func NewValidatorDelayedRewardInfo(lastTime time.Time) ValidatorDelayedRewardInfo {
	return ValidatorDelayedRewardInfo {
		LastTime: lastTime,
	}
}
