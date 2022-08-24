package cli

import (
	"errors"
	"io/ioutil"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

type (
	// RecommanderClassRatesJSON defines a slice of RecommanderClassRateJSON objects which can be
	// converted to a slice of RecommanderClassRate objects.
	RecommanderClassRatesJSON []RecommanderClassRateJSON

	// RecommanderClassRateJSON defines a RecommanderClassRate used
	// to parse parameter change proposals from a JSON file.
	RecommanderClassRateJSON struct {
		Index       uint32           `json:"index" yaml:"index"`
		Rate	    sdk.Dec          `json:"rate"  yaml:"rate"`
	}
)

func buildCommissionRates(rateStr, maxRateStr, maxChangeRateStr string) (commission types.CommissionRates, err error) {
	if rateStr == "" || maxRateStr == "" || maxChangeRateStr == "" {
		return commission, errors.New("must specify all validator commission parameters")
	}

	rate, err := sdk.NewDecFromStr(rateStr)
	if err != nil {
		return commission, err
	}

	maxRate, err := sdk.NewDecFromStr(maxRateStr)
	if err != nil {
		return commission, err
	}

	maxChangeRate, err := sdk.NewDecFromStr(maxChangeRateStr)
	if err != nil {
		return commission, err
	}

	commission = types.NewCommissionRates(rate, maxRate, maxChangeRate)

	return commission, nil
}

func buildReallocatedCommissionRule(validatorRateStr, recommandersRateStr string, incentiveDepth uint32, recommanderClassRatesJSON RecommanderClassRatesJSON) (rCommissionRule types.ReallocatedCommissionRule, err error) {

	if validatorRateStr == "" || recommandersRateStr == "" {
		return rCommissionRule, errors.New("must specify all reallocated commission parameters")
	}

	validatorRate, err := sdk.NewDecFromStr(validatorRateStr)
	if err != nil {
		return rCommissionRule, err
	}

	recommandersRate, err := sdk.NewDecFromStr(recommandersRateStr)
	if err != nil {
		return rCommissionRule, err
	}

	recommanderClassRates := buildRecommanderClassRates(recommanderClassRatesJSON)

	rCommissionRule = types.NewReallocatedCommissionRule(validatorRate, recommandersRate, incentiveDepth, recommanderClassRates)

	return rCommissionRule, nil
}

func buildRecommanderClassRates(recommanderClassRatesJSON RecommanderClassRatesJSON) types.RecommanderClassRates {
	recommanderClassRates := types.RecommanderClassRates{}
	for _, rate := range recommanderClassRatesJSON {
		recommanderClassRates = append(recommanderClassRates, types.RecommanderClassRate{Index: rate.Index, Rate: rate.Rate})
	}
	recommanderClassRates = recommanderClassRates.Sort()

	return recommanderClassRates
}

func parseRecommanderClassRatesJSON(cdc *codec.LegacyAmino, ruleFile string) (RecommanderClassRatesJSON, error) {
	rule := RecommanderClassRatesJSON{}

	if ruleFile != "" {
		contents, err := ioutil.ReadFile(ruleFile)
		if err != nil {
			return rule, err
		}

		if err := cdc.UnmarshalJSON(contents, &rule); err != nil {
			return rule, err
		}
	}

	return rule, nil
}
