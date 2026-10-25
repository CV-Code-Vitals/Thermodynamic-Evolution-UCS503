package main

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const maxSnapshotArchiveBytes int64 = 50 << 20

var snapshotIDPattern = regexp.MustCompile(`^[a-f0-9]{32}$`)

type SnapshotArchiveReader interface {
	GetObject(context.Context, string) (io.ReadCloser, error)
}

type SnapshotWorkspaceProvider interface {
	ExtractSnapshot(context.Context, string) (string, func(), error)
}

type SnapshotStore struct {
	reader   SnapshotArchiveReader
	tempRoot string
}

type snapshotS3Client interface {
	GetObject(context.Context, *s3.GetObjectInput, ...func(*s3.Options)) (*s3.GetObjectOutput, error)
}

type s3SnapshotArchiveReader struct {
	client snapshotS3Client
	bucket string
}

func NewSnapshotStore(reader SnapshotArchiveReader, tempRoot string) *SnapshotStore {
	return &SnapshotStore{reader: reader, tempRoot: tempRoot}
}

func NewS3SnapshotStore(client snapshotS3Client, bucket, tempRoot string) *SnapshotStore {
	if client == nil || strings.TrimSpace(bucket) == "" {
		return NewSnapshotStore(nil, tempRoot)
	}
	return NewSnapshotStore(&s3SnapshotArchiveReader{client: client, bucket: bucket}, tempRoot)
}

func (r *s3SnapshotArchiveReader) GetObject(ctx context.Context, key string) (io.ReadCloser, error) {
	if r.client == nil || strings.TrimSpace(r.bucket) == "" {
		return nil, errors.New("source S3 client and bucket are required")
	}
	if !strings.HasPrefix(key, "snapshots/") || !strings.HasSuffix(key, "/source.zip") {
		return nil, errors.New("invalid derived snapshot object key")
	}
	output, err := r.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("read source snapshot from S3: %w", err)
	}
	if output == nil || output.Body == nil {
		return nil, errors.New("source snapshot S3 response has no body")
	}
	return output.Body, nil
}

func (s *SnapshotStore) ExtractSnapshot(ctx context.Context, snapshotID string) (string, func(), error) {
	if !snapshotIDPattern.MatchString(snapshotID) {
		return "", nil, errors.New("invalid source snapshot ID")
	}
	if s.reader == nil {
		return "", nil, errors.New("source snapshot reader is not configured")
	}

	key := "snapshots/" + snapshotID + "/source.zip"
	body, err := s.reader.GetObject(ctx, key)
	if err != nil {
		return "", nil, fmt.Errorf("read source snapshot: %w", err)
	}
	archive, readErr := io.ReadAll(io.LimitReader(body, maxSnapshotArchiveBytes+1))
	closeErr := body.Close()
	if readErr != nil {
		return "", nil, fmt.Errorf("read source snapshot body: %w", readErr)
	}
	if closeErr != nil {
		return "", nil, fmt.Errorf("close source snapshot body: %w", closeErr)
	}
	if int64(len(archive)) > maxSnapshotArchiveBytes {
		return "", nil, fmt.Errorf("source snapshot exceeds the %d-byte limit", maxSnapshotArchiveBytes)
	}
	if err := validateSnapshotArchive(archive); err != nil {
		return "", nil, fmt.Errorf("validate source snapshot: %w", err)
	}

	tempDir, err := os.MkdirTemp(s.tempRoot, "ucs503-snapshot-*")
	if err != nil {
		return "", nil, fmt.Errorf("create source snapshot workspace: %w", err)
	}
	var once sync.Once
	cleanup := func() {
		once.Do(func() {
			if err := os.RemoveAll(tempDir); err != nil {
				fmt.Fprintf(os.Stderr, "remove source snapshot workspace: %v\n", err)
			}
		})
	}

	archivePath := filepath.Join(tempDir, "source.zip")
	if err := os.WriteFile(archivePath, archive, 0o600); err != nil {
		cleanup()
		return "", nil, fmt.Errorf("write temporary source archive: %w", err)
	}
	extractDir := filepath.Join(tempDir, "source")
	if err := os.Mkdir(extractDir, 0o700); err != nil {
		cleanup()
		return "", nil, fmt.Errorf("create extracted source workspace: %w", err)
	}
	if err := extractZip(archivePath, extractDir); err != nil {
		cleanup()
		return "", nil, fmt.Errorf("extract source snapshot: %w", err)
	}
	if err := os.Remove(archivePath); err != nil {
		cleanup()
		return "", nil, fmt.Errorf("remove temporary source archive: %w", err)
	}

	workspace, err := repositoryRoot(extractDir)
	if err != nil {
		cleanup()
		return "", nil, err
	}
	return workspace, cleanup, nil
}

func repositoryRoot(extractDir string) (string, error) {
	entries, err := os.ReadDir(extractDir)
	if err != nil {
		return "", fmt.Errorf("inspect extracted source workspace: %w", err)
	}
	if len(entries) == 1 && entries[0].IsDir() {
		return filepath.Join(extractDir, entries[0].Name()), nil
	}
	return extractDir, nil
}

func validateSnapshotArchive(archive []byte) error {
	reader, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
	if err != nil {
		return fmt.Errorf("open source snapshot ZIP: %w", err)
	}
	if len(reader.File) > 10_000 {
		return errors.New("source snapshot contains too many entries")
	}
	for _, file := range reader.File {
		if file.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("source snapshot contains symlink %q", file.Name)
		}
	}
	return nil
}
