package github_test

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alrayyes/pipeline-analytics/internal/ingestion"
	ghclient "github.com/alrayyes/pipeline-analytics/internal/ingestion/github"
	"github.com/stretchr/testify/require"
)

// httptest.Server.Close calls CloseIdleConnections on http.DefaultTransport, so
// a client that shares it loses its own idle connection, and an in-flight
// request, whenever any other test's server closes. The client has to keep a
// connection pool of its own.
func TestClient_KeepsItsConnectionsWhenTheDefaultTransportClosesIdleOnes(t *testing.T) {
	t.Parallel()

	var opened, idle atomic.Int64

	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}))
	server.Config.ConnState = func(_ net.Conn, state http.ConnState) {
		switch state {
		case http.StateNew:
			opened.Add(1)
		case http.StateIdle:
			idle.Add(1)
		default:
		}
	}
	server.Start()
	t.Cleanup(server.Close)

	client, err := ghclient.NewClient(server.URL + "/")
	require.NoError(t, err)

	req := ingestion.RunActionRequest{Identifier: "alrayyes/pipeline-analytics", Token: "t", ForgeRunID: "1"}
	require.NoError(t, client.CancelRun(context.Background(), req))
	require.Eventually(t, func() bool { return idle.Load() == 1 }, 2*time.Second, time.Millisecond, "the first connection goes idle")

	if transport, ok := http.DefaultTransport.(*http.Transport); ok {
		transport.CloseIdleConnections()
	}

	require.NoError(t, client.CancelRun(context.Background(), req))
	require.EqualValues(t, 1, opened.Load(), "the second request reused the client's own connection")
}
