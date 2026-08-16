package interfaces

import (
	"context"

	"github.com/sraj/addressbook/internal/features/contact/domain"
)

// contactService is the consumer-facing port the handler depends on. It is
// satisfied by *application.Service.
type contactService interface {
	List(ctx context.Context, userID uint, query string, page, size int) ([]domain.Contact, int64, error)
	GetByID(ctx context.Context, id, userID uint) (*domain.Contact, error)
	Create(ctx context.Context, userID uint, name string, emails, phones []string, addresses []domain.Address, notes string) (*domain.Contact, error)
	Update(ctx context.Context, id, userID uint, name string, emails, phones []string, addresses []domain.Address, notes string) (*domain.Contact, error)
	Delete(ctx context.Context, id, userID uint) error
}
