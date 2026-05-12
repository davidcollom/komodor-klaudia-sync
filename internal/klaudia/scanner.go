package klaudia

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/afero"
)

type Scanner struct {
	Directory        string
	Recursive        bool
	FileExtensions   []string
	Fs               afero.Fs
	MaxFileSizeBytes int64
	Logger           Logger
}

func (s Scanner) Scan() ([]LocalFile, error) {
	filesystem := s.Fs
	if filesystem == nil {
		filesystem = afero.NewOsFs()
	}

	logger := s.Logger
	if logger == nil {
		logger = NewStdLogger(os.Stdout)
	}
	maxFileSizeBytes := s.MaxFileSizeBytes
	if maxFileSizeBytes <= 0 {
		maxFileSizeBytes = MaxFileSizeBytes
	}

	var files []LocalFile
	var oversized []string

	walkFn := func(currentPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if currentPath != s.Directory && !s.Recursive {
				return filepath.SkipDir
			}
			return nil
		}

		relative, err := filepath.Rel(s.Directory, currentPath)
		if err != nil {
			return err
		}
		relative = filepath.ToSlash(relative)

		if !ShouldInclude(relative, s.FileExtensions) {
			if IsSupportedExtension(relative) {
				logger.Debugf("Skipped %s (extension not in filter)", relative)
			} else {
				logger.Warnf("Skipped %s (unsupported format: %s)", relative, strings.ToLower(Extension(relative)))
			}
			return nil
		}

		if info.Size() > maxFileSizeBytes {
			oversized = append(oversized, fmt.Sprintf("%s (%d bytes)", relative, info.Size()))
			return nil
		}

		content, err := afero.ReadFile(filesystem, currentPath)
		if err != nil {
			return err
		}

		hash := sha256.Sum256(content)
		files = append(files, LocalFile{
			RelativePath: relative,
			AbsolutePath: currentPath,
			Size:         info.Size(),
			Hash:         hex.EncodeToString(hash[:]),
			Content:      content,
		})
		logger.Debugf("Found %s (%d bytes, hash: %.8s)", relative, info.Size(), hex.EncodeToString(hash[:]))
		return nil
	}

	if err := afero.Walk(filesystem, s.Directory, walkFn); err != nil {
		return nil, err
	}

	sort.Slice(files, func(i, j int) bool { return files[i].RelativePath < files[j].RelativePath })

	if len(oversized) > 0 {
		return nil, fmt.Errorf("found %d file(s) larger than 52.4 MB:\n- %s", len(oversized), strings.Join(oversized, "\n- "))
	}

	return files, nil
}
