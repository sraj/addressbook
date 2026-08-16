package application

import (
	"context"
	"fmt"

	"github.com/mobentum/xdb"
)

// billingEnsurer lets the admin service create a billing account when a user
// is activated. It is satisfied by *billing/application.Service.
type billingEnsurer interface {
	EnsureAccount(ctx context.Context, userID uint, email string) error
}

type Service struct {
	db      *xdb.DB
	billing billingEnsurer
}

func NewService(db *xdb.DB, billing billingEnsurer) *Service {
	return &Service{db: db, billing: billing}
}

type statsModel struct {
	Count int `db:"count"`
}

func (s *Service) IsAdmin(ctx context.Context, userID uint) (bool, error) {
	var role string
	err := s.db.Select("role").From("users").Where(xdb.Cond.Eq("id", userID)).One(ctx, &role)
	if err != nil {
		return false, err
	}
	return role == "admin", nil
}

func (s *Service) ListUsers(ctx context.Context) ([]UserRow, error) {
	type userPlanModel struct {
		ID                 uint    `db:"id"`
		Email              string  `db:"email"`
		Role               string  `db:"role"`
		Status             string  `db:"status"`
		CreatedAt          *string `db:"created_at"`
		PlanName           *string `db:"plan_name"`
		SubscriptionStatus *string `db:"subscription_status"`
		SubscriptionEnd    *string `db:"subscription_end"`
	}

	var rows []userPlanModel
	err := s.db.Select(
		"u.id", "u.email", "u.role", "u.status", "u.created_at",
		"p.name as plan_name",
		"s.status as subscription_status",
		"s.current_period_end as subscription_end",
	).
		From("users u").
		LeftJoin("account_members m ON m.user_id = u.id").
		LeftJoin("accounts a ON a.id = m.account_id").
		LeftJoin("plans p ON p.id = a.plan_id").
		LeftJoin("(SELECT account_id, status, current_period_end FROM subscriptions WHERE id IN (SELECT MAX(id) FROM subscriptions GROUP BY account_id)) s ON s.account_id = a.id").
		OrderBy("u.created_at", xdb.DESC).
		All(ctx, &rows)
	if err != nil {
		return nil, err
	}

	users := make([]UserRow, len(rows))
	for i := range rows {
		createdAt := ""
		if rows[i].CreatedAt != nil {
			createdAt = *rows[i].CreatedAt
		}
		planName := ""
		if rows[i].PlanName != nil {
			planName = *rows[i].PlanName
		}
		subStatus := ""
		if rows[i].SubscriptionStatus != nil {
			subStatus = *rows[i].SubscriptionStatus
		}
		subEnd := ""
		if rows[i].SubscriptionEnd != nil {
			subEnd = *rows[i].SubscriptionEnd
		}
		users[i] = UserRow{
			ID:                 rows[i].ID,
			Email:              rows[i].Email,
			Role:               rows[i].Role,
			Status:             rows[i].Status,
			CreatedAt:          createdAt,
			PlanName:           planName,
			SubscriptionStatus: subStatus,
			SubscriptionEnd:    subEnd,
		}
	}
	return users, nil
}

func (s *Service) UpdateUserStatus(ctx context.Context, userID uint, status string) error {
	// When activating a user, ensure they have an account + free subscription
	if status == "active" {
		var email string
		if err := s.db.Select("email").From("users").Where(xdb.Cond.Eq("id", userID)).One(ctx, &email); err != nil {
			return err
		}
		if err := s.billing.EnsureAccount(ctx, userID, email); err != nil {
			return err
		}
	}
	_, err := s.db.Update("users").Set("status", status).Where(xdb.Cond.Eq("id", userID)).Exec(ctx)
	return err
}

func (s *Service) Stats(ctx context.Context) (*StatsResponse, error) {
	resp := &StatsResponse{}
	var row statsModel

	if err := s.db.Select("COUNT(*)").From("users").One(ctx, &row); err != nil {
		return nil, fmt.Errorf("total users: %w", err)
	}
	resp.TotalUsers = row.Count

	// Count free accounts (accounts on the free plan)
	if err := s.db.Select("COUNT(*)").
		From("accounts a").
		Where(xdb.Cond.Raw("a.plan_id = (SELECT id FROM plans WHERE name = 'free')")).
		One(ctx, &row); err != nil {
		return nil, err
	}
	resp.FreeAccounts = row.Count

	// Count pro accounts (accounts on the pro plan)
	if err := s.db.Select("COUNT(*)").
		From("accounts a").
		Where(xdb.Cond.Raw("a.plan_id = (SELECT id FROM plans WHERE name = 'pro')")).
		One(ctx, &row); err != nil {
		return nil, err
	}
	resp.ProAccounts = row.Count

	return resp, nil
}
