// Package httpserver builds the pipeline-analytics HTTP handler tree.
package httpserver

import (
	"context"
	"io/fs"
	"log/slog"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/alrayyes/pipeline-analytics/internal/auth"
	"github.com/alrayyes/pipeline-analytics/internal/ingestion"
	"github.com/alrayyes/pipeline-analytics/internal/metrics"
	"github.com/alrayyes/pipeline-analytics/internal/settings"
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
	// ForgeTokens keeps the tokens saved for registering repos. nil is fine:
	// the routes aren't served and a token must then be given each time.
	ForgeTokens ingestion.ForgeTokenStore
	// JobLogs reads a job's log from its forge for GET /api/runs/{runId}/
	// jobs/{jobId}/log. nil is fine: the route then answers 404.
	JobLogs   *ingestion.JobLogService
	Auth      *auth.Service
	AuthStore auth.Store
	Settings  *settings.Service
	// GitHubRateLimits reports the rate-limit status last observed for a
	// GitHub token, for GET /api/insights/github-rate-limit. nil is fine --
	// the endpoint just reports every token with no status yet.
	GitHubRateLimits ingestion.RateLimitReporter
	// Ready reports whether the instance can serve, for GET /readyz --
	// in production, a database read. nil means always ready, so a Deps
	// built without one (most tests) still answers.
	Ready func(ctx context.Context) error
	// ReadyCacheTTL is how long /readyz reuses the last probe result, so a
	// router polling it can't hammer the dependency. Zero means
	// defaultReadyCacheTTL.
	ReadyCacheTTL time.Duration
	// Draining, once true, makes /readyz answer 503 without probing, so a
	// shutting-down instance leaves rotation while it keeps serving. nil
	// means never draining.
	Draining *atomic.Bool
	// Version is reported by GET /api/version -- the build's tagged
	// version, or "dev" for a local build.
	Version string
	// Assets is the built frontend (internal/webassets.FS()), served for
	// every path the API doesn't claim, falling back to index.html for the
	// SPA client router.
	Assets fs.FS
}

// mountTelemetry registers the reads behind the telemetry views: failure
// insights, the branch list and the flaky steps.
func mountTelemetry(mux *http.ServeMux, service *metrics.Service) {
	mux.HandleFunc("GET /api/insights/failures", (&failureInsightsHandler{service: service}).get)
	mux.HandleFunc("GET /api/branches", (&branchesHandler{service: service}).list)
	mux.HandleFunc("GET /api/steps/flaky", (&flakyStepsHandler{service: service}).list)
}

// mountForgeTokens registers the saved forge token routes, when there is a
// store to keep them in.
func mountForgeTokens(mux *http.ServeMux, store ingestion.ForgeTokenStore) {
	if store == nil {
		return
	}

	h := &forgeTokensHandler{store: store}
	mux.HandleFunc("GET /api/forge-tokens", h.list)
	mux.HandleFunc("PUT /api/forge-tokens", h.save)
	mux.HandleFunc("DELETE /api/forge-tokens/{tokenId}", h.delete)
}

// New returns the root HTTP handler.
func New(deps Deps) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/readyz", readyzHandler(deps.Ready, deps.ReadyCacheTTL, deps.Draining))
	mux.HandleFunc("GET /api/version", versionHandler(deps.Version))

	repos := &reposHandler{registrar: deps.Registrar, store: deps.IngestionStore, tokens: deps.ForgeTokens}
	mux.HandleFunc("GET /api/repos", repos.list)
	mux.HandleFunc("POST /api/repos", repos.register)
	mux.HandleFunc("GET /api/repos/identifiers", repos.identifiers)
	mux.HandleFunc("POST /api/repos/discover", repos.discover)
	mux.HandleFunc("DELETE /api/repos/{repoId}", repos.untrack)

	usage := &usageHandler{service: deps.Metrics, repos: deps.IngestionStore}
	mux.HandleFunc("GET /api/repos/{repoId}/usage", usage.get)

	pipelines := &pipelinesHandler{service: deps.Metrics}
	mux.HandleFunc("GET /api/pipelines", pipelines.list)
	mux.HandleFunc("GET /api/pipelines/{pipelineId}", pipelines.get)
	mux.HandleFunc("GET /api/pipelines/{pipelineId}/steps", pipelines.steps)
	mux.HandleFunc("GET /api/pipelines/{pipelineId}/flaky-runs", pipelines.flakyRuns)
	mux.HandleFunc("GET /api/runs/{runId}/steps", pipelines.runSteps)

	if deps.JobLogs != nil {
		mux.HandleFunc("GET /api/runs/{runId}/jobs/{jobId}/log", (&jobLogHandler{service: deps.JobLogs}).get)
	}

	mux.HandleFunc("GET /api/steps/unhealthy", pipelines.unhealthySteps)

	insights := &insightsHandler{repos: deps.IngestionStore, rateLimits: deps.GitHubRateLimits}
	mux.HandleFunc("GET /api/insights/github-rate-limit", insights.githubRateLimit)

	mountTelemetry(mux, deps.Metrics)

	mountForgeTokens(mux, deps.ForgeTokens)

	runList := &runListHandler{service: deps.Metrics}
	mux.HandleFunc("GET /api/runs", runList.list)

	registerAuthRoutes(mux, &authHandler{service: deps.Auth})

	settingsH := &settingsHandler{service: deps.Settings}
	mux.HandleFunc("GET /api/settings", settingsH.get)
	mux.HandleFunc("PATCH /api/settings", settingsH.patch)

	mux.Handle("/api/mcp", newMCPHandler(deps))

	webhooks := &webhooksHandler{store: deps.RunStore, reconciler: deps.Reconciler}
	mux.HandleFunc("POST /webhooks/github", webhooks.github)
	mux.HandleFunc("POST /webhooks/forgejo", webhooks.forgejo)

	mux.Handle("/", staticHandler(deps.Assets))

	return requestLogger(limitBodies(requireSession(deps.AuthStore, mux)))
}

