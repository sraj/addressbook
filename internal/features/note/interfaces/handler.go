package interfaces

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/mobentum/kern"
	"github.com/mobentum/kern/extensions/xvalidator"
	billingDomain "github.com/sraj/addressbook/internal/billing/domain"
	"github.com/sraj/addressbook/internal/features/note/application"
	"github.com/sraj/addressbook/internal/features/note/domain"
	"github.com/sraj/addressbook/internal/shared"
)

type Handler struct {
	svc noteService
}

func NewHandler(svc noteService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) List(c *kern.Context) {
	userID, ok := shared.UserIDFromCtx(c.Context())
	if !ok {
		c.NoContent(http.StatusUnauthorized)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}

	notes, total, err := h.svc.List(c.Context(), userID, page, size)
	if err != nil {
		shared.SendError(c, http.StatusInternalServerError, "failed to list notes", err)
		return
	}

	_ = c.JSON(http.StatusOK, shared.NewPaginatedResponse(notes, total, page, size))
}

func (h *Handler) Get(c *kern.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		shared.SendError(c, http.StatusBadRequest, "invalid note id")
		return
	}

	userID, ok := shared.UserIDFromCtx(c.Context())
	if !ok {
		c.NoContent(http.StatusUnauthorized)
		return
	}

	note, err := h.svc.GetByID(c.Context(), uint(id), userID)
	if err != nil {
		if err == domain.ErrNoteNotFound {
			shared.SendError(c, http.StatusNotFound, "note not found")
			return
		}
		shared.SendError(c, http.StatusInternalServerError, "failed to get note", err)
		return
	}

	_ = c.JSON(http.StatusOK, note)
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

	note, err := h.svc.Create(c.Context(), userID, req.Title, req.Content)
	if err != nil {
		if errors.Is(err, billingDomain.ErrQuotaExceeded) {
			shared.SendError(c, http.StatusForbidden, err.Error())
			return
		}
		shared.SendError(c, http.StatusInternalServerError, "failed to create note", err)
		return
	}

	_ = c.JSON(http.StatusCreated, note)
}

func (h *Handler) Update(c *kern.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		shared.SendError(c, http.StatusBadRequest, "invalid note id")
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

	note, err := h.svc.Update(c.Context(), uint(id), userID, req.Title, req.Content)
	if err != nil {
		if err == domain.ErrNoteNotFound {
			shared.SendError(c, http.StatusNotFound, "note not found")
			return
		}
		if l := kern.LoggerFromContext(c.Context()); l != nil {
			l.Error("failed to update note", "error", err, "note_id", id)
		}
		shared.SendError(c, http.StatusInternalServerError, "failed to update note", err)
		return
	}

	_ = c.JSON(http.StatusOK, note)
}

func (h *Handler) Delete(c *kern.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		shared.SendError(c, http.StatusBadRequest, "invalid note id")
		return
	}

	userID, ok := shared.UserIDFromCtx(c.Context())
	if !ok {
		c.NoContent(http.StatusUnauthorized)
		return
	}

	if err := h.svc.Delete(c.Context(), uint(id), userID); err != nil {
		if err == domain.ErrNoteNotFound {
			shared.SendError(c, http.StatusNotFound, "note not found")
			return
		}
		shared.SendError(c, http.StatusInternalServerError, "failed to delete note", err)
		return
	}

	c.NoContent(http.StatusNoContent)
}
