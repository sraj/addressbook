package application

import (
	"context"
	"errors"

	"github.com/sraj/addressbook/internal/features/contact/domain"
)

type Service struct {
	repo    domain.Repository
	billing quotaChecker
}

func NewService(repo domain.Repository, billing quotaChecker) *Service {
	return &Service{repo: repo, billing: billing}
}

func (s *Service) Create(ctx context.Context, userID, collectionID uint, name string, emails, phones []string, addresses []domain.Address, notes string) (*domain.Contact, error) {
	if err := s.billing.CheckQuota(ctx, userID, "contacts"); err != nil {
		return nil, err
	}

	contact := &domain.Contact{
		UserID:       userID,
		CollectionID: collectionID,
		Name:         name,
		Emails:       emails,
		Phones:       phones,
		Addresses:    addresses,
		Notes:        notes,
	}
	if err := s.repo.Create(ctx, contact); err != nil {
		return nil, err
	}

	if err := s.billing.IncrementUsage(ctx, userID, "contacts"); err != nil {
		return nil, err
	}

	return contact, nil
}

func (s *Service) GetByID(ctx context.Context, id, userID uint) (*domain.Contact, error) {
	contact, err := s.repo.FindByID(ctx, id, userID)
	if err != nil {
		if errors.Is(err, domain.ErrContactNotFound) {
			return nil, domain.ErrContactNotFound
		}
		return nil, err
	}
	return contact, nil
}

func (s *Service) Update(ctx context.Context, id, userID uint, name string, emails, phones []string, addresses []domain.Address, notes string) (*domain.Contact, error) {
	contact, err := s.repo.FindByID(ctx, id, userID)
	if err != nil {
		if errors.Is(err, domain.ErrContactNotFound) {
			return nil, domain.ErrContactNotFound
		}
		return nil, err
	}

	contact.Name = name
	contact.Emails = emails
	contact.Phones = phones
	contact.Addresses = addresses
	contact.Notes = notes

	if err := s.repo.Update(ctx, contact); err != nil {
		return nil, err
	}
	return contact, nil
}

func (s *Service) Delete(ctx context.Context, id, userID uint) error {
	_, err := s.repo.FindByID(ctx, id, userID)
	if err != nil {
		if errors.Is(err, domain.ErrContactNotFound) {
			return domain.ErrContactNotFound
		}
		return err
	}
	if err := s.repo.Delete(ctx, id, userID); err != nil {
		return err
	}
	return s.billing.DecrementUsage(ctx, userID, "contacts")
}

func (s *Service) List(ctx context.Context, userID uint, query string, page, size int) ([]domain.Contact, int64, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	if query != "" {
		return s.repo.Search(ctx, userID, query, page, size)
	}
	return s.repo.ListByUser(ctx, userID, page, size)
}

func (s *Service) ListByCollection(ctx context.Context, userID, collectionID uint, page, size int) ([]domain.Contact, int64, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	return s.repo.ListByCollection(ctx, userID, collectionID, page, size)
}

// SetCollection moves a contact into (or out of) a collection. The ownership
// guard lives in the repository WHERE clause; a missing/misowned contact
// surfaces as domain.ErrContactNotFound.
func (s *Service) SetCollection(ctx context.Context, id, userID, collectionID uint) error {
	return s.repo.SetCollection(ctx, id, userID, collectionID)
}

func (s *Service) ListAllByUser(ctx context.Context, userID, collectionID uint) ([]domain.Contact, error) {
	return s.repo.ListAllByUser(ctx, userID, collectionID)
}

// Export renders all matching contacts in the requested format ("csv"|"xlsx").
// Returns the file bytes and an HTTP content type.
func (s *Service) Export(ctx context.Context, userID, collectionID uint, format string) ([]byte, string, error) {
	contacts, err := s.repo.ListAllByUser(ctx, userID, collectionID)
	if err != nil {
		return nil, "", err
	}

	switch format {
	case "xlsx":
		data, err := ExportXLSX(contacts)
		return data, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", err
	default:
		data, err := ExportCSV(contacts)
		return data, "text/csv; charset=utf-8", err
	}
}
