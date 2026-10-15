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
