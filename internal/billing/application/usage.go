package application

import (
	"context"
	"fmt"

	"github.com/sraj/addressbook/internal/billing/domain"
)

// GetUserRole returns the role for a user (used for admin quota exemptions).
func (s *Service) GetUserRole(ctx context.Context, userID uint) (string, error) {
	return s.repo.GetUserRole(ctx, userID)
}

// CheckQuota verifies that the user hasn't exceeded their plan's limit
// for the given resource (e.g. "widgets").
// Returns domain.ErrQuotaExceeded if the limit has been reached.
func (s *Service) CheckQuota(ctx context.Context, userID uint, resource string) error {
	// Admins have unlimited access
	role, err := s.repo.GetUserRole(ctx, userID)
	if err == nil && role == "admin" {
		return nil
	}

	account, err := s.repo.GetAccountByUser(ctx, userID)
	if err != nil {
		return err
	}

	plan, err := s.repo.GetPlan(ctx, account.PlanID)
	if err != nil {
		return err
	}

	quota := plan.Quota(resource)
	if quota == -1 {
		return nil
	}

	usage, err := s.repo.GetUsage(ctx, account.ID)
	if err != nil {
		return err
	}

	if usage[resource] >= quota {
		return fmt.Errorf("%w: %s limit (%d) reached", domain.ErrQuotaExceeded, resource, quota)
	}

	return nil
}

// IncrementUsage increases the resource usage counter for the user's account.
func (s *Service) IncrementUsage(ctx context.Context, userID uint, resource string) error {
	account, err := s.repo.GetAccountByUser(ctx, userID)
	if err != nil {
		return err
	}
	return s.repo.IncrementUsage(ctx, account.ID, resource)
}

// DecrementUsage decreases the resource usage counter (e.g. on delete).
func (s *Service) DecrementUsage(ctx context.Context, userID uint, resource string) error {
	account, err := s.repo.GetAccountByUser(ctx, userID)
	if err != nil {
		return err
	}
	return s.repo.DecrementUsage(ctx, account.ID, resource)
}

// RemainingQuota returns the number of resources the user can still create
// for the given resource, or -1 if unlimited. Admins are always unlimited.
func (s *Service) RemainingQuota(ctx context.Context, userID uint, resource string) (int, error) {
	role, err := s.repo.GetUserRole(ctx, userID)
	if err == nil && role == "admin" {
		return -1, nil
	}

	account, err := s.repo.GetAccountByUser(ctx, userID)
	if err != nil {
		return 0, err
	}

	plan, err := s.repo.GetPlan(ctx, account.PlanID)
	if err != nil {
		return 0, err
	}

	quota := plan.Quota(resource)
	if quota == -1 {
		return -1, nil
	}

	usage, err := s.repo.GetUsage(ctx, account.ID)
	if err != nil {
		return 0, err
	}

	remaining := quota - usage[resource]
	if remaining < 0 {
		remaining = 0
	}
	return remaining, nil
}

// GetUsage returns the user's current resource usage and their plan details.
func (s *Service) GetUsage(ctx context.Context, userID uint) (map[string]int, *domain.Plan, error) {
	account, err := s.repo.GetAccountByUser(ctx, userID)
	if err != nil {
		return nil, nil, err
	}

	plan, err := s.repo.GetPlan(ctx, account.PlanID)
	if err != nil {
		return nil, nil, err
	}

	usage, err := s.repo.GetUsage(ctx, account.ID)
	if err != nil {
		return nil, nil, err
	}

	return usage, plan, nil
}
