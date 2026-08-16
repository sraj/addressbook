package interfaces

import (
	"context"

	"github.com/sraj/addressbook/internal/auth/domain"
)

// authService is the consumer-facing port the handler depends on. It is
// satisfied by *application.Service.
type authService interface {
	Register(ctx context.Context, email, password string) (*domain.User, error)
	Login(ctx context.Context, email, password string) (*domain.User, error)
	GenerateToken(userID uint) (string, error)
	UserByID(ctx context.Context, id uint) (*domain.User, error)
	ForgotPassword(ctx context.Context, email string) (string, error)
	ResetPassword(ctx context.Context, token, newPassword string) error
	UpdateProfile(ctx context.Context, userID uint, name string, prefs *domain.Preferences) (*domain.User, error)
	ChangePassword(ctx context.Context, userID uint, currentPassword, newPassword string) error
}

// billingIniter lets the auth context create a billing account on registration.
// It is satisfied by *billing/application.Service.
type billingIniter interface {
	InitAccount(ctx context.Context, userID uint, accountName, email string) error
}

// mailSender is satisfied by *mailer.Mailer (nil when email is not configured).
type mailSender interface {
	Send(to, subject, body string) error
}
