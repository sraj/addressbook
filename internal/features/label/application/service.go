package application

import (
	"context"
	"errors"
	"fmt"
	"html"
	"math"
	"strconv"
	"strings"

	"github.com/sraj/addressbook/internal/features/contact/domain"
	labelDomain "github.com/sraj/addressbook/internal/features/label/domain"
)

// ContactsReader is satisfied by *contact/application.Service.
type ContactsReader interface {
	ListAllByUser(ctx context.Context, userID, collectionID uint) ([]domain.Contact, error)
}

// CheckoutCreator is satisfied by *billing/infrastructure.StripeService when
// Stripe is configured; nil otherwise.
type CheckoutCreator interface {
	CreateOneTimeCheckoutSession(ctx context.Context, userEmail string, unitAmount int64, currency, productName string, quantity int64, metadata map[string]string) (string, string, error)
	GetCheckoutSession(ctx context.Context, sessionID string) (string, error)
}

type Service struct {
	repo     labelDomain.Repository
	contacts ContactsReader
	stripe   CheckoutCreator
	appURL   string
	// priceCents is the per-sheet price in minor units.
	priceCents int64
	currency   string
	// labelsPerSheet mirrors a standard Avery 5160 sheet (30 labels).
	labelsPerSheet int
}

type Options struct {
	PriceCents     int64
	Currency       string
	LabelsPerSheet int
}

func NewService(repo labelDomain.Repository, contacts ContactsReader, stripe CheckoutCreator, appURL string, opts Options) *Service {
	if opts.LabelsPerSheet <= 0 {
		opts.LabelsPerSheet = 30
	}
	if opts.Currency == "" {
		opts.Currency = "usd"
	}
	return &Service{
		repo:           repo,
		contacts:       contacts,
		stripe:         stripe,
		appURL:         appURL,
		priceCents:     opts.PriceCents,
		currency:       opts.Currency,
		labelsPerSheet: opts.LabelsPerSheet,
	}
}

// GenerateSheet builds a printable HTML label sheet for the given Avery
// template. If collectionID is 0, all of the user's contacts are used.
// Contacts without an address are skipped.
func (s *Service) GenerateSheet(ctx context.Context, userID, collectionID uint, format string) (string, error) {
	f, err := labelDomain.FormatByCode(format)
	if err != nil {
		return "", err
	}

	contacts, err := s.contacts.ListAllByUser(ctx, userID, collectionID)
	if err != nil {
		return "", err
	}

	var labels []string
	for _, c := range contacts {
		if len(c.Addresses) == 0 {
			continue
		}
		a := c.Addresses[0]
		lines := []string{html.EscapeString(c.Name), html.EscapeString(a.Line1)}
		if a.Line2 != "" {
			lines = append(lines, html.EscapeString(a.Line2))
		}
		lines = append(lines, html.EscapeString(fmt.Sprintf("%s, %s %s", a.City, a.State, a.Zip)))
		if a.Country != "" {
			lines = append(lines, html.EscapeString(a.Country))
		}
		labels = append(labels, strings.Join(lines, "<br>"))
	}

	return renderSheet(labels, f), nil
}

// CreateOrder creates a Stripe checkout for printed labels and records the
// order. It returns the order and the Stripe checkout URL.
func (s *Service) CreateOrder(ctx context.Context, userID, collectionID uint, userEmail, format string) (*labelDomain.LabelOrder, string, error) {
	if s.stripe == nil {
		return nil, "", errors.New("label printing is not configured")
	}

	f, err := labelDomain.FormatByCode(format)
	if err != nil {
		return nil, "", err
	}

	contacts, err := s.contacts.ListAllByUser(ctx, userID, collectionID)
	if err != nil {
		return nil, "", err
	}
	contactCount := 0
	for _, c := range contacts {
		if len(c.Addresses) > 0 {
			contactCount++
		}
	}
	if contactCount == 0 {
		return nil, "", errors.New("no contacts with addresses to print")
	}

	labelsPerSheet := f.LabelsPerSheet()
	if labelsPerSheet <= 0 {
		labelsPerSheet = s.labelsPerSheet
	}
	sheetCount := int(math.Ceil(float64(contactCount) / float64(labelsPerSheet)))
	if sheetCount < 1 {
		sheetCount = 1
	}
	amountCents := s.priceCents * int64(sheetCount)

	// Reserve the order row so the webhook can find it by session id.
	order := &labelDomain.LabelOrder{
		UserID:       userID,
		CollectionID: collectionID,
		ContactCount: contactCount,
		SheetCount:   sheetCount,
		AmountCents:  amountCents,
		Currency:     s.currency,
		Status:       labelDomain.StatusPending,
		LabelType:    f.Code,
	}
	if err := s.repo.Create(ctx, order); err != nil {
		return nil, "", err
	}

	meta := map[string]string{
		"order_type": "labels",
		"order_id":   fmt.Sprintf("%d", order.ID),
	}

	url, sessionID, err := s.stripe.CreateOneTimeCheckoutSession(ctx, userEmail, s.priceCents, s.currency, "Printed address labels", int64(sheetCount), meta)
	if err != nil {
		return nil, "", err
	}

	if err := s.repo.SetSessionID(ctx, order.ID, sessionID); err != nil {
		return nil, "", err
	}
	order.StripeSessionID = sessionID

	return order, url, nil
}

