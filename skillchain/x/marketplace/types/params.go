package types

import (
	"fmt"

	"cosmossdk.io/math"
)

// Default parameter values
var (
	DefaultPlatformFeePercent  = uint64(5)      // 5%
	DefaultMinContractDuration = uint64(86400)  // 1 days in secondes
	DefaultMinGigPrice         = math.NewInt(100_000_000) // 100 SKILL
)

// NewParams creates a new Params instance.
func NewParams(feePercent, minDuration uint64, minPrice math.Int) Params {
	return Params{
		PlatformFeePercent:  feePercent,
		MinContractDuration: minDuration,
		MinGigPrice:         minPrice,
	}
}

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	return NewParams(
		DefaultPlatformFeePercent,
		DefaultMinContractDuration,
		DefaultMinGigPrice,
	)
}

// Validate validates the set of params.
func (p Params) Validate() error {
	if p.PlatformFeePercent > 100 {
		return fmt.Errorf("platform fee cannot exceed 100%")
	}
	if p.MinContractDuration == 0 {
		return fmt.Errorf("min contract duration (in seconds) must be greater than zero")
	}
	if p.MinGigPrice.IsNegative() {
		return fmt.Errorf("min gig price must be a positive integer")
	}

	return nil
}
