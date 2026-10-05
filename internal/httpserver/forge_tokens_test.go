package httpserver_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alrayyes/pipeline-analytics/internal/auth"
	"github.com/alrayyes/pipeline-analytics/internal/httpserver"
	"github.com/alrayyes/pipeline-analytics/internal/ingestion"
	ingestionsqlite "github.com/alrayyes/pipeline-analytics/internal/ingestion/sqlite"
	"github.com/stretchr/testify/require"
)

// tokenRecorder is a forge client that remembers which token each call
// carried, so a test can see whether the saved one was used.
type tokenRecorder struct {
	fakeForgeClient

	getRepoToken, webhookToken, discoverToken string
}

func (r *tokenRecorder) GetRepo(_ context.Context, req ingestion.GetRepoRequest) (ingestion.RepoMetadata, error) {
	r.getRepoToken = req.Token

	return ingestion.RepoMetadata{}, nil
}

func (r *tokenRecorder) CreateWebhook(_ context.Context, req ingestion.CreateWebhookRequest) error {
	r.webhookToken = req.Token

	return nil
}

func (r *tokenRecorder) ListAccessibleRepos(_ context.Context, req ingestion.ListAccessibleReposRequest) ([]string, error) {
	r.discoverToken = req.Token

	return []string{"o/n"}, nil
}

func newTokenServer(t *testing.T) (testServer, *tokenRecorder) {
	t.Helper()

	srv, recorder, _ := newTokenServerWithAuthStore(t)

	return srv, recorder
}

// newTokenServerWithAuthStore also returns the auth store, for the test that
// needs a real API token.
func newTokenServerWithAuthStore(t *testing.T) (testServer, *tokenRecorder, auth.Store) {
	t.Helper()

	var authStore auth.Store

	recorder := &tokenRecorder{}

	srv := newTestServerWithDeps(t, func(deps *httpserver.Deps) {
		authStore = deps.AuthStore

		store, ok := deps.IngestionStore.(*ingestionsqlite.Store)
		require.True(t, ok)

		deps.ForgeTokens = store
		deps.Registrar = ingestion.NewRegistrar(store, map[ingestion.Forge]ingestion.ForgeClient{
			ingestion.ForgeGitHub:  recorder,
			ingestion.ForgeForgejo: recorder,
		}, "https://example.com")
	})

	return srv, recorder, authStore
}

func do(srv testServer, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, srv.authenticated(req))

	return rec
}

func saveToken(t *testing.T, srv testServer, body string) map[string]any {
	t.Helper()

	rec := do(srv, http.MethodPut, "/api/forge-tokens", body)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	var out map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &out))

	return out
}

