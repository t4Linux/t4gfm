package internal

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/t4Linux/t4gfm/src/internal/common"
	"github.com/t4Linux/t4gfm/src/internal/ui/notify"
	"github.com/t4Linux/t4gfm/src/internal/utils"

	variable "github.com/t4Linux/t4gfm/src/config"
)

var openExternalURL = func(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
}

func (m *model) openEasterEggURL() tea.Cmd {
	const easterEggURL = "https://github.com/t4Linux"
	reqID := m.ioReqCnt
	m.ioReqCnt++
	return func() tea.Msg {
		if err := openExternalURL(easterEggURL); err != nil {
			return NewNotifyModalMsg(notify.New(true, "Failed to open URL", err.Error(), notify.NoAction), reqID)
		}
		return nil
	}
}

// Back to parent directory
func (m *model) parentDirectory() {
	err := m.getFocusedFilePanel().ParentDirectory()
	if err != nil {
		slog.Error("Error while changing to parent directory", "error", err)
	}
}

func (m *model) parentDirectoryFromSymlinkSource() {
	currentPath := m.getFocusedFilePanel().Location
	if currentPath == "" {
		return
	}

	resolvedPath, err := filepath.EvalSymlinks(currentPath)
	if err != nil {
		m.parentDirectory()
		return
	}

	parentPath := filepath.Dir(resolvedPath)
	if parentPath == "" || parentPath == "." {
		parentPath = string(filepath.Separator)
	}

	if updateErr := m.updateCurrentFilePanelDir(parentPath); updateErr != nil {
		slog.Error("Error while changing to symlink source parent", "error", updateErr, "target", parentPath)
	}
}

// Enter directory or open file with default application
// TODO: Unit test this
func (m *model) enterPanel() tea.Cmd {
	panel := m.getFocusedFilePanel()

	if panel.Empty() {
		return nil
	}
	selectedItem := panel.GetFocusedItem()
	if selectedItem.Directory {
		targetPath := selectedItem.Location

		if selectedItem.Info.Mode()&os.ModeSymlink != 0 {
			targetInfo, statErr := os.Stat(targetPath)
			if statErr != nil || !targetInfo.IsDir() {
				return nil
			}
		}
		// TODO : Propagate error out from this this function. Return here, instead of logging
		err := m.updateCurrentFilePanelDir(targetPath)
		if err != nil {
			slog.Error("Error while changing to directory", "error", err, "target", targetPath)
		}
		return nil
	}

	if variable.ChooserFile != "" {
		chooserErr := m.chooserFileWriteAndQuit(panel.GetFocusedItem().Location)
		if chooserErr == nil {
			return nil
		}
		// Continue with preview if file is not writable
		slog.Error("Error while writing to chooser file, continuing with file open", "error", chooserErr)
	}
	return m.executeOpenCommand()
}

func (m *model) executeOpenCommand() tea.Cmd {
	panel := m.getFocusedFilePanel()

	filePath := panel.GetFocusedItem().Location
	if m.blockUnsafeOpenPath(filePath) {
		return nil
	}
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(filePath), "."))
	if extEditor, ok := common.Config.OpenWith[ext]; ok {
		cmd := exec.Command(extEditor, filePath)
		utils.DetachFromTerminal(cmd)
		if err := cmd.Start(); err != nil {
			slog.Error("Error while open file with extension mapping", "error", err)
		}
		return nil
	}

	return m.openFileWithEditor()
}

func (m *model) openShellAtCurrentDir() tea.Cmd {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "bash"
	}
	cmd := exec.Command(shell)
	cmd.Dir = m.getFocusedFilePanel().Location
	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		return editorFinishedMsg{err}
	})
}

func (m *model) openLazyGitIfRepo() tea.Cmd {
	cwd := m.getFocusedFilePanel().Location
	check := exec.Command("git", "-C", cwd, "rev-parse", "--is-inside-work-tree")
	output, err := check.Output()
	if err != nil || strings.TrimSpace(string(output)) != "true" {
		return nil
	}

	cmd := exec.Command("lazygit")
	cmd.Dir = cwd
	return tea.ExecProcess(cmd, func(execErr error) tea.Msg {
		return editorFinishedMsg{execErr}
	})
}

