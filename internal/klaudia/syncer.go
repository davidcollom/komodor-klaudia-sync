package klaudia

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"sort"

	"github.com/spf13/afero"
)

type Syncer struct {
	Client *Client
	Logger Logger
	FS     afero.Fs
}

func (s Syncer) Run(ctx context.Context, cfg Config) (SyncSummary, error) {
	logger := s.Logger
	if logger == nil {
		logger = NewStdLogger(os.Stdout)
	}
	if s.Client == nil {
		return SyncSummary{}, fmt.Errorf("client is required")
	}
	if err := ValidateConfig(cfg); err != nil {
		return SyncSummary{}, err
	}

	logger.Infof("Starting Klaudia sync")
	logger.Infof("Directory: %s", cfg.Directory)
	logger.Infof("File type: %s", cfg.FileType)
	logger.Infof("Recursive: %t", cfg.Recursive)
	logger.Infof("Dry run: %t", cfg.DryRun)
	logger.Infof("Supported formats: markdown, pdf, txt, doc/docx, csv, json, yaml")
	logger.Infof("Maximum file size: %d bytes", MaxFileSizeBytes)

	files, err := Scanner{
		Directory:        cfg.Directory,
		Recursive:        cfg.Recursive,
		FileExtensions:   NormalizeExtensions(cfg.FileExtensions),
		Fs:               s.FS,
		MaxFileSizeBytes: MaxFileSizeBytes,
		Logger:           logger,
	}.Scan()
	if err != nil {
		return SyncSummary{}, err
	}
	logger.Infof("Found %d local files", len(files))

	remoteFiles, err := s.Client.ListFiles(ctx, cfg.FileType)
	if err != nil {
		return SyncSummary{}, err
	}
	logger.Infof("Found %d remote files", len(remoteFiles))

	remoteByName := make(map[string]RemoteFile, len(remoteFiles))
	for _, remoteFile := range remoteFiles {
		remoteByName[remoteFile.Name] = remoteFile
	}

	var summary SyncSummary
	seenRemote := make(map[string]struct{}, len(remoteFiles))

	for _, file := range files {
		remoteFile, exists := remoteByName[file.RelativePath]
		if !exists {
			logger.Infof("✓ Uploading %s", file.RelativePath)
			if cfg.DryRun {
				summary.Uploaded++
				continue
			}
			if err := s.Client.UploadFile(ctx, cfg.FileType, file.RelativePath, file.Content); err != nil {
				logger.Errorf("Failed to upload %s: %v", file.RelativePath, err)
				summary.Failed++
				continue
			}
			summary.Uploaded++
			continue
		}

		seenRemote[remoteFile.ID] = struct{}{}
		remoteContent, err := s.Client.DownloadFile(ctx, cfg.FileType, remoteFile.ID)
		if err != nil {
			logger.Errorf("Failed to download %s for comparison: %v", file.RelativePath, err)
			summary.Failed++
			continue
		}
		remoteHash := sha256.Sum256(remoteContent)
		if file.Hash == hex.EncodeToString(remoteHash[:]) {
			logger.Debugf("Unchanged %s", file.RelativePath)
			summary.Unchanged++
			continue
		}

		logger.Infof("✓ Updating %s", file.RelativePath)
		if cfg.DryRun {
			summary.Updated++
			continue
		}
		if err := s.Client.UpdateFile(ctx, cfg.FileType, remoteFile.ID, file.RelativePath, file.Content); err != nil {
			logger.Errorf("Failed to update %s: %v", file.RelativePath, err)
			summary.Failed++
			continue
		}
		summary.Updated++
	}

	var deleteIDs []string
	for _, remoteFile := range remoteFiles {
		if _, ok := seenRemote[remoteFile.ID]; ok {
			continue
		}
		logger.Infof("✓ Deleting %s", remoteFile.Name)
		if cfg.DryRun {
			summary.Deleted++
			continue
		}
		deleteIDs = append(deleteIDs, remoteFile.ID)
	}

	if len(deleteIDs) > 0 && !cfg.DryRun {
		sort.Strings(deleteIDs)
		if err := s.Client.DeleteFiles(ctx, cfg.FileType, deleteIDs); err != nil {
			logger.Errorf("Failed to delete %d remote file(s): %v", len(deleteIDs), err)
			summary.Failed++
		} else {
			summary.Deleted += len(deleteIDs)
		}
	}

	logger.Infof("Sync summary: %d uploaded, %d updated, %d deleted, %d unchanged, %d failed", summary.Uploaded, summary.Updated, summary.Deleted, summary.Unchanged, summary.Failed)
	return summary, nil
}

func (s SyncSummary) String() string {
	return fmt.Sprintf("%d uploaded, %d updated, %d deleted, %d unchanged, %d failed", s.Uploaded, s.Updated, s.Deleted, s.Unchanged, s.Failed)
}
