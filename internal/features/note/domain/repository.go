package domain

import "context"

type Repository interface {
	Create(ctx context.Context, note *Note) error
	FindByID(ctx context.Context, id, userID uint) (*Note, error)
	Update(ctx context.Context, note *Note) error
	Delete(ctx context.Context, id, userID uint) error
	ListByUser(ctx context.Context, userID uint, page, size int) ([]Note, int64, error)
}
