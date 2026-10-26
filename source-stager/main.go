package main

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

const (
	maxRepositoryURLBytes = 2048
	maxMetadataBytes      = 1 << 20
	snapshotTTL           = 24 * time.Hour
	snapshotPrefix        = "snapshots/"
)

var (
	ErrInvalidRepositoryURL = errors.New("repository URL must identify a public GitHub repository")
	ErrUnsafeGitHubAddress  = errors.New("GitHub hostname resolved to a non-public address")
)

type SourceLimits struct {
	MaxArchiveBytes  int64
	MaxExpandedBytes int64
	MaxEntries       int
	MaxFileBytes     uint64
}

func defaultSourceLimits() SourceLimits {
	return SourceLimits{
		MaxArchiveBytes:  50 << 20,
		MaxExpandedBytes: 256 << 20,
		MaxEntries:       10_000,
		MaxFileBytes:     16 << 20,
	}
}

type DNSResolver interface {
	LookupIPAddr(context.Context, string) ([]net.IPAddr, error)
}

type SnapshotObjectStore interface {
	PutSnapshot(context.Context, string, []byte) error
}

type RepositoryStagerConfig struct {
	Client   *http.Client
	Store    SnapshotObjectStore
	Resolver DNSResolver
	Limits   SourceLimits
}

type RepositoryStager struct {
	client   *http.Client
	store    SnapshotObjectStore
	resolver DNSResolver
	limits   SourceLimits
}

type repositoryMetadata struct {
	DefaultBranch string `json:"default_branch"`
	Size          int64  `json:"size"`
}

type StageRequest struct {
	RepositoryURL string `json:"repository_url"`
}

type StageResponse struct {
	SnapshotID string    `json:"snapshot_id"`
	ExpiresAt  time.Time `json:"expires_at"`
}

type s3PutObjectClient interface {
	PutObject(context.Context, *s3.PutObjectInput, ...func(*s3.Options)) (*s3.PutObjectOutput, error)
}

type s3SnapshotStore struct {
	client s3PutObjectClient
	bucket string
}

type defaultResolver struct{}

func (defaultResolver) LookupIPAddr(ctx context.Context, host string) ([]net.IPAddr, error) {
	return net.DefaultResolver.LookupIPAddr(ctx, host)
}

func main() {
	bucket := strings.TrimSpace(os.Getenv("SOURCE_BUCKET"))
	if bucket == "" {
		panic("SOURCE_BUCKET must be configured")
	}
	awsConfig, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		panic(fmt.Errorf("load AWS configuration: %w", err))
	}
	store := &s3SnapshotStore{
		client: s3.NewFromConfig(awsConfig),
		bucket: bucket,
	}
	stager := NewRepositoryStager(RepositoryStagerConfig{Store: store})
	lambda.Start(NewStageHandler(stager, time.Now))
}

func ValidateGitHubRepositoryURL(rawURL string) (string, string, error) {
	rawURL = strings.TrimSpace(rawURL)
	if len(rawURL) == 0 || len(rawURL) > maxRepositoryURLBytes {
		return "", "", ErrInvalidRepositoryURL
	}
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Scheme != "https" || !strings.EqualFold(parsed.Hostname(), "github.com") ||
		parsed.User != nil || parsed.Port() != "" || parsed.RawQuery != "" || parsed.Fragment != "" ||
		parsed.Opaque != "" || parsed.ForceQuery {
		return "", "", ErrInvalidRepositoryURL
	}

	segments := strings.Split(strings.Trim(parsed.EscapedPath(), "/"), "/")
	if len(segments) != 2 || strings.Contains(parsed.EscapedPath(), "%") {
		return "", "", ErrInvalidRepositoryURL
	}
	owner, repository := segments[0], strings.TrimSuffix(segments[1], ".git")
	if !validGitHubSegment(owner) || !validGitHubSegment(repository) {
		return "", "", ErrInvalidRepositoryURL
	}
	return owner, repository, nil
}

