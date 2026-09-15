// Command pipeline-analytics runs the pipeline-analytics server.
package main

import (
	"context"
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
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

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

	if err := viper.BindPFlag("addr", cmd.Flags().Lookup("addr")); err != nil {
		panic(err)
	}

	if err := viper.BindPFlag("db", cmd.Flags().Lookup("db")); err != nil {
		panic(err)
	}

	viper.SetEnvPrefix("PIPELINE_ANALYTICS")
	viper.AutomaticEnv()

	return cmd
}

func runServe(ctx context.Context) error {
	cfg := config.Config{
		Addr:   viper.GetString("addr"),
		DBPath: viper.GetString("db"),
	}
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid config: %w", err)
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

	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           httpserver.New(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)

	go func() {
		slog.Info("listening", "addr", cfg.Addr)

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