// registerAuthRoutes wires every WebAuthn, token, and credential route --
// split out of New so that function's own route wiring stays under
// funlen's statement limit as the route list grows.
func registerAuthRoutes(mux *http.ServeMux, authH *authHandler) {
	mux.HandleFunc("POST /api/auth/register/options", authH.registerOptions)
	mux.HandleFunc("POST /api/auth/register", authH.register)
	mux.HandleFunc("POST /api/auth/login/options", authH.loginOptions)
	mux.HandleFunc("POST /api/auth/login", authH.login)
	mux.HandleFunc("POST /api/auth/logout", authH.logout)
	mux.HandleFunc("POST /api/auth/tokens", authH.issueToken)
	mux.HandleFunc("DELETE /api/auth/tokens/{tokenId}", authH.revokeToken)
	mux.HandleFunc("POST /api/auth/credentials/options", authH.addCredentialOptions)
	mux.HandleFunc("POST /api/auth/credentials", authH.addCredential)
	mux.HandleFunc("GET /api/auth/credentials", authH.listCredentials)
	mux.HandleFunc("DELETE /api/auth/credentials/{credentialId}", authH.revokeCredential)
}

// readyzTimeout bounds the readiness probe. It's under the healthcheck
// subcommand's own 2s request timeout, so a hung database still gets a 503
// back to the checker instead of the checker giving up first.
const readyzTimeout = time.Second

// defaultReadyCacheTTL is how long a readiness result is reused: long enough
// that polling every second costs one probe per window, short enough that a
// recovery or an outage shows within a probe interval or two.
const defaultReadyCacheTTL = 3 * time.Second

// readyzHandler answers whether the instance can serve: 200 when ready
// reports nil and the instance isn't draining, 503 otherwise. The result is
// cached for ttl. The cause goes to the log, never the response, since the
// endpoint is unauthenticated.
func readyzHandler(ready func(ctx context.Context) error, ttl time.Duration, draining *atomic.Bool) http.HandlerFunc {
	if ttl <= 0 {
		ttl = defaultReadyCacheTTL
	}

	var (
		mu        sync.Mutex
		checkedAt time.Time
		lastErr   error
	)

	// probe runs ready at most once per ttl. The lock is held through the
	// probe, so concurrent polls share one check instead of each running it.
	probe := func(ctx context.Context) error {
		mu.Lock()
		defer mu.Unlock()

		if !checkedAt.IsZero() && time.Since(checkedAt) < ttl {
			return lastErr
		}

		lastErr = ready(ctx)
		checkedAt = time.Now()

		return lastErr
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if draining != nil && draining.Load() {
			writeError(w, http.StatusServiceUnavailable, "not_ready", "not ready")

			return
		}

		if ready != nil {
			// Detached from the request: the result is shared through the
			// cache, so one client hanging up mustn't cache a cancellation
			// as an outage for everyone else.
			ctx, cancel := context.WithTimeout(context.WithoutCancel(r.Context()), readyzTimeout)
			defer cancel()

			if err := probe(ctx); err != nil {
				slog.ErrorContext(r.Context(), "readiness check failed", "error", err)
				writeError(w, http.StatusServiceUnavailable, "not_ready", "not ready")

				return
			}
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}
}
