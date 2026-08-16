package domain

import "context"

type Repository interface {
	Create(ctx context.Context, user *User) error
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByID(ctx context.Context, id uint) (*User, error)
	Exists(ctx context.Context, email string) (bool, error)
	UpdateName(ctx context.Context, userID uint, name string) error
	UpdatePreferences(ctx context.Context, userID uint, preferences string) error
	UpdatePassword(ctx context.Context, userID uint, passwordHash string) error
	CreateResetToken(ctx context.Context, token *PasswordResetToken) error
	FindResetToken(ctx context.Context, token string) (*PasswordResetToken, error)
	MarkResetTokenUsed(ctx context.Context, tokenID uint) error
}
