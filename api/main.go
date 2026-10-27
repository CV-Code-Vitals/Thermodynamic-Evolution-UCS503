package main

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/jackc/pgx/v5/pgxpool"
)

// HealthResponse defines the structure for our health-check endpoint.
type HealthResponse struct {
	Status  string `json:"status"`
	Module  string `json:"module"`
	Version string `json:"version"`
}

var (
	archivesFile = func() string {
		if v := os.Getenv("ARCHIVES_FILE"); v != "" {
			return v
		}
		return "data/archives.json"
	}()
	uploadsDir = func() string {
		if v := os.Getenv("UPLOADS_DIR"); v != "" {
			return v
		}
		return "uploads"
	}()
	mu                sync.Mutex
	archiveMutationMu sync.Mutex // serializes read-modify-write archive updates
	adminPasskey      = func() string {
		if v := os.Getenv("ADMIN_PASSKEY"); v != "" {
			return v
		}
		if v := os.Getenv("API_TOKEN"); v != "" {
			return v
		}
		return ""
	}()

	s3Bucket string
	s3Region string
	s3Client s3API
	useS3    bool

	sessionTTL         = time.Hour
	sessionMu          sync.Mutex
	sessionStore                 = map[string]time.Time{}
	sessionCookie                = "ucs503_session"
	sessionTokenRandom io.Reader = rand.Reader

	dbPool *pgxpool.Pool
	useDB  bool
)

type s3API interface {
	PutObject(context.Context, *s3.PutObjectInput, ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	GetObject(context.Context, *s3.GetObjectInput, ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	DeleteObject(context.Context, *s3.DeleteObjectInput, ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
}

func randomSessionToken() (string, error) {
	b := make([]byte, 32)
	if _, err := io.ReadFull(sessionTokenRandom, b); err != nil {
		return "", fmt.Errorf("generate session token: %w", err)
	}
	return hex.EncodeToString(b), nil
}

func hasValidSession(r *http.Request) bool {
	if r == nil {
		return false
	}
	cookie, err := r.Cookie(sessionCookie)
	if err != nil || cookie.Value == "" {
		return false
	}

	sessionMu.Lock()
	defer sessionMu.Unlock()

	expiresAt, ok := sessionStore[cookie.Value]
	if !ok {
		return false
	}
	if time.Now().After(expiresAt) {
		delete(sessionStore, cookie.Value)
		return false
	}
	return true
}

func constantTimeEquals(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

func sessionSameSite(r *http.Request) http.SameSite {
	// TLS is commonly terminated by a trusted reverse proxy. Do not infer this
	// from arbitrary client headers unless TRUST_PROXY is explicitly enabled.
	if r != nil && (r.TLS != nil || (os.Getenv("TRUST_PROXY") == "true" && r.Header.Get("X-Forwarded-Proto") == "https")) {
		return http.SameSiteNoneMode
	}
	return http.SameSiteLaxMode
}

func requestIsHTTPS(r *http.Request) bool {
	return r.TLS != nil || (os.Getenv("TRUST_PROXY") == "true" && r.Header.Get("X-Forwarded-Proto") == "https")
}

func isAllowedOrigin(origin string) bool {
	if origin == "" {
		return false
	}
	origin = strings.TrimRight(origin, "/")
	for _, configured := range strings.Split(os.Getenv("CORS_ALLOWED_ORIGINS"), ",") {
		if strings.TrimSpace(configured) != "" &&
			origin == strings.TrimRight(strings.TrimSpace(configured), "/") {
			return true
		}
	}
	return false
}

func validateCORSAllowedOrigins(value string) error {
	validOrigins := 0
	for index, configured := range strings.Split(value, ",") {
		origin := strings.TrimRight(strings.TrimSpace(configured), "/")
		if origin == "" {
			continue
		}
		parsed, err := url.Parse(origin)
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") ||
			parsed.Host == "" || parsed.User != nil || parsed.Path != "" ||
			parsed.RawQuery != "" || parsed.Fragment != "" {
			return fmt.Errorf("CORS_ALLOWED_ORIGINS entry %d must be an HTTP(S) origin without a path", index+1)
		}
		validOrigins++
	}
	if validOrigins == 0 {
		return fmt.Errorf("CORS_ALLOWED_ORIGINS must contain at least one HTTP(S) origin")
	}
	return nil
}

// Browser form posts bypass CORS. Require a configured Origin for cookie-authenticated mutations.
func requireTrustedOrigin(w http.ResponseWriter, r *http.Request) bool {
	if r.Header.Get("Origin") == "" || !isAllowedOrigin(r.Header.Get("Origin")) {
		writeJSONError(w, http.StatusForbidden, "origin_not_allowed")
		return false
	}
	return true
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := strings.TrimRight(r.Header.Get("Origin"), "/")
		if isAllowedOrigin(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-API-Token, Authorization")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
			w.Header().Add("Vary", "Origin")
		}
		if r.Method == http.MethodOptions {
			if origin == "" {
				http.Error(w, "Origin required", http.StatusForbidden)
				return
			}
			if !isAllowedOrigin(origin) {
				http.Error(w, "Origin not allowed", http.StatusForbidden)
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		next.ServeHTTP(w, r)
	})
}

func ensureUploadsDir() error {
	if _, err := os.Stat(uploadsDir); os.IsNotExist(err) {
		return os.MkdirAll(uploadsDir, 0o755)
	}
	return nil
}

func loadArchives() ([]ArchiveEntry, error) {
	if useDB && dbPool != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		rows, err := dbPool.Query(ctx, "SELECT filename, original_name, title, version, summary, uploaded_at, url FROM archives ORDER BY uploaded_at DESC")
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var entries []ArchiveEntry
		for rows.Next() {
			var e ArchiveEntry
			var uploaded time.Time
			if err := rows.Scan(&e.Filename, &e.OriginalName, &e.Title, &e.Version, &e.Summary, &uploaded, &e.URL); err != nil {
				return nil, err
			}
			e.UploadedAt = uploaded.Format(time.RFC3339)
			entries = append(entries, e)
		}
		return entries, nil
	}

	mu.Lock()
	defer mu.Unlock()

	if _, err := os.Stat(archivesFile); os.IsNotExist(err) {
		return []ArchiveEntry{}, nil
	}

	f, err := os.Open(archivesFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var entries []ArchiveEntry
	if err := json.NewDecoder(f).Decode(&entries); err != nil {
		return nil, err
	}
	return entries, nil
}

func saveArchives(entries []ArchiveEntry) error {
	if useDB && dbPool != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		for _, e := range entries {
			uploaded, _ := time.Parse(time.RFC3339, e.UploadedAt)
			_, err := dbPool.Exec(ctx, "INSERT INTO archives (filename, original_name, title, version, summary, uploaded_at, url) VALUES ($1,$2,$3,$4,$5,$6,$7) ON CONFLICT (filename) DO UPDATE SET original_name=EXCLUDED.original_name, title=EXCLUDED.title, version=EXCLUDED.version, summary=EXCLUDED.summary, uploaded_at=EXCLUDED.uploaded_at, url=EXCLUDED.url",
				e.Filename, e.OriginalName, e.Title, e.Version, e.Summary, uploaded, e.URL)
			if err != nil {
				return err
			}
		}
		return nil
	}

	mu.Lock()
	defer mu.Unlock()

	if err := ensureArchivesDir(); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(archivesFile), ".archives-*.json")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)

	enc := json.NewEncoder(tmp)
	enc.SetIndent("", "  ")
	if err := enc.Encode(entries); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, archivesFile)
}