func (m *model) displayFileLikeRanger() tea.Cmd {
	const maxPreviewWithEditorSize = int64(300 * 1024 * 1024)

	panel := m.getFocusedFilePanel()
	if panel.Empty() {
		return nil
	}

	item := panel.GetFocusedItem()
	if item.Directory {
		shell := os.Getenv("SHELL")
		if shell == "" {
			shell = "bash"
		}
		cmd := exec.Command(shell, "-lc", "ls -la --color=always -- \"$1\" | less -R", "_", item.Location)
		cmd.Dir = panel.Location
		return tea.ExecProcess(cmd, func(err error) tea.Msg { return editorFinishedMsg{err} })
	}

	if item.Info.Size() > maxPreviewWithEditorSize {
		m.notifyModel = notify.New(
			true,
			"Preview blocked",
			fmt.Sprintf("File is larger than 300M (%s).", common.FormatFileSize(item.Info.Size())),
			notify.NoAction,
		)
		return nil
	}

	if batPath, err := exec.LookPath("bat"); err == nil {
		cmd := exec.Command(batPath, "--paging=always", "--style=plain", "--color=always", item.Location)
		cmd.Dir = panel.Location
		return tea.ExecProcess(cmd, func(err error) tea.Msg { return editorFinishedMsg{err} })
	}

	if batcatPath, err := exec.LookPath("batcat"); err == nil {
		cmd := exec.Command(batcatPath, "--paging=always", "--style=plain", "--color=always", item.Location)
		cmd.Dir = panel.Location
		return tea.ExecProcess(cmd, func(err error) tea.Msg { return editorFinishedMsg{err} })
	}

	if lessPath, err := exec.LookPath("less"); err == nil {
		cmd := exec.Command(lessPath, "-R", item.Location)
		cmd.Dir = panel.Location
		return tea.ExecProcess(cmd, func(err error) tea.Msg { return editorFinishedMsg{err} })
	}

	return nil
}

// Switch to the directory where the sidebar cursor is located
func (m *model) sidebarSelectDirectory() {
	// We can't do this when we have only divider directories
	// m.sidebarModel.directories[m.sidebarModel.cursor].location would point to a divider dir.
	if m.sidebarModel.NoActualDir() {
		return
	}
	// TODO(Refactor): Move this to a function m.ResetFocus()
	m.focusPanel = nonePanelFocus
	panel := m.getFocusedFilePanel()
	target := m.sidebarModel.GetCurrentDirectoryLocation()
	if info, statErr := os.Stat(target); statErr != nil {
		slog.Error("Error switching to sidebar item", "error", statErr)
		return
	} else if !info.IsDir() {
		panel.IsFocused = true
		return
	}

	err := m.updateCurrentFilePanelDir(target)
	if err != nil {
		slog.Error("Error switching to sidebar directory", "error", err)
	}
	panel.IsFocused = true
}

// Toggle dotfile display or not
func (m *model) toggleDotFileController() {
	m.fileModel.ToggleDotFile()
	err := utils.WriteBoolFile(variable.ToggleDotFile, m.fileModel.DisplayDotFiles)
	if err != nil {
		slog.Error("Error while updating toggleDotFile data", "error", err)
	}
}

// Toggle dotfile display or not
func (m *model) toggleFooterController() tea.Cmd {
	if !m.toggleFooter {
		m.toggleFooter = true
		m.compactFooter = false
	} else {
		m.compactFooter = !m.compactFooter
	}
	err := utils.WriteBoolFile(variable.ToggleFooter, m.toggleFooter)
	if err != nil {
		slog.Error("Error while updating toggleFooter data", "error", err)
	}
	err = utils.WriteBoolFile(variable.ToggleCompactFooter, m.compactFooter)
	if err != nil {
		slog.Error("Error while updating compact footer data", "error", err)
	}
	m.setHeightValues()
	return m.updateComponentDimensions()
}

// Focus on search bar
func (m *model) searchBarFocus() {
	panel := m.getFocusedFilePanel()
	if panel.SearchBar.Focused() {
		panel.SearchBar.Blur()
	} else {
		panel.SearchBar.Focus()
		m.firstTextInput = true
	}

	// config search bar width
	panel.SearchBar.Width = m.fileModel.SinglePanelWidth - common.InnerPadding
}

func (m *model) sidebarSearchBarFocus() {
	if m.sidebarModel.SearchBarFocused() {
		// Ideally Code should never reach here. Once sidebar is focused, we should
		// not cause sidebarSearchBarFocus() event by pressing search key
		slog.Error("sidebarSearchBarFocus() called on Focused sidebar")
		m.sidebarModel.SearchBarBlur()
		return
	}
	m.sidebarModel.SearchBarFocus()
	m.firstTextInput = true
}
