package internal

import (
	"log/slog"

	tea "github.com/charmbracelet/bubbletea"

	variable "github.com/t4Linux/t4gfm/src/config"
	"github.com/t4Linux/t4gfm/src/internal/common"
	"github.com/t4Linux/t4gfm/src/internal/ui/sidebar"
	"github.com/t4Linux/t4gfm/src/internal/utils"
)

// Pinned directory
func (m *model) pinnedDirectory() {
	panel := m.getFocusedFilePanel()
	err := m.sidebarModel.TogglePinnedDirectory(panel.Location)
	if err != nil {
		slog.Error("Error while toggling pinned directory", "error", err)
	}
}

// Focus on sidebar
func (m *model) focusOnSideBar() {
	if common.Config.SidebarWidth == 0 {
		return
	}
	if m.focusPanel == sidebarFocus {
		m.focusPanel = nonePanelFocus
		m.getFocusedFilePanel().IsFocused = true
	} else {
		m.focusPanel = sidebarFocus
		m.getFocusedFilePanel().IsFocused = false
	}
}

func (m *model) toggleSidebarController() tea.Cmd {
	if m.sidebarVisibleWidth < 5 || m.sidebarVisibleWidth > 20 {
		m.sidebarVisibleWidth = 20
	}

	if common.Config.SidebarWidth == 0 {
		common.Config.SidebarWidth = m.sidebarVisibleWidth
		if m.sidebarModel.Disabled() {
			m.sidebarModel = sidebar.New()
		}
	} else {
		m.sidebarVisibleWidth = common.Config.SidebarWidth
		common.Config.SidebarWidth = 0
		if m.focusPanel == sidebarFocus {
			m.focusPanel = nonePanelFocus
			m.getFocusedFilePanel().IsFocused = true
		}
	}
	err := utils.WriteBoolFile(variable.ToggleSidebar, common.Config.SidebarWidth != 0)
	if err != nil {
		slog.Error("Error while updating toggleSidebar data", "error", err)
	}
	return m.updateComponentDimensions()
}

// Focus on processbar
func (m *model) focusOnProcessBar() {
	if !m.toggleFooter {
		return
	}

	if m.focusPanel == processBarFocus {
		m.focusPanel = nonePanelFocus
		m.getFocusedFilePanel().IsFocused = true
	} else {
		m.focusPanel = processBarFocus
		m.getFocusedFilePanel().IsFocused = false
	}
}

// focus on metadata
func (m *model) focusOnMetadata() {
	if !m.toggleFooter {
		return
	}

	if m.focusPanel == metadataFocus {
		m.focusPanel = nonePanelFocus
		m.getFocusedFilePanel().IsFocused = true
	} else {
		m.focusPanel = metadataFocus
		m.getFocusedFilePanel().IsFocused = false
	}
}

func (m *model) focusOnMainPanel() {
	m.focusPanel = nonePanelFocus
	m.getFocusedFilePanel().IsFocused = true
}

func (m *model) focusOnGit() {
	if !m.toggleFooter {
		return
	}

	if m.focusPanel == gitPanelFocus {
		m.focusPanel = nonePanelFocus
		m.getFocusedFilePanel().IsFocused = true
	} else {
		m.focusPanel = gitPanelFocus
		m.getFocusedFilePanel().IsFocused = false
	}
}

func (m *model) jumpToSidebarSectionPinned() {
	if common.Config.SidebarWidth == 0 {
		return
	}
	m.focusPanel = sidebarFocus
	m.getFocusedFilePanel().IsFocused = false
	m.sidebarModel.ClearSearch()
	m.sidebarModel.UpdateDirectories(m.getFocusedFilePanel().Location)
	m.sidebarModel.JumpToPinned()
}

func (m *model) jumpToSidebarSectionDisks() {
	if common.Config.SidebarWidth == 0 {
		return
	}
	m.focusPanel = sidebarFocus
	m.getFocusedFilePanel().IsFocused = false
	m.sidebarModel.ClearSearch()
	m.sidebarModel.UpdateDirectories(m.getFocusedFilePanel().Location)
	m.sidebarModel.JumpToDisks()
}

func (m *model) jumpToSidebarSectionList() {
	if common.Config.SidebarWidth == 0 {
		return
	}
	m.focusPanel = sidebarFocus
	m.getFocusedFilePanel().IsFocused = false
	m.sidebarModel.ClearSearch()
	m.sidebarModel.UpdateDirectories(m.getFocusedFilePanel().Location)
	m.sidebarModel.JumpToList()
}
