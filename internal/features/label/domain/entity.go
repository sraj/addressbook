package domain

import (
	"context"
	"errors"
)

var ErrOrderNotFound = errors.New("label order not found")

const (
	StatusPending  = "pending"
	StatusPaid     = "paid"
	StatusCanceled = "canceled"
)

type LabelOrder struct {
	ID              uint   `json:"id"`
	UserID          uint   `json:"user_id"`
	CollectionID    uint   `json:"collection_id,omitempty"`
	ContactCount    int    `json:"contact_count"`
	SheetCount      int    `json:"sheet_count"`
	AmountCents     int64  `json:"amount_cents"`
	Currency        string `json:"currency"`
	Status          string `json:"status"`
	LabelType       string `json:"label_type,omitempty"`
	StripeSessionID string `json:"stripe_session_id,omitempty"`
	CreatedAt       string `json:"created_at,omitempty"`
}

type Repository interface {
	Create(ctx context.Context, o *LabelOrder) error
	FindBySessionID(ctx context.Context, sessionID string) (*LabelOrder, error)
	ListByUser(ctx context.Context, userID uint) ([]LabelOrder, error)
	SetSessionID(ctx context.Context, orderID uint, sessionID string) error
	UpdateStatusBySessionID(ctx context.Context, sessionID, status string) error
}
