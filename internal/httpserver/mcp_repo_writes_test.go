package httpserver_test

import (
	"context"
	"encoding/json"
	"maps"
	"net/http"
	"strings"
	"testing"

	"github.com/alrayyes/pipeline-analytics/internal/auth"
	"github.com/alrayyes/pipeline-analytics/internal/httpserver"
	"github.com/alrayyes/pipeline-analytics/internal/ingestion"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

// bearerSession connects an MCP client holding only an API token, for the same
// account as srv's session.
func bearerSession(t *testing.T, srv testServer, authStore auth.Store) *mcp.ClientSession {
	t.Helper()

	userID, err := authStore.Session(context.Background(), srv.sessionCookie.Value)
	require.NoError(t, err)

	_, raw, err := authStore.CreateToken(context.Background(), userID, 0)
	require.NoError(t, err)

	return connectMCPWithBearer(t, srv, raw)
}

// listedRepoIdentifiers is what GET /api/repos reports, so a test can see that
// a tool's write is visible through the endpoint.
func listedRepoIdentifiers(t *testing.T, srv testServer) []string {
	t.Helper()

	rec := do(srv, http.MethodGet, "/api/repos", "")
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	var body struct {
		Repos []struct {
			Identifier string `json:"identifier"`
		} `json:"repos"`
	}

	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))

	identifiers := make([]string, 0, len(body.Repos))
	for _, repo := range body.Repos {
		identifiers = append(identifiers, repo.Identifier)
	}

	return identifiers
}

type repoResult struct {
	ID          string `json:"id"`
	Forge       string `json:"forge"`
	Identifier  string `json:"identifier"`
	TokenMasked string `json:"tokenMasked"`
}

func TestMCPRegisterRepo(t *testing.T) {
	t.Parallel()

	args := func(extra map[string]any) map[string]any {
		in := map[string]any{"forge": "github", "identifier": "alrayyes/new-repo", "token": "ghp_supersecrettoken1234"}
		maps.Copy(in, extra)

		return in
	}

	t.Run("registers a repo and reports it with the token masked", func(t *testing.T) {
		t.Parallel()

		srv := newTestServer(t, nil)

		got := callTool[repoResult](t, connectMCP(t, srv), "register_repo", args(nil))

		require.Equal(t, "alrayyes/new-repo", got.Identifier)
		require.NotEmpty(t, got.ID)
		require.NotContains(t, got.TokenMasked, "supersecret")
		require.Contains(t, listedRepoIdentifiers(t, srv), "alrayyes/new-repo", "the endpoint sees what the tool registered")
	})

	t.Run("a repo that is already tracked is a conflict, as on the endpoint", func(t *testing.T) {
		t.Parallel()

		srv := newTestServer(t, nil)
		session := connectMCP(t, srv)
		callTool[repoResult](t, session, "register_repo", args(nil))

		require.Contains(t, toolFailure(t, session, "register_repo", args(nil)), "already_tracked")
	})

	t.Run("an archived repo is refused", func(t *testing.T) {
		t.Parallel()

		srv := newTestServerWithRepoMeta(t, ingestion.RepoMetadata{Archived: true})

		require.Contains(t, toolFailure(t, connectMCP(t, srv), "register_repo", args(nil)), "repo_archived")
	})

	t.Run("no token given and none saved says so", func(t *testing.T) {
		t.Parallel()

		srv := newTestServer(t, nil)

		in := args(nil)
		delete(in, "token")

		require.Contains(t, toolFailure(t, connectMCP(t, srv), "register_repo", in), "no_saved_token")
	})

	t.Run("an empty identifier is refused by the handler's own check", func(t *testing.T) {
		t.Parallel()

		srv := newTestServer(t, nil)

		require.Contains(t, toolFailure(t, connectMCP(t, srv), "register_repo", args(map[string]any{"identifier": ""})), "invalid_body")
	})

	t.Run("an identifier past the length limit is refused", func(t *testing.T) {
		t.Parallel()

		srv := newTestServer(t, nil)

		msg := toolFailure(t, connectMCP(t, srv), "register_repo", args(map[string]any{"identifier": strings.Repeat("a", 201)}))

		require.Contains(t, msg, "invalid_body")
	})

	t.Run("an API token may register, as the endpoint allows", func(t *testing.T) {
		t.Parallel()

		var authStore auth.Store

		srv := newTestServerWithDeps(t, func(deps *httpserver.Deps) { authStore = deps.AuthStore })

		got := callTool[repoResult](t, bearerSession(t, srv, authStore), "register_repo", args(nil))

		require.Equal(t, "alrayyes/new-repo", got.Identifier)
	})
}

func TestMCPDiscoverRepos(t *testing.T) {
	t.Parallel()

	t.Run("lists the repos the token can reach", func(t *testing.T) {
		t.Parallel()

		srv := newTestServerWithDiscovery(t, []string{"alrayyes/a", "alrayyes/b"}, nil)

		got := callTool[struct {
			Identifiers []string `json:"identifiers"`
		}](t, connectMCP(t, srv), "discover_repos", map[string]any{"forge": "github", "token": "ghp_x"})

		require.Equal(t, []string{"alrayyes/a", "alrayyes/b"}, got.Identifiers)
	})

	t.Run("a forge failure is reported as one", func(t *testing.T) {
		t.Parallel()

		srv := newTestServerWithDiscovery(t, nil, context.DeadlineExceeded)

		require.Contains(t, toolFailure(t, connectMCP(t, srv), "discover_repos", map[string]any{"forge": "github", "token": "ghp_x"}), "forge_error")
	})

	t.Run("an unknown forge is refused", func(t *testing.T) {
		t.Parallel()

		srv := newTestServerWithDiscovery(t, nil, nil)

		require.Contains(t, toolFailure(t, connectMCP(t, srv), "discover_repos", map[string]any{"forge": "gitlab", "token": "ghp_x"}), "invalid_body")
	})
}

func TestMCPUntrackRepo(t *testing.T) {
	t.Parallel()

	t.Run("removes the repo", func(t *testing.T) {
		t.Parallel()

		srv := newTestServer(t, nil)
		id := seedRepo(t, srv)

		got := callTool[map[string]any](t, connectMCP(t, srv), "untrack_repo", map[string]any{"repoId": id})

		require.Equal(t, true, got["untracked"])
		require.Empty(t, listedRepoIdentifiers(t, srv))
	})

	t.Run("an unknown repo is not found", func(t *testing.T) {
		t.Parallel()

		srv := newTestServer(t, nil)

		require.Contains(t, toolFailure(t, connectMCP(t, srv), "untrack_repo", map[string]any{"repoId": "no-such-repo"}), "not_found")
	})

	t.Run("an API token may untrack, as the endpoint allows", func(t *testing.T) {
		t.Parallel()

		var authStore auth.Store

		srv := newTestServerWithDeps(t, func(deps *httpserver.Deps) { authStore = deps.AuthStore })
		id := seedRepo(t, srv)

		got := callTool[map[string]any](t, bearerSession(t, srv, authStore), "untrack_repo", map[string]any{"repoId": id})

		require.Equal(t, true, got["untracked"])
	})
}
