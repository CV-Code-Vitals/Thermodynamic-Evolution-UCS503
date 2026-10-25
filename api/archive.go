package main

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode/utf16"
	"unicode/utf8"
)

const maxArchiveSize int64 = 50 * 1024 * 1024
const archiveURLLifetime = 5 * time.Minute
const pendingArchiveLifetime = 48 * time.Hour
const storageOperationTimeout = 20 * time.Second

var (
	ErrArchiveNotFound   = errors.New("archive not found")
	ErrArchiveNotPending = errors.New("archive upload is not pending")
	ErrArchiveNotReady   = errors.New("archive is not available")
	ErrInvalidUpload     = errors.New("uploaded file does not match the requested metadata")
)

type ArchiveEntry struct {
	ID           string `json:"id" dynamodbav:"id"`
	Filename     string `json:"filename,omitempty" dynamodbav:"filename,omitempty"`
	OriginalName string `json:"originalName" dynamodbav:"originalName"`
	ContentType  string `json:"contentType" dynamodbav:"contentType"`
	Size         int64  `json:"size" dynamodbav:"size"`
	Title        string `json:"title" dynamodbav:"title"`
	Version      string `json:"version" dynamodbav:"version"`
	Date         string `json:"date" dynamodbav:"date"`
	Summary      string `json:"summary" dynamodbav:"summary"`
	UploadedAt   string `json:"uploadedAt" dynamodbav:"uploadedAt"`
	State        string `json:"state" dynamodbav:"state"`
	URL          string `json:"url,omitempty" dynamodbav:"url,omitempty"`
	PendingKey   string `json:"-" dynamodbav:"pendingKey,omitempty"`
	ObjectKey    string `json:"-" dynamodbav:"objectKey"`
	ExpiresAt    int64  `json:"-" dynamodbav:"expiresAt,omitempty"`
}

type UploadMetadata struct {
	OriginalName string `json:"originalName"`
	ContentType  string `json:"contentType"`
	Title        string `json:"title"`
	Version      string `json:"version"`
	Date         string `json:"date"`
	Summary      string `json:"summary"`
	Size         int64  `json:"size"`
}

type PresignedUpload struct {
	UploadID  string            `json:"uploadId"`
	URL       string            `json:"url"`
	Fields    map[string]string `json:"fields,omitempty"`
	ExpiresAt time.Time         `json:"expiresAt"`
}

type PresignedDownload struct {
	URL       string    `json:"url"`
	ExpiresAt time.Time `json:"expiresAt"`
}

type ArchiveService interface {
	BeginUpload(context.Context, UploadMetadata) (PresignedUpload, error)
	CompleteUpload(context.Context, string) (ArchiveEntry, error)
	List(context.Context) ([]ArchiveEntry, error)
	Download(context.Context, string) (PresignedDownload, error)
	// Delete is idempotent: an already-absent archive is a successful deletion.
	Delete(context.Context, string) error
}

type ArchiveRepository interface {
	CreatePending(context.Context, ArchiveEntry) error
	Get(context.Context, string) (ArchiveEntry, error)
	ListAvailable(context.Context) ([]ArchiveEntry, error)
	SetStatus(context.Context, string, string) error
	Delete(context.Context, string) error
}

type ObjectStore interface {
	CreateUpload(context.Context, string, string, int64, time.Duration) (PresignedUpload, error)
	VerifyUpload(context.Context, string, string, string, int64) error
	PromoteUpload(context.Context, string, string) error
	PresignDownload(context.Context, string, time.Duration) (string, error)
	DeleteObject(context.Context, string) error
}

type archiveService struct {
	repository ArchiveRepository
	objects    ObjectStore
	logger     *slog.Logger
	now        func() time.Time
	newID      func() (string, error)
}

func newArchiveService(repository ArchiveRepository, objects ObjectStore) *archiveService {
	return newArchiveServiceWithLogger(repository, objects, slog.Default())
}

func newArchiveServiceWithLogger(repository ArchiveRepository, objects ObjectStore, logger *slog.Logger) *archiveService {
	if logger == nil {
		logger = slog.Default()
	}
	return &archiveService{
		repository: repository,
		objects:    objects,
		logger:     logger,
		now:        time.Now,
		newID:      newArchiveID,
	}
}

