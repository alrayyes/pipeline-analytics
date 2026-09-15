package httpserver_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alrayyes/pipeline-analytics/internal/httpserver"
	"github.com/stretchr/testify/require"
)

func TestHealthz(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	httpserver.New().ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "ok", rec.Body.String())
}
