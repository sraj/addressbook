package interfaces

import (
	"context"

	"github.com/sraj/addressbook/internal/features/bookmark/domain"
)

// bookmarkService is the consumer-facing port the handler depends on. It is
// satisfied by *application.Service.
type bookmarkService interface {
	List(ctx context.Context, userID uint, category string, page, size int) ([]domain.Bookmark, int64, error)
	GetByID(ctx context.Context, id, userID uint) (*domain.Bookmark, error)
	Create(ctx context.Context, userID uint, url, title, description, faviconURL, category string) (*domain.Bookmark, error)
	Update(ctx context.Context, id, userID uint, url, title, description, faviconURL, category string) (*domain.Bookmark, error)
	Delete(ctx context.Context, id, userID uint) error
	Import(ctx context.Context, userID uint, items []domain.Bookmark) (imported, skipped int, err error)
	ListCategories(ctx context.Context, userID uint) ([]string, error)
}
