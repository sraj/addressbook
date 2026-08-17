package application

import "context"

// quotaChecker is the provider-side port the service depends on for billing
// quota enforcement. It is satisfied by *billing/application.Service.
type quotaChecker interface {
	CheckQuota(ctx context.Context, userID uint, resource string) error
	IncrementUsage(ctx context.Context, userID uint, resource string) error
	DecrementUsage(ctx context.Context, userID uint, resource string) error
	RemainingQuota(ctx context.Context, userID uint, resource string) (int, error)
}
