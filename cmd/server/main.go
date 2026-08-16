package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/sraj/addressbook/internal/app"
	"github.com/sraj/addressbook/internal/config"
	"github.com/sraj/addressbook/internal/shared"
	"github.com/sraj/addressbook/migrations"
	"github.com/mobentum/kern"
	"github.com/mobentum/kern/extensions/xlog"
	"github.com/mobentum/kern/extensions/xotel"
	"github.com/mobentum/kern/middleware"
)

func main() {

	slogger := xlog.NewLogger(xlog.Config{Format: "json", Level: slog.LevelInfo})
	slog.SetDefault(slogger)

	cfg, err := config.Load()
	if err != nil {
		slogger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	shutdownOTel, err := xotel.Setup(xotel.SetupConfig{
		ServiceName: "addressbook",
		Endpoint:    cfg.OTLPEndpoint,
		Insecure:    true,
		Metrics:     true,
	})
	if err != nil {
		slogger.Error("failed to setup OpenTelemetry", "error", err)
		os.Exit(1)
	}
	defer shutdownOTel()

	db, err := shared.NewDB(cfg.DatabaseDriver, cfg.DatabasePath)
	if err != nil {
		slogger.Error("failed to open database", "error", err)
		os.Exit(1)
	}

	if err := migrations.MigrateUp(db, cfg.DatabaseDriver); err != nil {
		slogger.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	a := app.New(db, cfg)

	server := kern.New(kern.WithSlogLogger(slogger))

	server.Use(middleware.RequestID())
	server.Use(xotel.Middleware(xotel.Config{ServiceName: "addressbook"}))
	server.Use(kern.Logger(kern.LoggerConfig{
		SLogger: slogger,
		Format:  "json",
		Fields:  map[string]interface{}{"log_type": "access"},
	}))
	server.Use(shared.Recovery())
	server.Use(kern.CORS(cfg.CORSOrigins))
	server.Use(middleware.SecurityHeaders(middleware.SecurityHeadersConfig{
		StrictTransportSecurity: "max-age=31536000; includeSubDomains",
		ContentSecurityPolicy:   "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self'; connect-src 'self'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'",
		ReferrerPolicy:          "strict-origin-when-cross-origin",
		PermissionsPolicy:       "camera=(), microphone=(), geolocation=(), payment=()",
	}))

	// Global rate limit: 200 requests/min per IP
	server.Use(middleware.RateLimiter(middleware.RateLimiterConfig{
		Requests: 200,
		Window:   time.Minute,
		Message:  "Too Many Requests",
	}))

	server.Use(middleware.Timeout(middleware.TimeoutConfig{Duration: 30 * time.Second}))

	// Health and readiness checks
	api := server.Group("/api")
	api.GET("/health", func(c *kern.Context) { _ = c.JSON(http.StatusOK, map[string]string{"status": "ok"}) })
	api.GET("/ready", func(c *kern.Context) {
		ctx := c.Context()
		if err := db.Ping(ctx); err != nil {
			_ = c.JSON(http.StatusServiceUnavailable, map[string]string{"status": "unavailable"})
			return
		}
		_ = c.JSON(http.StatusOK, map[string]string{"status": "ready"})
	})

	// Go runtime + process metrics for the observability stack
	registerMetrics(server)

	// Auth rate limit: 10 requests/min per IP (login, register, forgot-password)
	authRateLimit := middleware.RateLimiter(middleware.RateLimiterConfig{
		Requests: 10,
		Window:   time.Minute,
		Message:  "Too many attempts. Try again later.",
	})

	jwtAuth := shared.JWTAuth(a.TokenValidator)

	a.RegisterRoutes(server, jwtAuth, authRateLimit)

	server.Static("/", "./web/dist")

	slogger.Info("server started", "addr", cfg.Addr)
	if err := server.Run(cfg.Addr); err != nil {
		slogger.Error("server error", "error", err)
		os.Exit(1)
	}
}
