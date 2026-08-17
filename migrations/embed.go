package migrations

import (
	"embed"
	"io/fs"

	"github.com/mobentum/xdb"
)

// Context describes one bounded context's migration set. Each context tracks
// its own migration version independently via its own bookkeeping table, so
// teams adding migrations never collide on version numbers or ordering.
type Context struct {
	// Name is the directory name under sqlite/ and postgres/.
	Name string
	// Table is the per-context migration bookkeeping table name.
	Table string
}

// Contexts lists every bounded context, in migration run order. Auth must run
// before billing (billing's account_members references users).
var Contexts = []Context{
	{Name: "auth", Table: "schema_migrations_auth"},
	{Name: "contact", Table: "schema_migrations_contact"},
	{Name: "note", Table: "schema_migrations_note"},
	{Name: "bookmark", Table: "schema_migrations_bookmark"},
	{Name: "billing", Table: "schema_migrations_billing"},
	{Name: "collection", Table: "schema_migrations_collection"},
}

//go:embed sqlite
var sqliteFS embed.FS

//go:embed postgres
var postgresFS embed.FS

// FS returns the embedded filesystem for the given driver.
func FS(driver string) fs.FS {
	switch norm(driver) {
	case "postgres":
		return postgresFS
	default:
		return sqliteFS
	}
}

// Path returns the fs path for a context within the driver's embedded tree.
func Path(driver, context string) string {
	return norm(driver) + "/" + context
}

// norm maps the config driver value (sqlite3|postgres) to the embed directory
// name (sqlite|postgres).
func norm(driver string) string {
	switch driver {
	case "postgres":
		return "postgres"
	default:
		return "sqlite"
	}
}

// Table returns the migration bookkeeping table for a context.
func Table(context string) string {
	for _, c := range Contexts {
		if c.Name == context {
			return c.Table
		}
	}
	return ""
}

// MigrateUp runs every context's pending migrations, each against its own
// bookkeeping table, in the order listed in Contexts.
func MigrateUp(db *xdb.DB, driver string) error {
	for _, ctx := range Contexts {
		if err := db.MigrateUpOpts(FS(driver), Path(driver, ctx.Name), xdb.MigrateOptions{Table: ctx.Table}); err != nil {
			return &ContextError{Context: ctx.Name, Err: err}
		}
	}
	return nil
}

// MigrateDown rolls back every context's migrations in reverse order.
func MigrateDown(db *xdb.DB, driver string) error {
	for i := len(Contexts) - 1; i >= 0; i-- {
		ctx := Contexts[i]
		if err := db.MigrateDownOpts(FS(driver), Path(driver, ctx.Name), xdb.MigrateOptions{Table: ctx.Table}); err != nil {
			return &ContextError{Context: ctx.Name, Err: err}
		}
	}
	return nil
}

// ContextError wraps a migration failure with the context it happened in.
type ContextError struct {
	Context string
	Err     error
}

func (e *ContextError) Error() string {
	return "migrations[" + e.Context + "]: " + e.Err.Error()
}

func (e *ContextError) Unwrap() error { return e.Err }