func validGitHubSegment(value string) bool {
	if value == "" || value == "." || value == ".." {
		return false
	}
	for _, character := range value {
		if (character >= 'a' && character <= 'z') || (character >= 'A' && character <= 'Z') ||
			(character >= '0' && character <= '9') || character == '-' || character == '_' || character == '.' {
			continue
		}
		return false
	}
	return true
}

func NewRepositoryStager(config RepositoryStagerConfig) *RepositoryStager {
	resolver := config.Resolver
	if resolver == nil {
		resolver = defaultResolver{}
	}
	limits := config.Limits
	defaults := defaultSourceLimits()
	if limits.MaxArchiveBytes <= 0 {
		limits.MaxArchiveBytes = defaults.MaxArchiveBytes
	}
	if limits.MaxExpandedBytes <= 0 {
		limits.MaxExpandedBytes = defaults.MaxExpandedBytes
	}
	if limits.MaxEntries <= 0 {
		limits.MaxEntries = defaults.MaxEntries
	}
	if limits.MaxFileBytes == 0 {
		limits.MaxFileBytes = defaults.MaxFileBytes
	}

	stager := &RepositoryStager{
		store:    config.Store,
		resolver: resolver,
		limits:   limits,
	}
	if config.Client == nil {
		stager.client = &http.Client{
			Timeout:   90 * time.Second,
			Transport: newPublicGitHubTransport(resolver),
			CheckRedirect: func(request *http.Request, via []*http.Request) error {
				return stager.validateRedirect(request, via)
			},
		}
	} else {
		clientCopy := *config.Client
		clientCopy.CheckRedirect = func(request *http.Request, via []*http.Request) error {
			return stager.validateRedirect(request, via)
		}
		stager.client = &clientCopy
	}
	return stager
}

func (s *RepositoryStager) StageRepository(ctx context.Context, repositoryURL string) (string, error) {
	if s.store == nil {
		return "", errors.New("source snapshot object store is not configured")
	}
	owner, repository, err := ValidateGitHubRepositoryURL(repositoryURL)
	if err != nil {
		return "", err
	}

	metadataURL := "https://api.github.com/repos/" + url.PathEscape(owner) + "/" + url.PathEscape(repository)
	metadataBody, err := s.getLimited(ctx, metadataURL, maxMetadataBytes, "repository metadata")
	if err != nil {
		return "", err
	}
	var metadata repositoryMetadata
	if err := json.Unmarshal(metadataBody, &metadata); err != nil {
		return "", fmt.Errorf("decode GitHub repository metadata: %w", err)
	}
	if strings.TrimSpace(metadata.DefaultBranch) == "" || metadata.Size < 0 {
		return "", errors.New("GitHub repository metadata is missing a valid default branch or size")
	}
	if metadata.Size > s.limits.MaxExpandedBytes/1024 {
		return "", errors.New("GitHub repository exceeds the expanded source limit")
	}

	archiveURL := metadataURL + "/zipball/" + url.PathEscape(metadata.DefaultBranch)
	archive, err := s.getLimited(ctx, archiveURL, s.limits.MaxArchiveBytes, "repository archive")
	if err != nil {
		return "", err
	}
	if err := s.validateArchive(archive); err != nil {
		return "", fmt.Errorf("validate GitHub source archive: %w", err)
	}

	idBytes := make([]byte, 16)
	if _, err := rand.Read(idBytes); err != nil {
		return "", fmt.Errorf("generate source snapshot ID: %w", err)
	}
	snapshotID := hex.EncodeToString(idBytes)
	objectKey := snapshotPrefix + snapshotID + "/source.zip"
	if err := s.store.PutSnapshot(ctx, objectKey, archive); err != nil {
		return "", fmt.Errorf("store GitHub source snapshot: %w", err)
	}
	return snapshotID, nil
}

