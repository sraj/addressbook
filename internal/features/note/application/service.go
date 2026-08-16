package application

import (
	"context"
	"errors"

	"github.com/sraj/addressbook/internal/features/note/domain"
)

type Service struct {
	repo    domain.Repository
	billing quotaChecker
}

func NewService(repo domain.Repository, billing quotaChecker) *Service {
	return &Service{repo: repo, billing: billing}
}

func (s *Service) Create(ctx context.Context, userID uint, title, content string) (*domain.Note, error) {
	if err := s.billing.CheckQuota(ctx, userID, "notes"); err != nil {
		return nil, err
	}

	note := &domain.Note{
		UserID:  userID,
		Title:   title,
		Content: content,
	}
	if err := s.repo.Create(ctx, note); err != nil {
		return nil, err
	}

	if err := s.billing.IncrementUsage(ctx, userID, "notes"); err != nil {
		return nil, err
	}

	return note, nil
}

func (s *Service) GetByID(ctx context.Context, id, userID uint) (*domain.Note, error) {
	note, err := s.repo.FindByID(ctx, id, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNoteNotFound) {
			return nil, domain.ErrNoteNotFound
		}
		return nil, err
	}
	return note, nil
}

func (s *Service) Update(ctx context.Context, id, userID uint, title, content string) (*domain.Note, error) {
	note, err := s.repo.FindByID(ctx, id, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNoteNotFound) {
			return nil, domain.ErrNoteNotFound
		}
		return nil, err
	}

	note.Title = title
	note.Content = content

	if err := s.repo.Update(ctx, note); err != nil {
		return nil, err
	}
	return note, nil
}

func (s *Service) Delete(ctx context.Context, id, userID uint) error {
	_, err := s.repo.FindByID(ctx, id, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNoteNotFound) {
			return domain.ErrNoteNotFound
		}
		return err
	}
	if err := s.repo.Delete(ctx, id, userID); err != nil {
		return err
	}
	return s.billing.DecrementUsage(ctx, userID, "notes")
}

func (s *Service) List(ctx context.Context, userID uint, page, size int) ([]domain.Note, int64, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	return s.repo.ListByUser(ctx, userID, page, size)
}
