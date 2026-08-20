package infrastructure

import (
	"context"
	"errors"
	"time"

	"github.com/mobentum/xdb"
	authDomain "github.com/sraj/addressbook/internal/auth/domain"
	"golang.org/x/crypto/bcrypt"
)

type userModel struct {
	ID           uint   `db:"id"`
	Email        string `db:"email"`
	Name         string `db:"name"`
	Preferences  string `db:"preferences"`
	PasswordHash string `db:"password_hash"`
	Role         string `db:"role"`
	Status       string `db:"status"`
}

type passwordResetTokenModel struct {
	ID        uint      `db:"id"`
	UserID    uint      `db:"user_id"`
	Token     string    `db:"token"`
	ExpiresAt time.Time `db:"expires_at"`
	Used      bool      `db:"used"`
}

type sqliteRepo struct {
	db *xdb.DB
}

func NewSQLiteRepo(db *xdb.DB) authDomain.Repository {
	return &sqliteRepo{db: db}
}

func Seed(ctx context.Context, db *xdb.DB) error {
	// Check if admin user exists
	type userRow struct {
		ID   uint   `db:"id"`
		Role string `db:"role"`
	}
	var existing userRow
	err := db.Select("id", "role").From("users").Where(xdb.Cond.Eq("email", "admin@example.com")).One(ctx, &existing)
	if err == nil {
		// User exists — ensure role is admin
		if existing.Role != "admin" {
			_, err = db.Update("users").Set("role", "admin").Where(xdb.Cond.Eq("id", existing.ID)).Exec(ctx)
			return err
		}
		return nil
	}

	// Not found — create admin user
	hash, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = db.Insert("users").Columns("email", "password_hash", "role").Values("admin@example.com", string(hash), "admin").Exec(ctx)
	return err
}

func (r *sqliteRepo) Create(ctx context.Context, user *authDomain.User) error {
	return r.db.Insert("users").Columns("email", "password_hash").Values(user.Email, user.PasswordHash).Returning("id").One(ctx, &user.ID)
}

func (r *sqliteRepo) FindByEmail(ctx context.Context, email string) (*authDomain.User, error) {
	var m userModel
	err := r.db.Select("id", "email", "name", "preferences", "password_hash", "role", "status").From("users").Where(xdb.Cond.Eq("email", email)).One(ctx, &m)
	if err != nil {
		if errors.Is(err, xdb.ErrNotFound) {
			return nil, authDomain.ErrUserNotFound
		}
		return nil, err
	}
	return toDomainUser(&m), nil
}

func (r *sqliteRepo) FindByID(ctx context.Context, id uint) (*authDomain.User, error) {
	var m userModel
	err := r.db.Select("id", "email", "name", "preferences", "password_hash", "role", "status").From("users").Where(xdb.Cond.Eq("id", id)).One(ctx, &m)
	if err != nil {
		if errors.Is(err, xdb.ErrNotFound) {
			return nil, authDomain.ErrUserNotFound
		}
		return nil, err
	}
	return toDomainUser(&m), nil
}

func toDomainUser(m *userModel) *authDomain.User {
	return &authDomain.User{
		ID:           m.ID,
		Email:        m.Email,
		Name:         m.Name,
		Preferences:  m.Preferences,
		PasswordHash: m.PasswordHash,
		Role:         m.Role,
		Status:       m.Status,
	}
}

func (r *sqliteRepo) Exists(ctx context.Context, email string) (bool, error) {
	return r.db.Select("1").From("users").Where(xdb.Cond.Eq("email", email)).Exists(ctx)
}

func (r *sqliteRepo) UpdateName(ctx context.Context, userID uint, name string) error {
	_, err := r.db.Update("users").Set("name", name).Where(xdb.Cond.Eq("id", userID)).Exec(ctx)
	return err
}

func (r *sqliteRepo) UpdatePreferences(ctx context.Context, userID uint, preferences string) error {
	_, err := r.db.Update("users").Set("preferences", preferences).Where(xdb.Cond.Eq("id", userID)).Exec(ctx)
	return err
}

func (r *sqliteRepo) UpdatePassword(ctx context.Context, userID uint, passwordHash string) error {
	_, err := r.db.Update("users").Set("password_hash", passwordHash).Where(xdb.Cond.Eq("id", userID)).Exec(ctx)
	return err
}

func (r *sqliteRepo) CreateResetToken(ctx context.Context, token *authDomain.PasswordResetToken) error {
	_, err := r.db.Insert("password_reset_tokens").Columns("user_id", "token", "expires_at", "used").
		Values(token.UserID, token.Token, token.ExpiresAt, token.Used).Exec(ctx)
	return err
}

func (r *sqliteRepo) FindResetToken(ctx context.Context, tokenStr string) (*authDomain.PasswordResetToken, error) {
	var m passwordResetTokenModel
	err := r.db.Select("id", "user_id", "token", "expires_at", "used").From("password_reset_tokens").
		Where(xdb.Cond.Eq("token", tokenStr)).One(ctx, &m)
	if err != nil {
		if errors.Is(err, xdb.ErrNotFound) {
			return nil, authDomain.ErrInvalidToken
		}
		return nil, err
	}
	return &authDomain.PasswordResetToken{
		ID:        m.ID,
		UserID:    m.UserID,
		Token:     m.Token,
		ExpiresAt: m.ExpiresAt,
		Used:      m.Used,
	}, nil
}

func (r *sqliteRepo) MarkResetTokenUsed(ctx context.Context, tokenID uint) error {
	_, err := r.db.Update("password_reset_tokens").Set("used", true).Where(xdb.Cond.Eq("id", tokenID)).Exec(ctx)
	return err
}
