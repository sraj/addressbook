package infrastructure

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/sraj/addressbook/internal/billing/domain"
	"github.com/mobentum/xdb"
	"github.com/stripe/stripe-go/v79"
	"github.com/stripe/stripe-go/v79/price"
)

var validResources = map[string]bool{"contacts": true, "notes": true, "bookmarks": true}

func Seed(ctx context.Context, db *xdb.DB, stripeSecretKey string) error {
	r := &sqliteRepo{db: db}
	if err := r.SeedPlans(ctx); err != nil {
		return err
	}
	if stripeSecretKey != "" {
		if err := r.syncPricesFromStripe(ctx, stripeSecretKey); err != nil {
			return err
		}
	}
	return r.seedExistingAccounts(ctx)
}

var accountOnce sync.Once

func (r *sqliteRepo) seedExistingAccounts(ctx context.Context) error {
	var ret error
	accountOnce.Do(func() {
		ret = r.seedExistingAccountsOnce(ctx)
	})
	return ret
}

func (r *sqliteRepo) syncPricesFromStripe(ctx context.Context, stripeSecretKey string) error {
	stripe.Key = stripeSecretKey

	type planRow struct {
		ID            uint   `db:"id"`
		StripePriceID string `db:"stripe_price_id"`
	}
	var rows []planRow
	err := r.db.Select("id", "stripe_price_id").
		From("plans").Where(xdb.Cond.Raw("stripe_price_id != ''")).All(ctx, &rows)
	if err != nil {
		return err
	}

	for _, row := range rows {
		p, err := price.Get(row.StripePriceID, nil)
		if err != nil {
			return err
		}
		if _, err := r.db.Update("plans").Set("price_monthly", p.UnitAmount).
			Where(xdb.Cond.Eq("id", row.ID)).Exec(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (r *sqliteRepo) seedExistingAccountsOnce(ctx context.Context) error {
	type userRow struct {
		ID    uint   `db:"id"`
		Email string `db:"email"`
	}
	var plan planModel
	err := r.db.Select("id").From("plans").Where(xdb.Cond.Eq("name", "free")).One(ctx, &plan)
	if err != nil {
		return err
	}

	var users []userRow
	err = r.db.Select("id", "email").From("users").OrderBy("id", xdb.ASC).All(ctx, &users)
	if err != nil {
		return err
	}

	for _, u := range users {
		exists, err := r.db.Select("1").From("account_members").
			Where(xdb.Cond.Eq("user_id", u.ID)).Exists(ctx)
		if err != nil {
			return err
		}
		if exists {
			continue
		}

		var accountID uint
		err = r.db.Insert("accounts").Columns("name", "plan_id", "usage").
			Values(u.Email, plan.ID, "{}").Returning("id").One(ctx, &accountID)
		if err != nil {
			return err
		}

		_, err = r.db.Insert("account_members").Columns("account_id", "user_id", "role").
			Values(accountID, u.ID, "owner").Exec(ctx)
		if err != nil {
			return err
		}

		_, err = r.db.Insert("subscriptions").Columns("account_id", "plan_id", "status").
			Values(accountID, plan.ID, "active").Exec(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

type planModel struct {
	ID            uint   `db:"id"`
	Name          string `db:"name"`
	Limits        string `db:"limits"`
	PriceMonthly  int    `db:"price_monthly"`
	StripePriceID string `db:"stripe_price_id"`
}

type accountModel struct {
	ID               uint   `db:"id"`
	Name             string `db:"name"`
	PlanID           uint   `db:"plan_id"`
	Usage            string `db:"usage"`
	StripeCustomerID string `db:"stripe_customer_id"`
	CreatedAt        string `db:"created_at"`
}

type memberModel struct {
	ID        uint   `db:"id"`
	AccountID uint   `db:"account_id"`
	UserID    uint   `db:"user_id"`
	Role      string `db:"role"`
}

type subscriptionModel struct {
	ID                   uint    `db:"id"`
	AccountID            uint    `db:"account_id"`
	PlanID               uint    `db:"plan_id"`
	StripeSubscriptionID string  `db:"stripe_subscription_id"`
	Status               string  `db:"status"`
	CurrentPeriodStart   *string `db:"current_period_start"`
	CurrentPeriodEnd     *string `db:"current_period_end"`
	CanceledAt           *string `db:"canceled_at"`
}

type sqliteRepo struct {
	db *xdb.DB
}

func NewSQLiteRepo(db *xdb.DB) domain.Repository {
	return &sqliteRepo{db: db}
}

func (r *sqliteRepo) SeedPlans(ctx context.Context) error {
	defaultPlans := []struct {
		name, limits string
		price        int
	}{
		{"free", `{"contacts":50,"notes":50,"bookmarks":100}`, 0},
		{"pro", `{"contacts":-1,"notes":-1,"bookmarks":-1}`, 0},
		{"business", `{"contacts":-1,"notes":-1,"bookmarks":-1}`, 0},
	}

	for _, p := range defaultPlans {
		exists, err := r.db.Select("1").From("plans").Where(xdb.Cond.Eq("name", p.name)).Exists(ctx)
		if err != nil {
			return err
		}
		if exists {
			continue
		}
		if _, err := r.db.Insert("plans").Columns("name", "limits", "price_monthly").
			Values(p.name, p.limits, p.price).Exec(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (r *sqliteRepo) GetPlan(ctx context.Context, planID uint) (*domain.Plan, error) {
	var m planModel
	err := r.db.Select("id", "name", "limits", "price_monthly", "stripe_price_id").
		From("plans").Where(xdb.Cond.Eq("id", planID)).One(ctx, &m)
	if err != nil {
		if errors.Is(err, xdb.ErrNotFound) {
			return nil, domain.ErrPlanNotFound
		}
		return nil, err
	}
	return toPlan(&m), nil
}

func (r *sqliteRepo) GetPlanByName(ctx context.Context, name string) (*domain.Plan, error) {
	var m planModel
	err := r.db.Select("id", "name", "limits", "price_monthly", "stripe_price_id").
		From("plans").Where(xdb.Cond.Eq("name", name)).One(ctx, &m)
	if err != nil {
		if errors.Is(err, xdb.ErrNotFound) {
			return nil, domain.ErrPlanNotFound
		}
		return nil, err
	}
	return toPlan(&m), nil
}

func (r *sqliteRepo) GetPlans(ctx context.Context) ([]domain.Plan, error) {
	var rows []planModel
	err := r.db.Select("id", "name", "limits", "price_monthly", "stripe_price_id").
		From("plans").OrderBy("price_monthly", xdb.ASC).All(ctx, &rows)
	if err != nil {
		return nil, err
	}
	plans := make([]domain.Plan, len(rows))
	for i := range rows {
		plans[i] = *toPlan(&rows[i])
	}
	return plans, nil
}

func (r *sqliteRepo) CreateAccount(ctx context.Context, account *domain.Account) error {
	return r.db.Insert("accounts").Columns("name", "plan_id", "usage").
		Values(account.Name, account.PlanID, account.UsageJSON).
		Returning("id").One(ctx, &account.ID)
}

func (r *sqliteRepo) GetUserRole(ctx context.Context, userID uint) (string, error) {
	var role string
	err := r.db.Select("role").From("users").Where(xdb.Cond.Eq("id", userID)).One(ctx, &role)
	if err != nil {
		return "", err
	}
	return role, nil
}

func (r *sqliteRepo) GetAccountByUser(ctx context.Context, userID uint) (*domain.Account, error) {
	var a accountModel
	err := r.db.Select("a.id", "a.name", "a.plan_id", "a.usage", "a.stripe_customer_id").
		From("accounts a").
		Join("account_members m ON m.account_id = a.id").
		Where(xdb.Cond.Eq("m.user_id", userID)).
		One(ctx, &a)
	if err != nil {
		if errors.Is(err, xdb.ErrNotFound) {
			return nil, domain.ErrAccountNotFound
		}
		return nil, err
	}
	return toAccount(&a), nil
}

func (r *sqliteRepo) GetAccountByCustomerID(ctx context.Context, customerID string) (*domain.Account, error) {
	var a accountModel
	err := r.db.Select("id", "name", "plan_id", "usage", "stripe_customer_id").
		From("accounts").Where(xdb.Cond.Eq("stripe_customer_id", customerID)).One(ctx, &a)
	if err != nil {
		if errors.Is(err, xdb.ErrNotFound) {
			return nil, domain.ErrAccountNotFound
		}
		return nil, err
	}
	return toAccount(&a), nil
}

func (r *sqliteRepo) GetAccount(ctx context.Context, accountID uint) (*domain.Account, error) {
	var a accountModel
	err := r.db.Select("id", "name", "plan_id", "usage", "stripe_customer_id").
		From("accounts").Where(xdb.Cond.Eq("id", accountID)).One(ctx, &a)
	if err != nil {
		if errors.Is(err, xdb.ErrNotFound) {
			return nil, domain.ErrAccountNotFound
		}
		return nil, err
	}
	return toAccount(&a), nil
}

func (r *sqliteRepo) SetStripeCustomerID(ctx context.Context, accountID uint, customerID string) error {
	_, err := r.db.Update("accounts").Set("stripe_customer_id", customerID).Where(xdb.Cond.Eq("id", accountID)).Exec(ctx)
	return err
}

func (r *sqliteRepo) SetPlanPriceID(ctx context.Context, planID uint, priceID string) error {
	_, err := r.db.Update("plans").Set("stripe_price_id", priceID).Where(xdb.Cond.Eq("id", planID)).Exec(ctx)
	return err
}

func (r *sqliteRepo) UpdatePlanPrice(ctx context.Context, planID uint, priceMonthly int64) error {
	_, err := r.db.Update("plans").Set("price_monthly", priceMonthly).Where(xdb.Cond.Eq("id", planID)).Exec(ctx)
	return err
}

func (r *sqliteRepo) SetAccountPlan(ctx context.Context, accountID, planID uint) error {
	_, err := r.db.Update("accounts").Set("plan_id", planID).
		Set("updated_at", time.Now().Format("2006-01-02 15:04:05")).
		Where(xdb.Cond.Eq("id", accountID)).Exec(ctx)
	return err
}

func (r *sqliteRepo) CreateMember(ctx context.Context, member *domain.AccountMember) error {
	return r.db.Insert("account_members").Columns("account_id", "user_id", "role").
		Values(member.AccountID, member.UserID, member.Role).Returning("id").One(ctx, &member.ID)
}

func (r *sqliteRepo) CreateSubscription(ctx context.Context, sub *domain.Subscription) error {
	_, _ = r.db.Delete("subscriptions").Where(xdb.Cond.Eq("account_id", sub.AccountID)).Exec(ctx)
	return r.db.Insert("subscriptions").Columns("account_id", "plan_id", "stripe_subscription_id", "status").
		Values(sub.AccountID, sub.PlanID, sub.StripeSubscriptionID, sub.Status).Returning("id").One(ctx, &sub.ID)
}

func (r *sqliteRepo) GetSubscriptionByStripeID(ctx context.Context, stripeID string) (*domain.Subscription, error) {
	var s subscriptionModel
	err := r.db.Select("id", "account_id", "plan_id", "stripe_subscription_id", "status", "current_period_start", "current_period_end", "canceled_at").
		From("subscriptions").Where(xdb.Cond.Eq("stripe_subscription_id", stripeID)).OrderBy("created_at", xdb.DESC).One(ctx, &s)
	if err != nil {
		if errors.Is(err, xdb.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toSubscription(&s), nil
}

func (r *sqliteRepo) GetSubscription(ctx context.Context, accountID uint) (*domain.Subscription, error) {
	var s subscriptionModel
	err := r.db.Select("id", "account_id", "plan_id", "stripe_subscription_id", "status", "current_period_start", "current_period_end", "canceled_at").
		From("subscriptions").Where(xdb.Cond.Eq("account_id", accountID)).One(ctx, &s)
	if err != nil {
		if errors.Is(err, xdb.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toSubscription(&s), nil
}

func (r *sqliteRepo) GetUsage(ctx context.Context, accountID uint) (map[string]int, error) {
	var a accountModel
	err := r.db.Select("usage").From("accounts").Where(xdb.Cond.Eq("id", accountID)).One(ctx, &a)
	if err != nil {
		return nil, err
	}
	var m map[string]int
	if err := json.Unmarshal([]byte(a.Usage), &m); err != nil {
		return map[string]int{}, nil
	}
	return m, nil
}

func (r *sqliteRepo) IncrementUsage(ctx context.Context, accountID uint, resource string) error {
	if !validResources[resource] {
		return errors.New("invalid resource: " + resource)
	}
	m, err := r.GetUsage(ctx, accountID)
	if err != nil {
		return err
	}
	m[resource]++
	b, _ := json.Marshal(m)
	_, err = r.db.Update("accounts").Set("usage", string(b)).Where(xdb.Cond.Eq("id", accountID)).Exec(ctx)
	return err
}

func (r *sqliteRepo) DecrementUsage(ctx context.Context, accountID uint, resource string) error {
	if !validResources[resource] {
		return errors.New("invalid resource: " + resource)
	}
	m, err := r.GetUsage(ctx, accountID)
	if err != nil {
		return err
	}
	if m[resource] > 0 {
		m[resource]--
	}
	b, _ := json.Marshal(m)
	_, err = r.db.Update("accounts").Set("usage", string(b)).Where(xdb.Cond.Eq("id", accountID)).Exec(ctx)
	return err
}

func toPlan(m *planModel) *domain.Plan {
	return &domain.Plan{
		ID:            m.ID,
		Name:          m.Name,
		LimitsJSON:    m.Limits,
		PriceMonthly:  m.PriceMonthly,
		StripePriceID: m.StripePriceID,
	}
}

func toAccount(m *accountModel) *domain.Account {
	return &domain.Account{
		ID:               m.ID,
		Name:             m.Name,
		PlanID:           m.PlanID,
		UsageJSON:        m.Usage,
		StripeCustomerID: m.StripeCustomerID,
	}
}

func toSubscription(m *subscriptionModel) *domain.Subscription {
	s := &domain.Subscription{
		ID:                   m.ID,
		AccountID:            m.AccountID,
		PlanID:               m.PlanID,
		StripeSubscriptionID: m.StripeSubscriptionID,
		Status:               m.Status,
	}
	if m.CurrentPeriodStart != nil {
		s.CurrentPeriodStart = *m.CurrentPeriodStart
	}
	if m.CurrentPeriodEnd != nil {
		s.CurrentPeriodEnd = *m.CurrentPeriodEnd
	}
	if m.CanceledAt != nil {
		s.CanceledAt = *m.CanceledAt
	}
	return s
}
