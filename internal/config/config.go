package config

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/mobentum/kern/extensions/xconfig"
)

type Config struct {
	DatabaseDriver    string
	DatabasePath      string
	Addr              string
	JWTSecret         string
	CORSOrigins       []string
	SecureCookie      bool
	MailProvider      string
	MailAPIKey        string
	MailDomain        string
	MailFrom          string
	MailFromName      string
	AppURL            string
	StripeSecretKey   string
	StripeWebhookKey  string
	OTLPEndpoint      string
	EnableOTLP         bool
}

func Load() (*Config, error) {
	// Load .env if it exists — don't fail if missing (Docker/production)
	if _, err := os.Stat(".env"); err == nil {
		_ = godotenv.Load()
	}

	loader, err := xconfig.New()
	if err != nil {
		return nil, err
	}

	return &Config{
		DatabaseDriver: loader.String("DATABASE_DRIVER", "sqlite3"),
		DatabasePath:   loader.String("DATABASE_PATH", "kern-addressbook.db"),
		Addr:           loader.String("ADDR", ":8080"),
		JWTSecret:      loader.String("JWT_SECRET", "kern-addressbook-dev-secret-change-in-production"),
		CORSOrigins:    loader.Strings("CORS_ORIGINS", []string{"http://localhost:5173", "http://localhost:3000"}),
		SecureCookie: func() bool {
			v, _ := loader.Bool("SECURE_COOKIE", false)
			return v
		}(),
		MailProvider: loader.String("MAIL_PROVIDER", ""),
		MailAPIKey:   loader.String("MAIL_API_KEY", ""),
		MailDomain:   loader.String("MAIL_DOMAIN", ""),
		MailFrom:     loader.String("MAIL_FROM", "noreply@addressbook.app"),
		MailFromName: loader.String("MAIL_FROM_NAME", "Address Book"),
		AppURL:            loader.String("APP_URL", "http://localhost:5173"),
		StripeSecretKey:   loader.String("STRIPE_SECRET_KEY", ""),
		StripeWebhookKey:  loader.String("STRIPE_WEBHOOK_SECRET", ""),
		EnableOTLP:        func() bool {
			v, _ := loader.Bool("ENABLE_OTLP", false)
			return v
		}(),
		OTLPEndpoint:      loader.String("OTLP_ENDPOINT", "localhost:4318"),
	}, nil
}
