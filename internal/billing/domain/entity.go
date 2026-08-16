package domain

import "encoding/json"

type Plan struct {
	ID            uint   `json:"id"`
	Name          string `json:"name"`
	LimitsJSON    string `json:"-"`
	PriceMonthly  int    `json:"price_monthly"`
	StripePriceID string `json:"stripe_price_id"`
}

func (p *Plan) Limits() map[string]int {
	var m map[string]int
	_ = json.Unmarshal([]byte(p.LimitsJSON), &m)
	return m
}

func (p *Plan) Quota(resource string) int {
	return p.Limits()[resource]
}

type Account struct {
	ID               uint   `json:"id"`
	Name             string `json:"name"`
	PlanID           uint   `json:"plan_id"`
	UsageJSON        string `json:"-"`
	StripeCustomerID string `json:"stripe_customer_id,omitempty"`
}

func (a *Account) Usage() map[string]int {
	var m map[string]int
	_ = json.Unmarshal([]byte(a.UsageJSON), &m)
	return m
}

type AccountMember struct {
	ID        uint   `json:"id"`
	AccountID uint   `json:"account_id"`
	UserID    uint   `json:"user_id"`
	Role      string `json:"role"`
}

type Subscription struct {
	ID                   uint   `json:"id"`
	AccountID            uint   `json:"account_id"`
	PlanID               uint   `json:"plan_id"`
	StripeSubscriptionID string `json:"stripe_subscription_id,omitempty"`
	Status               string `json:"status"`
	CurrentPeriodStart   string `json:"current_period_start,omitempty"`
	CurrentPeriodEnd     string `json:"current_period_end,omitempty"`
	CanceledAt           string `json:"canceled_at,omitempty"`
}
