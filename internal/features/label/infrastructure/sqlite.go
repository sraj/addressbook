package infrastructure

import (
	"context"
	"errors"

	"github.com/mobentum/xdb"
	"github.com/sraj/addressbook/internal/features/label/domain"
)

type labelOrderModel struct {
	ID              uint   `db:"id"`
	UserID          uint   `db:"user_id"`
	CollectionID    uint   `db:"collection_id"`
	ContactCount    int    `db:"contact_count"`
	SheetCount      int    `db:"sheet_count"`
	AmountCents     int64  `db:"amount_cents"`
	Currency        string `db:"currency"`
	Status          string `db:"status"`
	LabelType       string `db:"label_type"`
	StripeSessionID string `db:"stripe_session_id"`
}

type sqliteRepo struct {
	db *xdb.DB
}

func NewSQLiteRepo(db *xdb.DB) domain.Repository {
	return &sqliteRepo{db: db}
}

func (r *sqliteRepo) Create(ctx context.Context, o *domain.LabelOrder) error {
	return r.db.Insert("label_orders").
		Columns("user_id", "collection_id", "contact_count", "sheet_count", "amount_cents", "currency", "status", "label_type", "stripe_session_id").
		Values(o.UserID, o.CollectionID, o.ContactCount, o.SheetCount, o.AmountCents, o.Currency, o.Status, o.LabelType, o.StripeSessionID).
		Returning("id").One(ctx, &o.ID)
}

func (r *sqliteRepo) FindBySessionID(ctx context.Context, sessionID string) (*domain.LabelOrder, error) {
	var m labelOrderModel
	err := r.db.Select("id", "user_id", "collection_id", "contact_count", "sheet_count", "amount_cents", "currency", "status", "label_type", "stripe_session_id").
		From("label_orders").
		Where(xdb.Cond.Eq("stripe_session_id", sessionID)).
		One(ctx, &m)
	if err != nil {
		if errors.Is(err, xdb.ErrNotFound) {
			return nil, domain.ErrOrderNotFound
		}
		return nil, err
	}
	return toDomain(&m), nil
}

func (r *sqliteRepo) ListByUser(ctx context.Context, userID uint) ([]domain.LabelOrder, error) {
	var rows []labelOrderModel
	err := r.db.Select("id", "user_id", "collection_id", "contact_count", "sheet_count", "amount_cents", "currency", "status", "label_type", "stripe_session_id").
		From("label_orders").
		Where(xdb.Cond.Eq("user_id", userID)).
		OrderBy("created_at", xdb.DESC).
		All(ctx, &rows)
	if err != nil {
		return nil, err
	}
	orders := make([]domain.LabelOrder, len(rows))
	for i := range rows {
		orders[i] = *toDomain(&rows[i])
	}
	return orders, nil
}

func (r *sqliteRepo) SetSessionID(ctx context.Context, orderID uint, sessionID string) error {
	_, err := r.db.Update("label_orders").
		Set("stripe_session_id", sessionID).
		Where(xdb.Cond.Eq("id", orderID)).
		Exec(ctx)
	return err
}

func (r *sqliteRepo) UpdateStatusBySessionID(ctx context.Context, sessionID, status string) error {
	_, err := r.db.Update("label_orders").
		Set("status", status).
		Where(xdb.Cond.Eq("stripe_session_id", sessionID)).
		Exec(ctx)
	return err
}

func toDomain(m *labelOrderModel) *domain.LabelOrder {
	return &domain.LabelOrder{
		ID:              m.ID,
		UserID:          m.UserID,
		CollectionID:    m.CollectionID,
		ContactCount:    m.ContactCount,
		SheetCount:      m.SheetCount,
		AmountCents:     m.AmountCents,
		Currency:        m.Currency,
		Status:          m.Status,
		LabelType:       m.LabelType,
		StripeSessionID: m.StripeSessionID,
	}
}
