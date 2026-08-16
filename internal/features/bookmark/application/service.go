package application

import (
	"context"
	"errors"

	"github.com/mobentum/kern"
	"github.com/sraj/addressbook/internal/features/bookmark/domain"
	"github.com/sraj/addressbook/internal/features/bookmark/infrastructure"
)

type Service struct {
	repo    domain.Repository
	billing quotaChecker
}

func NewService(repo domain.Repository, billing quotaChecker) *Service {
	return &Service{repo: repo, billing: billing}
}

func (s *Service) enrich(ctx context.Context, rawURL, title, description, faviconURL string) (string, string, string) {
	if title != "" && description != "" && faviconURL != "" {
		return title, description, faviconURL
	}

	data, err := infrastructure.EnrichURL(rawURL)
	if err != nil {
		kern.LoggerFromContext(ctx).Warn("bookmark enrichment failed", "url", rawURL, "error", err)
		return title, description, faviconURL
	}

	if title == "" {
		title = data.Title
	}
	if description == "" {
		description = data.Description
	}
	if faviconURL == "" {
		faviconURL = data.FaviconURL
	}
	return title, description, faviconURL
}

func (s *Service) Create(ctx context.Context, userID uint, rawURL, title, description, faviconURL, category string) (*domain.Bookmark, error) {
	if err := s.billing.CheckQuota(ctx, userID, "bookmarks"); err != nil {
		return nil, err
	}

	title, description, faviconURL = s.enrich(ctx, rawURL, title, description, faviconURL)

	bookmark := &domain.Bookmark{
		UserID:      userID,
		URL:         rawURL,
		Title:       title,
		Description: description,
		FaviconURL:  faviconURL,
		Category:    category,
	}
	if err := s.repo.Create(ctx, bookmark); err != nil {
		return nil, err
	}

	if err := s.billing.IncrementUsage(ctx, userID, "bookmarks"); err != nil {
		return nil, err
	}

	return bookmark, nil
}

func (s *Service) GetByID(ctx context.Context, id, userID uint) (*domain.Bookmark, error) {
	bookmark, err := s.repo.FindByID(ctx, id, userID)
	if err != nil {
		if errors.Is(err, domain.ErrBookmarkNotFound) {
			return nil, domain.ErrBookmarkNotFound
		}
		return nil, err
	}
	return bookmark, nil
}

func (s *Service) Update(ctx context.Context, id, userID uint, rawURL, title, description, faviconURL, category string) (*domain.Bookmark, error) {
	bookmark, err := s.repo.FindByID(ctx, id, userID)
	if err != nil {
		if errors.Is(err, domain.ErrBookmarkNotFound) {
			return nil, domain.ErrBookmarkNotFound
		}
		return nil, err
	}

	if rawURL != bookmark.URL || title == "" || description == "" || faviconURL == "" {
		title, description, faviconURL = s.enrich(ctx, rawURL, title, description, faviconURL)
	}

	bookmark.URL = rawURL
	bookmark.Title = title
	bookmark.Description = description
	bookmark.FaviconURL = faviconURL
	bookmark.Category = category

	if err := s.repo.Update(ctx, bookmark); err != nil {
		return nil, err
	}
	return bookmark, nil
}

func (s *Service) Delete(ctx context.Context, id, userID uint) error {
	_, err := s.repo.FindByID(ctx, id, userID)
	if err != nil {
		if errors.Is(err, domain.ErrBookmarkNotFound) {
			return domain.ErrBookmarkNotFound
		}
		return err
	}
	if err := s.repo.Delete(ctx, id, userID); err != nil {
		return err
	}
	return s.billing.DecrementUsage(ctx, userID, "bookmarks")
}

func (s *Service) List(ctx context.Context, userID uint, category string, page, size int) ([]domain.Bookmark, int64, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	if category != "" {
		return s.repo.ListByCategory(ctx, userID, category, page, size)
	}
	return s.repo.ListByUser(ctx, userID, page, size)
}

func (s *Service) Import(ctx context.Context, userID uint, items []domain.Bookmark) (imported, skipped int, err error) {
	for _, item := range items {
		if err := s.billing.CheckQuota(ctx, userID, "bookmarks"); err != nil {
			return imported, skipped, err
		}

		exists, err := s.repo.ExistsByURL(ctx, userID, item.URL)
		if err != nil {
			return imported, skipped, err
		}
		if exists {
			skipped++
			continue
		}
		item.UserID = userID
		if err := s.repo.Create(ctx, &item); err != nil {
			return imported, skipped, err
		}
		if err := s.billing.IncrementUsage(ctx, userID, "bookmarks"); err != nil {
			return imported, skipped, err
		}
		imported++
	}
	return imported, skipped, nil
}

func (s *Service) ListCategories(ctx context.Context, userID uint) ([]string, error) {
	return s.repo.ListCategories(ctx, userID)
}
