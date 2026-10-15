package httpserver

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/alrayyes/pipeline-analytics/internal/ingestion"
)

type rateLimitStatusDTO struct {
	Limit     int       `json:"limit"`
	Remaining int       `json:"remaining"`
	Used      int       `json:"used"`
	ResetAt   time.Time `json:"resetAt"`
}

type githubTokenUsageDTO struct {
	TokenMasked string              `json:"tokenMasked"`
	Repos       []string            `json:"repos"`
	Status      *rateLimitStatusDTO `json:"status,omitempty"`
}

type insightsHandler struct {
	repos      ingestion.Store
	rateLimits ingestion.RateLimitReporter
}

// githubTokenGroup accumulates every repo tracked under one GitHub token,
// plus that token's most recently observed rate-limit status.
type githubTokenGroup struct {
	tokenMasked string
	identifiers []string
	status      *ingestion.RateLimitSnapshot
}

// githubRateLimit reports, per distinct GitHub token this app holds (one
// row per token, not per repo -- the same token often tracks more than one
// repo), which repos it's used for and its most recently observed
// rate-limit status. That status comes from a real API response's headers
// (ingestion.RateLimitReporter), never a dedicated poll -- GitHub's own
// guidance prefers this over calling GET /rate_limit.
func (h *insightsHandler) githubRateLimit(w http.ResponseWriter, r *http.Request) {
	dtos, err := h.tokenUsage(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "list repos")

		return
	}

	writeJSON(w, http.StatusOK, dtos)
}

// tokenUsage is the endpoint's body, shared with the MCP get_github_rate_limit
// tool so both group repos by token and report its status the same way.
func (h *insightsHandler) tokenUsage(ctx context.Context) ([]githubTokenUsageDTO, error) {
	repos, _, err := h.repos.ListRepos(ctx, ingestion.RepoListFilter{})
	if err != nil {
		return nil, fmt.Errorf("list repos: %w", err)
	}

	order, byToken := h.groupGitHubReposByToken(ctx, repos)

	return toGitHubTokenUsageDTOs(order, byToken), nil
}

// groupGitHubReposByToken decrypts each GitHub repo's token and groups repo
// identifiers by it, attaching whatever rate-limit status is known for that
// token. order preserves first-seen order, since a map alone wouldn't.
func (h *insightsHandler) groupGitHubReposByToken(
	ctx context.Context, repos []ingestion.Repo,
) (order []string, byToken map[string]*githubTokenGroup) {
	byToken = make(map[string]*githubTokenGroup)
	order = make([]string, 0, len(repos))

	for _, repo := range repos {
		if repo.Forge != ingestion.ForgeGitHub {
			continue
		}

		token, err := h.repos.RepoToken(ctx, repo.ID)
		if err != nil {
			continue
		}

		group, ok := byToken[token]
		if !ok {
			group = &githubTokenGroup{tokenMasked: repo.TokenMasked}
			byToken[token] = group
			order = append(order, token)
		}

		group.identifiers = append(group.identifiers, repo.Identifier)
		h.attachStatus(group, token)
	}

	return order, byToken
}

func (h *insightsHandler) attachStatus(group *githubTokenGroup, token string) {
	if group.status != nil || h.rateLimits == nil {
		return
	}

	if snapshot, ok := h.rateLimits.RateLimitFor(token); ok {
		group.status = &snapshot
	}
}

func toGitHubTokenUsageDTOs(order []string, byToken map[string]*githubTokenGroup) []githubTokenUsageDTO {
	dtos := make([]githubTokenUsageDTO, 0, len(order))

	for _, token := range order {
		group := byToken[token]
		dto := githubTokenUsageDTO{TokenMasked: group.tokenMasked, Repos: group.identifiers}

		if group.status != nil {
			dto.Status = &rateLimitStatusDTO{
				Limit:     group.status.Limit,
				Remaining: group.status.Remaining,
				Used:      group.status.Used,
				ResetAt:   group.status.ResetAt,
			}
		}

		dtos = append(dtos, dto)
	}

	return dtos
}
