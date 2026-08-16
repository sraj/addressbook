package interfaces

import (
	"context"

	"github.com/sraj/addressbook/internal/features/note/domain"
)

// noteService is the consumer-facing port the handler depends on. It is
// satisfied by *application.Service.
type noteService interface {
	List(ctx context.Context, userID uint, page, size int) ([]domain.Note, int64, error)
	GetByID(ctx context.Context, id, userID uint) (*domain.Note, error)
	Create(ctx context.Context, userID uint, title, content string) (*domain.Note, error)
	Update(ctx context.Context, id, userID uint, title, content string) (*domain.Note, error)
	Delete(ctx context.Context, id, userID uint) error
}
