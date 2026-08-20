package infrastructure

import (
	"context"
	"errors"
	"time"

	"github.com/mobentum/xdb"
	"github.com/sraj/addressbook/internal/features/bookmark/domain"
)

type bookmarkModel struct {
	ID          uint   `db:"id"`
	UserID      uint   `db:"user_id"`
	URL         string `db:"url"`
	Title       string `db:"title"`
	Description string `db:"description"`
	FaviconURL  string `db:"favicon_url"`
	Category    string `db:"category"`
	CreatedAt   string `db:"created_at"`
	UpdatedAt   string `db:"updated_at"`
}

type sqliteRepo struct {
	db *xdb.DB
}

func NewSQLiteRepo(db *xdb.DB) domain.Repository {
	return &sqliteRepo{db: db}
}

func (r *sqliteRepo) Create(ctx context.Context, bookmark *domain.Bookmark) error {
	return r.db.Insert("bookmarks").Columns("user_id", "url", "title", "description", "favicon_url", "category").
		Values(bookmark.UserID, bookmark.URL, bookmark.Title, bookmark.Description, bookmark.FaviconURL, bookmark.Category).
		Returning("id").One(ctx, &bookmark.ID)
}

func (r *sqliteRepo) FindByID(ctx context.Context, id, userID uint) (*domain.Bookmark, error) {
	var m bookmarkModel
	err := r.db.Select("id", "user_id", "url", "title", "description", "favicon_url", "category", "created_at", "updated_at").
		From("bookmarks").
		Where(xdb.Cond.And(xdb.Cond.Eq("id", id), xdb.Cond.Eq("user_id", userID))).
		One(ctx, &m)
	if err != nil {
		if errors.Is(err, xdb.ErrNotFound) {
			return nil, domain.ErrBookmarkNotFound
		}
		return nil, err
	}
	return toDomain(&m), nil
}

func (r *sqliteRepo) Update(ctx context.Context, bookmark *domain.Bookmark) error {
	_, err := r.db.Update("bookmarks").
		Set("url", bookmark.URL).
		Set("title", bookmark.Title).
		Set("description", bookmark.Description).
		Set("favicon_url", bookmark.FaviconURL).
		Set("category", bookmark.Category).
		Set("updated_at", time.Now().Format("2006-01-02 15:04:05")).
		Where(xdb.Cond.Eq("id", bookmark.ID)).
		Exec(ctx)
	return err
}

func (r *sqliteRepo) Delete(ctx context.Context, id, userID uint) error {
	_, err := r.db.Delete("bookmarks").
		Where(xdb.Cond.And(xdb.Cond.Eq("id", id), xdb.Cond.Eq("user_id", userID))).
		Exec(ctx)
	return err
}

func (r *sqliteRepo) ListByUser(ctx context.Context, userID uint, page, size int) ([]domain.Bookmark, int64, error) {
	result, err := xdb.Paginate[bookmarkModel](ctx,
		r.db.Select("id", "user_id", "url", "title", "description", "favicon_url", "category", "created_at", "updated_at").
			From("bookmarks").
			Where(xdb.Cond.Eq("user_id", userID)).
			OrderBy("updated_at", xdb.DESC),
		xdb.Page{Number: page, Size: size},
	)
	if err != nil {
		return nil, 0, err
	}

	items := make([]domain.Bookmark, len(result.Items))
	for i := range result.Items {
		items[i] = *toDomain(&result.Items[i])
	}
	return items, int64(result.Total), nil
}

func (r *sqliteRepo) ListByCategory(ctx context.Context, userID uint, category string, page, size int) ([]domain.Bookmark, int64, error) {
	result, err := xdb.Paginate[bookmarkModel](ctx,
		r.db.Select("id", "user_id", "url", "title", "description", "favicon_url", "category", "created_at", "updated_at").
			From("bookmarks").
			Where(xdb.Cond.And(xdb.Cond.Eq("user_id", userID), xdb.Cond.Eq("category", category))).
			OrderBy("updated_at", xdb.DESC),
		xdb.Page{Number: page, Size: size},
	)
	if err != nil {
		return nil, 0, err
	}

	items := make([]domain.Bookmark, len(result.Items))
	for i := range result.Items {
		items[i] = *toDomain(&result.Items[i])
	}
	return items, int64(result.Total), nil
}

func (r *sqliteRepo) ListCategories(ctx context.Context, userID uint) ([]string, error) {
	type categoryRow struct {
		Category string `db:"category"`
	}
	var rows []categoryRow
	err := r.db.Select("DISTINCT category").From("bookmarks").
		Where(xdb.Cond.And(xdb.Cond.Eq("user_id", userID), xdb.Cond.Raw("category != ''"))).
		OrderBy("category", xdb.ASC).
		All(ctx, &rows)
	if err != nil {
		return nil, err
	}
	categories := make([]string, len(rows))
	for i := range rows {
		categories[i] = rows[i].Category
	}
	return categories, nil
}

func (r *sqliteRepo) BulkCreate(ctx context.Context, bookmarks []domain.Bookmark) error {
	for i := range bookmarks {
		b := &bookmarks[i]
		if err := r.Create(ctx, b); err != nil {
			return err
		}
	}
	return nil
}

func (r *sqliteRepo) ExistsByURL(ctx context.Context, userID uint, url string) (bool, error) {
	return r.db.Select("1").From("bookmarks").
		Where(xdb.Cond.And(xdb.Cond.Eq("user_id", userID), xdb.Cond.Eq("url", url))).
		Exists(ctx)
}

func toDomain(m *bookmarkModel) *domain.Bookmark {
	return &domain.Bookmark{
		ID:          m.ID,
		UserID:      m.UserID,
		URL:         m.URL,
		Title:       m.Title,
		Description: m.Description,
		FaviconURL:  m.FaviconURL,
		Category:    m.Category,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}
