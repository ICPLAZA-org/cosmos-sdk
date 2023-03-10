package types

import (
	"time"

	yaml "gopkg.in/yaml.v2"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewReallocatedCommissionRule returns an initialized validator reallocated commission rule.
func NewReallocatedCommissionRule(validatorRate, recommandersRate sdk.Dec, incentiveDepth uint32, recommanderClassRates []RecommanderClassRate) ReallocatedCommissionRule {
	return ReallocatedCommissionRule{
		ValidatorRate:         validatorRate,
		RecommandersRate:      recommandersRate,
		IncentiveDepth:        incentiveDepth,
		RecommanderClassRates: recommanderClassRates,
		UpdateTime:            time.Unix(0, 0).UTC(),
	}
}

// NewReallocatedCommissionRuleWithTime returns an initialized validator reallocated commission rule with a specified
// update time which should be the current block BFT time.
func NewReallocatedCommissionRuleWithTime(validatorRate, recommandersRate sdk.Dec, incentiveDepth uint32, recommanderClassRates []RecommanderClassRate, updatedAt time.Time) ReallocatedCommissionRule {
	return ReallocatedCommissionRule{
		ValidatorRate:         validatorRate,
		RecommandersRate:      recommandersRate,
		IncentiveDepth:        incentiveDepth,
		RecommanderClassRates: recommanderClassRates,
		UpdateTime:            updatedAt,
	}
}

// String implements the Stringer interface for a ReallocatedCommissionRule object.
func (c ReallocatedCommissionRule) String() string {
	out, _ := yaml.Marshal(c)
	return string(out)
}

// Validate performs basic sanity validation checks of initial reallocated commission rule
// parameters. If validation fails, an SDK error is returned.
func (c ReallocatedCommissionRule) Validate() error {
	switch {
	case c.ValidatorRate.IsNegative():
		// rate cannot be negative
		return ErrCommissionNegative

	case c.RecommandersRate.IsNegative():
		// rate cannot be negative
		return ErrCommissionNegative

	case c.ValidatorRate.GT(sdk.OneDec()):
		// rate cannot be greater than 1
		return ErrCommissionHuge

	case c.RecommandersRate.GT(sdk.OneDec()):
		// rate cannot be greater than 1
		return ErrCommissionHuge

	case c.ValidatorRate.Add(c.RecommandersRate).GT(sdk.OneDec()):
		// rate cannot be greater than 1
		return ErrCommissionHuge

	case int(c.IncentiveDepth) != len(c.RecommanderClassRates):
		// mismatch
		return ErrMismatchRecommanderClass
	}

	if err := c.validateRecommanderClassRates(); err != nil {
		return err
	}

	return nil
}

func (c ReallocatedCommissionRule) validateRecommanderClassRates() error {
	return validateRecommanderClassRates(c.RecommanderClassRates)
}

func (c ReallocatedCommissionRule) ValidateRCommissionNewRate(newValidatorRate, newRecommanderRate sdk.Dec, blockTime time.Time) error {
	switch {
	case blockTime.Sub(c.UpdateTime).Hours() < 24:
		// new rate cannot be changed more than once within 24 hours
		return ErrCommissionUpdateTime

	case newValidatorRate.IsNegative():
		// new rate cannot be negative
		return ErrCommissionNegative

	case newRecommanderRate.IsNegative():
		// new rate cannot be negative
		return ErrCommissionNegative

	case newValidatorRate.Add(newRecommanderRate).GT(sdk.OneDec()):
		// new rate % points change cannot be greater than the max change rate
		return ErrCommissionGTMaxChangeRate
	}

	return nil
}


func (c ReallocatedCommissionRule) ValidateRecommanderNewRule(incentiveDepth uint32, recommanderClassRates []RecommanderClassRate, blockTime time.Time) error {
	switch {
	case blockTime.Sub(c.UpdateTime).Hours() < 24:
		// new rate cannot be changed more than once within 24 hours
		return ErrCommissionUpdateTime

	case int(incentiveDepth) != len(recommanderClassRates) :
		// mismatch
		return ErrMismatchRecommanderClass
		
	}

	return nil
}
