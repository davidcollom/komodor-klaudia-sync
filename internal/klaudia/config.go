package klaudia

import (
	"fmt"
	"strings"
)

func ValidateConfig(cfg Config) error {
	if strings.TrimSpace(cfg.Directory) == "" {
		return fmt.Errorf("directory is required")
	}
	if strings.TrimSpace(cfg.APIKey) == "" {
		return fmt.Errorf("api key is required")
	}
	if cfg.FileType != FileTypeKnowledgeBase && cfg.FileType != FileTypeBlueprints {
		return fmt.Errorf("file type must be %q or %q", FileTypeKnowledgeBase, FileTypeBlueprints)
	}
	if strings.TrimSpace(cfg.APIBaseURL) == "" {
		cfg.APIBaseURL = DefaultAPIBaseURL
	}
	return nil
}

func NormalizeExtensions(values []string) []string {
	if len(values) == 0 {
		return nil
	}

	normalized := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		value = strings.TrimSpace(strings.ToLower(value))
		if value == "" {
			continue
		}
		if !strings.HasPrefix(value, ".") {
			value = "." + value
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		normalized = append(normalized, value)
	}
	return normalized
}

func IsSupportedExtension(filename string) bool {
	_, ok := SupportedExtensions[strings.ToLower(Extension(filename))]
	return ok
}

func Extension(filename string) string {
	idx := strings.LastIndex(filename, ".")
	if idx < 0 {
		return ""
	}
	return filename[idx:]
}

func ShouldInclude(filename string, fileExtensions []string) bool {
	ext := strings.ToLower(Extension(filename))
	if !IsSupportedExtension(filename) {
		return false
	}
	if len(fileExtensions) == 0 {
		return true
	}
	for _, allowed := range fileExtensions {
		if ext == allowed {
			return true
		}
	}
	return false
}