func (s *RepositoryStager) getLimited(ctx context.Context, rawURL string, maximum int64, description string) ([]byte, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Scheme != "https" || parsed.User != nil {
		return nil, fmt.Errorf("%s URL is invalid", description)
	}
	if err := s.validatePublicHost(ctx, parsed.Hostname()); err != nil {
		return nil, err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, parsed.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create %s request: %w", description, err)
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("User-Agent", "ucs503-source-stager")
	response, err := s.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", description, err)
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("fetch %s returned HTTP %d", description, response.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, maximum+1))
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", description, err)
	}
	if int64(len(body)) > maximum {
		return nil, fmt.Errorf("%s exceeds the %d-byte limit", description, maximum)
	}
	return body, nil
}

func (s *RepositoryStager) validateRedirect(request *http.Request, via []*http.Request) error {
	if len(via) >= 3 {
		return errors.New("GitHub archive exceeded the redirect limit")
	}
	if request.URL.Scheme != "https" || request.URL.User != nil {
		return errors.New("GitHub archive redirect must use HTTPS without user info")
	}
	if port := request.URL.Port(); port != "" && port != "443" {
		return errors.New("GitHub archive redirect must use the standard HTTPS port")
	}
	return s.validatePublicHost(request.Context(), request.URL.Hostname())
}

func (s *RepositoryStager) validatePublicHost(ctx context.Context, host string) error {
	host = strings.ToLower(strings.TrimSuffix(host, "."))
	switch host {
	case "api.github.com", "github.com", "codeload.github.com":
	default:
		return fmt.Errorf("GitHub source host %q is not allowlisted", host)
	}
	addresses, err := s.resolver.LookupIPAddr(ctx, host)
	if err != nil {
		return fmt.Errorf("resolve GitHub source host %q: %w", host, err)
	}
	if len(addresses) == 0 {
		return fmt.Errorf("%w: host %q has no addresses", ErrUnsafeGitHubAddress, host)
	}
	for _, address := range addresses {
		if !isPublicAddress(address.IP) {
			return fmt.Errorf("%w: %s", ErrUnsafeGitHubAddress, address.IP)
		}
	}
	return nil
}

func isPublicAddress(ip net.IP) bool {
	if ip == nil {
		return false
	}
	address, ok := netip.AddrFromSlice(ip)
	if !ok {
		return false
	}
	address = address.Unmap()
	if !address.IsGlobalUnicast() || address.IsPrivate() || address.IsLoopback() ||
		address.IsLinkLocalUnicast() || address.IsLinkLocalMulticast() || address.IsMulticast() ||
		address.IsUnspecified() {
		return false
	}
	for _, rawPrefix := range []string{
		"0.0.0.0/8", "100.64.0.0/10", "192.0.0.0/24", "192.0.2.0/24",
		"192.88.99.0/24", "198.18.0.0/15", "198.51.100.0/24", "203.0.113.0/24",
		"240.0.0.0/4", "64:ff9b:1::/48", "100::/64", "2001::/23",
		"2001:db8::/32", "2002::/16",
	} {
		prefix, err := netip.ParsePrefix(rawPrefix)
		if err == nil && prefix.Contains(address) {
			return false
		}
	}
	return true
}

func newPublicGitHubTransport(resolver DNSResolver) *http.Transport {
	return &http.Transport{
		Proxy: nil,
		DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
			host, port, err := net.SplitHostPort(address)
			if err != nil {
				return nil, fmt.Errorf("parse GitHub source address: %w", err)
			}
			host = strings.ToLower(strings.TrimSuffix(host, "."))
			switch host {
			case "api.github.com", "github.com", "codeload.github.com":
			default:
				return nil, fmt.Errorf("GitHub source host %q is not allowlisted", host)
			}
			addresses, err := resolver.LookupIPAddr(ctx, host)
			if err != nil {
				return nil, fmt.Errorf("resolve GitHub source host %q: %w", host, err)
			}
			for _, resolved := range addresses {
				if !isPublicAddress(resolved.IP) {
					return nil, fmt.Errorf("%w: %s", ErrUnsafeGitHubAddress, resolved.IP)
				}
			}
			if len(addresses) == 0 {
				return nil, fmt.Errorf("%w: host %q has no addresses", ErrUnsafeGitHubAddress, host)
			}
			dialer := &net.Dialer{Timeout: 15 * time.Second, KeepAlive: 30 * time.Second}
			return dialer.DialContext(ctx, network, net.JoinHostPort(addresses[0].IP.String(), port))
		},
		TLSClientConfig:   &tls.Config{MinVersion: tls.VersionTLS12},
		ForceAttemptHTTP2: true,
	}
}

