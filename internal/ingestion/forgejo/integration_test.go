package forgejo_test

import (
	"context"
	"testing"

	"code.gitea.io/sdk/gitea"
	"github.com/alrayyes/pipeline-analytics/internal/ingestion"
	forgejoclient "github.com/alrayyes/pipeline-analytics/internal/ingestion/forgejo"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcforgejo "github.com/testcontainers/testcontainers-go/modules/forgejo"
)

// TestIntegration_ForgejoWebhookEvents pins github.com/alrayyes/pipeline-
// analytics#120's finding against a real Forgejo instance rather than only
// the mocked HTTP tests in client_test.go, which can't catch the forge
// itself silently dropping an event name from a hook it otherwise reports
// creating successfully.
func TestIntegration_ForgejoWebhookEvents(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping container-backed integration test in -short mode")
	}

	testcontainers.SkipIfProviderIsNotHealthy(t)

	ctx := context.Background()

	container, err := tcforgejo.Run(ctx, "codeberg.org/forgejo/forgejo:11")
	require.NoError(t, err)
	testcontainers.CleanupContainer(t, container)

	instanceURL, err := container.ConnectionString(ctx)
	require.NoError(t, err)

	owner := container.AdminUsername()
	token := adminAccessToken(t, instanceURL, owner, container.AdminPassword())

	t.Run("push persists as a subscribed event -- the current fix", func(t *testing.T) {
		t.Parallel()

		const repo = "webhook-probe-push"
		createRepo(t, instanceURL, token, repo)

		client, err := forgejoclient.NewClient()
		require.NoError(t, err)

		err = client.CreateWebhook(ctx, ingestion.CreateWebhookRequest{
			InstanceURL: instanceURL,
			Identifier:  owner + "/" + repo,
			Token:       token,
			CallbackURL: "https://example.com/webhooks/forgejo",
			Secret:      "shh",
		})
		require.NoError(t, err)

		require.Equal(t, []string{"push"}, repoHookEvents(t, instanceURL, token, owner, repo))
	})

	t.Run("workflow_run and workflow_job silently persist as zero events -- why this project no longer uses them", func(t *testing.T) {
		t.Parallel()

		const repo = "webhook-probe-legacy-events"
		createRepo(t, instanceURL, token, repo)

		api := forgejoAPI(t, instanceURL, token)

		_, _, err := api.CreateRepoHook(owner, repo, gitea.CreateHookOption{
			Type: gitea.HookTypeGitea,
			Config: map[string]string{
				"url":          "https://example.com/webhooks/forgejo",
				"content_type": "json",
			},
			Events: []string{"workflow_run", "workflow_job"},
			Active: true,
		})
		require.NoError(t, err)

		require.Empty(t, repoHookEvents(t, instanceURL, token, owner, repo))
	})
}

func forgejoAPI(t *testing.T, instanceURL, token string) *gitea.Client {
	t.Helper()

	api, err := gitea.NewClient(instanceURL,
		gitea.SetToken(token),
		gitea.SetGiteaVersion(""), // skip the live version-check request
	)
	require.NoError(t, err)

	return api
}

// adminAccessToken creates a fresh API token for the container's admin
// user via basic auth, the same way a real user would mint one to hand to
// this app -- CreateWebhook and friends are always called with a token,
// never basic auth.
func adminAccessToken(t *testing.T, instanceURL, username, password string) string {
	t.Helper()

	api, err := gitea.NewClient(instanceURL,
		gitea.SetBasicAuth(username, password),
		gitea.SetGiteaVersion(""),
	)
	require.NoError(t, err)

	token, _, err := api.CreateAccessToken(gitea.CreateAccessTokenOption{
		Name:   "pipeline-analytics-integration-test",
		Scopes: []gitea.AccessTokenScope{gitea.AccessTokenScopeAll},
	})
	require.NoError(t, err)

	return token.Token
}

func createRepo(t *testing.T, instanceURL, token, name string) {
	t.Helper()

	_, _, err := forgejoAPI(t, instanceURL, token).CreateRepo(gitea.CreateRepoOption{
		Name:     name,
		AutoInit: true,
	})
	require.NoError(t, err)
}

func repoHookEvents(t *testing.T, instanceURL, token, owner, repo string) []string {
	t.Helper()

	hooks, _, err := forgejoAPI(t, instanceURL, token).ListRepoHooks(owner, repo, gitea.ListHooksOptions{})
	require.NoError(t, err)
	require.Len(t, hooks, 1)

	return hooks[0].Events
}
