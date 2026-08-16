package domain

import "errors"

var (
	ErrQuotaExceeded   = errors.New("quota exceeded for this resource")
	ErrAccountNotFound = errors.New("account not found")
	ErrPlanNotFound    = errors.New("plan not found")
)
