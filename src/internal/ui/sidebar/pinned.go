package sidebar

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/t4Linux/t4gfm/src/internal/utils"
)

type PinnedManager struct {
	filePath string
}

func NewPinnedFileManager(filePath string) PinnedManager {
	if err := utils.InitJSONFile(filePath); err != nil {
		slog.Error("Error initializing pinned JSON file", "error", err)
	}

	return PinnedManager{
		filePath: filePath,
	}
}

// Load reads the pinned directories from file and cleans non-existing ones
func (mgr *PinnedManager) Load() []directory {
	directories := []directory{}

	jsonData, err := os.ReadFile(mgr.filePath)
	if err != nil {
		slog.Error("Error reading pinned directories file", "error", err)
		return directories
	}

	// Check for the old format has been dropped in this manager
	if err := json.Unmarshal(jsonData, &directories); err != nil {
		slog.Error("Error parsing pinned directories data", "error", err)
		return directories
	}

	// Clean non-existing directories
	cleanedDirs := mgr.Clean(directories)

	return cleanedDirs
}

// Save marshals and writes the pinned directories to file.
func (mgr *PinnedManager) Save(dirs []directory) error {
	data, err := json.Marshal(dirs)
	if err != nil {
		return fmt.Errorf("error marshaling pinned directories: %w", err)
	}

	if err := os.WriteFile(mgr.filePath, data, utils.ConfigFilePerm); err != nil {
		return fmt.Errorf("error writing pinned directories file: %w", err)
	}

	return nil
}

// Toggle adds or removes a directory from the pinned directories list
func (mgr *PinnedManager) Toggle(dir string) error {
	dirs := mgr.Load()
	unPinned := false

	for i, other := range dirs {
		if other.Location == dir {
			dirs = append(dirs[:i], dirs[i+1:]...)
			unPinned = true
			break
		}
	}

	if !unPinned {
		dirs = append(dirs, directory{
			Location: dir,
			Name:     filepath.Base(dir),
		})
	}

	if err := mgr.Save(dirs); err != nil {
		return fmt.Errorf("error saving pinned directories: %w", err)
	}

	return nil
}

func (mgr *PinnedManager) Move(dir string, delta int) (bool, error) {
	if delta == 0 {
		return false, nil
	}

	dirs := mgr.Load()
	idx := -1
	for i := range dirs {
		if dirs[i].Location == dir {
			idx = i
			break
		}
	}
	if idx == -1 {
		return false, nil
	}

	target := idx + delta
	if target < 0 || target >= len(dirs) {
		return false, nil
	}

	dirs[idx], dirs[target] = dirs[target], dirs[idx]
	if err := mgr.Save(dirs); err != nil {
		return false, fmt.Errorf("error saving reordered pinned directories: %w", err)
	}

	return true, nil
}

// Clean removes non-existing directories and optionally saves the updated list
func (mgr *PinnedManager) Clean(dirs []directory) []directory {
	cleanedDirs := make([]directory, 0, len(dirs))
	for _, dir := range dirs {
		if _, err := os.Stat(dir.Location); err == nil {
			cleanedDirs = append(cleanedDirs, dir)
		} else if !os.IsNotExist(err) {
			slog.Warn("error while checking pinned directory", "directory", dir.Location, "error", err)
		}
	}

	if len(cleanedDirs) == len(dirs) {
		return cleanedDirs
	}

	if err := mgr.Save(cleanedDirs); err != nil {
		slog.Error("error saving pinned directories", "error", err)
	}

	return cleanedDirs
}
