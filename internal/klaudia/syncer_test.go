package klaudia

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spf13/afero"
)

func TestSyncerUploadsUpdatesAndDeletes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v2/klaudia/files/knowledge-base":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"files":[{"id":"old-id","name":"old.md","size":3},{"id":"same-id","name":"same.md","size":4}]}`))
		case r.Method == http.MethodGet && r.URL.Path == "/api/v2/klaudia/files/knowledge-base/same-id":
			_, _ = w.Write([]byte("same"))
		case r.Method == http.MethodGet && r.URL.Path == "/api/v2/klaudia/files/knowledge-base/old-id":
			_, _ = w.Write([]byte("old"))
		case r.Method == http.MethodPost && r.URL.Path == "/api/v2/klaudia/files/knowledge-base":
			w.WriteHeader(http.StatusCreated)
		case r.Method == http.MethodPut && r.URL.Path == "/api/v2/klaudia/files/knowledge-base/same-id":
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodDelete && r.URL.Path == "/api/v2/klaudia/files/knowledge-base":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"deletedFiles":["old-id"],"failedFiles":[]}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	fs := afero.NewMemMapFs()
	dir := "/workspace"
	mustWriteSyncFile(t, fs, dir+"/new.md", []byte("new content"))
	mustWriteSyncFile(t, fs, dir+"/same.md", []byte("same"))

	summary, err := Syncer{Client: NewClient(server.URL, "token", server.Client()), Logger: NewStdLogger(io.Discard), FS: fs}.Run(context.Background(), Config{
		Directory:  dir,
		FileType:   FileTypeKnowledgeBase,
		APIKey:     "token",
		APIBaseURL: server.URL,
		Recursive:  true,
	})
	if err != nil {
		t.Fatalf("sync failed: %v", err)
	}
	if summary.Uploaded != 1 || summary.Updated != 0 || summary.Deleted != 1 || summary.Unchanged != 1 || summary.Failed != 0 {
		t.Fatalf("unexpected summary: %+v", summary)
	}
}

func mustWriteSyncFile(t *testing.T, fs afero.Fs, path string, content []byte) {
	t.Helper()
	if err := afero.WriteFile(fs, path, content, 0o600); err != nil {
		t.Fatalf("write failed: %v", err)
	}
}