func TestForgeTokens(t *testing.T) {
	t.Parallel()

	t.Run("saves a token and returns only its masked form, in every response", func(t *testing.T) {
		t.Parallel()

		srv, _ := newTokenServer(t)

		saved := saveToken(t, srv, `{"forge":"github","token":"ghp_supersecret1234"}`)
		require.Equal(t, "github", saved["forge"])
		require.Equal(t, "****1234", saved["tokenMasked"])
		require.NotEmpty(t, saved["id"])

		list := do(srv, http.MethodGet, "/api/forge-tokens", "")
		require.Equal(t, http.StatusOK, list.Code)
		require.NotContains(t, list.Body.String(), "ghp_supersecret1234")
		require.Contains(t, list.Body.String(), "****1234")
	})

	t.Run("saving again for the same forge replaces it", func(t *testing.T) {
		t.Parallel()

		srv, _ := newTokenServer(t)

		first := saveToken(t, srv, `{"forge":"github","token":"ghp_old00000"}`)
		second := saveToken(t, srv, `{"forge":"github","token":"ghp_new11111"}`)
		require.Equal(t, first["id"], second["id"])

		var list struct {
			Tokens []map[string]any `json:"tokens"`
		}

		require.NoError(t, json.Unmarshal(do(srv, http.MethodGet, "/api/forge-tokens", "").Body.Bytes(), &list))
		require.Len(t, list.Tokens, 1)
		require.Equal(t, "****1111", list.Tokens[0]["tokenMasked"])
	})

	t.Run("an empty list is an empty array, not null", func(t *testing.T) {
		t.Parallel()

		srv, _ := newTokenServer(t)

		require.JSONEq(t, `{"tokens":[]}`, do(srv, http.MethodGet, "/api/forge-tokens", "").Body.String())
	})

	t.Run("deleting removes it, and an unknown id is not found", func(t *testing.T) {
		t.Parallel()

		srv, _ := newTokenServer(t)

		saved := saveToken(t, srv, `{"forge":"github","token":"ghp_supersecret1234"}`)
		id, _ := saved["id"].(string)

		require.Equal(t, http.StatusNoContent, do(srv, http.MethodDelete, "/api/forge-tokens/"+id, "").Code)
		require.Equal(t, http.StatusNotFound, do(srv, http.MethodDelete, "/api/forge-tokens/"+id, "").Code)
		require.JSONEq(t, `{"tokens":[]}`, do(srv, http.MethodGet, "/api/forge-tokens", "").Body.String())
	})

	t.Run("rejects a bad forge, no token or an over-long one", func(t *testing.T) {
		t.Parallel()

		srv, _ := newTokenServer(t)

		for _, body := range []string{
			`{"forge":"gitlab","token":"x"}`,
			`{"forge":"github"}`,
			`{"forge":"github","token":""}`,
			`{"forge":"github","token":"` + strings.Repeat("x", 1025) + `"}`,
			`{"forge":"forgejo","forgejoInstanceUrl":"` + strings.Repeat("x", 2049) + `","token":"t"}`,
		} {
			require.Equal(t, http.StatusBadRequest, do(srv, http.MethodPut, "/api/forge-tokens", body).Code, body)
		}
	})

	t.Run("a Forgejo token needs its instance URL", func(t *testing.T) {
		t.Parallel()

		srv, _ := newTokenServer(t)

		require.Equal(t, http.StatusBadRequest, do(srv, http.MethodPut, "/api/forge-tokens", `{"forge":"forgejo","token":"t"}`).Code)
	})

	t.Run("requires a session, and an API token isn't one", func(t *testing.T) {
		t.Parallel()

		srv, _, authStore := newTokenServerWithAuthStore(t)

		anon := httptest.NewRecorder()
		srv.ServeHTTP(anon, httptest.NewRequest(http.MethodGet, "/api/forge-tokens", nil))
		require.Equal(t, http.StatusUnauthorized, anon.Code)

		// A real API token authenticates elsewhere but isn't a session, so
		// none of the three routes answers it.
		userID, err := authStore.Session(context.Background(), srv.sessionCookie.Value)
		require.NoError(t, err)

		_, raw, err := authStore.CreateToken(context.Background(), userID)
		require.NoError(t, err)

		for _, call := range []struct{ method, path, body string }{
			{http.MethodGet, "/api/forge-tokens", ""},
			{http.MethodPut, "/api/forge-tokens", `{"forge":"github","token":"t"}`},
			{http.MethodDelete, "/api/forge-tokens/x", ""},
		} {
			req := httptest.NewRequest(call.method, call.path, strings.NewReader(call.body))
			req.Header.Set("Authorization", "Bearer "+raw)

			rec := httptest.NewRecorder()
			srv.ServeHTTP(rec, req)
			require.Equal(t, http.StatusUnauthorized, rec.Code, call.method+" "+call.path)
		}

		other := httptest.NewRequest(http.MethodGet, "/api/repos", nil)
		other.Header.Set("Authorization", "Bearer "+raw)

		otherRec := httptest.NewRecorder()
		srv.ServeHTTP(otherRec, other)
		require.NotEqual(t, http.StatusUnauthorized, otherRec.Code, "the token itself is valid")
	})
}

