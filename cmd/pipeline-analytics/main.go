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
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alrayyes/pipeline-analytics/internal/config"
	"github.com/alrayyes/pipeline-analytics/internal/db"
	"github.com/alrayyes/pipeline-analytics/internal/httpserver"
	"github.com/alrayyes/pipeline-analytics/internal/ingestion"
	forgejoclient "github.com/alrayyes/pipeline-analytics/internal/ingestion/forgejo"
	githubclient "github.com/alrayyes/pipeline-analytics/internal/ingestion/github"
	"github.com/alrayyes/pipeline-analytics/internal/ingestion/sqlite"
	"github.com/alrayyes/pipeline-analytics/internal/webassets"
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

	for _, name := range []string{"addr", "db", "callback-url", "encryption-key"} {
		if err := viper.BindPFlag(name, cmd.Flags().Lookup(name)); err != nil {
			panic(err)
		}
	}

	viper.SetEnvPrefix("PIPELINE_ANALYTICS")
	viper.AutomaticEnv()

	return cmd
}

func runServe(ctx context.Context) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

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

	handler, err := buildHandler(cfg, conn)
	if err != nil {
		return err
	}

	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	return serveUntilDone(ctx, srv)
}

func loadConfig() (config.Config, error) {
	encryptionKey, err := hex.DecodeString(viper.GetString("encryption-key"))
	if err != nil {
		return config.Config{}, fmt.Errorf("decode encryption key: %w", err)
	}

	cfg := config.Config{
		Addr:          viper.GetString("addr"),
		DBPath:        viper.GetString("db"),
		CallbackURL:   viper.GetString("callback-url"),
		EncryptionKey: encryptionKey,
	}
	if err := cfg.Validate(); err != nil {
		return config.Config{}, fmt.Errorf("invalid config: %w", err)
	}

	return cfg, nil
}

func buildHandler(cfg config.Config, conn *sql.DB) (http.Handler, error) {
	store := sqlite.NewStore(conn, cfg.EncryptionKey)

	githubForgeClient, err := githubclient.NewClient("")
	if err != nil {
		return nil, fmt.Errorf("create github client: %w", err)
	}

	forgejoForgeClient, err := forgejoclient.NewClient()
	if err != nil {
		return nil, fmt.Errorf("create forgejo client: %w", err)
	}

	registrar := ingestion.NewRegistrar(store, map[ingestion.Forge]ingestion.ForgeClient{
		ingestion.ForgeGitHub:  githubForgeClient,
		ingestion.ForgeForgejo: forgejoForgeClient,
	}, cfg.CallbackURL)

	assets, err := webassets.FS()
	if err != nil {
		return nil, fmt.Errorf("load embedded frontend: %w", err)
	}

	return httpserver.New(registrar, store, version, assets), nil
}

func serveUntilDone(ctx context.Context, srv *http.Server) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

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
