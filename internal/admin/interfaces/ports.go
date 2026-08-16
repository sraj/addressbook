package interfaces

import (
	"context"

	adminApp "github.com/sraj/addressbook/internal/admin/application"
	billingDomain "github.com/sraj/addressbook/internal/billing/domain"
)

// adminChecker is the consumer-facing port for user administration. It is
// satisfied by *application.Service.
type adminChecker interface {
	IsAdmin(ctx context.Context, userID uint) (bool, error)
	ListUsers(ctx context.Context) ([]adminApp.UserRow, error)
	UpdateUserStatus(ctx context.Context, userID uint, status string) error
	Stats(ctx context.Context) (*adminApp.StatsResponse, error)
}

// planManager lets the admin panel manage billing plans. It is satisfied by
// *billing/application.Service.
type planManager interface {
	ListPlans(ctx context.Context) ([]billingDomain.Plan, error)
	UpdatePlanPriceID(ctx context.Context, planID uint, priceID string) error
	SyncPricesFromStripe(ctx context.Context) error
}
