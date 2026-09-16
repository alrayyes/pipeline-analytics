// Command pipeline-analytics runs the pipeline-analytics server.
package main

import (
	"context"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/alrayyes/pipeline-analytics/internal/auth"
	authsqlite "github.com/alrayyes/pipeline-analytics/internal/auth/sqlite"
	"github.com/alrayyes/pipeline-analytics/internal/config"
	"github.com/alrayyes/pipeline-analytics/internal/db"
	"github.com/alrayyes/pipeline-analytics/internal/httpserver"
	"github.com/alrayyes/pipeline-analytics/internal/ingestion"
	forgejoclient "github.com/alrayyes/pipeline-analytics/internal/ingestion/forgejo"
	githubclient "github.com/alrayyes/pipeline-analytics/internal/ingestion/github"
	ingestionsqlite "github.com/alrayyes/pipeline-analytics/internal/ingestion/sqlite"
	"github.com/alrayyes/pipeline-analytics/internal/metrics"
	metricssqlite "github.com/alrayyes/pipeline-analytics/internal/metrics/sqlite"
	"github.com/alrayyes/pipeline-analytics/internal/webassets"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// version is set via goreleaser's ldflags (-X main.version={{.Version}}) on
// a release build; a local `go build` leaves it at "dev".
var version = "dev"

func main() {
	if err := newRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "pipeline-analytics",
		Short: "Self-hosted pipeline analytics for GitHub Actions and Forgejo Actions",
	}
	root.AddCommand(newServeCmd())

	return root
}

func newServeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Run the pipeline-analytics server",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runServe(cmd.Context())
		},
	}

	cmd.Flags().String("addr", ":8080", "address to listen on")
	cmd.Flags().String("db", "pipeline-analytics.db", "path to the SQLite database file")
	cmd.Flags().String("callback-url", "", "this server's own public base URL, used for forge webhook callbacks")
	cmd.Flags().String("encryption-key", "", "hex-encoded 32-byte key repo tokens are encrypted under at rest")
	cmd.Flags().Duration("reconcile-interval", time.Hour, "how often to poll tracked repos for reconciliation (forge-ingestion/spec.md requires at least hourly)")
	cmd.Flags().String("log-level", "info", "log verbosity: debug, info, warn, or error")

	for _, name := range []string{"addr", "db", "callback-url", "encryption-key", "reconcile-interval", "log-level"} {
		if err := viper.BindPFlag(name, cmd.Flags().Lookup(name)); err != nil {
			panic(err)
		}
	}

	viper.SetEnvPrefix("PIPELINE_ANALYTICS")
	// A flag like "callback-url" otherwise maps to the literal env var name
	// PIPELINE_ANALYTICS_CALLBACK-URL, which no shell, Docker, or systemd
	// unit can set -- environment variable names can't contain a hyphen.
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	return cmd
}

func runServe(ctx context.Context) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: cfg.LogLevel})))

	conn, err := db.Open(cfg.DBPath)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}

	defer func() {
		if cerr := conn.Close(); cerr != nil {
			slog.Error("close database", "error", cerr)
		}
	}()

	if err := db.Migrate(ctx, conn); err != nil {
		return fmt.Errorf("migrate database: %w", err)
	}

	ingestionStore := ingestionsqlite.NewStore(conn, cfg.EncryptionKey)

	forgeClients, err := newForgeClients()
	if err != nil {
		return err
	}

	handler, err := buildHandler(cfg, conn, ingestionStore, forgeClients)
	if err != nil {
		return err
	}

	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	reconciler := ingestion.NewReconciler(ingestionStore, ingestionStore, forgeClients)
	go reconciler.Run(ctx, cfg.ReconcileInterval)

	return serveUntilDone(ctx, srv)
}

// newForgeClients returns the real GitHub and Forgejo ForgeClient adapters,
// shared between the HTTP handler's repo registrar and the reconciliation
// scheduler.
func newForgeClients() (map[ingestion.Forge]ingestion.ForgeClient, error) {
	githubForgeClient, err := githubclient.NewClient("")
	if err != nil {
		return nil, fmt.Errorf("create github client: %w", err)
	}

	forgejoForgeClient, err := forgejoclient.NewClient()
	if err != nil {
		return nil, fmt.Errorf("create forgejo client: %w", err)
	}

	return map[ingestion.Forge]ingestion.ForgeClient{
		ingestion.ForgeGitHub:  githubForgeClient,
		ingestion.ForgeForgejo: forgejoForgeClient,
	}, nil
}

