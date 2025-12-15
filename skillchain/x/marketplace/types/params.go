package types

import (
	"fmt"

	"cosmossdk.io/math"
)

// Default parameter values
var (
	DefaultPlatformFeePercent   = uint64(5)        // 5%
	DefaultMinContractDuration  = uint64(86400)    // 1 days in secondes
	DefaultMinGigPrice          = math.NewInt(100) // 100 SKILL
	DefaultDisputeDuration      = uint64(604800)   // 7 days in seconds
	DefaultMinArbitersRequired  = uint64(3)        // 3 arbiters
	DefaultArbiterStakeRequired = uint64(1000)     // 1000 SKILL
)

// NewParams creates a new Params instance.
func NewParams(feePercent, minDuration uint64, minPrice math.Int, disputeDuration, minArbitersRequired, arbiterStakeRequired uint64) Params {
	return Params{
		PlatformFeePercent:   feePercent,
		MinContractDuration:  minDuration,
		MinGigPrice:          minPrice,
		DisputeDuration:      disputeDuration,
		MinArbitersRequired:  minArbitersRequired,
		ArbiterStakeRequired: arbiterStakeRequired,
	}
}

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	return NewParams(
		DefaultPlatformFeePercent,
		DefaultMinContractDuration,
		DefaultMinGigPrice,
		DefaultDisputeDuration,
		DefaultMinArbitersRequired,
		DefaultArbiterStakeRequired,
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
	if p.ArbiterStakeRequired < 1 {
		return fmt.Errorf("arbiter stake required must be greater than zero")
	}
	if p.MinArbitersRequired < 1 {
		return fmt.Errorf("min arbiters required must be at least 1")
	}
	if p.DisputeDuration < 86400 {
		return fmt.Errorf("dispute duration must be at least 1 day")
	}

	return nil
}
