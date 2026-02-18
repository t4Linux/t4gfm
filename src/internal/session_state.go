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

type panelSessionState struct {
	PanelPaths         []string `json:"panel_paths"`
	FocusedPanelIndex  int      `json:"focused_panel_index"`
	FilePreviewVisible bool     `json:"file_preview_visible"`
}

func loadPanelSessionState() (*panelSessionState, error) {
	data, err := os.ReadFile(variable.PanelSessionFile)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var state *panelSessionState
	if err = json.Unmarshal(data, &state); err != nil {
		return nil, err
	}
	return state, nil
}

func (m *model) savePanelSessionState() {
	if err := os.MkdirAll(filepath.Dir(variable.PanelSessionFile), utils.ConfigDirPerm); err != nil {
		slog.Error("Failed to create panel session directory", "error", err)
		return
	}

	state := panelSessionState{
		FocusedPanelIndex:  m.fileModel.FocusedPanelIndex,
		FilePreviewVisible: m.fileModel.FilePreview.IsOpen(),
	}

	state.PanelPaths = make([]string, len(m.fileModel.FilePanels))
	for i, panel := range m.fileModel.FilePanels {
		state.PanelPaths[i] = panel.Location
	}

	data, err := json.Marshal(state)
	if err != nil {
		slog.Error("Failed to marshal panel session", "error", err)
		return
	}

	if err = os.WriteFile(variable.PanelSessionFile, data, utils.ConfigFilePerm); err != nil {
		slog.Error("Failed to save panel session", "error", err)
	}
}

func (m *model) persistUIState() {
	if err := utils.WriteBoolFile(variable.ToggleFilePreview, m.fileModel.FilePreview.IsOpen()); err != nil {
		slog.Error("Error while updating file preview toggle state", "error", err)
	}
	m.savePanelSessionState()
}
