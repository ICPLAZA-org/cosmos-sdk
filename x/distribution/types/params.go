package types

import (
	"fmt"
	"time"

	yaml "gopkg.in/yaml.v2"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

const (
	DefaultMinDelayedPeriod int64 = 120 // 120 days
	DefaultMaxDelayedInterval int64 = 60 // 120 days
	DefaultDelayedUnit time.Duration = time.Hour * 24 * 1 // 1 days
)

// Parameter keys
var (
	ParamStoreKeyCommunityTax        = []byte("communitytax")
	ParamStoreKeyBaseProposerReward  = []byte("baseproposerreward")
	ParamStoreKeyBonusProposerReward = []byte("bonusproposerreward")
	ParamStoreKeyWithdrawAddrEnabled = []byte("withdrawaddrenabled")
	ParamStoreKeyDelayedRewardProportion = []byte("delayedrewardproportion")
	ParamStoreKeyMinDelayedRewardPeriod = []byte("mindelayedrewardperiod")
	ParamStoreKeyMaxDelayedRewardInterval = []byte("maxdelayedrewardinterval")
	ParamStoreKeyDelayedRewardUnit = []byte("delayedrewardunit")
)

// ParamKeyTable returns the parameter key table.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// DefaultParams returns default distribution parameters
func DefaultParams() Params {
	return Params{
		CommunityTax:        sdk.NewDecWithPrec(2, 2), // 2%
		BaseProposerReward:  sdk.NewDecWithPrec(1, 2), // 1%
		BonusProposerReward: sdk.NewDecWithPrec(4, 2), // 4%
		WithdrawAddrEnabled: true,
		DelayedRewardProportion:  sdk.NewDecWithPrec(7,1), // 70%
		MinDelayedRewardPeriod:   DefaultMinDelayedPeriod,
		MaxDelayedRewardInterval: DefaultMaxDelayedInterval,
		DelayedRewardUnit:        DefaultDelayedUnit,
	}
}

func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

// ParamSetPairs returns the parameter set pairs.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(ParamStoreKeyCommunityTax, &p.CommunityTax, validateCommunityTax),
		paramtypes.NewParamSetPair(ParamStoreKeyBaseProposerReward, &p.BaseProposerReward, validateBaseProposerReward),
		paramtypes.NewParamSetPair(ParamStoreKeyBonusProposerReward, &p.BonusProposerReward, validateBonusProposerReward),
		paramtypes.NewParamSetPair(ParamStoreKeyWithdrawAddrEnabled, &p.WithdrawAddrEnabled, validateWithdrawAddrEnabled),
		paramtypes.NewParamSetPair(ParamStoreKeyDelayedRewardProportion, &p.DelayedRewardProportion, validateDelayedRewardProportion),
		paramtypes.NewParamSetPair(ParamStoreKeyMinDelayedRewardPeriod, &p.MinDelayedRewardPeriod, validateMinDelayedRewardPeriod),
		paramtypes.NewParamSetPair(ParamStoreKeyMaxDelayedRewardInterval, &p.MaxDelayedRewardInterval, validateMaxDelayedRewardInterval),
		paramtypes.NewParamSetPair(ParamStoreKeyDelayedRewardUnit, &p.DelayedRewardUnit, validateDelayedRewardUnit),
	}
}

// ValidateBasic performs basic validation on distribution parameters.
func (p Params) ValidateBasic() error {
	if p.CommunityTax.IsNegative() || p.CommunityTax.GT(sdk.OneDec()) {
		return fmt.Errorf(
			"community tax should be non-negative and less than one: %s", p.CommunityTax,
		)
	}
	if p.BaseProposerReward.IsNegative() {
		return fmt.Errorf(
			"base proposer reward should be positive: %s", p.BaseProposerReward,
		)
	}
	if p.BonusProposerReward.IsNegative() {
		return fmt.Errorf(
			"bonus proposer reward should be positive: %s", p.BonusProposerReward,
		)
	}
	if v := p.BaseProposerReward.Add(p.BonusProposerReward).Add(p.CommunityTax); v.GT(sdk.OneDec()) {
		return fmt.Errorf(
			"sum of base, bonus proposer rewards, and community tax cannot be greater than one: %s", v,
		)
	}

	if err := validateDelayedRewardProportion(p.DelayedRewardProportion); err != nil {
		return err
	}
	if err := validateMinDelayedRewardPeriod(p.MinDelayedRewardPeriod); err != nil {
		return err
	}
	if err := validateMaxDelayedRewardInterval(p.MaxDelayedRewardInterval); err != nil {
		return err
	}
	if err := validateDelayedRewardUnit(p.DelayedRewardUnit); err != nil {
		return err
	}

	return nil
}

func validateCommunityTax(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNil() {
		return fmt.Errorf("community tax must be not nil")
	}
	if v.IsNegative() {
		return fmt.Errorf("community tax must be positive: %s", v)
	}
	if v.GT(sdk.OneDec()) {
		return fmt.Errorf("community tax too large: %s", v)
	}

	return nil
}

func validateBaseProposerReward(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNil() {
		return fmt.Errorf("base proposer reward must be not nil")
	}
	if v.IsNegative() {
		return fmt.Errorf("base proposer reward must be positive: %s", v)
	}
	if v.GT(sdk.OneDec()) {
		return fmt.Errorf("base proposer reward too large: %s", v)
	}

	return nil
}

func validateBonusProposerReward(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNil() {
		return fmt.Errorf("bonus proposer reward must be not nil")
	}
	if v.IsNegative() {
		return fmt.Errorf("bonus proposer reward must be positive: %s", v)
	}
	if v.GT(sdk.OneDec()) {
		return fmt.Errorf("bonus proposer reward too large: %s", v)
	}

	return nil
}

func validateWithdrawAddrEnabled(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

func validateDelayedRewardProportion(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNil() {
		return fmt.Errorf("community tax must be not nil")
	}
	if v.IsNegative() {
		return fmt.Errorf("community tax must be positive: %s", v)
	}
	if v.GT(sdk.OneDec()) {
		return fmt.Errorf("community tax too large: %s", v)
	}

	return nil
}

func validateMinDelayedRewardPeriod(i interface{}) error {
	v, ok := i.(int64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v <= 0 {
		return fmt.Errorf(
			"min delayed reward period must be positive: %s", v,
		)
	}
	
	return nil
}

func validateMaxDelayedRewardInterval(i interface{}) error {
	v, ok := i.(int64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v <= 0 {
		return fmt.Errorf("delayed reward interval must be positive: %s", v)
	}

	return nil
}

func validateDelayedRewardUnit(i interface{}) error {
	v, ok := i.(time.Duration)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v <= 0 {
		return fmt.Errorf(
			"delayed reward unit must be positive: %s", v,
		)
	}

	return nil
}
