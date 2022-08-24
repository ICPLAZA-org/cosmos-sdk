package types

import (
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

//-----------------------------------------------------------------------------
// Sorting

type RecommanderClassRates []RecommanderClassRate

func validateRecommanderClassRates(r RecommanderClassRates) error {
	totalRate := sdk.ZeroDec()

	for _, rate := range r {
		if rate.Rate.IsNegative() {
			return ErrCommissionNegative
		}else if rate.Rate.GT(sdk.OneDec()) {
			return ErrCommissionHuge
		}

		totalRate = totalRate.Add(rate.Rate)
	}

	if totalRate.GT(sdk.OneDec()) {
		return ErrCommissionHuge
	}

	return nil
}

var _ sort.Interface = RecommanderClassRates{}

// Len implements sort.Interface for RecommanderClassRates
func (r RecommanderClassRates) Len() int { return len(r) }

// Less implements sort.Interface for RecommanderClassRates
func (r RecommanderClassRates) Less(i, j int) bool { return r[i].Index < r[j].Index }

// Swap implements sort.Interface for RecommanderClassRates
func (r RecommanderClassRates) Swap(i, j int) { r[i], r[j] = r[j], r[i] }

// Sort is a helper function to sort the set of decimal RecommanderClassRates in-place.
func (r RecommanderClassRates) Sort() RecommanderClassRates {
        sort.Sort(r)
        return r
}
