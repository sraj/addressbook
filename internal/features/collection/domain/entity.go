package domain

import (
	"context"
	"errors"
)

var (
	ErrCollectionNotFound = errors.New("collection not found")
	ErrCollectionName     = errors.New("invalid collection name")
	ErrInvalidToken       = errors.New("invalid invite token")
)

type Collection struct {
	ID           uint   `json:"id"`
	UserID       uint   `json:"user_id"`
	Name         string `json:"name"`
	InviteToken  string `json:"invite_token"`
	ContactCount int64  `json:"contact_count,omitempty"`
	CreatedAt    string `json:"created_at,omitempty"`
	UpdatedAt    string `json:"updated_at,omitempty"`
}

type Repository interface {
	Create(ctx context.Context, c *Collection) error
	FindByID(ctx context.Context, id, userID uint) (*Collection, error)
	FindByToken(ctx context.Context, token string) (*Collection, error)
	ListByUser(ctx context.Context, userID uint) ([]Collection, error)
	Update(ctx context.Context, c *Collection) error
	Delete(ctx context.Context, id, userID uint) error
	RegenerateToken(ctx context.Context, id, userID uint, token string) error
	CountContacts(ctx context.Context, collectionID uint) (int64, error)
	UnlinkContacts(ctx context.Context, collectionID uint) error
}