func TestSavedTokenUsedForRegistrationAndDiscovery(t *testing.T) {
	t.Parallel()

	t.Run("registering with no token uses the saved one for the forge", func(t *testing.T) {
		t.Parallel()

		srv, recorder := newTokenServer(t)
		saveToken(t, srv, `{"forge":"github","token":"ghp_supersecret1234"}`)

		rec := do(srv, http.MethodPost, "/api/repos", `{"forge":"github","identifier":"o/n"}`)
		require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
		require.Equal(t, "ghp_supersecret1234", recorder.getRepoToken)
		require.Equal(t, "ghp_supersecret1234", recorder.webhookToken)
		require.NotContains(t, rec.Body.String(), "ghp_supersecret1234")
		require.Contains(t, rec.Body.String(), "****1234")
	})

	t.Run("a token in the request wins over the saved one", func(t *testing.T) {
		t.Parallel()

		srv, recorder := newTokenServer(t)
		saveToken(t, srv, `{"forge":"github","token":"ghp_saved00000"}`)

		rec := do(srv, http.MethodPost, "/api/repos", `{"forge":"github","identifier":"o/n","token":"ghp_typed11111"}`)
		require.Equal(t, http.StatusCreated, rec.Code)
		require.Equal(t, "ghp_typed11111", recorder.getRepoToken)
	})

	t.Run("discovery with no token uses the saved one", func(t *testing.T) {
		t.Parallel()

		srv, recorder := newTokenServer(t)
		saveToken(t, srv, `{"forge":"github","token":"ghp_supersecret1234"}`)

		rec := do(srv, http.MethodPost, "/api/repos/discover", `{"forge":"github"}`)
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		require.Equal(t, "ghp_supersecret1234", recorder.discoverToken)
	})

	t.Run("a Forgejo token is looked up by instance", func(t *testing.T) {
		t.Parallel()

		srv, recorder := newTokenServer(t)
		saveToken(t, srv, `{"forge":"forgejo","forgejoInstanceUrl":"https://git.example","token":"fj_instance1111"}`)

		rec := do(srv, http.MethodPost, "/api/repos/discover", `{"forge":"forgejo","forgejoInstanceUrl":"https://git.example/"}`)
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		require.Equal(t, "fj_instance1111", recorder.discoverToken)

		other := do(srv, http.MethodPost, "/api/repos/discover", `{"forge":"forgejo","forgejoInstanceUrl":"https://other.example"}`)
		require.Equal(t, http.StatusBadRequest, other.Code)
	})

	t.Run("no token and none saved is a 400 with its own code", func(t *testing.T) {
		t.Parallel()

		srv, _ := newTokenServer(t)

		for _, call := range []struct{ path, body string }{
			{"/api/repos", `{"forge":"github","identifier":"o/n"}`},
			{"/api/repos/discover", `{"forge":"github"}`},
		} {
			rec := do(srv, http.MethodPost, call.path, call.body)
			require.Equal(t, http.StatusBadRequest, rec.Code, call.path)
			require.Contains(t, rec.Body.String(), "no_saved_token")
		}
	})

	t.Run("deleting the saved token means asking for one again", func(t *testing.T) {
		t.Parallel()

		srv, _ := newTokenServer(t)
		id, _ := saveToken(t, srv, `{"forge":"github","token":"ghp_supersecret1234"}`)["id"].(string)
		require.Equal(t, http.StatusNoContent, do(srv, http.MethodDelete, "/api/forge-tokens/"+id, "").Code)

		require.Equal(t, http.StatusBadRequest, do(srv, http.MethodPost, "/api/repos/discover", `{"forge":"github"}`).Code)
	})
}

var errTokenStoreDown = errors.New("token store down")

// failingTokenStore fails every call, for the 500 paths.
type failingTokenStore struct{}

func (failingTokenStore) SaveForgeToken(context.Context, ingestion.Forge, string, string) (ingestion.SavedToken, error) {
	return ingestion.SavedToken{}, errTokenStoreDown
}

func (failingTokenStore) ListForgeTokens(context.Context) ([]ingestion.SavedToken, error) {
	return nil, errTokenStoreDown
}

func (failingTokenStore) DeleteForgeToken(context.Context, string) error { return errTokenStoreDown }

func (failingTokenStore) ForgeToken(context.Context, ingestion.Forge, string) (string, error) {
	return "", errTokenStoreDown
}

func TestForgeTokens_StoreFailure(t *testing.T) {
	t.Parallel()

	srv := newTestServerWithDeps(t, func(deps *httpserver.Deps) { deps.ForgeTokens = failingTokenStore{} })

	for _, call := range []struct{ method, path, body string }{
		{http.MethodGet, "/api/forge-tokens", ""},
		{http.MethodPut, "/api/forge-tokens", `{"forge":"github","token":"t"}`},
		{http.MethodDelete, "/api/forge-tokens/x", ""},
		{http.MethodPost, "/api/repos", `{"forge":"github","identifier":"o/n"}`},
		{http.MethodPost, "/api/repos/discover", `{"forge":"github"}`},
	} {
		require.Equal(t, http.StatusInternalServerError, do(srv, call.method, call.path, call.body).Code, call.method+" "+call.path)
	}
}

func TestSavedTokens_NotServedWithoutAStore(t *testing.T) {
	t.Parallel()

	srv := newTestServer(t, nil)

	// An unrouted path falls through to the static shell, never to a handler.
	require.NotContains(t, do(srv, http.MethodGet, "/api/forge-tokens", "").Body.String(), `"tokens"`)
	require.Equal(t, http.StatusBadRequest, do(srv, http.MethodPost, "/api/repos/discover", `{"forge":"github"}`).Code, "a token must be given each time")
}
