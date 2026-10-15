package httpserver

import (
	"context"
	"errors"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// errRepoWritesUnavailable is returned by the repo-write tools in a build with
// no registrar.
var errRepoWritesUnavailable = errors.New("repo registration is not available")

// addRepoWriteTools registers the tools that start and stop tracking a repo and
// list what a token could track. They run the same functions as the endpoints
// (registerRepo, discoverRepos, untrackRepo in repos.go), so the validation, the
// token lookup and the conflict codes are the endpoints'. The endpoints are not
// session-only, so neither are the tools: any caller the MCP gate lets in may
// use them, an API token included. A registration without a token uses the one
// saved for the forge, exactly as the endpoint does.
func (h *mcpHandler) addRepoWriteTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "register_repo",
		Description: "Start tracking a repository: register it with a webhook and begin ingesting its runs. Pass a repo-scoped token, or omit it to use the one saved for the forge and instance. The token is never echoed back. Refused for an archived repo, a fork or a mirror, or one already tracked. Matches POST /api/repos.",
	}, h.registerRepoTool)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "discover_repos",
		Description: "List the repositories a token can reach on a forge, as identifiers to register. Pass a token, or omit it to use the one saved for the forge and instance. Matches POST /api/repos/discover.",
	}, h.discoverReposTool)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "untrack_repo",
		Description: "Stop tracking a repository and delete its stored runs, jobs and steps. This does not remove the webhook created on the forge. Matches DELETE /api/repos/{repoId}.",
	}, h.untrackRepoTool)
}

func problemError(p *apiProblem) error {
	return mcpToolError{code: p.code, message: p.message}
}

type registerRepoInput struct {
	Forge              string `json:"forge" jsonschema:"github or forgejo"`
	Identifier         string `json:"identifier" jsonschema:"the repository, as owner/name; at most 200 characters"`
	ForgejoInstanceURL string `json:"forgejoInstanceUrl,omitempty" jsonschema:"for forgejo, the instance's URL"`
	Token              string `json:"token,omitempty" jsonschema:"a repo-scoped personal access token; omitted uses the token saved for this forge and instance"`
}

func (h *mcpHandler) registerRepoTool(ctx context.Context, _ *mcp.CallToolRequest, in registerRepoInput) (*mcp.CallToolResult, repoDTO, error) {
	if h.repoWrites == nil || h.repoWrites.registrar == nil {
		return nil, repoDTO{}, errRepoWritesUnavailable
	}

	repo, problem := h.repoWrites.registerRepo(ctx, repoRegistrationDTO(in))
	if problem != nil {
		return nil, repoDTO{}, problemError(problem)
	}

	return nil, toRepoDTO(repo), nil
}

type discoverReposInput struct {
	Forge              string `json:"forge" jsonschema:"github or forgejo"`
	ForgejoInstanceURL string `json:"forgejoInstanceUrl,omitempty" jsonschema:"for forgejo, the instance's URL"`
	Token              string `json:"token,omitempty" jsonschema:"a personal access token; omitted uses the token saved for this forge and instance"`
}

func (h *mcpHandler) discoverReposTool(ctx context.Context, _ *mcp.CallToolRequest, in discoverReposInput) (*mcp.CallToolResult, repoIdentifiersDTO, error) {
	if h.repoWrites == nil || h.repoWrites.registrar == nil {
		return nil, repoIdentifiersDTO{}, errRepoWritesUnavailable
	}

	identifiers, problem := h.repoWrites.discoverRepos(ctx, repoDiscoveryDTO(in))
	if problem != nil {
		return nil, repoIdentifiersDTO{}, problemError(problem)
	}

	return nil, repoIdentifiersDTO{Identifiers: identifiers}, nil
}

type untrackRepoInput struct {
	RepoID string `json:"repoId" jsonschema:"the repository's id, as list_repos reports it"`
}

type untrackRepoResult struct {
	RepoID    string `json:"repoId"`
	Untracked bool   `json:"untracked"`
}

func (h *mcpHandler) untrackRepoTool(ctx context.Context, _ *mcp.CallToolRequest, in untrackRepoInput) (*mcp.CallToolResult, untrackRepoResult, error) {
	if h.repoWrites == nil || h.repoWrites.store == nil {
		return nil, untrackRepoResult{}, errRepoWritesUnavailable
	}

	if problem := h.repoWrites.untrackRepo(ctx, in.RepoID); problem != nil {
		return nil, untrackRepoResult{}, problemError(problem)
	}

	return nil, untrackRepoResult{RepoID: in.RepoID, Untracked: true}, nil
}
