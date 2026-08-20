package domain

import "time"

type User struct {
	ID           uint   `json:"id"`
	Email        string `json:"email"`
	Name         string `json:"name"`
	Preferences  string `json:"-"`
	PasswordHash string `json:"-"`
	Role         string `json:"role"`
	Status       string `json:"status"`
}

type Preferences struct {
	DefaultPage     string `json:"default_page"`
	PageSize        int    `json:"page_size"`
	Compact         bool   `json:"compact"`
	NotifyBilling   bool   `json:"notify_billing"`
	NotifyMarketing bool   `json:"notify_marketing"`
}

type PasswordResetToken struct {
	ID        uint      `json:"id"`
	UserID    uint      `json:"user_id"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	Used      bool      `json:"used"`
}
