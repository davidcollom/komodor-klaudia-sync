package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/davidcollom/komodor-klaudia-sync/internal/klaudia"
	"github.com/davidcollom/komodor-klaudia-sync/internal/version"
	"github.com/spf13/cobra"
)

func main() {
	if err := newRootCmd().ExecuteContext(context.Background()); err != nil {
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "klaudia-sync",
		Short:         "Sync a directory to Komodor Klaudia",
		Long:          "klaudia-sync synchronises a local directory with the Komodor Klaudia API (knowledge-base or blueprint) using full CRUD operations.",
		SilenceUsage:  true,
		SilenceErrors: false,
		// Default behaviour when no subcommand is given: run the sync.
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSync(cmd)
		},
	}
	root.PersistentFlags().StringP("log-level", "l", "info", "Logrus log level (trace, debug, info, warn, error, fatal, panic)")

	addSyncFlags(root)
	root.AddCommand(newVersionCmd())
	root.AddCommand(newSyncCmd())

	return root
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version and build information",
		Run: func(cmd *cobra.Command, args []string) {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "klaudia-sync %s (commit: %s, built: %s)\n",
				version.Version, version.Commit, version.Date)
		},
	}
}

func newSyncCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync a directory to Komodor Klaudia",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSync(cmd)
		},
	}
	addSyncFlags(cmd)
	return cmd
}

// addSyncFlags registers all sync-related flags on cmd, resolving defaults
// from environment variables. Sensitive values (api-key) use an empty default
// and are resolved at execution time to avoid exposing secrets in --help output.
func addSyncFlags(cmd *cobra.Command) {
	cmd.Flags().String("directory", envOr("", "KLAUDIA_DIRECTORY", "INPUT_DIRECTORY"), "Directory to sync")
	cmd.Flags().String("file-type", envOr(klaudia.FileTypeKnowledgeBase, "KLAUDIA_FILE_TYPE", "INPUT_FILE_TYPE", "INPUT_FILE-TYPE"), "knowledge-base or blueprint")
	cmd.Flags().String("api-key", "", "Komodor API key (or set KOMODOR_API_KEY)")
	cmd.Flags().String("api-base-url", envOr(klaudia.DefaultAPIBaseURL, "KOMODOR_API_BASE_URL", "INPUT_API_BASE_URL", "INPUT_API-BASE-URL"), "Komodor API base URL")
	cmd.Flags().Bool("recursive", envBool(true, "KLAUDIA_RECURSIVE", "INPUT_RECURSIVE"), "Recurse into subdirectories")
	cmd.Flags().Bool("dry-run", envBool(false, "KLAUDIA_DRY_RUN", "INPUT_DRY_RUN", "INPUT_DRY-RUN"), "Preview changes without applying them")
	cmd.Flags().Bool("debug", envBool(false, "KLAUDIA_DEBUG", "INPUT_DEBUG"), "Enable debug logging")
	cmd.Flags().String("file-extensions", envOr("", "KLAUDIA_FILE_EXTENSIONS", "INPUT_FILE_EXTENSIONS", "INPUT_FILE-EXTENSIONS"), "Comma-separated extension allowlist")
}

func runSync(cmd *cobra.Command) error {
	flags := cmd.Flags()

	directory, _ := flags.GetString("directory")
	fileType, _ := flags.GetString("file-type")
	apiKey, _ := flags.GetString("api-key")
	// Fall back to environment variable so Docker/CI callers can pass via env only.
	if strings.TrimSpace(apiKey) == "" {
		apiKey = envOr("", "KOMODOR_API_KEY", "INPUT_API_KEY", "INPUT_API-KEY")
	}
	apiBaseURL, _ := flags.GetString("api-base-url")
	recursive, _ := flags.GetBool("recursive")
	dryRun, _ := flags.GetBool("dry-run")
	logLevel, _ := flags.GetString("log-level")
	fileExtensionsStr, _ := flags.GetString("file-extensions")

	cfg := klaudia.Config{
		Directory:      strings.TrimSpace(directory),
		FileType:       strings.TrimSpace(fileType),
		APIKey:         strings.TrimSpace(apiKey),
		APIBaseURL:     strings.TrimSpace(apiBaseURL),
		Recursive:      recursive,
		DryRun:         dryRun,
		FileExtensions: splitExtensions(fileExtensionsStr),
	}

	logger := klaudia.NewLogger(os.Stdout, logLevel)
	client := klaudia.NewClient(cfg.APIBaseURL, cfg.APIKey, klaudia.NewRetryableHTTPClient(logger, logLevel))
	summary, err := klaudia.Syncer{Client: client, Logger: logger}.Run(cmd.Context(), cfg)
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Sync complete: %s\n", summary.String())
	return nil
}

func envOr(fallback string, keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return value
		}
	}
	return fallback
}

func envBool(fallback bool, keys ...string) bool {
	for _, key := range keys {
		switch strings.ToLower(strings.TrimSpace(os.Getenv(key))) {
		case "1", "true", "yes", "y", "on":
			return true
		case "0", "false", "no", "n", "off":
			return false
		}
	}
	return fallback
}

func splitExtensions(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return strings.Split(value, ",")
}
