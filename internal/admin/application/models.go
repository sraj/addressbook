package application

type UserRow struct {
	ID                 uint   `json:"id"`
	Email              string `json:"email"`
	Role               string `json:"role"`
	Status             string `json:"status"`
	CreatedAt          string `json:"created_at"`
	PlanName           string `json:"plan_name"`
	SubscriptionStatus string `json:"subscription_status"`
	SubscriptionEnd    string `json:"subscription_end"`
}

type StatsResponse struct {
	TotalUsers    int `json:"total_users"`
	ActiveToday   int `json:"active_today"`
	FreeAccounts  int `json:"free_accounts"`
	ProAccounts   int `json:"pro_accounts"`
}
