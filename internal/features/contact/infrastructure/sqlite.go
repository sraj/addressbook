package infrastructure

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/mobentum/xdb"
	"github.com/sraj/addressbook/internal/features/contact/domain"
)

type contactModel struct {
	ID           uint   `db:"id"`
	UserID       uint   `db:"user_id"`
	CollectionID uint   `db:"collection_id"`
	Name         string `db:"name"`
	Emails       string `db:"emails"`
	Phones       string `db:"phones"`
	Addresses    string `db:"addresses"`
	Notes        string `db:"notes"`
}

type sqliteRepo struct {
	db *xdb.DB
}

func NewSQLiteRepo(db *xdb.DB) domain.Repository {
	return &sqliteRepo{db: db}
}

func (r *sqliteRepo) Create(ctx context.Context, contact *domain.Contact) error {
	m, err := toModel(contact)
	if err != nil {
		return err
	}
	return r.db.Insert("contacts").Columns("user_id", "collection_id", "name", "emails", "phones", "addresses", "notes").
		Values(m.UserID, m.CollectionID, m.Name, m.Emails, m.Phones, m.Addresses, m.Notes).Returning("id").One(ctx, &contact.ID)
}

func (r *sqliteRepo) FindByID(ctx context.Context, id, userID uint) (*domain.Contact, error) {
	var m contactModel
	err := r.db.Select("id", "user_id", "collection_id", "name", "emails", "phones", "addresses", "notes").
		From("contacts").
		Where(xdb.Cond.And(xdb.Cond.Eq("id", id), xdb.Cond.Eq("user_id", userID))).
		One(ctx, &m)
	if err != nil {
		if errors.Is(err, xdb.ErrNotFound) {
			return nil, domain.ErrContactNotFound
		}
		return nil, err
	}
	return toDomain(&m)
}

func (r *sqliteRepo) Update(ctx context.Context, contact *domain.Contact) error {
	m, err := toModel(contact)
	if err != nil {
		return err
	}
	_, err = r.db.Update("contacts").
		Set("name", m.Name).
		Set("emails", m.Emails).
		Set("phones", m.Phones).
		Set("addresses", m.Addresses).
		Set("notes", m.Notes).
		Where(xdb.Cond.Eq("id", contact.ID)).
		Exec(ctx)
	return err
}

func (r *sqliteRepo) Delete(ctx context.Context, id, userID uint) error {
	_, err := r.db.Delete("contacts").
		Where(xdb.Cond.And(xdb.Cond.Eq("id", id), xdb.Cond.Eq("user_id", userID))).
		Exec(ctx)
	return err
}

func (r *sqliteRepo) SetCollection(ctx context.Context, id, userID, collectionID uint) error {
	affected, err := r.db.Update("contacts").
		Set("collection_id", collectionID).
		Where(xdb.Cond.And(xdb.Cond.Eq("id", id), xdb.Cond.Eq("user_id", userID))).
		Exec(ctx)
	if err != nil {
		return err
	}
	if affected == 0 {
		return domain.ErrContactNotFound
	}
	return nil
}

func (r *sqliteRepo) ListByUser(ctx context.Context, userID uint, page, size int) ([]domain.Contact, int64, error) {
	return r.list(ctx, userID, 0, page, size)
}

func (r *sqliteRepo) ListByCollection(ctx context.Context, userID, collectionID uint, page, size int) ([]domain.Contact, int64, error) {
	return r.list(ctx, userID, collectionID, page, size)
}

func (r *sqliteRepo) list(ctx context.Context, userID, collectionID uint, page, size int) ([]domain.Contact, int64, error) {
	where := xdb.Cond.Eq("user_id", userID)
	if collectionID > 0 {
		where = xdb.Cond.And(where, xdb.Cond.Eq("collection_id", collectionID))
	}

	result, err := xdb.Paginate[contactModel](ctx,
		r.db.Select("id", "user_id", "collection_id", "name", "emails", "phones", "addresses", "notes").
			From("contacts").
			Where(where).
			OrderBy("updated_at", xdb.DESC),
		xdb.Page{Number: page, Size: size},
	)
	if err != nil {
		return nil, 0, err
	}

	contacts := make([]domain.Contact, len(result.Items))
	for i := range result.Items {
		c, err := toDomain(&result.Items[i])
		if err != nil {
			return nil, 0, err
		}
		contacts[i] = *c
	}
	return contacts, int64(result.Total), nil
}

