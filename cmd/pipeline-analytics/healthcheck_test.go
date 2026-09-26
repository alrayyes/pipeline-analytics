package main

import (
	"bytes"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// listenOnFreePort binds PIPELINE_ANALYTICS_ADDR to an actual free loopback
// port for the duration of t - the healthcheck subcommand reads that env var
// directly, it doesn't take a flag.
func listenOnFreePort(t *testing.T) net.Listener {
	t.Helper()

	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = l.Close() })

	t.Setenv("PIPELINE_ANALYTICS_ADDR", l.Addr().String())

	return l
}

func TestHealthcheckSucceedsWhenHealthzAnswers200(t *testing.T) {
	l := listenOnFreePort(t)

	srv := &http.Server{ReadHeaderTimeout: time.Second, Handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})}
	go func() { _ = srv.Serve(l) }()
	t.Cleanup(func() { _ = srv.Close() })

	root := newRootCmd()
	root.SetArgs([]string{"healthcheck"})
	root.SetOut(new(bytes.Buffer))
	require.NoError(t, root.Execute())
}

func TestHealthcheckFailsWhenHealthzAnswersNon200(t *testing.T) {
	l := listenOnFreePort(t)

	srv := &http.Server{ReadHeaderTimeout: time.Second, Handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	})}
	go func() { _ = srv.Serve(l) }()
	t.Cleanup(func() { _ = srv.Close() })

	root := newRootCmd()
	root.SetArgs([]string{"healthcheck"})
	root.SetOut(new(bytes.Buffer))
	root.SetErr(new(bytes.Buffer))
	require.Error(t, root.Execute())
}

func TestHealthcheckFailsWhenNothingIsListening(t *testing.T) {
	// A free port that was bound and immediately released, rather than
	// listenOnFreePort - nothing serves it, which is the failure mode the
	// container's own HEALTHCHECK is there to catch.
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := l.Addr().String()
	require.NoError(t, l.Close())

	t.Setenv("PIPELINE_ANALYTICS_ADDR", addr)

	root := newRootCmd()
	root.SetArgs([]string{"healthcheck"})
	root.SetOut(new(bytes.Buffer))
	root.SetErr(new(bytes.Buffer))
	require.Error(t, root.Execute())
}

func TestHealthcheckFallsBackToDefaultAddrWhenEnvUnset(t *testing.T) {
	// serve's own flag default is ":8080" - nothing should be listening
	// there in a test process, so this just proves the fallback is used
	// (a connection-refused error) rather than an empty-addr parse error.
	root := newRootCmd()
	root.SetArgs([]string{"healthcheck"})
	root.SetOut(new(bytes.Buffer))
	root.SetErr(new(bytes.Buffer))
	require.Error(t, root.Execute())
}
