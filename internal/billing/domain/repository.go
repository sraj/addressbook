package domain

import "context"

type Repository interface {
	GetPlans(ctx context.Context) ([]Plan, error)
	GetPlan(ctx context.Context, id uint) (*Plan, error)
	GetPlanByName(ctx context.Context, name string) (*Plan, error)
	SetPlanPriceID(ctx context.Context, planID uint, priceID string) error
	UpdatePlanPrice(ctx context.Context, planID uint, priceMonthly int64) error
	CreateAccount(ctx context.Context, account *Account) error
	GetAccount(ctx context.Context, accountID uint) (*Account, error)
	GetAccountByUser(ctx context.Context, userID uint) (*Account, error)
	GetAccountByCustomerID(ctx context.Context, customerID string) (*Account, error)
	CreateMember(ctx context.Context, member *AccountMember) error
	SetAccountPlan(ctx context.Context, accountID, planID uint) error
	SetStripeCustomerID(ctx context.Context, accountID uint, customerID string) error
	GetSubscription(ctx context.Context, accountID uint) (*Subscription, error)
	GetSubscriptionByStripeID(ctx context.Context, stripeID string) (*Subscription, error)
	CreateSubscription(ctx context.Context, sub *Subscription) error
	GetUsage(ctx context.Context, accountID uint) (map[string]int, error)
	IncrementUsage(ctx context.Context, accountID uint, resource string) error
	DecrementUsage(ctx context.Context, accountID uint, resource string) error
	GetUserRole(ctx context.Context, userID uint) (string, error)
}