func (r *sqliteRepo) ListAllByUser(ctx context.Context, userID, collectionID uint) ([]domain.Contact, error) {
	where := xdb.Cond.Eq("user_id", userID)
	if collectionID > 0 {
		where = xdb.Cond.And(where, xdb.Cond.Eq("collection_id", collectionID))
	}

	var rows []contactModel
	err := r.db.Select("id", "user_id", "collection_id", "name", "emails", "phones", "addresses", "notes").
		From("contacts").
		Where(where).
		OrderBy("name", xdb.ASC).
		All(ctx, &rows)
	if err != nil {
		return nil, err
	}

	contacts := make([]domain.Contact, len(rows))
	for i := range rows {
		c, err := toDomain(&rows[i])
		if err != nil {
			return nil, err
		}
		contacts[i] = *c
	}
	return contacts, nil
}

func (r *sqliteRepo) Search(ctx context.Context, userID uint, query string, page, size int) ([]domain.Contact, int64, error) {
	escaped := sanitizeFTSTerm(query)
	if escaped == "" {
		return r.ListByUser(ctx, userID, page, size)
	}

	ftsQuery := fmt.Sprintf("%s*", escaped)

	cond := xdb.Cond.And(
		xdb.Cond.Eq("user_id", userID),
		xdb.Cond.Raw("id IN (SELECT rowid FROM contacts_fts WHERE contacts_fts MATCH ?)", ftsQuery),
	)

	result, err := xdb.Paginate[contactModel](ctx,
		r.db.Select("id", "user_id", "collection_id", "name", "emails", "phones", "addresses", "notes").
			From("contacts").
			Where(cond).
			OrderBy("updated_at", xdb.DESC),
		xdb.Page{Number: page, Size: size},
	)
	if err != nil {
		return nil, 0, err
	}

	contacts := make([]domain.Contact, len(result.Items))
	for i := range result.Items {
		c, err := toDomain(&result.Items[i])
		if err != nil {
			return nil, 0, err
		}
		contacts[i] = *c
	}
	return contacts, int64(result.Total), nil
}

func toModel(c *domain.Contact) (*contactModel, error) {
	emails, err := json.Marshal(c.Emails)
	if err != nil {
		return nil, err
	}
	phones, err := json.Marshal(c.Phones)
	if err != nil {
		return nil, err
	}
	addrs, err := json.Marshal(c.Addresses)
	if err != nil {
		return nil, err
	}
	return &contactModel{
		ID:           c.ID,
		UserID:       c.UserID,
		CollectionID: c.CollectionID,
		Name:         c.Name,
		Emails:       string(emails),
		Phones:       string(phones),
		Addresses:    string(addrs),
		Notes:        c.Notes,
	}, nil
}

func toDomain(m *contactModel) (*domain.Contact, error) {
	var emails []string
	if err := json.Unmarshal([]byte(m.Emails), &emails); err != nil {
		return nil, err
	}
	var phones []string
	if err := json.Unmarshal([]byte(m.Phones), &phones); err != nil {
		return nil, err
	}
	var addrs []domain.Address
	if err := json.Unmarshal([]byte(m.Addresses), &addrs); err != nil {
		return nil, err
	}
	return &domain.Contact{
		ID:           m.ID,
		UserID:       m.UserID,
		CollectionID: m.CollectionID,
		Name:         m.Name,
		Emails:       emails,
		Phones:       phones,
		Addresses:    addrs,
		Notes:        m.Notes,
	}, nil
}

func sanitizeFTSTerm(q string) string {
	q = strings.TrimSpace(q)
	if q == "" {
		return ""
	}
	var b strings.Builder
	for _, r := range q {
		if r > 127 || ('a' <= r && r <= 'z') || ('A' <= r && r <= 'Z') || ('0' <= r && r <= '9') || r == ' ' || r == '-' || r == '_' || r == '@' || r == '.' || r == '+' || r == '#' {
			if r == '"' {
				b.WriteRune(' ')
			} else {
				b.WriteRune(r)
			}
		} else {
			b.WriteRune(' ')
		}
	}
	return b.String()
}