func (s *RepositoryStager) validateArchive(archive []byte) error {
	if int64(len(archive)) > s.limits.MaxArchiveBytes {
		return fmt.Errorf("archive exceeds %d compressed bytes", s.limits.MaxArchiveBytes)
	}
	reader, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
	if err != nil {
		return fmt.Errorf("open ZIP: %w", err)
	}
	if len(reader.File) > s.limits.MaxEntries {
		return fmt.Errorf("archive has more than %d entries", s.limits.MaxEntries)
	}

	var expandedBytes uint64
	seen := make(map[string]struct{}, len(reader.File))
	for _, file := range reader.File {
		name := file.Name
		if strings.Contains(name, "\\") || path.IsAbs(name) {
			return fmt.Errorf("unsafe archive path %q", name)
		}
		cleanName := path.Clean(strings.TrimSuffix(name, "/"))
		if cleanName == "." || cleanName == ".." || strings.HasPrefix(cleanName, "../") {
			return fmt.Errorf("unsafe archive path %q", name)
		}
		if _, exists := seen[cleanName]; exists {
			return fmt.Errorf("duplicate archive path %q", cleanName)
		}
		seen[cleanName] = struct{}{}
		if file.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("archive contains symlink %q", name)
		}
		if file.UncompressedSize64 > s.limits.MaxFileBytes {
			return fmt.Errorf("archive entry %q exceeds the per-file limit", name)
		}
		if file.UncompressedSize64 > uint64(s.limits.MaxExpandedBytes)-expandedBytes {
			return fmt.Errorf("archive exceeds the expanded size limit")
		}

		entry, err := file.Open()
		if err != nil {
			return fmt.Errorf("open archive entry %q: %w", name, err)
		}
		remaining := s.limits.MaxExpandedBytes - int64(expandedBytes)
		copied, copyErr := io.Copy(io.Discard, io.LimitReader(entry, remaining+1))
		closeErr := entry.Close()
		if copyErr != nil {
			return fmt.Errorf("read archive entry %q: %w", name, copyErr)
		}
		if closeErr != nil {
			return fmt.Errorf("close archive entry %q: %w", name, closeErr)
		}
		if copied < 0 || copied > remaining || uint64(copied) != file.UncompressedSize64 {
			return fmt.Errorf("archive entry %q exceeds declared or total expanded size", name)
		}
		expandedBytes += uint64(copied)
	}
	return nil
}

func NewStageHandler(stager *RepositoryStager, now func() time.Time) func(context.Context, StageRequest) (StageResponse, error) {
	return func(ctx context.Context, request StageRequest) (StageResponse, error) {
		snapshotID, err := stager.StageRepository(ctx, request.RepositoryURL)
		if err != nil {
			return StageResponse{}, err
		}
		return StageResponse{
			SnapshotID: snapshotID,
			ExpiresAt:  now().UTC().Add(snapshotTTL),
		}, nil
	}
}

func (s *s3SnapshotStore) PutSnapshot(ctx context.Context, key string, body []byte) error {
	if s.client == nil || strings.TrimSpace(s.bucket) == "" {
		return errors.New("source S3 bucket/client is not configured")
	}
	if !strings.HasPrefix(key, snapshotPrefix) || !strings.HasSuffix(key, "/source.zip") {
		return errors.New("invalid generated source snapshot key")
	}
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:               aws.String(s.bucket),
		Key:                  aws.String(key),
		Body:                 bytes.NewReader(body),
		ContentType:          aws.String("application/zip"),
		ServerSideEncryption: types.ServerSideEncryptionAes256,
	})
	if err != nil {
		return fmt.Errorf("put source snapshot %s: %w", key, err)
	}
	return nil
}
