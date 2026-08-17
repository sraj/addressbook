package app

import (
	"github.com/mobentum/kern"
	"github.com/mobentum/xdb"
	"github.com/samber/do/v2"
	adminIntf "github.com/sraj/addressbook/internal/admin/interfaces"
	authIntf "github.com/sraj/addressbook/internal/auth/interfaces"
	billingInfra "github.com/sraj/addressbook/internal/billing/infrastructure"
	billingIntf "github.com/sraj/addressbook/internal/billing/interfaces"
	"github.com/sraj/addressbook/internal/config"
	bookmarkIntf "github.com/sraj/addressbook/internal/features/bookmark/interfaces"
	collectionIntf "github.com/sraj/addressbook/internal/features/collection/interfaces"
	contactIntf "github.com/sraj/addressbook/internal/features/contact/interfaces"
	labelIntf "github.com/sraj/addressbook/internal/features/label/interfaces"
	noteIntf "github.com/sraj/addressbook/internal/features/note/interfaces"
	"github.com/sraj/addressbook/internal/mailer"
	"github.com/sraj/addressbook/internal/shared"
)

type App struct {
	AuthHandler       *authIntf.Handler
	ContactHandler    *contactIntf.Handler
	NoteHandler       *noteIntf.Handler
	BookmarkHandler   *bookmarkIntf.Handler
	BillingHandler    *billingIntf.Handler
	AdminHandler      *adminIntf.Handler
	CollectionHandler *collectionIntf.Handler
	LabelHandler      *labelIntf.Handler
	TokenValidator    shared.TokenValidator
}

// New is the composition root. It builds a samber/do container where each
// bounded context registers its own service + handler providers (see each
// context's register.go), then resolves the handlers by type.
func New(db *xdb.DB, cfg *config.Config) *App {
	injector := do.New(
		sharedInfra(db, cfg),
		authIntf.Provide,
		billingIntf.Provide,
		contactIntf.Provide,
		noteIntf.Provide,
		bookmarkIntf.Provide,
		collectionIntf.Provide,
		labelIntf.Provide,
		adminIntf.Provide,
	)

	return &App{
		AuthHandler:       do.MustInvoke[*authIntf.Handler](injector),
		ContactHandler:    do.MustInvoke[*contactIntf.Handler](injector),
		NoteHandler:       do.MustInvoke[*noteIntf.Handler](injector),
		BookmarkHandler:   do.MustInvoke[*bookmarkIntf.Handler](injector),
		BillingHandler:    do.MustInvoke[*billingIntf.Handler](injector),
		CollectionHandler: do.MustInvoke[*collectionIntf.Handler](injector),
		LabelHandler:      do.MustInvoke[*labelIntf.Handler](injector),
		AdminHandler:      do.MustInvoke[*adminIntf.Handler](injector),
		TokenValidator:    do.MustInvoke[shared.TokenValidator](injector),
	}
}

// sharedInfra registers dependencies shared across contexts: DB, config,
// mailer (when configured), and the Stripe client (when configured).
func sharedInfra(db *xdb.DB, cfg *config.Config) func(do.Injector) {
	return func(i do.Injector) {
		do.ProvideValue(i, db)
		do.ProvideValue(i, cfg)

		if m := buildMailer(cfg); m != nil {
			do.ProvideValue(i, m)
		}
		if cfg.StripeSecretKey != "" {
			do.ProvideValue(i, billingInfra.NewStripeService(cfg.StripeSecretKey, cfg.StripeWebhookKey, cfg.AppURL))
		}
	}
}

// buildMailer returns a mailer when email config is present, else nil.
func buildMailer(cfg *config.Config) *mailer.Mailer {
	if cfg.MailProvider == "" || cfg.MailAPIKey == "" {
		return nil
	}
	return mailer.New(mailer.Config{
		Provider: mailer.Provider(cfg.MailProvider),
		APIKey:   cfg.MailAPIKey,
		Domain:   cfg.MailDomain,
		From:     cfg.MailFrom,
		FromName: cfg.MailFromName,
		BaseURL:  cfg.AppURL,
	})
}

// RegisterRoutes wires every bounded context's HTTP routes into the server.
// This is the single place that lists contexts; adding a new context only
// touches this method (plus New for wiring).
func (a *App) RegisterRoutes(server *kern.App, jwtAuth kern.MiddlewareFunc, authRateLimit kern.MiddlewareFunc) {
	a.AuthHandler.RegisterRoutes(server, jwtAuth, authRateLimit)
	a.ContactHandler.RegisterRoutes(server, jwtAuth)
	a.NoteHandler.RegisterRoutes(server, jwtAuth)
	a.BookmarkHandler.RegisterRoutes(server, jwtAuth)
	a.CollectionHandler.RegisterRoutes(server, jwtAuth)
	a.LabelHandler.RegisterRoutes(server, jwtAuth)
	a.BillingHandler.RegisterRoutes(server, jwtAuth)
	a.AdminHandler.RegisterRoutes(server, jwtAuth)
}
