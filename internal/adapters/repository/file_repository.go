package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	
	"github.com/panoramablock/thegraph-data-extraction/internal/domain/entity"
)

// FileRepository is an adapter that implements the ports.Repository interface using the file system
type FileRepository struct {
	baseDir      string
	metadataDir  string
	entityDir    string
	cursorCache  map[string]string
	cursorMu     sync.RWMutex
	flushTimeout time.Duration
	encoder      *json.Encoder
}

// FileRepositoryConfig holds the configuration for the file repository
type FileRepositoryConfig struct {
	BaseDir      string
	FlushTimeout time.Duration
}

// NewFileRepository creates a new file repository
func NewFileRepository(config FileRepositoryConfig) (*FileRepository, error) {
	// Set default base directory if not provided
	if config.BaseDir == "" {
		config.BaseDir = "data"
	}
	
	// Set default flush timeout if not provided
	if config.FlushTimeout <= 0 {
		config.FlushTimeout = 5 * time.Second
	}
	
	// Create directories
	metadataDir := filepath.Join(config.BaseDir, "metadata")
	entityDir := filepath.Join(config.BaseDir, "entities")
	
	for _, dir := range []string{metadataDir, entityDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}
	
	repo := &FileRepository{
		baseDir:      config.BaseDir,
		metadataDir:  metadataDir,
		entityDir:    entityDir,
		cursorCache:  make(map[string]string),
		flushTimeout: config.FlushTimeout,
	}
	
	// Load existing cursors into cache
	if err := repo.loadCursors(); err != nil {
		log.Warn().Err(err).Msg("Failed to load cursors from disk")
	}
	
	return repo, nil
}

// loadCursors loads existing cursor files into the cache
func (r *FileRepository) loadCursors() error {
	files, err := os.ReadDir(r.metadataDir)
	if err != nil {
		return fmt.Errorf("failed to read metadata directory: %w", err)
	}
	
	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".cursor" {
			continue
		}
		
		path := filepath.Join(r.metadataDir, file.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			log.Error().
				Str("file", path).
				Err(err).
				Msg("Failed to read cursor file")
			continue
		}
		
		key := file.Name()[:len(file.Name())-7] // Remove .cursor extension
		r.cursorMu.Lock()
		r.cursorCache[key] = string(data)
		r.cursorMu.Unlock()
		
		log.Debug().
			Str("key", key).
			Str("cursor", string(data)).
			Msg("Loaded cursor from file")
	}
	
	return nil
}

// SaveEntity saves an entity to the repository
func (r *FileRepository) SaveEntity(ctx context.Context, e *entity.Entity) error {
	if e == nil {
		return fmt.Errorf("cannot save nil entity")
	}
	
	// Generate a filename based on entity type, deployment, and ID
	key := fmt.Sprintf("%s_%s", e.Type, e.Deployment)
	filename := fmt.Sprintf("%s_%s_%d.json", e.Type, e.ID, e.Timestamp.UnixNano())
	path := filepath.Join(r.entityDir, filename)
	
	// Marshal the entity to JSON
	data, err := json.MarshalIndent(e, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling entity: %w", err)
	}
	
	// Write the entity to a file
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("error writing entity file: %w", err)
	}
	
	// Update cursor cache if this entity has an ID
	if e.ID != "" {
		r.cursorMu.Lock()
		r.cursorCache[key] = e.ID
		r.cursorMu.Unlock()
		
		// Write cursor to a file asynchronously
		go func() {
			cursorPath := filepath.Join(r.metadataDir, key+".cursor")
			if err := os.WriteFile(cursorPath, []byte(e.ID), 0644); err != nil {
				log.Error().
					Str("key", key).
					Str("path", cursorPath).
					Err(err).
					Msg("Failed to write cursor file")
			}
		}()
	}
	
	return nil
}

// SaveEntityStream saves entities using a streaming JSON encoder to reduce memory usage
func (r *FileRepository) SaveEntityStream(ctx context.Context, entityType, deployment string, entities []*entity.Entity) error {
	if len(entities) == 0 {
		return nil
	}
	
	// Generate a filename based on entity type, deployment, and timestamp
	timestamp := time.Now().UTC().Format("20060102_150405")
	filename := fmt.Sprintf("%s_%s_%s.jsonl", entityType, deployment, timestamp)
	path := filepath.Join(r.entityDir, filename)
	
	// Create or truncate the file
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("error creating entity file: %w", err)
	}
	defer file.Close()
	
	// Create a JSON encoder that writes to the file
	encoder := json.NewEncoder(file)
	
	// Write each entity as a separate JSON line
	for _, e := range entities {
		if err := encoder.Encode(e); err != nil {
			return fmt.Errorf("error encoding entity: %w", err)
		}
	}
	
	// Update cursor cache if there are entities with IDs
	if len(entities) > 0 && entities[len(entities)-1].ID != "" {
		key := fmt.Sprintf("%s_%s", entityType, deployment)
		lastID := entities[len(entities)-1].ID
		
		r.cursorMu.Lock()
		r.cursorCache[key] = lastID
		r.cursorMu.Unlock()
		
		// Write cursor to a file asynchronously
		go func() {
			cursorPath := filepath.Join(r.metadataDir, key+".cursor")
			if err := os.WriteFile(cursorPath, []byte(lastID), 0644); err != nil {
				log.Error().
					Str("key", key).
					Str("path", cursorPath).
					Err(err).
					Msg("Failed to write cursor file")
			}
		}()
	}
	
	return nil
}

// GetLatestCursor gets the latest cursor for a given entity type and deployment
func (r *FileRepository) GetLatestCursor(ctx context.Context, entityType, deployment string) (string, error) {
	key := fmt.Sprintf("%s_%s", entityType, deployment)
	
	// Try to get cursor from cache
	r.cursorMu.RLock()
	cursor, exists := r.cursorCache[key]
	r.cursorMu.RUnlock()
	
	if exists {
		return cursor, nil
	}
	
	// If not in cache, try to read from file
	cursorPath := filepath.Join(r.metadataDir, key+".cursor")
	data, err := os.ReadFile(cursorPath)
	if err != nil {
		if os.IsNotExist(err) {
			// No cursor file exists, return empty string
			return "", nil
		}
		return "", fmt.Errorf("error reading cursor file: %w", err)
	}
	
	// Update cache
	cursor = string(data)
	r.cursorMu.Lock()
	r.cursorCache[key] = cursor
	r.cursorMu.Unlock()
	
	return cursor, nil
}

// Close flushes any pending data and closes the repository
func (r *FileRepository) Close() error {
	// Nothing to close for file repository
	return nil
} 