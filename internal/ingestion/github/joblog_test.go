package github_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/alrayyes/pipeline-analytics/internal/ingestion"
	ghclient "github.com/alrayyes/pipeline-analytics/internal/ingestion/github"
	"github.com/stretchr/testify/require"
)

func logRequest(lines int) ingestion.JobLogRequest {
	return ingestion.JobLogRequest{
		Identifier: "alrayyes/pipeline-analytics",
		Token:      "ghp_test",
		ForgeJobID: "5001",
		Lines:      lines,
	}
}

func TestClient_JobLogTail(t *testing.T) {
	t.Parallel()

	t.Run("follows GitHub's redirect and returns the last lines, oldest first", func(t *testing.T) {
		t.Parallel()

		var gotAuth, gotRedirectAuth string

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/repos/alrayyes/pipeline-analytics/actions/jobs/5001/logs":
				gotAuth = r.Header.Get("Authorization")
				http.Redirect(w, r, "/download/5001", http.StatusFound)
			case "/download/5001":
				gotRedirectAuth = r.Header.Get("Authorization")
				_, _ = w.Write([]byte("one\ntwo\n\x1b[31mthree\x1b[0m\nfour\n"))
			default:
				t.Errorf("unexpected path %s", r.URL.Path)
			}
		}))
		defer server.Close()

		client, err := ghclient.NewClient(server.URL + "/")
		require.NoError(t, err)

		tail, err := client.JobLogTail(context.Background(), logRequest(2))
		require.NoError(t, err)

		require.Equal(t, []string{"\x1b[31mthree\x1b[0m", "four"}, tail.Lines)
		require.True(t, tail.Truncated)
		require.Equal(t, "Bearer ghp_test", gotAuth)
		require.NotEmpty(t, gotRedirectAuth, "same-host redirects keep the header; cross-host ones drop it")
	})

	t.Run("a short log isn't truncated", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte("only\n"))
		}))
		defer server.Close()

		client, err := ghclient.NewClient(server.URL + "/")
		require.NoError(t, err)

		tail, err := client.JobLogTail(context.Background(), logRequest(200))
		require.NoError(t, err)
		require.Equal(t, []string{"only"}, tail.Lines)
		require.False(t, tail.Truncated)
	})

	t.Run("a very long line is cut", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(strings.Repeat("x", ingestion.MaxLogLineLength+500) + "\n"))
		}))
		defer server.Close()

		client, err := ghclient.NewClient(server.URL + "/")
		require.NoError(t, err)

		tail, err := client.JobLogTail(context.Background(), logRequest(5))
		require.NoError(t, err)
		require.Len(t, tail.Lines[0], ingestion.MaxLogLineLength)
	})

	for _, tc := range []struct {
		name   string
		status int
		want   error
	}{
		{"a 404 means the log expired", http.StatusNotFound, ingestion.ErrLogExpired},
		{"a 410 means the log expired", http.StatusGone, ingestion.ErrLogExpired},
		{"a 403 means the token can't read it", http.StatusForbidden, ingestion.ErrLogForbidden},
		{"a 401 means the token can't read it", http.StatusUnauthorized, ingestion.ErrLogForbidden},
		{"a 500 means the forge is unreachable", http.StatusInternalServerError, ingestion.ErrLogUnreachable},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tc.status)
			}))
			defer server.Close()

			client, err := ghclient.NewClient(server.URL + "/")
			require.NoError(t, err)

			_, err = client.JobLogTail(context.Background(), logRequest(5))
			require.ErrorIs(t, err, tc.want)
		})
	}

	t.Run("a dead forge is unreachable", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.NotFoundHandler())
		url := server.URL + "/"
		server.Close()

		client, err := ghclient.NewClient(url)
		require.NoError(t, err)

		_, err = client.JobLogTail(context.Background(), logRequest(5))
		require.ErrorIs(t, err, ingestion.ErrLogUnreachable)
	})

	t.Run("an identifier that isn't owner/name is rejected", func(t *testing.T) {
		t.Parallel()

		client, err := ghclient.NewClient("")
		require.NoError(t, err)

		req := logRequest(5)
		req.Identifier = "no-slash"

		_, err = client.JobLogTail(context.Background(), req)
		require.ErrorIs(t, err, ghclient.ErrInvalidIdentifier)
	})

	t.Run("a job id that can't form a URL is an error, not a request", func(t *testing.T) {
		t.Parallel()

		client, err := ghclient.NewClient("http://127.0.0.1:1/")
		require.NoError(t, err)

		req := logRequest(5)
		req.ForgeJobID = "5\x7f1"

		_, err = client.JobLogTail(context.Background(), req)
		require.Error(t, err)
		require.NotErrorIs(t, err, ingestion.ErrLogUnreachable)
	})

	t.Run("a log with no final newline and CRLF endings keeps every line clean", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte("a\r\nb\r\nlast"))
		}))
		defer server.Close()

		client, err := ghclient.NewClient(server.URL + "/")
		require.NoError(t, err)

		tail, err := client.JobLogTail(context.Background(), logRequest(10))
		require.NoError(t, err)
		require.Equal(t, []string{"a", "b", "last"}, tail.Lines)
	})

	t.Run("asking for no lines still returns the last one", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte("a\nb\n"))
		}))
		defer server.Close()

		client, err := ghclient.NewClient(server.URL + "/")
		require.NoError(t, err)

		tail, err := client.JobLogTail(context.Background(), logRequest(0))
		require.NoError(t, err)
		require.Equal(t, []string{"b"}, tail.Lines)
	})

	t.Run("a cut never splits a multibyte character", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(strings.Repeat("x", ingestion.MaxLogLineLength-1) + "\u00e9 and more\n"))
		}))
		defer server.Close()

		client, err := ghclient.NewClient(server.URL + "/")
		require.NoError(t, err)

		tail, err := client.JobLogTail(context.Background(), logRequest(5))
		require.NoError(t, err)
		require.True(t, utf8.ValidString(tail.Lines[0]))
		require.Equal(t, strings.Repeat("x", ingestion.MaxLogLineLength-1), tail.Lines[0])
	})

	t.Run("a body cut off mid-stream is unreachable", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Length", "1000")
			_, _ = w.Write([]byte("partial\n"))
		}))
		defer server.Close()

		client, err := ghclient.NewClient(server.URL + "/")
		require.NoError(t, err)

		_, err = client.JobLogTail(context.Background(), logRequest(5))
		require.ErrorIs(t, err, ingestion.ErrLogUnreachable)
	})
}
