package domain

import "context"

type Repository interface {
	Create(ctx context.Context, contact *Contact) error
	FindByID(ctx context.Context, id, userID uint) (*Contact, error)
	Update(ctx context.Context, contact *Contact) error
	Delete(ctx context.Context, id, userID uint) error
	SetCollection(ctx context.Context, id, userID, collectionID uint) error
	ListByUser(ctx context.Context, userID uint, page, size int) ([]Contact, int64, error)
	ListByCollection(ctx context.Context, userID, collectionID uint, page, size int) ([]Contact, int64, error)
	ListAllByUser(ctx context.Context, userID, collectionID uint) ([]Contact, error)
	Search(ctx context.Context, userID uint, query string, page, size int) ([]Contact, int64, error)
}
