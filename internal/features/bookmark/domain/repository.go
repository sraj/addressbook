package domain

import "context"

type Repository interface {
	Create(ctx context.Context, bookmark *Bookmark) error
	FindByID(ctx context.Context, id, userID uint) (*Bookmark, error)
	Update(ctx context.Context, bookmark *Bookmark) error
	Delete(ctx context.Context, id, userID uint) error
	ExistsByURL(ctx context.Context, userID uint, url string) (bool, error)
	ListByUser(ctx context.Context, userID uint, page, size int) ([]Bookmark, int64, error)
	ListByCategory(ctx context.Context, userID uint, category string, page, size int) ([]Bookmark, int64, error)
	ListCategories(ctx context.Context, userID uint) ([]string, error)
}
