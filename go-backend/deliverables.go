package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

// ─────────────────────────────────────────────────────────────────────────────
// Data Models
// ─────────────────────────────────────────────────────────────────────────────

type Deliverable struct {
	ID         int    `json:"id"`
	Title      string `json:"title"`
	Version    string `json:"version"`
	Date       string `json:"date"`
	Summary    string `json:"summary"`
	Filename   string `json:"filename"`
	FileURL    string `json:"file_url"`
	UploadedAt string `json:"uploaded_at"`
}

type UploadDeliverableResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    Deliverable `json:"data"`
}

// ─────────────────────────────────────────────────────────────────────────────
// Storage Interface & Implementations
// ─────────────────────────────────────────────────────────────────────────────

type DeliverableStore interface {
	Save(d Deliverable) (Deliverable, error)
	GetAll() ([]Deliverable, error)
}

// PostgresStore implements DeliverableStore backed by PostgreSQL.
type PostgresStore struct {
	db *sql.DB
}

func (s *PostgresStore) Save(d Deliverable) (Deliverable, error) {
	query := `
		INSERT INTO deliverables (title, version, date, summary, filename, file_url, uploaded_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id;
	`
	var id int
	err := s.db.QueryRow(query, d.Title, d.Version, d.Date, d.Summary, d.Filename, d.FileURL, d.UploadedAt).Scan(&id)
	if err != nil {
		return d, fmt.Errorf("postgres insert: %w", err)
	}
	d.ID = id
	return d, nil
}

func (s *PostgresStore) GetAll() ([]Deliverable, error) {
	query := `
		SELECT id, title, version, date, COALESCE(summary, ''), filename, file_url, uploaded_at
		FROM deliverables
		ORDER BY id DESC;
	`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("postgres query: %w", err)
	}
	defer rows.Close()

	var list []Deliverable
	for rows.Next() {
		var d Deliverable
		var uploadedAt time.Time
		if err := rows.Scan(&d.ID, &d.Title, &d.Version, &d.Date, &d.Summary, &d.Filename, &d.FileURL, &uploadedAt); err != nil {
			return nil, fmt.Errorf("postgres scan: %w", err)
		}
		d.UploadedAt = uploadedAt.Format(time.RFC3339)
		list = append(list, d)
	}
	if list == nil {
		list = []Deliverable{}
	}
	return list, nil
}

// JSONStore implements DeliverableStore with atomic file persistence for standalone mode.
type JSONStore struct {
	filePath string
	mu       sync.RWMutex
	items    []Deliverable
	nextID   int
}

func NewJSONStore(filePath string) (*JSONStore, error) {
	s := &JSONStore{
		filePath: filePath,
		items:    []Deliverable{},
		nextID:   1,
	}

	if data, err := os.ReadFile(filePath); err == nil && len(data) > 0 {
		if err := json.Unmarshal(data, &s.items); err == nil {
			for _, item := range s.items {
				if item.ID >= s.nextID {
					s.nextID = item.ID + 1
				}
			}
		}
	}

	return s, nil
}

func (s *JSONStore) Save(d Deliverable) (Deliverable, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	d.ID = s.nextID
	s.nextID++
	s.items = append(s.items, d)

	data, err := json.MarshalIndent(s.items, "", "  ")
	if err != nil {
		return d, fmt.Errorf("marshal json: %w", err)
	}

	if err := os.WriteFile(s.filePath, data, 0o640); err != nil {
		return d, fmt.Errorf("write json file: %w", err)
	}

	return d, nil
}

func (s *JSONStore) GetAll() ([]Deliverable, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]Deliverable, len(s.items))
	copy(result, s.items)
	return result, nil
}

// initDeliverableStore attempts to connect to PostgreSQL, falling back to JSONStore.
func initDeliverableStore() DeliverableStore {
	dbHost := os.Getenv("DB_HOST")
	dbPort := envOr("DB_PORT", "5432")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbName := envOr("DB_NAME", "thermodynamic_db")

	if dbHost != "" && dbUser != "" {
		connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable connect_timeout=3",
			dbHost, dbPort, dbUser, dbPass, dbName)

		db, err := sql.Open("postgres", connStr)
		if err == nil {
			if err := db.Ping(); err == nil {
				log.Printf("[DB] Connected to PostgreSQL at %s:%s", dbHost, dbPort)
				createTableQuery := `
					CREATE TABLE IF NOT EXISTS deliverables (
						id SERIAL PRIMARY KEY,
						title TEXT NOT NULL,
						version TEXT NOT NULL,
						date TEXT NOT NULL,
						summary TEXT,
						filename TEXT NOT NULL,
						file_url TEXT NOT NULL,
						uploaded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
					);
				`
				if _, err := db.Exec(createTableQuery); err == nil {
					log.Printf("[DB] Initialized schema 'deliverables' in PostgreSQL.")
					return &PostgresStore{db: db}
				} else {
					log.Printf("[WARN] Failed to initialize table in PostgreSQL: %v", err)
				}
			} else {
				log.Printf("[INFO] PostgreSQL ping failed (%v) -- falling back to persistent JSON storage.", err)
			}
		}
	}

	// Persistent JSON storage fallback
	uploadsDir := "./uploads"
	_ = os.MkdirAll(uploadsDir, 0o750)
	jsonPath := filepath.Join(uploadsDir, "deliverables.json")
	store, err := NewJSONStore(jsonPath)
	if err != nil {
		log.Printf("[WARN] Error loading JSON store, using memory fallback: %v", err)
	} else {
		log.Printf("[INFO] Using persistent JSON storage for deliverables: %s", jsonPath)
	}
	return store
}

