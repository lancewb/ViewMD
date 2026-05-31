package preferences

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"ViewMD/internal/models"
)

const (
	defaultViewMode = "split"
	configDirName   = "ViewMD"
	configFileName  = "workspace.json"
)

type Store struct {
	path string
	mu   sync.Mutex
}

type Preferences struct {
	LastProjectPath string   `json:"lastProjectPath"`
	OpenFiles       []string `json:"openFiles"`
	ActiveFile      string   `json:"activeFile"`
	ViewMode        string   `json:"viewMode"`
}

func NewStore() *Store {
	baseDir, err := os.UserConfigDir()
	if err != nil || baseDir == "" {
		baseDir = "."
	}

	return &Store{
		path: filepath.Join(baseDir, configDirName, configFileName),
	}
}

func (s *Store) Load() (Preferences, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	prefs := Preferences{ViewMode: defaultViewMode}
	data, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		return prefs, nil
	}
	if err != nil {
		return prefs, err
	}
	if len(data) == 0 {
		return prefs, nil
	}
	if err := json.Unmarshal(data, &prefs); err != nil {
		return Preferences{ViewMode: defaultViewMode}, nil
	}

	prefs.ViewMode = NormalizeViewMode(prefs.ViewMode)
	prefs.OpenFiles = compactPaths(prefs.OpenFiles)
	return prefs, nil
}

func (s *Store) Save(prefs Preferences) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	prefs.ViewMode = NormalizeViewMode(prefs.ViewMode)
	prefs.OpenFiles = compactPaths(prefs.OpenFiles)

	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(prefs, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o644)
}

func (p Preferences) Workspace() models.WorkspaceState {
	return models.WorkspaceState{
		ProjectPath: p.LastProjectPath,
		OpenFiles:   compactPaths(p.OpenFiles),
		ActiveFile:  filepath.ToSlash(p.ActiveFile),
		ViewMode:    NormalizeViewMode(p.ViewMode),
	}
}

func NormalizeViewMode(mode string) string {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "code", "preview", "split":
		return strings.ToLower(strings.TrimSpace(mode))
	default:
		return defaultViewMode
	}
}

func compactPaths(paths []string) []string {
	seen := make(map[string]struct{}, len(paths))
	result := make([]string, 0, len(paths))

	for _, item := range paths {
		clean := filepath.ToSlash(strings.TrimSpace(item))
		if clean == "" {
			continue
		}
		if _, ok := seen[clean]; ok {
			continue
		}
		seen[clean] = struct{}{}
		result = append(result, clean)
	}

	return result
}
