package klaudia

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClientListFilesIncludesRequestContextOnAPIError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":{"code":4004,"message":"bad request: 400 Bad Request","requestID":"req-123"}}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", server.Client())

	_, err := client.ListFiles(context.Background(), FileTypeBlueprints)
	require.Error(t, err)
	require.Contains(t, err.Error(), "api GET /api/v2/klaudia/files/blueprint failed")
	require.Contains(t, err.Error(), "request_id=req-123")
	require.NotContains(t, strings.ToLower(err.Error()), `{"error"`)
}

func TestNewRetryableHTTPClientRetriesServerErrors(t *testing.T) {
	t.Parallel()

	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusBadGateway)
			_, _ = w.Write([]byte(`temporary upstream failure`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"files":[]}`))
	}))
	defer server.Close()

	logger := NewLogger(io.Discard, "debug")
	client := NewClient(server.URL, "token", NewRetryableHTTPClient(logger, "debug"))

	files, err := client.ListFiles(context.Background(), FileTypeBlueprints)
	require.NoError(t, err)
	require.Empty(t, files)
	require.Equal(t, 3, attempts)
}
