// Package httpserver builds the pipeline-analytics HTTP handler tree.
package httpserver

import (
	"io/fs"
	"net/http"

	"github.com/alrayyes/pipeline-analytics/internal/auth"
	"github.com/alrayyes/pipeline-analytics/internal/ingestion"
	"github.com/alrayyes/pipeline-analytics/internal/metrics"
)

// Deps are New's dependencies.
type Deps struct {
	Registrar      *ingestion.Registrar
	IngestionStore ingestion.Store
	RunStore       ingestion.RunStore
	// Reconciler is called to poll a repo immediately when a Forgejo
	// webhook fires -- see webhooksHandler.forgejo.
	Reconciler ingestion.RepoReconciler
	Metrics    *metrics.Service
	Auth       *auth.Service
	AuthStore  auth.Store
	// GitHubRateLimits reports the rate-limit status last observed for a
	// GitHub token, for GET /api/insights/github-rate-limit. nil is fine --
	// the endpoint just reports every token with no status yet.
	GitHubRateLimits ingestion.RateLimitReporter
	// Version is reported by GET /api/version -- the build's tagged
	// version, or "dev" for a local build.
	Version string
	// Assets is the built frontend (internal/webassets.FS()), served for
	// every path the API doesn't claim, falling back to index.html for the
	// SPA client router.
	Assets fs.FS
}

// New returns the root HTTP handler.
func New(deps Deps) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("GET /api/version", versionHandler(deps.Version))

	repos := &reposHandler{registrar: deps.Registrar, store: deps.IngestionStore}
	mux.HandleFunc("GET /api/repos", repos.list)
	mux.HandleFunc("POST /api/repos", repos.register)
	mux.HandleFunc("POST /api/repos/discover", repos.discover)
	mux.HandleFunc("DELETE /api/repos/{repoId}", repos.untrack)

	usage := &usageHandler{service: deps.Metrics, repos: deps.IngestionStore}
	mux.HandleFunc("GET /api/repos/{repoId}/usage", usage.get)

	pipelines := &pipelinesHandler{service: deps.Metrics}
	mux.HandleFunc("GET /api/pipelines", pipelines.list)
	mux.HandleFunc("GET /api/pipelines/{pipelineId}", pipelines.get)
	mux.HandleFunc("GET /api/pipelines/{pipelineId}/steps", pipelines.steps)

	insights := &insightsHandler{repos: deps.IngestionStore, rateLimits: deps.GitHubRateLimits}
	mux.HandleFunc("GET /api/insights/github-rate-limit", insights.githubRateLimit)

	authH := &authHandler{service: deps.Auth}
	mux.HandleFunc("POST /api/auth/register/options", authH.registerOptions)
	mux.HandleFunc("POST /api/auth/register", authH.register)
	mux.HandleFunc("POST /api/auth/login/options", authH.loginOptions)
	mux.HandleFunc("POST /api/auth/login", authH.login)
	mux.HandleFunc("POST /api/auth/logout", authH.logout)

	webhooks := &webhooksHandler{store: deps.RunStore, reconciler: deps.Reconciler}
	mux.HandleFunc("POST /webhooks/github", webhooks.github)
	mux.HandleFunc("POST /webhooks/forgejo", webhooks.forgejo)

	mux.Handle("/", staticHandler(deps.Assets))

	return requireSession(deps.AuthStore, mux)
}