func loadConfig() (config.Config, error) {
	encryptionKey, err := hex.DecodeString(viper.GetString("encryption-key"))
	if err != nil {
		return config.Config{}, fmt.Errorf("decode encryption key: %w", err)
	}

	logLevel, err := parseLogLevel(viper.GetString("log-level"))
	if err != nil {
		return config.Config{}, err
	}

	cfg := config.Config{
		Addr:              viper.GetString("addr"),
		DBPath:            viper.GetString("db"),
		CallbackURL:       viper.GetString("callback-url"),
		EncryptionKey:     encryptionKey,
		ReconcileInterval: viper.GetDuration("reconcile-interval"),
		LogLevel:          logLevel,
	}
	if err := cfg.Validate(); err != nil {
		return config.Config{}, fmt.Errorf("invalid config: %w", err)
	}

	return cfg, nil
}

// parseLogLevel parses the --log-level flag's value into a slog.Level.
// slog.Level's own UnmarshalText already accepts "debug"/"info"/"warn"/
// "error" case-insensitively (and a "level+offset" form, e.g. "info+2") --
// this just gives an invalid value a clearer error than the raw parse
// failure would.
func parseLogLevel(s string) (slog.Level, error) {
	var level slog.Level
	if err := level.UnmarshalText([]byte(s)); err != nil {
		return 0, fmt.Errorf("invalid log level %q (want debug, info, warn, or error): %w", s, err)
	}

	return level, nil
}

func buildHandler(cfg config.Config, conn *sql.DB, ingestionStore *ingestionsqlite.Store, forgeClients map[ingestion.Forge]ingestion.ForgeClient) (http.Handler, error) {
	registrar := ingestion.NewRegistrar(ingestionStore, forgeClients, cfg.CallbackURL)

	authStore := authsqlite.NewStore(conn)

	webAuthn, err := newWebAuthn(cfg.CallbackURL)
	if err != nil {
		return nil, fmt.Errorf("configure webauthn: %w", err)
	}

	assets, err := webassets.FS()
	if err != nil {
		return nil, fmt.Errorf("load embedded frontend: %w", err)
	}

	return httpserver.New(httpserver.Deps{
		Registrar:      registrar,
		IngestionStore: ingestionStore,
		RunStore:       ingestionStore,
		Metrics:        metrics.NewService(metricssqlite.NewStore(conn)),
		Auth:           auth.NewService(webAuthn, authStore),
		AuthStore:      authStore,
		Version:        version,
		Assets:         assets,
	}), nil
}

// newWebAuthn derives the WebAuthn Relying Party ID (the effective domain,
// no scheme or port) and allowed origin from the server's own public
// callback URL -- the dashboard is same-origin with itself, so there's
// nothing else to configure here.
func newWebAuthn(callbackURL string) (*webauthn.WebAuthn, error) {
	u, err := url.Parse(callbackURL)
	if err != nil {
		return nil, fmt.Errorf("parse callback url: %w", err)
	}

	web, err := webauthn.New(&webauthn.Config{
		RPID:          u.Hostname(),
		RPDisplayName: "pipeline-analytics",
		RPOrigins:     []string{callbackURL},
	})
	if err != nil {
		return nil, fmt.Errorf("create webauthn instance: %w", err)
	}

	return web, nil
}

// serveUntilDone runs srv until ctx is done, then shuts it down gracefully.
// ctx is expected to already carry signal-driven cancellation (runServe
// arms it once, shared with the reconciliation scheduler).
func serveUntilDone(ctx context.Context, srv *http.Server) error {
	errCh := make(chan error, 1)

	go func() {
		slog.Info("listening", "addr", srv.Addr)

		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("listen and serve: %w", err)

			return
		}

		errCh <- nil
	}()

	select {
	case <-ctx.Done():
		if err := srv.Shutdown(context.Background()); err != nil {
			return fmt.Errorf("shutdown: %w", err)
		}

		return nil
	case err := <-errCh:
		return err
	}
}