func (s *archiveService) BeginUpload(ctx context.Context, metadata UploadMetadata) (PresignedUpload, error) {
	if err := validateUploadMetadata(metadata.OriginalName, metadata.ContentType, metadata.Size); err != nil {
		return PresignedUpload{}, err
	}
	if err := validateArchiveMetadata(metadata.Title, metadata.Version, metadata.Date, metadata.Summary); err != nil {
		return PresignedUpload{}, err
	}

	id, err := s.newID()
	if err != nil {
		return PresignedUpload{}, fmt.Errorf("generate archive ID: %w", err)
	}
	objectID, err := s.newID()
	if err != nil {
		return PresignedUpload{}, fmt.Errorf("generate archive object key: %w", err)
	}
	now := s.now().UTC()
	entry := ArchiveEntry{
		ID:           id,
		Filename:     metadata.OriginalName,
		OriginalName: metadata.OriginalName,
		ContentType:  metadata.ContentType,
		Size:         metadata.Size,
		Title:        metadata.Title,
		Version:      metadata.Version,
		Date:         metadata.Date,
		Summary:      metadata.Summary,
		UploadedAt:   now.Format("2006-01-02T15:04:05.000000000Z"),
		State:        "pending",
		PendingKey:   "pending/" + id,
		ObjectKey:    "archives/" + objectID,
		ExpiresAt:    now.Add(pendingArchiveLifetime).Unix(),
	}
	if err := s.repositoryCall(ctx, func(callCtx context.Context) error {
		return s.repository.CreatePending(callCtx, entry)
	}); err != nil {
		return PresignedUpload{}, fmt.Errorf("create pending archive record: %w", err)
	}

	upload, err := s.objectsCall(ctx, func(callCtx context.Context) (PresignedUpload, error) {
		return s.objects.CreateUpload(callCtx, entry.PendingKey, entry.ContentType, entry.Size, archiveURLLifetime)
	})
	if err != nil {
		compensationErr := s.compensatePending(ctx, entry)
		return PresignedUpload{}, errors.Join(fmt.Errorf("create presigned upload: %w", err), compensationErr)
	}
	upload.UploadID = entry.ID
	return upload, nil
}

func (s *archiveService) CompleteUpload(ctx context.Context, id string) (ArchiveEntry, error) {
	entry, err := s.getEntry(ctx, id)
	if err != nil {
		return ArchiveEntry{}, err
	}
	if entry.State == "available" {
		if err := s.deleteObject(ctx, entry.PendingKey); err != nil {
			return ArchiveEntry{}, fmt.Errorf("clean up staged archive object: %w", err)
		}
		return entry, nil
	}
	if entry.State != "pending" {
		return ArchiveEntry{}, ErrArchiveNotPending
	}
	if !validGeneratedObjectKey(entry.PendingKey, "pending/") || !validGeneratedObjectKey(entry.ObjectKey, "archives/") {
		return ArchiveEntry{}, errors.New("pending archive record contains invalid object keys")
	}
	if err := s.objectsCallErr(ctx, func(callCtx context.Context) error {
		return s.objects.VerifyUpload(callCtx, entry.PendingKey, entry.OriginalName, entry.ContentType, entry.Size)
	}); err != nil {
		return ArchiveEntry{}, fmt.Errorf("verify staged archive object: %w", err)
	}
	if err := s.objectsCallErr(ctx, func(callCtx context.Context) error {
		return s.objects.PromoteUpload(callCtx, entry.PendingKey, entry.ObjectKey)
	}); err != nil {
		return ArchiveEntry{}, fmt.Errorf("promote staged archive object; retain object for a safe retry: %w", err)
	}
	if err := s.repositoryCall(ctx, func(callCtx context.Context) error {
		return s.repository.SetStatus(callCtx, entry.ID, "available")
	}); err != nil {
		persisted, readErr := s.getEntry(context.WithoutCancel(ctx), entry.ID)
		if readErr != nil {
			return ArchiveEntry{}, errors.Join(fmt.Errorf("mark archive available: %w", err), fmt.Errorf("reconcile archive state: %w", readErr))
		}
		if persisted.State != "available" {
			return ArchiveEntry{}, fmt.Errorf("mark archive available: %w", err)
		}
		entry = persisted
	} else {
		entry.State = "available"
		entry.ExpiresAt = 0
	}
	if err := s.deleteObject(ctx, entry.PendingKey); err != nil {
		return entry, fmt.Errorf("clean up staged archive object: %w", err)
	}
	return entry, nil
}

