package klaudia

import (
	"testing"

	"github.com/spf13/afero"
)

func TestScannerSkipsUnsupportedAndAppliesFilter(t *testing.T) {
	fs := afero.NewMemMapFs()
	dir := "/docs"
	mustWriteMemFile(t, fs, dir+"/keep.md", []byte("hello"))
	mustWriteMemFile(t, fs, dir+"/skip.exe", []byte("binary"))
	mustWriteMemFile(t, fs, dir+"/ignore.txt", []byte("text"))

	files, err := (Scanner{Directory: dir, Recursive: true, FileExtensions: []string{".md"}, Fs: fs}).Scan()
	if err != nil {
		t.Fatalf("scan failed: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}
	if files[0].RelativePath != "keep.md" {
		t.Fatalf("unexpected file: %s", files[0].RelativePath)
	}
}

func TestScannerFailsOnOversizedFile(t *testing.T) {
	fs := afero.NewMemMapFs()
	dir := "/docs"
	mustWriteMemFile(t, fs, dir+"/large.md", []byte("too-big"))

	_, err := (Scanner{Directory: dir, Recursive: true, Fs: fs, MaxFileSizeBytes: 1}).Scan()
	if err == nil {
		t.Fatal("expected error for oversized file")
	}
}

func mustWriteMemFile(t *testing.T, fs afero.Fs, path string, content []byte) {
	t.Helper()
	if err := afero.WriteFile(fs, path, content, 0o600); err != nil {
		t.Fatalf("write failed: %v", err)
	}
}
