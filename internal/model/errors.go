package model

import "errors"

var (
	ErrInsufficientFunds     = errors.New("insufficient funds")
	ErrInvalidSum            = errors.New("sum must be positive")
	ErrOrderNumberRequired   = errors.New("order number required")
	ErrOrderAlreadyWithdrawn = errors.New("order already withdrawn")
)
