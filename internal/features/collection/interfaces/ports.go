package interfaces

import (
	"context"

	collectionDomain "github.com/sraj/addressbook/internal/features/collection/domain"
	contactDomain "github.com/sraj/addressbook/internal/features/contact/domain"
)

// collectionService is the consumer-facing port the handler depends on.
type collectionService interface {
	Create(ctx context.Context, userID uint, name string) (*collectionDomain.Collection, error)
	List(ctx context.Context, userID uint) ([]collectionDomain.Collection, error)
	Get(ctx context.Context, id, userID uint) (*collectionDomain.Collection, error)
	Rename(ctx context.Context, id, userID uint, name string) (*collectionDomain.Collection, error)
	Delete(ctx context.Context, id, userID uint) error
	RegenerateToken(ctx context.Context, id, userID uint) (*collectionDomain.Collection, error)
	GetByToken(ctx context.Context, token string) (*collectionDomain.Collection, error)
	SubmitAddress(ctx context.Context, token, name, email, phone string, address contactDomain.Address) (*contactDomain.Contact, error)
	ListContacts(ctx context.Context, userID, collectionID uint, page, size int) ([]contactDomain.Contact, int64, error)
	AddContact(ctx context.Context, userID, collectionID uint, name string, emails, phones []string, addresses []contactDomain.Address, notes string) (*contactDomain.Contact, error)
	MoveContact(ctx context.Context, userID, collectionID, contactID uint) error
	RemoveContact(ctx context.Context, userID, collectionID, contactID uint) error
}
