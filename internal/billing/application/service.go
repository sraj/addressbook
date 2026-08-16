// Package billing provides account management, plan limits, usage tracking,
// and quota enforcement across the application.
package application

import (
	"context"
	"errors"

	"github.com/sraj/addressbook/internal/billing/domain"
)

// Service implements billing operations: plan management, account creation,
// usage quotas, and Stripe price synchronization.
type Service struct {
	repo            domain.Repository
	stripeSecretKey string
}

// NewService creates a billing service with the given repository and Stripe key.
func NewService(repo domain.Repository, stripeSecretKey string) *Service {
	return &Service{repo: repo, stripeSecretKey: stripeSecretKey}
}

// InitAccount creates a free-tier account for a newly registered user.
// Creates the account row, member record, and an active subscription.
func (s *Service) InitAccount(ctx context.Context, userID uint, accountName, email string) error {
	plan, err := s.repo.GetPlanByName(ctx, "free")
	if err != nil {
		return err
	}

	account := &domain.Account{
		Name:      accountName,
		PlanID:    plan.ID,
		UsageJSON: "{}",
	}
	if err := s.repo.CreateAccount(ctx, account); err != nil {
		return err
	}

	member := &domain.AccountMember{
		AccountID: account.ID,
		UserID:    userID,
		Role:      "owner",
	}
	if err := s.repo.CreateMember(ctx, member); err != nil {
		return err
	}

	sub := &domain.Subscription{
		AccountID: account.ID,
		PlanID:    plan.ID,
		Status:    "active",
	}
	if err := s.repo.CreateSubscription(ctx, sub); err != nil {
		return err
	}

	return nil
}

// EnsureAccount creates a free-tier account for the user if they don't already
// have one. Used by the admin panel when activating a user.
func (s *Service) EnsureAccount(ctx context.Context, userID uint, email string) error {
	_, err := s.repo.GetAccountByUser(ctx, userID)
	if err == nil {
		return nil // already has an account
	}
	if !errors.Is(err, domain.ErrAccountNotFound) {
		return err
	}
	return s.InitAccount(ctx, userID, email, email)
}

// GetAccount returns the account associated with a user.
func (s *Service) GetAccount(ctx context.Context, userID uint) (*domain.Account, error) {
	return s.repo.GetAccountByUser(ctx, userID)
}

// GetAccountByCustomerID finds an account by its Stripe customer ID.
func (s *Service) GetAccountByCustomerID(ctx context.Context, customerID string) (*domain.Account, error) {
	return s.repo.GetAccountByCustomerID(ctx, customerID)
}

// GetAccountByID finds an account by its internal ID.
func (s *Service) GetAccountByID(ctx context.Context, accountID uint) (*domain.Account, error) {
	return s.repo.GetAccount(ctx, accountID)
}

// GetSubscription returns the active subscription for an account.
func (s *Service) GetSubscription(ctx context.Context, accountID uint) (*domain.Subscription, error) {
	return s.repo.GetSubscription(ctx, accountID)
}

// GetSubscriptionByStripeID looks up a subscription by its Stripe subscription ID.
func (s *Service) GetSubscriptionByStripeID(ctx context.Context, stripeID string) (*domain.Subscription, error) {
	return s.repo.GetSubscriptionByStripeID(ctx, stripeID)
}

// SetStripeCustomerID saves the Stripe customer ID on the account.
func (s *Service) SetStripeCustomerID(ctx context.Context, accountID uint, customerID string) error {
	return s.repo.SetStripeCustomerID(ctx, accountID, customerID)
}

// SetPlan changes the account's plan and creates a new active subscription record.
func (s *Service) SetPlan(ctx context.Context, accountID, planID uint) error {
	if err := s.repo.SetAccountPlan(ctx, accountID, planID); err != nil {
		return err
	}
	return s.repo.CreateSubscription(ctx, &domain.Subscription{
		AccountID: accountID,
		PlanID:    planID,
		Status:    "active",
	})
}

// SetPlanWithStripeID is like SetPlan but also stores the Stripe subscription ID.
func (s *Service) SetPlanWithStripeID(ctx context.Context, accountID, planID uint, stripeSubID string) error {
	if err := s.repo.SetAccountPlan(ctx, accountID, planID); err != nil {
		return err
	}
	return s.repo.CreateSubscription(ctx, &domain.Subscription{
		AccountID:            accountID,
		PlanID:               planID,
		StripeSubscriptionID: stripeSubID,
		Status:               "active",
	})
}
