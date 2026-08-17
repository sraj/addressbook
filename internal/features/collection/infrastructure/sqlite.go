package infrastructure

import (
	"context"
	"errors"

	"github.com/mobentum/xdb"
	"github.com/sraj/addressbook/internal/features/collection/domain"
)

type collectionModel struct {
	ID           uint   `db:"id"`
	UserID       uint   `db:"user_id"`
	Name         string `db:"name"`
	InviteToken  string `db:"invite_token"`
	ContactCount int64  `db:"contact_count"`
}

type sqliteRepo struct {
	db *xdb.DB
}

func NewSQLiteRepo(db *xdb.DB) domain.Repository {
	return &sqliteRepo{db: db}
}

func (r *sqliteRepo) Create(ctx context.Context, c *domain.Collection) error {
	return r.db.Insert("collections").Columns("user_id", "name", "invite_token").
		Values(c.UserID, c.Name, c.InviteToken).Returning("id").One(ctx, &c.ID)
}

func (r *sqliteRepo) FindByID(ctx context.Context, id, userID uint) (*domain.Collection, error) {
	var m collectionModel
	err := r.db.Select("id", "user_id", "name", "invite_token").
		From("collections").
		Where(xdb.Cond.And(xdb.Cond.Eq("id", id), xdb.Cond.Eq("user_id", userID))).
		One(ctx, &m)
	if err != nil {
		if errors.Is(err, xdb.ErrNotFound) {
			return nil, domain.ErrCollectionNotFound
		}
		return nil, err
	}
	return toDomain(&m), nil
}

func (r *sqliteRepo) FindByToken(ctx context.Context, token string) (*domain.Collection, error) {
	var m collectionModel
	err := r.db.Select("id", "user_id", "name", "invite_token").
		From("collections").
		Where(xdb.Cond.Eq("invite_token", token)).
		One(ctx, &m)
	if err != nil {
		if errors.Is(err, xdb.ErrNotFound) {
			return nil, domain.ErrCollectionNotFound
		}
		return nil, err
	}
	return toDomain(&m), nil
}

func (r *sqliteRepo) ListByUser(ctx context.Context, userID uint) ([]domain.Collection, error) {
	var rows []collectionModel
	err := r.db.Select("c.id", "c.user_id", "c.name", "c.invite_token", "(SELECT COUNT(*) FROM contacts k WHERE k.collection_id = c.id) AS contact_count").
		From("collections c").
		Where(xdb.Cond.Eq("c.user_id", userID)).
		OrderBy("c.updated_at", xdb.DESC).
		All(ctx, &rows)
	if err != nil {
		return nil, err
	}

	collections := make([]domain.Collection, len(rows))
	for i := range rows {
		collections[i] = *toDomain(&rows[i])
	}
	return collections, nil
}

func (r *sqliteRepo) Update(ctx context.Context, c *domain.Collection) error {
	_, err := r.db.Update("collections").
		Set("name", c.Name).
		Where(xdb.Cond.Eq("id", c.ID)).
		Exec(ctx)
	return err
}

func (r *sqliteRepo) Delete(ctx context.Context, id, userID uint) error {
	_, err := r.db.Delete("collections").
		Where(xdb.Cond.And(xdb.Cond.Eq("id", id), xdb.Cond.Eq("user_id", userID))).
		Exec(ctx)
	return err
}

func (r *sqliteRepo) RegenerateToken(ctx context.Context, id, userID uint, token string) error {
	_, err := r.db.Update("collections").
		Set("invite_token", token).
		Where(xdb.Cond.And(xdb.Cond.Eq("id", id), xdb.Cond.Eq("user_id", userID))).
		Exec(ctx)
	return err
}

func (r *sqliteRepo) CountContacts(ctx context.Context, collectionID uint) (int64, error) {
	var count int64
	err := r.db.Select("COUNT(*)").From("contacts").
		Where(xdb.Cond.Eq("collection_id", collectionID)).One(ctx, &count)
	return count, err
}

func (r *sqliteRepo) UnlinkContacts(ctx context.Context, collectionID uint) error {
	_, err := r.db.Update("contacts").
		Set("collection_id", 0).
		Where(xdb.Cond.Eq("collection_id", collectionID)).
		Exec(ctx)
	return err
}

func toDomain(m *collectionModel) *domain.Collection {
	return &domain.Collection{
		ID:           m.ID,
		UserID:       m.UserID,
		Name:         m.Name,
		InviteToken:  m.InviteToken,
		ContactCount: m.ContactCount,
	}
}
