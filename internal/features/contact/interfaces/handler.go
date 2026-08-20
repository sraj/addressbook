package interfaces

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/mobentum/kern"
	"github.com/mobentum/kern/extensions/xvalidator"
	billingDomain "github.com/sraj/addressbook/internal/billing/domain"
	"github.com/sraj/addressbook/internal/features/contact/application"
	"github.com/sraj/addressbook/internal/features/contact/domain"
	"github.com/sraj/addressbook/internal/shared"
)

type Handler struct {
	svc contactService
}

func NewHandler(svc contactService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) List(c *kern.Context) {
	userID, ok := shared.UserIDFromCtx(c.Context())
	if !ok {
		c.NoContent(http.StatusUnauthorized)
		return
	}

	q := c.DefaultQuery("q", "")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}

	contacts, total, err := h.svc.List(c.Context(), userID, q, page, size)
	if err != nil {
		shared.SendError(c, http.StatusInternalServerError, "failed to list contacts", err)
		return
	}

	_ = c.JSON(http.StatusOK, shared.NewPaginatedResponse(contacts, total, page, size))
}

func (h *Handler) Get(c *kern.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		shared.SendError(c, http.StatusBadRequest, "invalid contact id")
		return
	}

	userID, ok := shared.UserIDFromCtx(c.Context())
	if !ok {
		c.NoContent(http.StatusUnauthorized)
		return
	}

	contact, err := h.svc.GetByID(c.Context(), uint(id), userID)
	if err != nil {
		if err == domain.ErrContactNotFound {
			shared.SendError(c, http.StatusNotFound, "contact not found")
			return
		}
		shared.SendError(c, http.StatusInternalServerError, "failed to get contact", err)
		return
	}

	_ = c.JSON(http.StatusOK, contact)
}

func (h *Handler) Create(c *kern.Context) {
	req, ok := xvalidator.ValidatedFromContext[application.CreateRequest](c.Context())
	if !ok {
		return
	}

	userID, ok := shared.UserIDFromCtx(c.Context())
	if !ok {
		c.NoContent(http.StatusUnauthorized)
		return
	}

	addresses := make([]domain.Address, len(req.Addresses))
	for i, a := range req.Addresses {
		addresses[i] = a.ToDomain()
	}

	contact, err := h.svc.Create(c.Context(), userID, 0, req.Name, req.Emails, req.Phones, addresses, req.Notes)
	if err != nil {
		if errors.Is(err, billingDomain.ErrQuotaExceeded) {
			shared.SendError(c, http.StatusForbidden, err.Error())
			return
		}
		shared.SendError(c, http.StatusInternalServerError, "failed to create contact", err)
		return
	}

	_ = c.JSON(http.StatusCreated, contact)
}

func (h *Handler) Update(c *kern.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		shared.SendError(c, http.StatusBadRequest, "invalid contact id")
		return
	}

	req, ok := xvalidator.ValidatedFromContext[application.UpdateRequest](c.Context())
	if !ok {
		return
	}

	userID, ok := shared.UserIDFromCtx(c.Context())
	if !ok {
		c.NoContent(http.StatusUnauthorized)
		return
	}

	addresses := make([]domain.Address, len(req.Addresses))
	for i, a := range req.Addresses {
		addresses[i] = a.ToDomain()
	}

	contact, err := h.svc.Update(c.Context(), uint(id), userID, req.Name, req.Emails, req.Phones, addresses, req.Notes)
	if err != nil {
		if err == domain.ErrContactNotFound {
			shared.SendError(c, http.StatusNotFound, "contact not found")
			return
		}
		if l := kern.LoggerFromContext(c.Context()); l != nil {
			l.Error("failed to update contact", "error", err, "contact_id", id)
		}
		shared.SendError(c, http.StatusInternalServerError, "failed to update contact", err)
		return
	}

	_ = c.JSON(http.StatusOK, contact)
}

