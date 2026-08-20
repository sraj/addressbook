package interfaces

import (
	"github.com/mobentum/xdb"
	"github.com/samber/do/v2"
	adminApp "github.com/sraj/addressbook/internal/admin/application"
	billingApp "github.com/sraj/addressbook/internal/billing/application"
)

// Provide registers the admin service and handler into the injector.
func Provide(i do.Injector) {
	do.Provide(i, func(i do.Injector) (*adminApp.Service, error) {
		db := do.MustInvoke[*xdb.DB](i)
		billing := do.MustInvoke[*billingApp.Service](i)
		return adminApp.NewService(db, billing), nil
	})

	do.Provide(i, func(i do.Injector) (*Handler, error) {
		svc := do.MustInvoke[*adminApp.Service](i)
		plans := do.MustInvoke[*billingApp.Service](i)
		return NewHandler(svc, plans), nil
	})
}
