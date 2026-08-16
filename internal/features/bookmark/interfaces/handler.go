package interfaces

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/mobentum/kern"
	"github.com/mobentum/kern/extensions/xvalidator"
	billingDomain "github.com/sraj/addressbook/internal/billing/domain"
	"github.com/sraj/addressbook/internal/features/bookmark/application"
	"github.com/sraj/addressbook/internal/features/bookmark/domain"
	"github.com/sraj/addressbook/internal/features/bookmark/infrastructure"
	"github.com/sraj/addressbook/internal/shared"
)

type Handler struct {
	svc bookmarkService
}

func NewHandler(svc bookmarkService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) List(c *kern.Context) {
	userID, ok := shared.UserIDFromCtx(c.Context())
	if !ok {
		c.NoContent(http.StatusUnauthorized)
		return
	}

	category := c.DefaultQuery("category", "")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}

	bookmarks, total, err := h.svc.List(c.Context(), userID, category, page, size)
	if err != nil {
		shared.SendError(c, http.StatusInternalServerError, "failed to list bookmarks", err)
		return
	}

	_ = c.JSON(http.StatusOK, shared.NewPaginatedResponse(bookmarks, total, page, size))
}

func (h *Handler) Get(c *kern.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		shared.SendError(c, http.StatusBadRequest, "invalid bookmark id")
		return
	}

	userID, ok := shared.UserIDFromCtx(c.Context())
	if !ok {
		c.NoContent(http.StatusUnauthorized)
		return
	}

	bookmark, err := h.svc.GetByID(c.Context(), uint(id), userID)
	if err != nil {
		if err == domain.ErrBookmarkNotFound {
			shared.SendError(c, http.StatusNotFound, "bookmark not found")
			return
		}
		shared.SendError(c, http.StatusInternalServerError, "failed to get bookmark", err)
		return
	}

	_ = c.JSON(http.StatusOK, bookmark)
}

func (h *Handler) Create(c *kern.Context) {
	req, ok := xvalidator.Validated[application.CreateRequest](c.Context())
	if !ok {
		return
	}

	userID, ok := shared.UserIDFromCtx(c.Context())
	if !ok {
		c.NoContent(http.StatusUnauthorized)
		return
	}

	bookmark, err := h.svc.Create(c.Context(), userID, req.URL, req.Title, req.Description, req.FaviconURL, req.Category)
	if err != nil {
		if errors.Is(err, billingDomain.ErrQuotaExceeded) {
			shared.SendError(c, http.StatusForbidden, err.Error())
			return
		}
		shared.SendError(c, http.StatusInternalServerError, "failed to create bookmark", err)
		return
	}

	_ = c.JSON(http.StatusCreated, bookmark)
}

func (h *Handler) Update(c *kern.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		shared.SendError(c, http.StatusBadRequest, "invalid bookmark id")
		return
	}

	req, ok := xvalidator.Validated[application.UpdateRequest](c.Context())
	if !ok {
		return
	}

	userID, ok := shared.UserIDFromCtx(c.Context())
	if !ok {
		c.NoContent(http.StatusUnauthorized)
		return
	}

	bookmark, err := h.svc.Update(c.Context(), uint(id), userID, req.URL, req.Title, req.Description, req.FaviconURL, req.Category)
	if err != nil {
		if err == domain.ErrBookmarkNotFound {
			shared.SendError(c, http.StatusNotFound, "bookmark not found")
			return
		}
		if l := kern.LoggerFromContext(c.Context()); l != nil {
			l.Error("failed to update bookmark", "error", err, "bookmark_id", id)
		}
		shared.SendError(c, http.StatusInternalServerError, "failed to update bookmark", err)
		return
	}

	_ = c.JSON(http.StatusOK, bookmark)
}

func (h *Handler) Delete(c *kern.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		shared.SendError(c, http.StatusBadRequest, "invalid bookmark id")
		return
	}

	userID, ok := shared.UserIDFromCtx(c.Context())
	if !ok {
		c.NoContent(http.StatusUnauthorized)
		return
	}

	if err := h.svc.Delete(c.Context(), uint(id), userID); err != nil {
		if err == domain.ErrBookmarkNotFound {
			shared.SendError(c, http.StatusNotFound, "bookmark not found")
			return
		}
		shared.SendError(c, http.StatusInternalServerError, "failed to delete bookmark", err)
		return
	}

	c.NoContent(http.StatusNoContent)
}

func (h *Handler) Import(c *kern.Context) {
	req, ok := xvalidator.Validated[application.ImportRequest](c.Context())
	if !ok {
		return
	}

	userID, ok := shared.UserIDFromCtx(c.Context())
	if !ok {
		c.NoContent(http.StatusUnauthorized)
		return
	}

	items := make([]domain.Bookmark, len(req.Bookmarks))
	for i, item := range req.Bookmarks {
		items[i] = domain.Bookmark{
			URL:         item.URL,
			Title:       item.Title,
			Description: item.Description,
			FaviconURL:  item.FaviconURL,
			Category:    item.Category,
		}
	}

	imported, skipped, err := h.svc.Import(c.Context(), userID, items)
	if err != nil {
		if errors.Is(err, billingDomain.ErrQuotaExceeded) {
			shared.SendError(c, http.StatusForbidden, err.Error())
			return
		}
		shared.SendError(c, http.StatusInternalServerError, "failed to import bookmarks", err)
		return
	}

	_ = c.JSON(http.StatusOK, application.ImportResponse{Imported: imported, Skipped: skipped})
}

func (h *Handler) ImportHTML(c *kern.Context) {
	var req application.ImportHTMLRequest
	if err := c.DecodeJSON(&req); err != nil {
		shared.SendError(c, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := xvalidator.Validate(req); err != nil {
		shared.ValidationError(c, err)
		return
	}

	userID, ok := shared.UserIDFromCtx(c.Context())
	if !ok {
		c.NoContent(http.StatusUnauthorized)
		return
	}

	parsed := infrastructure.ParseBookmarkHTML(req.HTML)
	if len(parsed) == 0 {
		shared.SendError(c, http.StatusBadRequest, "no bookmarks found in HTML")
		return
	}

	items := make([]domain.Bookmark, 0, len(parsed))
	for _, p := range parsed {
		if p.Title == "" {
			p.Title = p.URL
		}
		if p.FaviconURL == "" {
			p.FaviconURL = p.URL
		}
		items = append(items, domain.Bookmark{
			URL:        p.URL,
			Title:      p.Title,
			Category:   p.Category,
			FaviconURL: infrastructure.FaviconURLFromURL(p.FaviconURL),
		})
	}

	imported, skipped, err := h.svc.Import(c.Context(), userID, items)
	if err != nil {
		if errors.Is(err, billingDomain.ErrQuotaExceeded) {
			shared.SendError(c, http.StatusForbidden, err.Error())
			return
		}
		shared.SendError(c, http.StatusInternalServerError, "failed to import bookmarks", err)
		return
	}

	_ = c.JSON(http.StatusOK, application.ImportResponse{Imported: imported, Skipped: skipped})
}

func (h *Handler) ListCategories(c *kern.Context) {
	userID, ok := shared.UserIDFromCtx(c.Context())
	if !ok {
		c.NoContent(http.StatusUnauthorized)
		return
	}

	categories, err := h.svc.ListCategories(c.Context(), userID)
	if err != nil {
		shared.SendError(c, http.StatusInternalServerError, "failed to list categories", err)
		return
	}

	_ = c.JSON(http.StatusOK, map[string][]string{"categories": categories})
}
