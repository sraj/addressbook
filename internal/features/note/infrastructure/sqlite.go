package infrastructure

import (
	"context"
	"errors"
	"time"

	"github.com/sraj/addressbook/internal/features/note/domain"
	"github.com/mobentum/xdb"
)

type noteModel struct {
	ID        uint   `db:"id"`
	UserID    uint   `db:"user_id"`
	Title     string `db:"title"`
	Content   string `db:"content"`
	CreatedAt string `db:"created_at"`
	UpdatedAt string `db:"updated_at"`
}

type sqliteRepo struct {
	db *xdb.DB
}

func NewSQLiteRepo(db *xdb.DB) domain.Repository {
	return &sqliteRepo{db: db}
}

func (r *sqliteRepo) Create(ctx context.Context, note *domain.Note) error {
	return r.db.Insert("notes").Columns("user_id", "title", "content").
		Values(note.UserID, note.Title, note.Content).Returning("id").One(ctx, &note.ID)
}

func (r *sqliteRepo) FindByID(ctx context.Context, id, userID uint) (*domain.Note, error) {
	var m noteModel
	err := r.db.Select("id", "user_id", "title", "content", "created_at", "updated_at").
		From("notes").
		Where(xdb.Cond.And(xdb.Cond.Eq("id", id), xdb.Cond.Eq("user_id", userID))).
		One(ctx, &m)
	if err != nil {
		if errors.Is(err, xdb.ErrNotFound) {
			return nil, domain.ErrNoteNotFound
		}
		return nil, err
	}
	return toDomain(&m), nil
}

func (r *sqliteRepo) Update(ctx context.Context, note *domain.Note) error {
	_, err := r.db.Update("notes").
		Set("title", note.Title).
		Set("content", note.Content).
		Set("updated_at", time.Now().Format("2006-01-02 15:04:05")).
		Where(xdb.Cond.Eq("id", note.ID)).
		Exec(ctx)
	return err
}

func (r *sqliteRepo) Delete(ctx context.Context, id, userID uint) error {
	_, err := r.db.Delete("notes").
		Where(xdb.Cond.And(xdb.Cond.Eq("id", id), xdb.Cond.Eq("user_id", userID))).
		Exec(ctx)
	return err
}

func (r *sqliteRepo) ListByUser(ctx context.Context, userID uint, page, size int) ([]domain.Note, int64, error) {
	result, err := xdb.Paginate[noteModel](ctx,
		r.db.Select("id", "user_id", "title", "content", "created_at", "updated_at").
			From("notes").
			Where(xdb.Cond.Eq("user_id", userID)).
			OrderBy("updated_at", xdb.DESC),
		xdb.Page{Number: page, Size: size},
	)
	if err != nil {
		return nil, 0, err
	}

	notes := make([]domain.Note, len(result.Items))
	for i := range result.Items {
		notes[i] = *toDomain(&result.Items[i])
	}
	return notes, int64(result.Total), nil
}

func toDomain(m *noteModel) *domain.Note {
	return &domain.Note{
		ID:        m.ID,
		UserID:    m.UserID,
		Title:     m.Title,
		Content:   m.Content,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}
