package types

// DONTCOVER

import (
	"cosmossdk.io/errors"
)

// x/marketplace module sentinel errors
var (
	ErrInvalidSigner     = errors.Register(ModuleName, 1100, "expected gov account as only signer for proposal message")
	ErrProfileNotFound   = errors.Register(ModuleName, 1101, "profile not found")
	ErrProfileExists     = errors.Register(ModuleName, 1102, "profile already exists")
	ErrGigNotFound       = errors.Register(ModuleName, 1200, "gig not found")
	ErrInvalidGigStatus  = errors.Register(ModuleName, 1201, "invalid gig status transition")
	ErrUnauthorized      = errors.Register(ModuleName, 1300, "unauthorized")
	ErrInsufficientFunds = errors.Register(ModuleName, 1400, "insufficient funds")
	ErrInvalidPrice      = errors.Register(ModuleName, 1401, "invalid price")
)
