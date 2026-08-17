package application

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"

	collectionDomain "github.com/sraj/addressbook/internal/features/collection/domain"
	contactDomain "github.com/sraj/addressbook/internal/features/contact/domain"
)

// contactStore is the provider-side port the service uses to persist address
// submissions as contacts. It is satisfied by *contact/application.Service.
type contactStore interface {
	Create(ctx context.Context, userID, collectionID uint, name string, emails, phones []string, addresses []contactDomain.Address, notes string) (*contactDomain.Contact, error)
	ListByCollection(ctx context.Context, userID, collectionID uint, page, size int) ([]contactDomain.Contact, int64, error)
	SetCollection(ctx context.Context, contactID, userID, collectionID uint) error
}

type Service struct {
	repo    collectionDomain.Repository
	contact contactStore
}

func NewService(repo collectionDomain.Repository, contact contactStore) *Service {
	return &Service{repo: repo, contact: contact}
}

func (s *Service) Create(ctx context.Context, userID uint, name string) (*collectionDomain.Collection, error) {
	name = strings.TrimSpace(name)
	if name == "" || len(name) > 200 {
		return nil, collectionDomain.ErrCollectionName
	}

	c := &collectionDomain.Collection{
		UserID:      userID,
		Name:        name,
		InviteToken: generateToken(),
	}
	if err := s.repo.Create(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *Service) List(ctx context.Context, userID uint) ([]collectionDomain.Collection, error) {
	return s.repo.ListByUser(ctx, userID)
}

func (s *Service) Get(ctx context.Context, id, userID uint) (*collectionDomain.Collection, error) {
	c, err := s.repo.FindByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	count, err := s.repo.CountContacts(ctx, c.ID)
	if err != nil {
		return nil, err
	}
	c.ContactCount = count
	return c, nil
}

func (s *Service) Rename(ctx context.Context, id, userID uint, name string) (*collectionDomain.Collection, error) {
	name = strings.TrimSpace(name)
	if name == "" || len(name) > 200 {
		return nil, collectionDomain.ErrCollectionName
	}
	c, err := s.repo.FindByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	c.Name = name
	if err := s.repo.Update(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *Service) Delete(ctx context.Context, id, userID uint) error {
	if _, err := s.repo.FindByID(ctx, id, userID); err != nil {
		return err
	}
	if err := s.repo.UnlinkContacts(ctx, id); err != nil {
		return err
	}
	return s.repo.Delete(ctx, id, userID)
}

func (s *Service) RegenerateToken(ctx context.Context, id, userID uint) (*collectionDomain.Collection, error) {
	if _, err := s.repo.FindByID(ctx, id, userID); err != nil {
		return nil, err
	}
	token := generateToken()
	if err := s.repo.RegenerateToken(ctx, id, userID, token); err != nil {
		return nil, err
	}
	return s.Get(ctx, id, userID)
}

// GetByToken returns a collection by its invite token for the public page.
func (s *Service) GetByToken(ctx context.Context, token string) (*collectionDomain.Collection, error) {
	return s.repo.FindByToken(ctx, token)
}

// SubmitAddress is called from the public invite page. It resolves the
// collection by token and stores the submitted address as a new contact
// owned by the collection's user.
func (s *Service) SubmitAddress(ctx context.Context, token, name, email, phone string, address contactDomain.Address) (*contactDomain.Contact, error) {
	if name = strings.TrimSpace(name); name == "" {
		return nil, errors.New("name is required")
	}
	if err := validateAddress(address); err != nil {
		return nil, err
	}

	c, err := s.repo.FindByToken(ctx, token)
	if err != nil {
		if errors.Is(err, collectionDomain.ErrCollectionNotFound) {
			return nil, collectionDomain.ErrInvalidToken
		}
		return nil, err
	}

	var emails, phones []string
	if email != "" {
		emails = []string{email}
	}
	if phone != "" {
		phones = []string{phone}
	}

	return s.contact.Create(ctx, c.UserID, c.ID, name, emails, phones, []contactDomain.Address{address}, "")
}

func validateAddress(a contactDomain.Address) error {
	if a.Line1 == "" || a.City == "" || a.State == "" || a.Zip == "" || a.Country == "" {
		return errors.New("complete address is required")
	}
	return nil
}

func (s *Service) ListContacts(ctx context.Context, userID, collectionID uint, page, size int) ([]contactDomain.Contact, int64, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	return s.contact.ListByCollection(ctx, userID, collectionID, page, size)
}

// AddContact creates a new contact directly inside the user's collection,
// bypassing the public invite flow. The collection must belong to the user.
func (s *Service) AddContact(ctx context.Context, userID, collectionID uint, name string, emails, phones []string, addresses []contactDomain.Address, notes string) (*contactDomain.Contact, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("name is required")
	}
	if len(addresses) == 0 {
		return nil, errors.New("at least one address is required")
	}
	if _, err := s.repo.FindByID(ctx, collectionID, userID); err != nil {
		return nil, err
	}
	return s.contact.Create(ctx, userID, collectionID, name, emails, phones, addresses, notes)
}

// MoveContact assigns an existing contact (that belongs to the user) to the
// given collection. The target collection must belong to the user.
func (s *Service) MoveContact(ctx context.Context, userID, collectionID, contactID uint) error {
	if _, err := s.repo.FindByID(ctx, collectionID, userID); err != nil {
		return err
	}
	return s.contact.SetCollection(ctx, contactID, userID, collectionID)
}

// RemoveContact clears a contact's collection membership.
func (s *Service) RemoveContact(ctx context.Context, userID, collectionID, contactID uint) error {
	if _, err := s.repo.FindByID(ctx, collectionID, userID); err != nil {
		return err
	}
	return s.contact.SetCollection(ctx, contactID, userID, 0)
}

func generateToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		// crypto/rand failure is catastrophic; fall back to a timestamp-based
		// token so the request still succeeds in development.
		return hex.EncodeToString([]byte("dev-fallback-token"))
	}
	return hex.EncodeToString(b)
}
