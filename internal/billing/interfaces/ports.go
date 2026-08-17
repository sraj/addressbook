package interfaces

import (
	"context"

	"github.com/sraj/addressbook/internal/billing/domain"
	"github.com/sraj/addressbook/internal/billing/infrastructure"
	"github.com/stripe/stripe-go/v79"
)

// billingService is the consumer-facing port the handler depends on. It is
// satisfied by *application.Service.
type billingService interface {
	GetAccount(ctx context.Context, userID uint) (*domain.Account, error)
	GetAccountByID(ctx context.Context, accountID uint) (*domain.Account, error)
	GetAccountByCustomerID(ctx context.Context, customerID string) (*domain.Account, error)
	GetPlan(ctx context.Context, planID uint) (*domain.Plan, error)
	GetPlanByName(ctx context.Context, name string) (*domain.Plan, error)
	GetSubscription(ctx context.Context, accountID uint) (*domain.Subscription, error)
	GetSubscriptionByStripeID(ctx context.Context, stripeID string) (*domain.Subscription, error)
	GetPlans(ctx context.Context) ([]domain.Plan, error)
	GetUsage(ctx context.Context, userID uint) (map[string]int, *domain.Plan, error)
	SetStripeCustomerID(ctx context.Context, accountID uint, customerID string) error
	SetPlan(ctx context.Context, accountID, planID uint) error
	SetPlanWithStripeID(ctx context.Context, accountID, planID uint, stripeSubID string) error
}

// stripeChecker is satisfied by *infrastructure.StripeService (nil when Stripe
// is not configured).
type stripeChecker interface {
	CreateCheckoutSession(ctx context.Context, account *domain.Account, userEmail, priceID, planName string) (string, error)
	CreatePortalSession(ctx context.Context, account *domain.Account) (string, error)
	ChangeSubscriptionPlan(ctx context.Context, subscriptionID, newPriceID string) error
	CancelSubscription(ctx context.Context, subscriptionID string) error
	ListInvoices(ctx context.Context, customerID string) ([]infrastructure.InvoiceRow, error)
	VerifyWebhook(payload []byte, sigHeader string) (stripe.Event, error)
}

// mailSender is satisfied by *mailer.Mailer (nil when email is not configured).
type mailSender interface {
	Send(to, subject, body string) error
}

// labelOrderUpdater is satisfied by *label/application.Service and is only
// wired when the label feature is available (always, in this app).
type labelOrderUpdater interface {
	MarkOrderPaidBySessionID(ctx context.Context, sessionID string) error
}
