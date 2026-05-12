package klaudia

import "time"

const (
	DefaultAPIBaseURL       = "https://api.komodor.com"
	MaxFileSizeBytes  int64 = 54_945_382

	FileTypeKnowledgeBase = "knowledge-base"
	FileTypeBlueprints    = "blueprint"
	Version               = "dev"
)

var SupportedExtensions = map[string]struct{}{
	".md":       {},
	".markdown": {},
	".pdf":      {},
	".txt":      {},
	".doc":      {},
	".docx":     {},
	".csv":      {},
	".json":     {},
	".yaml":     {},
	".yml":      {},
}

type Config struct {
	Directory      string
	FileType       string
	APIKey         string
	APIBaseURL     string
	Recursive      bool
	DryRun         bool
	FileExtensions []string
}

type LocalFile struct {
	RelativePath string
	AbsolutePath string
	Size         int64
	Hash         string
	Content      []byte
}

type RemoteFile struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Size           int64     `json:"size"`
	UploadedAt     time.Time `json:"uploadedAt"`
	CreatedByEmail string    `json:"createdByEmail"`
}

type ListFilesResponse struct {
	Files []RemoteFile `json:"files"`
}

type SyncSummary struct {
	Uploaded  int
	Updated   int
	Deleted   int
	Skipped   int
	Unchanged int
	Failed    int
}