func (h *Handler) Delete(c *kern.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		shared.SendError(c, http.StatusBadRequest, "invalid contact id")
		return
	}

	userID, ok := shared.UserIDFromCtx(c.Context())
	if !ok {
		c.NoContent(http.StatusUnauthorized)
		return
	}

	if err := h.svc.Delete(c.Context(), uint(id), userID); err != nil {
		if err == domain.ErrContactNotFound {
			shared.SendError(c, http.StatusNotFound, "contact not found")
			return
		}
		shared.SendError(c, http.StatusInternalServerError, "failed to delete contact", err)
		return
	}

	c.NoContent(http.StatusNoContent)
}

func (h *Handler) Import(c *kern.Context) {
	userID, ok := shared.UserIDFromCtx(c.Context())
	if !ok {
		c.NoContent(http.StatusUnauthorized)
		return
	}

	file, err := c.File("file")
	if err != nil {
		shared.SendError(c, http.StatusBadRequest, "file is required")
		return
	}

	format := c.DefaultQuery("format", "")
	if format == "" {
		format = fileExtToFormat(file.Filename)
	}
	if format != "csv" && format != "xlsx" {
		shared.SendError(c, http.StatusBadRequest, "unsupported format (use csv or xlsx)")
		return
	}

	f, err := file.Open()
	if err != nil {
		shared.SendError(c, http.StatusBadRequest, "failed to open file")
		return
	}
	defer func() { _ = f.Close() }()
	data, err := io.ReadAll(f)
	if err != nil {
		shared.SendError(c, http.StatusInternalServerError, "failed to read file", err)
		return
	}

	records, err := application.ParseImportRecords(data, format)
	if err != nil {
		shared.SendError(c, http.StatusBadRequest, "failed to parse file: "+err.Error())
		return
	}

	var collectionID uint
	if v, cerr := strconv.ParseUint(c.DefaultQuery("collection_id", "0"), 10, 64); cerr == nil {
		collectionID = uint(v)
	}

	result, err := h.svc.Import(c.Context(), userID, collectionID, records)
	if err != nil {
		if errors.Is(err, billingDomain.ErrQuotaExceeded) {
			shared.SendError(c, http.StatusForbidden, err.Error())
			return
		}
		shared.SendError(c, http.StatusInternalServerError, "failed to import contacts", err)
		return
	}

	_ = c.JSON(http.StatusOK, result)
}

func (h *Handler) Export(c *kern.Context) {
	userID, ok := shared.UserIDFromCtx(c.Context())
	if !ok {
		c.NoContent(http.StatusUnauthorized)
		return
	}

	format := c.DefaultQuery("format", "csv")
	if format != "csv" && format != "xlsx" {
		shared.SendError(c, http.StatusBadRequest, "unsupported format (use csv or xlsx)")
		return
	}

	var collectionID uint
	if v, cerr := strconv.ParseUint(c.DefaultQuery("collection_id", "0"), 10, 64); cerr == nil {
		collectionID = uint(v)
	}

	data, contentType, err := h.svc.Export(c.Context(), userID, collectionID, format)
	if err != nil {
		shared.SendError(c, http.StatusInternalServerError, "failed to export contacts", err)
		return
	}

	ext := "csv"
	if format == "xlsx" {
		ext = "xlsx"
	}
	c.SetHeader("Content-Disposition", fmt.Sprintf("attachment; filename=contacts.%s", ext))
	c.SetHeader("Content-Type", contentType)
	_ = c.Data(http.StatusOK, contentType, data)
}

func fileExtToFormat(filename string) string {
	switch strings.ToLower(filename) {
	case ".csv":
		return "csv"
	case ".xlsx":
		return "xlsx"
	}
	if i := strings.LastIndex(filename, "."); i >= 0 {
		switch strings.ToLower(filename[i+1:]) {
		case "csv":
			return "csv"
		case "xlsx":
			return "xlsx"
		}
	}
	return ""
}
