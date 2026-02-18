package internal

import (
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"path/filepath"

	variable "github.com/t4Linux/t4gfm/src/config"
	"github.com/t4Linux/t4gfm/src/internal/utils"
)

func loadRangerMarks() map[string]string {
	marks := map[string]string{}
	data, err := os.ReadFile(variable.RangerMarksFile)
	if errors.Is(err, os.ErrNotExist) {
		return marks
	}
	if err != nil {
		slog.Error("Failed to read ranger marks file", "error", err)
		return marks
	}

	if err := json.Unmarshal(data, &marks); err != nil {
		slog.Error("Failed to parse ranger marks file", "error", err)
		return map[string]string{}
	}

	return marks
}

func (m *model) saveRangerMarks() {
	if err := os.MkdirAll(filepath.Dir(variable.RangerMarksFile), utils.ConfigDirPerm); err != nil {
		slog.Error("Failed to create ranger marks directory", "error", err)
		return
	}

	data, err := json.Marshal(m.rangerMarks)
	if err != nil {
		slog.Error("Failed to marshal ranger marks", "error", err)
		return
	}

	if err := os.WriteFile(variable.RangerMarksFile, data, utils.ConfigFilePerm); err != nil {
		slog.Error("Failed to save ranger marks", "error", err)
	}
}
