// Package billing provides Stripe integration for subscription management,
// including checkout sessions, customer portal, invoice retrieval,
// webhook verification, and subscription cancellation.
package infrastructure

import (
	"context"
	"errors"
	"fmt"

	"github.com/sraj/addressbook/internal/billing/domain"
	"github.com/stripe/stripe-go/v79"
	portal "github.com/stripe/stripe-go/v79/billingportal/session"
	"github.com/stripe/stripe-go/v79/checkout/session"
	"github.com/stripe/stripe-go/v79/customer"
	"github.com/stripe/stripe-go/v79/invoice"
	stripesub "github.com/stripe/stripe-go/v79/subscription"
	"github.com/stripe/stripe-go/v79/webhook"
)

// InvoiceRow represents a Stripe invoice returned to the frontend.
type InvoiceRow struct {
	ID               string `json:"id"`
	AmountPaid       int64  `json:"amount_paid"`
	Currency         string `json:"currency"`
	Status           string `json:"status"`
	Created          int64  `json:"created"`
	PeriodStart      int64  `json:"period_start"`
	PeriodEnd        int64  `json:"period_end"`
	HostedInvoiceURL string `json:"hosted_invoice_url"`
	InvoicePDF       string `json:"invoice_pdf"`
	Number           string `json:"number"`
}

// StripeService wraps the Stripe API for subscription billing operations.
type StripeService struct {
	secretKey  string
	webhookKey string
	appURL     string
}

// NewStripeService creates a new Stripe client. Sets stripe.Key globally.
func NewStripeService(secretKey, webhookKey, appURL string) *StripeService {
	stripe.Key = secretKey
	return &StripeService{
		secretKey:  secretKey,
		webhookKey: webhookKey,
		appURL:     appURL,
	}
}

// CreateCheckoutSession creates a Stripe checkout session for a subscription plan.
// Returns the redirect URL to Stripe's hosted checkout page.
// Sets account_id and plan_name in metadata so the webhook can associate the purchase.
func (s *StripeService) CreateCheckoutSession(ctx context.Context, account *domain.Account, userEmail, priceID, planName string) (string, error) {
	if s == nil {
		return "", errors.New("stripe not configured")
	}
	meta := map[string]string{
		"account_id": fmt.Sprintf("%d", account.ID),
		"plan_name":  planName,
	}
	params := &stripe.CheckoutSessionParams{
		Mode:       stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		SuccessURL: stripe.String(s.appURL + "/settings/billing?session_id={CHECKOUT_SESSION_ID}"),
		CancelURL:  stripe.String(s.appURL + "/settings/billing"),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{Price: stripe.String(priceID), Quantity: stripe.Int64(1)},
		},
		Metadata: meta,
		SubscriptionData: &stripe.CheckoutSessionSubscriptionDataParams{
			Metadata: meta,
		},
	}

	if account.StripeCustomerID != "" {
		params.Customer = &account.StripeCustomerID
	} else {
		params.CustomerEmail = stripe.String(userEmail)
	}

	sess, err := session.New(params)
	if err != nil {
		return "", fmt.Errorf("create checkout session: %w", err)
	}

	if sess.Customer != nil && account.StripeCustomerID == "" {
		account.StripeCustomerID = sess.Customer.ID
	}

	return sess.URL, nil
}

// CreatePortalSession creates a Stripe Customer Portal session for managing
// subscriptions, payment methods, and invoices. Creates a Stripe customer
// first if one doesn't exist for this account.
func (s *StripeService) CreatePortalSession(ctx context.Context, account *domain.Account) (string, error) {
	if s == nil {
		return "", errors.New("stripe not configured")
	}
	if account.StripeCustomerID == "" {
		stripeCustomer, err := customer.New(&stripe.CustomerParams{
			Email: stripe.String(account.Name),
			Metadata: map[string]string{
				"account_id": fmt.Sprintf("%d", account.ID),
			},
		})
		if err != nil {
			return "", fmt.Errorf("create customer: %w", err)
		}
		account.StripeCustomerID = stripeCustomer.ID
	}

	params := &stripe.BillingPortalSessionParams{
		Customer:  stripe.String(account.StripeCustomerID),
		ReturnURL: stripe.String(s.appURL + "/settings/billing"),
	}

	ps, err := portal.New(params)
	if err != nil {
		return "", fmt.Errorf("create portal session: %w", err)
	}

	return ps.URL, nil
}

// ChangeSubscriptionPlan updates an existing Stripe subscription to a new price.
// Finds the first subscription item and updates its price with proration.
// Prevents duplicate subscriptions by modifying in-place.
func (s *StripeService) ChangeSubscriptionPlan(ctx context.Context, subscriptionID, newPriceID string) error {
	if s == nil {
		return errors.New("stripe not configured")
	}

	// Fetch the subscription to get the subscription item ID
	sub, err := stripesub.Get(subscriptionID, nil)
	if err != nil {
		return fmt.Errorf("get subscription: %w", err)
	}
	if len(sub.Items.Data) == 0 {
		return fmt.Errorf("subscription has no items")
	}

	itemID := sub.Items.Data[0].ID
	params := &stripe.SubscriptionParams{
		ProrationBehavior: stripe.String("always_invoice"),
		Items: []*stripe.SubscriptionItemsParams{
			{
				ID:    stripe.String(itemID),
				Price: stripe.String(newPriceID),
			},
		},
	}
	_, err = stripesub.Update(subscriptionID, params)
	return err
}

// CancelSubscription cancels the Stripe subscription immediately with proration.
// The customer receives a credit for unused time.
func (s *StripeService) CancelSubscription(ctx context.Context, subscriptionID string) error {
	if s == nil {
		return errors.New("stripe not configured")
	}
	params := &stripe.SubscriptionCancelParams{
		InvoiceNow: stripe.Bool(true),
		Prorate:    stripe.Bool(true),
	}
	_, err := stripesub.Cancel(subscriptionID, params)
	return err
}

// ListInvoices fetches the last 50 invoices for a Stripe customer.
func (s *StripeService) ListInvoices(ctx context.Context, customerID string) ([]InvoiceRow, error) {
	if s == nil {
		return nil, errors.New("stripe not configured")
	}

	params := &stripe.InvoiceListParams{
		Customer: &customerID,
	}
	params.Limit = stripe.Int64(50)

	var result []InvoiceRow
	it := invoice.List(params)
	for it.Next() {
		inv := it.Invoice()
		result = append(result, InvoiceRow{
			ID:               inv.ID,
			AmountPaid:       inv.AmountPaid,
			Currency:         string(inv.Currency),
			Status:           string(inv.Status),
			Created:          inv.Created,
			PeriodStart:      inv.PeriodStart,
			PeriodEnd:        inv.PeriodEnd,
			HostedInvoiceURL: inv.HostedInvoiceURL,
			InvoicePDF:       inv.InvoicePDF,
			Number:           inv.Number,
		})
	}
	return result, it.Err()
}

// VerifyWebhook validates a Stripe webhook signature and returns the parsed event.
// Uses IgnoreAPIVersionMismatch to handle version differences between Stripe CLI and SDK.
func (s *StripeService) VerifyWebhook(payload []byte, sigHeader string) (stripe.Event, error) {
	if s == nil {
		return stripe.Event{}, errors.New("stripe not configured")
	}
	event, err := webhook.ConstructEventWithOptions(payload, sigHeader, s.webhookKey, webhook.ConstructEventOptions{
		IgnoreAPIVersionMismatch: true,
	})
	if err != nil {
		return stripe.Event{}, fmt.Errorf("webhook verification failed: %w", err)
	}
	return event, nil
}
