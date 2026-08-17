package interfaces

import (
	"context"

	"github.com/sraj/addressbook/internal/features/contact/application"
	"github.com/sraj/addressbook/internal/features/contact/domain"
)

// contactService is the consumer-facing port the handler depends on. It is
// satisfied by *application.Service.
type contactService interface {
	List(ctx context.Context, userID uint, query string, page, size int) ([]domain.Contact, int64, error)
	ListByCollection(ctx context.Context, userID, collectionID uint, page, size int) ([]domain.Contact, int64, error)
	ListAllByUser(ctx context.Context, userID, collectionID uint) ([]domain.Contact, error)
	Export(ctx context.Context, userID, collectionID uint, format string) ([]byte, string, error)
	GetByID(ctx context.Context, id, userID uint) (*domain.Contact, error)
	Create(ctx context.Context, userID, collectionID uint, name string, emails, phones []string, addresses []domain.Address, notes string) (*domain.Contact, error)
	Update(ctx context.Context, id, userID uint, name string, emails, phones []string, addresses []domain.Address, notes string) (*domain.Contact, error)
	Delete(ctx context.Context, id, userID uint) error
	Import(ctx context.Context, userID, collectionID uint, records []application.ImportRecord) (application.ImportResult, error)
}