// MarkOrderPaidBySessionID updates an order's status after a successful
// Stripe payment.
func (s *Service) MarkOrderPaidBySessionID(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return labelDomain.ErrOrderNotFound
	}
	return s.repo.UpdateStatusBySessionID(ctx, sessionID, labelDomain.StatusPaid)
}

func (s *Service) ListOrders(ctx context.Context, userID uint) ([]labelDomain.LabelOrder, error) {
	return s.repo.ListByUser(ctx, userID)
}

// SupportedFormats lists the Avery templates available for print and ordering.
func (s *Service) SupportedFormats() []labelDomain.Format {
	return labelDomain.SupportedFormats()
}

// ConfirmOrder marks an order paid using a checkout session id. Called by the
// frontend after a successful redirect when the webhook is unavailable.
func (s *Service) ConfirmOrder(ctx context.Context, userID uint, sessionID string) (*labelDomain.LabelOrder, error) {
	if s.stripe == nil {
		return nil, errors.New("label printing is not configured")
	}
	status, err := s.stripe.GetCheckoutSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if status == "paid" {
		if err := s.repo.UpdateStatusBySessionID(ctx, sessionID, labelDomain.StatusPaid); err != nil {
			return nil, err
		}
	}
	order, err := s.repo.FindBySessionID(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if order.UserID != userID {
		return nil, labelDomain.ErrOrderNotFound
	}
	return order, nil
}

func renderSheet(labels []string, f labelDomain.Format) string {
	labelsPerSheet := f.LabelsPerSheet()
	if labelsPerSheet <= 0 {
		labelsPerSheet = 30
	}
	var b strings.Builder
	b.WriteString(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<title>Address Labels</title>
<style>
  @page { size: 8.5in 11in; margin: 0.5in; }
  * { box-sizing: border-box; }
  body { font-family: Arial, Helvetica, sans-serif; margin: 0; padding: 0.5in; }
  .sheet { width: 7.5in; }
  .row { display: flex; }
  .label {
    width: ` + f.Width + `;
    height: ` + f.Height + `;
    border: 1px solid #ddd;
    padding: ` + f.CellPadding + `;
    font-size: ` + strconv.Itoa(f.FontSizePx) + `px;
    line-height: 1.25;
    overflow: hidden;
    word-break: break-word;
  }
  .print-btn {
    position: fixed; top: 12px; right: 12px; z-index: 10;
    padding: 8px 16px; font-size: 14px; cursor: pointer;
  }
  @media print {
    .print-btn { display: none; }
    .label { border: none; }
  }
</style>
</head>
<body>
<button id="print-btn" class="print-btn">Print</button>
`)
	b.WriteString(`<div class="sheet">`)
	for i, label := range labels {
		if i%labelsPerSheet == 0 && i > 0 {
			b.WriteString(`</div><div class="sheet" style="page-break-before: always;">`)
		}
		if i%f.Columns == 0 {
			b.WriteString(`<div class="row">`)
		}
		b.WriteString(`<div class="label">`)
		b.WriteString(label)
		b.WriteString(`</div>`)
		if i%f.Columns == f.Columns-1 {
			b.WriteString(`</div>`)
		}
	}
	// Close an open row.
	if len(labels)%f.Columns != 0 {
		b.WriteString(`</div>`)
	}
	b.WriteString(`</div>`)
	b.WriteString(`<script>` + labelDomain.PrintScript + `</script>`)
	b.WriteString(`</body></html>`)
	return b.String()
}
