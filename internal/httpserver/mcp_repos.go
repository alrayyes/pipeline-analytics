package httpserver

import (
	"context"
	"fmt"

	"github.com/alrayyes/pipeline-analytics/internal/ingestion"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// addRepoTools registers the reads about the service itself and the repos it
// tracks, each matching a GET endpoint that serves the dashboard's settings
// and repos views. The listing and validation come from the same helpers the
// handlers use (repos.go, version.go), so a repo reads the same either way and
// the token stays masked.
func (h *mcpHandler) addRepoTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_version",
		Description: "Get the version of the running service. Matches GET /api/version.",
	}, h.getVersion)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_repos",
		Description: "List the repositories this service tracks, oldest first, each with its forge, identifier, ingestion status and masked token. Page with limit and offset. Matches GET /api/repos.",
	}, h.listRepos)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_repo_identifiers",
		Description: "List the identifiers of the repos tracked on one forge, for a Forgejo forge on one instance. A forge is required. Matches GET /api/repos/identifiers.",
	}, h.listRepoIdentifiers)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_github_rate_limit",
		Description: "Get, for each GitHub token this service holds, the repos it is used for and its most recently observed rate-limit status. The status comes from real API responses and is absent for a token not yet used. Matches GET /api/insights/github-rate-limit.",
	}, h.getGitHubRateLimit)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_settings",
		Description: "Get the signed-in account's dashboard settings, each resolved against its default, beside the server's defaults. Matches GET /api/settings.",
	}, h.getSettings)
}

type noInput struct{}

func (h *mcpHandler) getVersion(context.Context, *mcp.CallToolRequest, noInput) (*mcp.CallToolResult, versionDTO, error) {
	return nil, versionDTO{Version: h.version}, nil
}

type listReposInput struct {
	Forge  string `json:"forge,omitempty" jsonschema:"restrict to one forge, github or forgejo; omitted returns every forge"`
	Limit  int    `json:"limit,omitempty" jsonschema:"max repos to return; omitted returns every repo"`
	Offset int    `json:"offset,omitempty" jsonschema:"repos to skip before the returned page"`
}

func (h *mcpHandler) listRepos(ctx context.Context, _ *mcp.CallToolRequest, in listReposInput) (*mcp.CallToolResult, repoListDTO, error) {
	repos, hasMore, err := h.repos.ListRepos(ctx, ingestion.RepoListFilter{
		Forge:  ingestion.Forge(in.Forge),
		Limit:  in.Limit,
		Offset: in.Offset,
	})
	if err != nil {
		return nil, repoListDTO{}, fmt.Errorf("list repos: %w", err)
	}

	return nil, toRepoListDTO(repos, hasMore), nil
}

type listRepoIdentifiersInput struct {
	Forge              string `json:"forge" jsonschema:"github or forgejo"`
	ForgejoInstanceURL string `json:"forgejoInstanceUrl,omitempty" jsonschema:"for forgejo, restrict to repos on this instance"`
}

func (h *mcpHandler) listRepoIdentifiers(ctx context.Context, _ *mcp.CallToolRequest, in listRepoIdentifiersInput) (*mcp.CallToolResult, repoIdentifiersDTO, error) {
	if !isTrackedForge(in.Forge) {
		return nil, repoIdentifiersDTO{}, fmt.Errorf("%w: forge is required", errInvalidToolInput)
	}

	identifiers, err := h.repos.ListRepoIdentifiers(ctx, ingestion.Forge(in.Forge), in.ForgejoInstanceURL)
	if err != nil {
		return nil, repoIdentifiersDTO{}, fmt.Errorf("list repo identifiers: %w", err)
	}

	return nil, repoIdentifiersDTO{Identifiers: identifiers}, nil
}

// githubRateLimitResult wraps the endpoint's array, since an MCP tool's
// structured result is an object.
type githubRateLimitResult struct {
	Tokens []githubTokenUsageDTO `json:"tokens"`
}

func (h *mcpHandler) getGitHubRateLimit(ctx context.Context, _ *mcp.CallToolRequest, _ noInput) (*mcp.CallToolResult, githubRateLimitResult, error) {
	insights := &insightsHandler{repos: h.repos, rateLimits: h.rateLimits}

	dtos, err := insights.tokenUsage(ctx)
	if err != nil {
		return nil, githubRateLimitResult{}, fmt.Errorf("get github rate limit: %w", err)
	}

	return nil, githubRateLimitResult{Tokens: dtos}, nil
}

func (h *mcpHandler) getSettings(ctx context.Context, _ *mcp.CallToolRequest, _ noInput) (*mcp.CallToolResult, settingsDTO, error) {
	info, ok := authInfoFrom(ctx)
	if !ok || h.settings == nil {
		return nil, settingsDTO{}, fmt.Errorf("%w: settings are not available", errInvalidToolInput)
	}

	got, err := h.settings.Get(ctx, info.UserID)
	if err != nil {
		return nil, settingsDTO{}, fmt.Errorf("get settings: %w", err)
	}

	return nil, toSettingsDTO(got), nil
}
