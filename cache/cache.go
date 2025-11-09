package cache

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Entry struct {
	URL          string `json:"url"`
	Path         string `json:"path"`
	Filename     string `json:"filename"`
	ETag         string `json:"etag"`
	LastModified string `json:"last_modified"`
	Size         int64  `json:"size"`
	Completed    bool   `json:"completed"`
	UpdatedAt    int64  `json:"updated_at_unix"`
}

type Manager struct {
	root      string
	indexPath string
	mu        sync.Mutex
	entries   map[string]*Entry
}

func NewManager(root string) (*Manager, error) {
	if root == "" {
		root = ".cache"
	}
	if err := os.MkdirAll(root, 0755); err != nil {
		return nil, err
	}
	indexPath := filepath.Join(root, "index.json")
	mgr := &Manager{
		root:      root,
		indexPath: indexPath,
		entries:   map[string]*Entry{},
	}
	if err := mgr.load(); err != nil {
		return nil, err
	}
	return mgr, nil
}

func (m *Manager) load() error {
	data, err := os.ReadFile(m.indexPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	var list []*Entry
	if err := json.Unmarshal(data, &list); err != nil {
		return err
	}
	for _, e := range list {
		m.entries[e.URL] = e
	}
	return nil
}

func (m *Manager) save() error {
	list := make([]*Entry, 0, len(m.entries))
	for _, e := range m.entries {
		list = append(list, e)
	}
	data, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return err
	}
	tmp := m.indexPath + ".part"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, m.indexPath)
}

func (m *Manager) Get(url string) (*Entry, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	e, ok := m.entries[url]
	return e, ok
}

func (m *Manager) OpenFile(e *Entry) (*os.File, error) {
	return os.Open(e.Path)
}

func (m *Manager) CreateTempWriter(url string) (string, *os.File, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	name := uuid.New().String() + ".part"
	path := filepath.Join(m.root, name)
	f, err := os.Create(path)
	if err != nil {
		return "", nil, err
	}
	return path, f, nil
}

func (m *Manager) Commit(url, tempPath, filename, etag, lastModified string, size int64) (*Entry, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	finalPath := filepath.Join(m.root, filename)
	if err := os.Rename(tempPath, finalPath); err != nil {
		_ = os.Remove(finalPath)
		if err2 := os.Rename(tempPath, finalPath); err2 != nil {
			return nil, err
		}
	}

	entry := &Entry{
		URL:          url,
		Path:         finalPath,
		Filename:     filename,
		ETag:         etag,
		LastModified: lastModified,
		Size:         size,
		Completed:    true,
		UpdatedAt:    time.Now().Unix(),
	}
	m.entries[url] = entry
	if err := m.save(); err != nil {
		return nil, err
	}
	return entry, nil
}
