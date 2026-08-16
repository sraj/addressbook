package interfaces

import (
	billingApp "github.com/sraj/addressbook/internal/billing/application"
	"github.com/sraj/addressbook/internal/features/note/application"
	"github.com/sraj/addressbook/internal/features/note/infrastructure"
	"github.com/mobentum/xdb"
	"github.com/samber/do/v2"
)

// Provide registers the note service and handler into the injector.
func Provide(i do.Injector) {
	do.Provide(i, func(i do.Injector) (*application.Service, error) {
		db := do.MustInvoke[*xdb.DB](i)
		billing := do.MustInvoke[*billingApp.Service](i)
		return application.NewService(infrastructure.NewSQLiteRepo(db), billing), nil
	})

	do.Provide(i, func(i do.Injector) (*Handler, error) {
		svc := do.MustInvoke[*application.Service](i)
		return NewHandler(svc), nil
	})
}
