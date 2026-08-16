package application

import (
	"context"
	"fmt"

	"github.com/sraj/addressbook/internal/billing/domain"
	"github.com/stripe/stripe-go/v79"
	"github.com/stripe/stripe-go/v79/price"
)

// ListPlans returns all available plans.
func (s *Service) ListPlans(ctx context.Context) ([]domain.Plan, error) {
	return s.repo.GetPlans(ctx)
}

// UpdatePlanPriceID sets the Stripe price ID for a plan.
func (s *Service) UpdatePlanPriceID(ctx context.Context, planID uint, priceID string) error {
	return s.repo.SetPlanPriceID(ctx, planID, priceID)
}

// SyncPricesFromStripe fetches prices from Stripe for all plans that have
// a stripe_price_id configured and updates their price_monthly.
func (s *Service) SyncPricesFromStripe(ctx context.Context) error {
	if s.stripeSecretKey == "" {
		return fmt.Errorf("stripe not configured")
	}

	stripe.Key = s.stripeSecretKey
	plans, err := s.repo.GetPlans(ctx)
	if err != nil {
		return err
	}

	for _, plan := range plans {
		if plan.StripePriceID == "" {
			continue
		}
		p, err := price.Get(plan.StripePriceID, nil)
		if err != nil {
			return fmt.Errorf("fetch price %s: %w", plan.StripePriceID, err)
		}
		if err := s.repo.SetPlanPriceID(ctx, plan.ID, plan.StripePriceID); err != nil {
			return err
		}
		if err := s.repo.UpdatePlanPrice(ctx, plan.ID, p.UnitAmount); err != nil {
			return err
		}
	}
	return nil
}

// GetProPriceID returns the Stripe price ID for the Pro plan.
func (s *Service) GetProPriceID(ctx context.Context) (string, error) {
	plan, err := s.repo.GetPlanByName(ctx, "pro")
	if err != nil {
		return "", err
	}
	return plan.StripePriceID, nil
}

// GetPlanByName looks up a plan by name (e.g. "free", "pro").
func (s *Service) GetPlanByName(ctx context.Context, name string) (*domain.Plan, error) {
	return s.repo.GetPlanByName(ctx, name)
}

// GetPlan returns a plan by its internal ID.
func (s *Service) GetPlan(ctx context.Context, planID uint) (*domain.Plan, error) {
	return s.repo.GetPlan(ctx, planID)
}

// GetPlans lists all plans in ascending price order.
func (s *Service) GetPlans(ctx context.Context) ([]domain.Plan, error) {
	return s.repo.GetPlans(ctx)
}