func ensureArchivesDir() error {
	return os.MkdirAll(filepath.Dir(archivesFile), 0o750)
}

func validateProductionConfig(environment, bucket, region, databaseURL string) error {
	if !strings.EqualFold(strings.TrimSpace(environment), "production") {
		return nil
	}
	if strings.TrimSpace(bucket) == "" {
		return fmt.Errorf("S3_BUCKET is required in production")
	}
	if strings.TrimSpace(region) == "" {
		return fmt.Errorf("S3_REGION is required in production")
	}
	if strings.TrimSpace(databaseURL) == "" {
		return fmt.Errorf("DATABASE_URL is required in production")
	}
	return nil
}

func archiveDownloadURL(filename string) string {
	return "/api/download?filename=" + url.QueryEscape(filename)
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	if !authorized(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	serveDownload(w, r)
}

func serveDownload(w http.ResponseWriter, r *http.Request) {
	filename := r.URL.Query().Get("filename")
	if filename == "" || filepath.Base(filename) != filename || strings.ContainsAny(filename, `/\`) {
		http.Error(w, "filename required", http.StatusBadRequest)
		return
	}

	entries, err := loadArchives()
	if err != nil {
		log.Printf("[ERROR] unable to read archive metadata for download: %v", err)
		http.Error(w, "Server error: unable to read archive metadata", http.StatusInternalServerError)
		return
	}
	var entry *ArchiveEntry
	for i := range entries {
		if entries[i].Filename == filename {
			entry = &entries[i]
			break
		}
	}
	if entry == nil {
		http.Error(w, "Archive not found", http.StatusNotFound)
		return
	}

	var body io.ReadCloser
	contentType := mime.TypeByExtension(filepath.Ext(entry.OriginalName))
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	if useS3 {
		if s3Client == nil {
			http.Error(w, "Storage is unavailable", http.StatusServiceUnavailable)
			return
		}
		object, err := s3Client.GetObject(r.Context(), &s3.GetObjectInput{
			Bucket: &s3Bucket,
			Key:    &filename,
		})
		if err != nil {
			log.Printf("[ERROR] unable to download S3 object %q: %v", filename, err)
			http.Error(w, "Unable to download archive from storage", http.StatusBadGateway)
			return
		}
		if object == nil || object.Body == nil {
			http.Error(w, "Storage returned an empty archive", http.StatusBadGateway)
			return
		}
		body = object.Body
		if object.ContentType != nil && *object.ContentType != "" {
			contentType = *object.ContentType
		}
	} else {
		body, err = os.Open(filepath.Join(uploadsDir, filename))
		if err != nil {
			if os.IsNotExist(err) {
				http.NotFound(w, r)
				return
			}
			log.Printf("[ERROR] unable to open local archive %q: %v", filename, err)
			http.Error(w, "Unable to download archive", http.StatusInternalServerError)
			return
		}
	}
	defer body.Close()

	downloadName := filepath.Base(entry.OriginalName)
	if downloadName == "." || downloadName == "" {
		downloadName = filename
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": downloadName}))
	w.Header().Set("Cache-Control", "private, no-store")
	if _, err := io.Copy(w, body); err != nil {
		log.Printf("[ERROR] unable to stream archive %q: %v", filename, err)
	}
}

func authorized(r *http.Request) bool {
	if adminPasskey == "" {
		return false
	}
	if hasValidSession(r) {
		return true
	}
	if t := r.Header.Get("X-API-Token"); t != "" {
		return constantTimeEquals(t, adminPasskey)
	}
	if auth := r.Header.Get("Authorization"); auth != "" {
		const prefix = "Bearer "
		if len(auth) > len(prefix) && strings.HasPrefix(auth, prefix) {
			return constantTimeEquals(auth[len(prefix):], adminPasskey)
		}
	}
	return false
}

func ensureMigrationsTable(pool *pgxpool.Pool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := pool.Exec(ctx, "CREATE TABLE IF NOT EXISTS schema_migrations (version text PRIMARY KEY, applied_at timestamptz NOT NULL DEFAULT now())")
	return err
}

func runMigrations(pool *pgxpool.Pool) error {
	migrationsDir := "migrations"
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("read migrations directory: %w", err)
	}

	var names []string
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".sql") {
			names = append(names, f.Name())
		}
	}
	if len(names) == 0 {
		return nil
	}

	sort.Strings(names)
	ctx := context.Background()
	for _, name := range names {
		version := name
		var exists bool
		err := pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version=$1)", version).Scan(&exists)
		if err != nil {
			return err
		}
		if exists {
			continue
		}

		b, err := os.ReadFile(filepath.Join(migrationsDir, name))
		if err != nil {
			return err
		}
		if _, err := pool.Exec(ctx, string(b)); err != nil {
			return err
		}
		if _, err := pool.Exec(ctx, "INSERT INTO schema_migrations(version) VALUES($1)", version); err != nil {
			return err
		}
	}
	return nil
}

func requireConfiguredAdminPasskey() {
	if strings.TrimSpace(adminPasskey) == "" {
		log.Fatal("ADMIN_PASSKEY is required; set it before starting the API")
	}
}

func main() {
	if isLambdaRuntime() {
		startLambda()
		return
	}
	if err := validateCORSAllowedOrigins(os.Getenv("CORS_ALLOWED_ORIGINS")); err != nil {
		log.Fatal(err)
	}
	deps, err := newHTTPDependencies(context.Background())
	if err != nil {
		log.Fatalf("API startup failed: %v", err)
	}
	deps.Analysis, err = newAnalysisProxy(os.Getenv("ANALYSIS_SERVICE_URL"), os.Getenv("ANALYSIS_API_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}
	handler := newRouter(deps)
	port := "8080"
	if v := strings.TrimSpace(os.Getenv("PORT")); v != "" {
		if _, err := strconv.Atoi(v); err == nil {
			port = v
		} else {
			log.Printf("[WARN] Ignoring invalid PORT; using default 8080")
		}
	}
	addr := ":" + port
	log.Printf("[SYS] Starting portal API on %s", addr)
	server := &http.Server{
		Addr: addr, Handler: handler,
		ReadHeaderTimeout: 10 * time.Second, ReadTimeout: 2 * time.Minute,
		WriteTimeout: 10 * time.Minute, IdleTimeout: 60 * time.Second,
	}
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("[FATAL] Server failed to start: %v", err)
	}
}

func newHTTPDependencies(ctx context.Context) (Dependencies, error) {
	if strings.EqualFold(strings.TrimSpace(os.Getenv("APP_ENV")), "production") {
		if os.Getenv("API_RUNTIME") != "http" {
			return Dependencies{}, fmt.Errorf("production HTTP startup requires API_RUNTIME=http and AWS archive configuration")
		}
		return newAWSDependencies(ctx)
	}
	return newLocalDependencies(os.Getenv("ADMIN_PASSKEY"), os.Getenv("TOKEN_HMAC_KEY"), archivesFile, uploadsDir)
}