func (s *archiveService) List(ctx context.Context) ([]ArchiveEntry, error) {
	var entries []ArchiveEntry
	err := s.repositoryCall(ctx, func(callCtx context.Context) error {
		var err error
		entries, err = s.repository.ListAvailable(callCtx)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("list available archives: %w", err)
	}
	if entries == nil {
		entries = []ArchiveEntry{}
	}
	available := entries[:0]
	for _, entry := range entries {
		if entry.State == "available" && validGeneratedObjectKey(entry.ObjectKey, "archives/") {
			available = append(available, entry)
		}
	}
	sort.Slice(available, func(i, j int) bool {
		if available[i].UploadedAt == available[j].UploadedAt {
			return available[i].ID > available[j].ID
		}
		return available[i].UploadedAt > available[j].UploadedAt
	})
	return available, nil
}

func (s *archiveService) Download(ctx context.Context, id string) (PresignedDownload, error) {
	entry, err := s.getEntry(ctx, id)
	if err != nil {
		return PresignedDownload{}, err
	}
	if entry.State != "available" {
		return PresignedDownload{}, ErrArchiveNotReady
	}
	if !validGeneratedObjectKey(entry.ObjectKey, "archives/") {
		return PresignedDownload{}, errors.New("available archive record contains an invalid object key")
	}
	expiresAt := s.now().UTC().Add(archiveURLLifetime)
	url, err := s.objectsCallString(ctx, func(callCtx context.Context) (string, error) {
		return s.objects.PresignDownload(callCtx, entry.ObjectKey, archiveURLLifetime)
	})
	if err != nil {
		return PresignedDownload{}, fmt.Errorf("create presigned download: %w", err)
	}
	return PresignedDownload{URL: url, ExpiresAt: expiresAt}, nil
}

func (s *archiveService) Delete(ctx context.Context, id string) error {
	entry, err := s.getEntry(ctx, id)
	if err != nil {
		if errors.Is(err, ErrArchiveNotFound) {
			return s.finishDeletion(id, nil)
		}
		return s.finishDeletion(id, err)
	}
	if entry.State != "available" && entry.State != "deleting" {
		return s.finishDeletion(id, ErrArchiveNotReady)
	}
	if !validGeneratedObjectKey(entry.ObjectKey, "archives/") {
		return s.finishDeletion(id, errors.New("archive record contains an invalid object key"))
	}
	if entry.State != "deleting" {
		if err := s.repositoryCall(ctx, func(callCtx context.Context) error {
			return s.repository.SetStatus(callCtx, entry.ID, "deleting")
		}); err != nil {
			return s.finishDeletion(id, fmt.Errorf("mark archive deleting: %w", err))
		}
	}
	if err := s.deleteObject(ctx, entry.ObjectKey); err != nil {
		return s.finishDeletion(id, fmt.Errorf("delete archive object; archive remains in deleting state for retry: %w", err))
	}
	if err := s.repositoryCall(ctx, func(callCtx context.Context) error {
		return s.repository.Delete(callCtx, entry.ID)
	}); err != nil {
		_, readErr := s.getEntry(context.WithoutCancel(ctx), entry.ID)
		if errors.Is(readErr, ErrArchiveNotFound) {
			return s.finishDeletion(id, nil)
		}
		if readErr != nil {
			return s.finishDeletion(id, errors.Join(
				fmt.Errorf("delete archive metadata: %w", err),
				fmt.Errorf("reconcile archive metadata deletion: %w", readErr),
			))
		}
		return s.finishDeletion(id, fmt.Errorf("delete archive metadata: %w", err))
	}
	return s.finishDeletion(id, nil)
}

func (s *archiveService) finishDeletion(id string, err error) error {
	logID := id
	if !validGeneratedObjectKey("pending/"+id, "pending/") {
		logID = "invalid"
	}
	if err != nil {
		s.logger.Error("archive deletion failed", "archive_id", logID, "outcome", "failed", "error", err)
		return err
	}
	s.logger.Info("archive deletion completed", "archive_id", logID, "outcome", "deleted")
	return nil
}

func (s *archiveService) getEntry(ctx context.Context, id string) (ArchiveEntry, error) {
	var entry ArchiveEntry
	err := s.repositoryCall(ctx, func(callCtx context.Context) error {
		var err error
		entry, err = s.repository.Get(callCtx, id)
		return err
	})
	if err != nil {
		return ArchiveEntry{}, fmt.Errorf("get archive record: %w", err)
	}
	return entry, nil
}

func (s *archiveService) deleteObject(ctx context.Context, key string) error {
	if key == "" {
		return nil
	}
	return s.objectsCallErr(ctx, func(callCtx context.Context) error {
		return s.objects.DeleteObject(callCtx, key)
	})
}

func (s *archiveService) compensatePending(ctx context.Context, entry ArchiveEntry) error {
	compensationCtx := context.WithoutCancel(ctx)
	objectErr := s.deleteObject(compensationCtx, entry.PendingKey)
	recordErr := s.repositoryCall(compensationCtx, func(callCtx context.Context) error {
		return s.repository.Delete(callCtx, entry.ID)
	})
	return errors.Join(wrapOptionalError("delete staged object after upload setup failure", objectErr), wrapOptionalError("delete pending metadata after upload setup failure", recordErr))
}

func (s *archiveService) repositoryCall(ctx context.Context, call func(context.Context) error) error {
	callCtx, cancel := context.WithTimeout(ctx, storageOperationTimeout)
	defer cancel()
	return call(callCtx)
}

func (s *archiveService) objectsCall(ctx context.Context, call func(context.Context) (PresignedUpload, error)) (PresignedUpload, error) {
	callCtx, cancel := context.WithTimeout(ctx, storageOperationTimeout)
	defer cancel()
	return call(callCtx)
}

func (s *archiveService) objectsCallErr(ctx context.Context, call func(context.Context) error) error {
	callCtx, cancel := context.WithTimeout(ctx, storageOperationTimeout)
	defer cancel()
	return call(callCtx)
}

func (s *archiveService) objectsCallString(ctx context.Context, call func(context.Context) (string, error)) (string, error) {
	callCtx, cancel := context.WithTimeout(ctx, storageOperationTimeout)
	defer cancel()
	return call(callCtx)
}

func newArchiveID() (string, error) {
	var id [16]byte
	if _, err := rand.Read(id[:]); err != nil {
		return "", err
	}
	id[6] = (id[6] & 0x0f) | 0x40
	id[8] = (id[8] & 0x3f) | 0x80
	return strings.Join([]string{
		hex.EncodeToString(id[0:4]),
		hex.EncodeToString(id[4:6]),
		hex.EncodeToString(id[6:8]),
		hex.EncodeToString(id[8:10]),
		hex.EncodeToString(id[10:16]),
	}, "-"), nil
}

func wrapOptionalError(operation string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", operation, err)
}

func validateUpload(name, contentType string, size int64, signature []byte) error {
	if err := validateUploadMetadata(name, contentType, size); err != nil {
		return err
	}

	var expectedType string
	switch strings.ToLower(filepath.Ext(name)) {
	case ".pdf":
		expectedType = "application/pdf"
	case ".ppt":
		expectedType = "application/vnd.ms-powerpoint"
	case ".pptx":
		expectedType = "application/vnd.openxmlformats-officedocument.presentationml.presentation"
	}

	switch expectedType {
	case "application/pdf":
		if !bytes.HasPrefix(signature, []byte("%PDF-")) {
			return errors.New("file does not have a PDF signature")
		}
	case "application/vnd.ms-powerpoint":
		if !isPowerPointCompoundFile(signature) {
			return errors.New("file is not a PowerPoint compound document")
		}
	case "application/vnd.openxmlformats-officedocument.presentationml.presentation":
		if !isPowerPointZip(signature) {
			return errors.New("file is not a PowerPoint OOXML package")
		}
	}
	return nil
}

func validateUploadMetadata(name, contentType string, size int64) error {
	if size < 1 || size > maxArchiveSize {
		return fmt.Errorf("file size must be between 1 and %d bytes", maxArchiveSize)
	}
	if !utf8.ValidString(name) || len(name) > 255 {
		return errors.New("filename must be valid UTF-8 and at most 255 bytes")
	}
	if name == "" || name == "." || name == ".." || strings.ContainsAny(name, `/\`) {
		return errors.New("filename must not contain path separators or traversal")
	}

	var expectedType string
	switch strings.ToLower(filepath.Ext(name)) {
	case ".pdf":
		expectedType = "application/pdf"
	case ".ppt":
		expectedType = "application/vnd.ms-powerpoint"
	case ".pptx":
		expectedType = "application/vnd.openxmlformats-officedocument.presentationml.presentation"
	default:
		return errors.New("file extension is not supported")
	}
	if contentType != expectedType {
		return errors.New("content type does not match the file extension")
	}
	return nil
}

func isPowerPointZip(content []byte) bool {
	reader, err := zip.NewReader(bytes.NewReader(content), int64(len(content)))
	if err != nil {
		return false
	}
	var contentTypes, presentation *zip.File
	for _, file := range reader.File {
		switch file.Name {
		case "[Content_Types].xml":
			contentTypes = file
		case "ppt/presentation.xml":
			presentation = file
		}
	}
	if contentTypes == nil || presentation == nil {
		return false
	}

	typesBytes, ok := readZipMember(contentTypes)
	if !ok {
		return false
	}
	var types struct {
		XMLName   xml.Name `xml:"Types"`
		Overrides []struct {
			PartName    string `xml:"PartName,attr"`
			ContentType string `xml:"ContentType,attr"`
		} `xml:"Override"`
	}
	if xml.Unmarshal(typesBytes, &types) != nil ||
		types.XMLName.Local != "Types" ||
		types.XMLName.Space != "http://schemas.openxmlformats.org/package/2006/content-types" {
		return false
	}
	hasPresentationType := false
	for _, override := range types.Overrides {
		if override.PartName == "/ppt/presentation.xml" &&
			override.ContentType == "application/vnd.openxmlformats-officedocument.presentationml.presentation.main+xml" {
			hasPresentationType = true
			break
		}
	}
	if !hasPresentationType {
		return false
	}

	presentationBytes, ok := readZipMember(presentation)
	if !ok {
		return false
	}
	decoder := xml.NewDecoder(bytes.NewReader(presentationBytes))
	for {
		token, err := decoder.Token()
		if err != nil {
			return false
		}
		if root, ok := token.(xml.StartElement); ok {
			return root.Name.Local == "presentation" &&
				root.Name.Space == "http://schemas.openxmlformats.org/presentationml/2006/main"
		}
	}
}

func isPowerPointCompoundFile(content []byte) bool {
	if len(content) < 512 || !bytes.Equal(content[:8], []byte{0xd0, 0xcf, 0x11, 0xe0, 0xa1, 0xb1, 0x1a, 0xe1}) {
		return false
	}
	header := content[:512]
	if binary.LittleEndian.Uint16(header[28:]) != 0xfffe {
		return false
	}
	majorVersion := binary.LittleEndian.Uint16(header[26:])
	sectorShift := binary.LittleEndian.Uint16(header[30:])
	if (majorVersion != 3 || sectorShift != 9) && (majorVersion != 4 || sectorShift != 12) {
		return false
	}
	sectorSize := 1 << sectorShift
	if len(content) < sectorSize || (len(content)-sectorSize)%sectorSize != 0 {
		return false
	}
	sectorCount := uint32((len(content) - sectorSize) / sectorSize)
	fatSectorCount := binary.LittleEndian.Uint32(header[44:])
	if fatSectorCount == 0 || fatSectorCount > sectorCount {
		return false
	}
	const (
		freeSector    = uint32(0xffffffff)
		endOfChain    = uint32(0xfffffffe)
		fatSector     = uint32(0xfffffffd)
		difatSector   = uint32(0xfffffffc)
		headerDifatAt = 76
		headerDifatN  = 109
	)
	fatSectorIDs := make([]uint32, 0, fatSectorCount)
	for i := 0; i < headerDifatN && uint32(len(fatSectorIDs)) < fatSectorCount; i++ {
		id := binary.LittleEndian.Uint32(header[headerDifatAt+i*4:])
		if id != freeSector {
			fatSectorIDs = append(fatSectorIDs, id)
		}
	}

	nextDIFAT := binary.LittleEndian.Uint32(header[68:])
	difatSectorCount := binary.LittleEndian.Uint32(header[72:])
	if difatSectorCount > sectorCount {
		return false
	}
	seenDIFAT := make(map[uint32]struct{}, difatSectorCount)
	difatEntriesPerSector := uint32(sectorSize/4 - 1)
	for i := uint32(0); uint32(len(fatSectorIDs)) < fatSectorCount && i < difatSectorCount; i++ {
		if nextDIFAT >= sectorCount {
			return false
		}
		if _, exists := seenDIFAT[nextDIFAT]; exists {
			return false
		}
		seenDIFAT[nextDIFAT] = struct{}{}
		sector := compoundSector(content, nextDIFAT, sectorSize)
		if sector == nil {
			return false
		}
		for j := uint32(0); j < difatEntriesPerSector && uint32(len(fatSectorIDs)) < fatSectorCount; j++ {
			id := binary.LittleEndian.Uint32(sector[j*4:])
			if id != freeSector {
				fatSectorIDs = append(fatSectorIDs, id)
			}
		}
		nextDIFAT = binary.LittleEndian.Uint32(sector[sectorSize-4:])
	}
	if uint32(len(fatSectorIDs)) != fatSectorCount {
		return false
	}

	fat := make([]uint32, 0, int(fatSectorCount)*int(sectorSize/4))
	for _, id := range fatSectorIDs {
		if id >= sectorCount {
			return false
		}
		sector := compoundSector(content, id, sectorSize)
		if sector == nil {
			return false
		}
		for offset := 0; offset < len(sector); offset += 4 {
			fat = append(fat, binary.LittleEndian.Uint32(sector[offset:]))
		}
	}

	directoryStart := binary.LittleEndian.Uint32(header[48:])
	directory, ok := readCompoundSectorChain(content, fat, directoryStart, sectorCount, sectorSize)
	if !ok {
		return false
	}
	hasRoot := false
	hasPowerPointStream := false
	for offset := 0; offset+128 <= len(directory); offset += 128 {
		entry := directory[offset : offset+128]
		entryType := entry[66]
		if entryType == 5 {
			hasRoot = true
		}
		if entryType != 2 {
			continue
		}
		nameLength := int(binary.LittleEndian.Uint16(entry[64:]))
		if nameLength < 2 || nameLength > 64 || nameLength%2 != 0 {
			continue
		}
		nameUnits := make([]uint16, nameLength/2-1)
		for i := range nameUnits {
			nameUnits[i] = binary.LittleEndian.Uint16(entry[i*2:])
		}
		name := string(utf16.Decode(nameUnits))
		streamSize := binary.LittleEndian.Uint64(entry[120:])
		if majorVersion == 3 && streamSize>>32 != 0 {
			continue
		}
		if name == "PowerPoint Document" && streamSize > 0 {
			hasPowerPointStream = true
		}
	}
	return hasRoot && hasPowerPointStream
}

func compoundSector(content []byte, sectorID uint32, sectorSize int) []byte {
	offset := int64(sectorID+1) * int64(sectorSize)
	end := offset + int64(sectorSize)
	if offset < 0 || end > int64(len(content)) {
		return nil
	}
	return content[int(offset):int(end)]
}

func readCompoundSectorChain(content []byte, fat []uint32, firstSector, sectorCount uint32, sectorSize int) ([]byte, bool) {
	const (
		freeSector  = uint32(0xffffffff)
		endOfChain  = uint32(0xfffffffe)
		fatSector   = uint32(0xfffffffd)
		difatSector = uint32(0xfffffffc)
	)
	if firstSector >= sectorCount {
		return nil, false
	}
	result := make([]byte, 0)
	visited := make(map[uint32]struct{})
	for current := firstSector; current != endOfChain; {
		if current >= sectorCount || current >= uint32(len(fat)) {
			return nil, false
		}
		if _, exists := visited[current]; exists {
			return nil, false
		}
		visited[current] = struct{}{}
		sector := compoundSector(content, current, sectorSize)
		if sector == nil {
			return nil, false
		}
		result = append(result, sector...)
		next := fat[current]
		if next == freeSector || next == fatSector || next == difatSector {
			return nil, false
		}
		current = next
	}
	return result, true
}

func readZipMember(file *zip.File) ([]byte, bool) {
	reader, err := file.Open()
	if err != nil {
		return nil, false
	}
	defer reader.Close()
	content, err := io.ReadAll(io.LimitReader(reader, 1<<20+1))
	if err != nil || len(content) > 1<<20 {
		return nil, false
	}
	return content, true
}

func validateArchiveMetadata(title, version, date, summary string) error {
	for _, field := range []struct {
		name  string
		value string
		limit int
	}{
		{name: "title", value: title, limit: 200},
		{name: "version", value: version, limit: 64},
		{name: "date", value: date, limit: 10},
		{name: "summary", value: summary, limit: 2000},
	} {
		if !utf8.ValidString(field.value) || len(field.value) > field.limit {
			return fmt.Errorf("%s must be valid UTF-8 and at most %d bytes", field.name, field.limit)
		}
	}
	parsedDate, err := time.Parse("2006-01-02", date)
	if err != nil || parsedDate.Format("2006-01-02") != date {
		return errors.New("date must be a valid ISO calendar date (YYYY-MM-DD)")
	}
	return nil
}
