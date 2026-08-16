package interfaces

import (
	"github.com/sraj/addressbook/internal/auth/application"
	"github.com/sraj/addressbook/internal/auth/infrastructure"
	billingApp "github.com/sraj/addressbook/internal/billing/application"
	"github.com/sraj/addressbook/internal/config"
	"github.com/sraj/addressbook/internal/mailer"
	"github.com/sraj/addressbook/internal/shared"
	"github.com/mobentum/xdb"
	"github.com/samber/do/v2"
)

// Provide registers the auth service and handler into the injector.
func Provide(i do.Injector) {
	do.Provide(i, func(i do.Injector) (*application.Service, error) {
		db := do.MustInvoke[*xdb.DB](i)
		cfg := do.MustInvoke[*config.Config](i)
		return application.NewService(infrastructure.NewSQLiteRepo(db), cfg.JWTSecret), nil
	})

	do.Provide(i, func(i do.Injector) (*Handler, error) {
		svc := do.MustInvoke[*application.Service](i)
		billing := do.MustInvoke[*billingApp.Service](i)
		cfg := do.MustInvoke[*config.Config](i)

		var m mailSender
		if v, err := do.Invoke[*mailer.Mailer](i); err == nil {
			m = v
		}

		return NewHandler(svc, billing, m, cfg.AppURL, cfg.SecureCookie), nil
	})

	// The auth service doubles as the shared JWT token validator.
	do.As[*application.Service, shared.TokenValidator](i)
}