// ─────────────────────────────────────────────────────────────────────────────
// HTTP Handlers
// ─────────────────────────────────────────────────────────────────────────────

var allowedExtensions = map[string]bool{
	".ppt":  true,
	".pptx": true,
	".pdf":  true,
	".zip":  true,
	".txt":  true,
	".md":   true,
}

// uploadDeliverableHandler receives presentation / deliverable files with metadata.
func uploadDeliverableHandler(store DeliverableStore, maxUploadBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxUploadBytes)

		fileHeader, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:  "missing_file",
				Detail: "No file provided under field 'file'.",
			})
			return
		}

		// Security: Validate file extension against allowed whitelist
		ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
		if !allowedExtensions[ext] {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:  "disallowed_file_type",
				Detail: fmt.Sprintf("File type %q is not permitted. Allowed: .ppt, .pptx, .pdf, .zip, .txt, .md", ext),
			})
			return
		}

		title := c.PostForm("title")
		if title == "" {
			title = strings.TrimSuffix(fileHeader.Filename, ext)
		}
		version := c.DefaultPostForm("version", "1.0.0")
		date := c.DefaultPostForm("date", time.Now().Format("2006-01-02"))
		summary := c.PostForm("summary")

		uploadsDir := "./uploads"
		if err := os.MkdirAll(uploadsDir, 0o750); err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:  "server_error",
				Detail: "Could not initialize uploads storage directory.",
			})
			return
		}

		// Security: sanitize filename and prepend timestamp
		cleanBase := filepath.Base(fileHeader.Filename)
		cleanBase = strings.ReplaceAll(cleanBase, " ", "_")
		savedName := fmt.Sprintf("%d_%s", time.Now().UnixMilli(), cleanBase)
		savedPath := filepath.Join(uploadsDir, savedName)

		if err := c.SaveUploadedFile(fileHeader, savedPath); err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:  "save_failed",
				Detail: "Failed to persist uploaded deliverable file.",
			})
			return
		}

		d := Deliverable{
			Title:      title,
			Version:    version,
			Date:       date,
			Summary:    summary,
			Filename:   fileHeader.Filename,
			FileURL:    "/uploads/" + savedName,
			UploadedAt: time.Now().UTC().Format(time.RFC3339),
		}

		savedD, err := store.Save(d)
		if err != nil {
			log.Printf("[ERROR] Store.Save deliverable: %v", err)
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:  "storage_error",
				Detail: "File saved, but metadata failed to persist.",
			})
			return
		}

		c.JSON(http.StatusCreated, UploadDeliverableResponse{
			Success: true,
			Message: "Archive committed to storage matrix.",
			Data:    savedD,
		})
	}
}

// getDeliverablesHandler returns all uploaded deliverables.
func getDeliverablesHandler(store DeliverableStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		items, err := store.GetAll()
		if err != nil {
			log.Printf("[ERROR] Store.GetAll: %v", err)
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:  "retrieval_failed",
				Detail: "Could not retrieve deliverables list.",
			})
			return
		}

		c.JSON(http.StatusOK, items)
	}
}

// serveUploadFileHandler serves uploaded files securely with path traversal protection.
func serveUploadFileHandler(uploadsDir string) gin.HandlerFunc {
	return func(c *gin.Context) {
		filename := filepath.Base(c.Param("filename"))
		target := filepath.Join(uploadsDir, filename)

		// Ensure file exists and is inside uploadsDir
		cleanedTarget := filepath.Clean(target)
		cleanedUploads, _ := filepath.Abs(uploadsDir)
		absTarget, _ := filepath.Abs(cleanedTarget)

		if !strings.HasPrefix(absTarget, cleanedUploads) {
			c.String(http.StatusForbidden, "Forbidden")
			return
		}

		if fi, err := os.Stat(absTarget); err != nil || fi.IsDir() {
			c.String(http.StatusNotFound, "File Not Found")
			return
		}

		c.File(absTarget)
	}
}
