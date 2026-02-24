package internal

import (
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/t4Linux/t4gfm/src/internal/common"
	"github.com/t4Linux/t4gfm/src/internal/ui/filemodel"
	"github.com/t4Linux/t4gfm/src/internal/ui/filepanel"
	"github.com/t4Linux/t4gfm/src/internal/ui/sortmodel"

	"github.com/t4Linux/t4gfm/src/internal/ui/notify"

	tea "github.com/charmbracelet/bubbletea"

	variable "github.com/t4Linux/t4gfm/src/config"
)

// mainKey handles most of key commands in the regular state of the application. For
// keys that performs actions in multiple panels, like going up or down,
// check the state of model m and handle properly.
// TODO: This function has grown too big. It needs to be fixed, via major
// updates and fixes in key handling code
func (m *model) mainKey(msg string) tea.Cmd { //nolint: gocyclo,cyclop,funlen // See above
	switch {
	// If move up Key is pressed, check the current state and executes
	case slices.Contains(common.Hotkeys.ListUp, msg):
		switch m.focusPanel {
		case sidebarFocus:
			m.sidebarModel.ListUp()
		case processBarFocus:
			m.processBarModel.ListUp()
		case metadataFocus:
			m.fileMetaData.ListUp()
		case gitPanelFocus:
			m.gitPanel.ListUp()
		case nonePanelFocus:
			if m.getFocusedFilePanel().IsVisualSelectMode() {
				m.getFocusedFilePanel().ItemSelectUp()
			} else {
				m.getFocusedFilePanel().ListUp()
			}
		}

		// If move down Key is pressed, check the current state and executes
	case slices.Contains(common.Hotkeys.ListDown, msg):
		switch m.focusPanel {
		case sidebarFocus:
			m.sidebarModel.ListDown()
		case processBarFocus:
			m.processBarModel.ListDown()
		case metadataFocus:
			m.fileMetaData.ListDown()
		case gitPanelFocus:
			m.gitPanel.ListDown()
		case nonePanelFocus:
			if m.getFocusedFilePanel().IsVisualSelectMode() {
				m.getFocusedFilePanel().ItemSelectDown()
			} else {
				m.getFocusedFilePanel().ListDown()
			}
		}

	case m.focusPanel == gitPanelFocus && msg == "H":
		m.gitPanel.PrevTab()

	case m.focusPanel == gitPanelFocus && msg == "L":
		m.gitPanel.NextTab()

	case m.focusPanel == nonePanelFocus && msg == "J" && m.fileModel.FilePreview.IsOpen():
		if m.fileModel.FilePreview.ScrollTextDown(1) {
			return m.fileModel.GetFilePreviewCmd(true)
		}

	case m.focusPanel == nonePanelFocus && msg == "K" && m.fileModel.FilePreview.IsOpen():
		if m.fileModel.FilePreview.ScrollTextUp(1) {
			return m.fileModel.GetFilePreviewCmd(true)
		}

	case m.focusPanel == nonePanelFocus && msg == "ctrl+d" && m.fileModel.FilePreview.IsOpen():
		if m.fileModel.FilePreview.ScrollTextPageDown() {
			return m.fileModel.GetFilePreviewCmd(true)
		}

	case m.focusPanel == nonePanelFocus && msg == "ctrl+u" && m.fileModel.FilePreview.IsOpen():
		if m.fileModel.FilePreview.ScrollTextPageUp() {
			return m.fileModel.GetFilePreviewCmd(true)
		}

	case slices.Contains(common.Hotkeys.PageUp, msg):
		m.getFocusedFilePanel().PgUp()

	case slices.Contains(common.Hotkeys.PageDown, msg):
		m.getFocusedFilePanel().PgDown()

	case msg == " " || msg == "space" || msg == "V" || msg == "shift+v" || slices.Contains(common.Hotkeys.ChangePanelMode, msg):
		panel := m.getFocusedFilePanel()
		if msg == " " || msg == "space" {
			switch panel.PanelMode {
			case filepanel.BrowserMode:
				panel.ChangeFilePanelMode()
				if !panel.Empty() {
					panel.SingleItemSelect()
				}
			case filepanel.SelectMode:
				panel.SingleItemSelect()
			}
		} else if msg == "V" || msg == "shift+v" {
			panel.ToggleVisualSelectMode()
		} else {
			panel.ChangeFilePanelMode()
		}

	case slices.Contains(common.Hotkeys.NextFilePanel, msg):
		if m.focusPanel == nonePanelFocus {
			m.fileModel.NextFilePanel()
		}

	case slices.Contains(common.Hotkeys.PreviousFilePanel, msg):
		if m.focusPanel == nonePanelFocus {
			m.fileModel.PreviousFilePanel()
		}

	case slices.Contains(common.Hotkeys.CloseFilePanel, msg):
		cmd, err := m.fileModel.CloseFilePanel()
		if err != nil && !errors.Is(err, filemodel.ErrMinimumPanelCount) {
			slog.Error("unexpected error while closing new panel", "error", err)
		} else if err == nil {
			m.persistUIState()
		}
		return cmd
	case slices.Contains(common.Hotkeys.CreateNewFilePanel, msg):
		cmd, err := m.fileModel.CreateNewFilePanel(variable.HomeDir)
		if err != nil && !errors.Is(err, filemodel.ErrMaximumPanelCount) {
			slog.Error("unexpected error while creating new panel", "error", err)
		} else if err == nil {
			m.persistUIState()
		}
		return cmd
	case slices.Contains(common.Hotkeys.ToggleFilePreviewPanel, msg):
		cmd := m.fileModel.ToggleFilePreviewPanel()
		m.persistUIState()
		return cmd

	case slices.Contains(common.Hotkeys.FocusOnSidebar, msg):
		m.focusOnSideBar()

	case slices.Contains(common.Hotkeys.FocusOnProcessBar, msg):
		m.focusOnProcessBar()

	case slices.Contains(common.Hotkeys.FocusOnMetaData, msg):
		m.focusOnMetadata()

	case slices.Contains(common.Hotkeys.FocusOnGit, msg):
		m.focusOnGit()

	case slices.Contains(common.Hotkeys.JumpToPinned, msg):
		m.jumpToSidebarSectionPinned()

	case slices.Contains(common.Hotkeys.JumpToDisks, msg):
		m.jumpToSidebarSectionDisks()

	case slices.Contains(common.Hotkeys.PasteItems, msg):
		return m.getPasteItemCmd()

	case slices.Contains(common.Hotkeys.FilePanelItemCreate, msg):
		m.panelCreateNewFile()
	case slices.Contains(common.Hotkeys.PinnedDirectory, msg):
		m.pinnedDirectory()

	case slices.Contains(common.Hotkeys.ToggleDotFile, msg):
		m.toggleDotFileController()

	case msg == "b" || slices.Contains(common.Hotkeys.ToggleSidebar, msg):
		return m.toggleSidebarController()

	case slices.Contains(common.Hotkeys.ToggleFooter, msg):
		return m.toggleFooterController()

	case slices.Contains(common.Hotkeys.ExtractFile, msg):
		return m.getExtractFileCmd()

	case slices.Contains(common.Hotkeys.CompressFile, msg):
		return m.getCompressSelectedFilesCmd()

	case slices.Contains(common.Hotkeys.OpenCommandLine, msg):
		m.promptModal.Open(true)
	case slices.Contains(common.Hotkeys.OpenT4gfmPrompt, msg):
		m.promptModal.Open(false)
	case slices.Contains(common.Hotkeys.OpenZoxide, msg):
		return m.zoxideModal.Open()

	case slices.Contains(common.Hotkeys.OpenHelpMenu, msg):
		m.helpMenu.Open()

	case slices.Contains(common.Hotkeys.OpenSortOptionsMenu, msg):
		m.sortModal.Open(m.getFocusedFilePanel().SortKind)

	case slices.Contains(common.Hotkeys.ToggleReverseSort, msg):
		m.getFocusedFilePanel().ToggleReverseSort()

	case slices.Contains(common.Hotkeys.OpenFileWithEditor, msg):
		if msg == "R" {
			m.panelOpenWithPrompt()
			return nil
		}
		return m.openFileWithEditor()

	case slices.Contains(common.Hotkeys.OpenCurrentDirectoryWithEditor, msg):
		return m.openDirectoryWithEditor()

	default:
		return m.normalAndBrowserModeKey(msg)
	}

	return nil
}

func (m *model) normalAndBrowserModeKey(msg string) tea.Cmd {
	if msg == "C" {
		if m.getFocusedFilePanel().IsFocused && m.fileModel.PanelCount() == 2 {
			return m.getTransferToOtherPanelCmd(false)
		}
		m.notifyModel = notify.New(true, "Two panels required",
			"Shift+C and Shift+M work only with exactly two open main panels and active file panel focus.",
			notify.NoAction)
		return nil
	}
	if msg == "M" {
		if m.getFocusedFilePanel().IsFocused && m.fileModel.PanelCount() == 2 {
			return m.getTransferToOtherPanelCmd(true)
		}
		m.notifyModel = notify.New(true, "Two panels required",
			"Shift+C and Shift+M work only with exactly two open main panels and active file panel focus.",
			notify.NoAction)
		return nil
	}

	// if not focus on the filepanel return
	if !m.getFocusedFilePanel().IsFocused {
		if isRangerPrefixStart(msg) {
			m.rangerPrefix = normalizeRangerPrefixStart(msg)
			return nil
		}
		if m.focusPanel == sidebarFocus && (msg == "esc" || msg == "ctrl+[") && m.sidebarModel.HasSearchQuery() {
			m.sidebarModel.ClearSearch()
			return nil
		}
		if m.focusPanel == sidebarFocus && slices.Contains(common.Hotkeys.Confirm, msg) {
			m.sidebarSelectDirectory()
		}
		if m.focusPanel == sidebarFocus && (msg == "K" || slices.Contains(common.Hotkeys.MovePinnedItemUp, msg)) {
			m.sidebarModel.MoveCurrentPinned(-1, m.getFocusedFilePanel().Location)
			return nil
		}
		if m.focusPanel == sidebarFocus && (msg == "J" || slices.Contains(common.Hotkeys.MovePinnedItemDown, msg)) {
			m.sidebarModel.MoveCurrentPinned(1, m.getFocusedFilePanel().Location)
			return nil
		}
		if m.focusPanel == sidebarFocus && slices.Contains(common.Hotkeys.FilePanelItemRename, msg) {
			m.sidebarModel.PinnedItemRename()
		}
		if m.focusPanel == sidebarFocus && slices.Contains(common.Hotkeys.SearchBar, msg) {
			m.sidebarSearchBarFocus()
		}
		return nil
	}
	// Check if in the select mode and focusOn filepanel
	if m.getFocusedFilePanel().PanelMode == filepanel.SelectMode {
		panel := m.getFocusedFilePanel()
		switch {
		case msg == "o" && panel.IsVisualSelectMode():
			panel.SwapVisualSelectionEnds()
		case msg == "esc" || msg == "ctrl+[":
			panel.ChangeFilePanelMode()
		case msg == " " || msg == "space":
			panel.SingleItemSelect()
		case slices.Contains(common.Hotkeys.Confirm, msg):
			panel.SingleItemSelect()
		case isRangerPrefixStart(msg):
			m.rangerPrefix = normalizeRangerPrefixStart(msg)
		case slices.Contains(common.Hotkeys.FilePanelSelectModeItemsSelectUp, msg):
			panel.ItemSelectUp()
		case slices.Contains(common.Hotkeys.FilePanelSelectModeItemsSelectDown, msg):
			panel.ItemSelectDown()
		case msg == "G":
			m.moveCursorToBottom()
		case slices.Contains(common.Hotkeys.DeleteItems, msg):
			return m.getDeleteTriggerCmd(false)
		case slices.Contains(common.Hotkeys.PermanentlyDeleteItems, msg):
			return m.getDeleteTriggerCmd(true)
		case slices.Contains(common.Hotkeys.CopyItems, msg):
			m.copyMultipleItem(false)
			if panel.SelectedCount() > 0 {
				panel.ChangeFilePanelMode()
			}
		case slices.Contains(common.Hotkeys.CutItems, msg):
			m.copyMultipleItem(true)
		case slices.Contains(common.Hotkeys.FilePanelSelectAllItem, msg):
			panel.SelectAllItem()
		}
		return nil
	}

	switch {
	case msg == "esc" || msg == "ctrl+[":
		if m.getFocusedFilePanel().SearchBar.Value() != "" {
			m.cancelSearch()
			return nil
		}
	case slices.Contains(common.Hotkeys.Confirm, msg):
		return m.enterPanel()
	case isRangerPrefixStart(msg):
		m.rangerPrefix = normalizeRangerPrefixStart(msg)
	case msg == "G":
		m.moveCursorToBottom()
	case msg == "L":
		return m.openLazyGitIfRepo()
	case msg == "S":
		return m.openShellAtCurrentDir()
	case msg == "i":
		return m.displayFileLikeRanger()
	case msg == "shift+left":
		m.parentDirectoryFromSymlinkSource()
	case slices.Contains(common.Hotkeys.ParentDirectory, msg):
		if msg == "backspace" {
			m.jumpToPreviousLocation()
		} else {
			m.parentDirectory()
		}
	case slices.Contains(common.Hotkeys.DeleteItems, msg):
		return m.getDeleteTriggerCmd(false)
	case slices.Contains(common.Hotkeys.PermanentlyDeleteItems, msg):
		return m.getDeleteTriggerCmd(true)
	case slices.Contains(common.Hotkeys.CopyItems, msg):
		m.copySingleItem(false)
	case slices.Contains(common.Hotkeys.CutItems, msg):
		m.copySingleItem(true)
	case slices.Contains(common.Hotkeys.FilePanelItemRename, msg):
		m.panelItemRename()
	case slices.Contains(common.Hotkeys.SearchBar, msg):
		m.searchBarFocus()
	case slices.Contains(common.Hotkeys.CopyPath, msg):
		m.copyPath()
	case slices.Contains(common.Hotkeys.CopyPWD, msg):
		m.copyPWD()
	}
	return nil
}

func isRangerPrefixStart(msg string) bool {
	switch normalizeRangerPrefixStart(msg) {
	case "y", "d", "p", "g", "o", "z", "m", ";", "s", "c", "+", "-":
		return true
	default:
		return false
	}
}

func normalizeRangerPrefixStart(msg string) string {
	switch msg {
	case "=", "shift+=", "kp+":
		return "+"
	case "_", "shift+-", "kp-":
		return "-"
	default:
		return msg
	}
}

func (m *model) rangerPrefixKey(msg string) tea.Cmd {
	prefix := m.rangerPrefix
	if prefix == "" {
		return nil
	}

	if slices.Contains(common.Hotkeys.CancelTyping, msg) || slices.Contains(common.Hotkeys.Quit, msg) {
		m.rangerPrefix = ""
		return nil
	}

	var cmd tea.Cmd
	switch prefix {
	case "y":
		switch msg {
		case "y":
			if m.getFocusedFilePanel().IsVisualSelectMode() || m.getFocusedFilePanel().PanelMode == filepanel.SelectMode {
				m.copyMultipleItem(false)
				if m.getFocusedFilePanel().SelectedCount() > 0 {
					m.getFocusedFilePanel().ChangeFilePanelMode()
				}
			} else {
				m.copySingleItem(false)
			}
		case "p":
			m.copyPath()
		case "d":
			m.copyPWD()
		case "n":
			m.copyName()
		case ".":
			m.copyNameWithoutExtension()
		}
	case "d":
		switch msg {
		case "d":
			if m.getFocusedFilePanel().PanelMode == filepanel.SelectMode {
				m.copyMultipleItem(true)
			} else {
				m.copySingleItem(true)
			}
		case "u":
			m.clipboard.Reset(false)
		case "D":
			cmd = m.getDeleteTriggerCmd(true)
		case "T":
			cmd = m.getDeleteTriggerCmd(false)
		}
	case "p":
		switch msg {
		case "p", "o", "P", "O", "l", "L":
			cmd = m.getPasteVariantCmd("p" + msg)
		case "h":
			m.rangerPrefix = "ph"
			return nil
		}
	case "ph":
		switch msg {
		case "l", "t":
			cmd = m.getPasteVariantCmd("ph" + msg)
		}
	case "g":
		cmd = m.handleGoPrefix(msg)
	case "o":
		m.handleSortPrefix(msg)
	case "z":
		switch msg {
		case "h":
			m.toggleDotFileController()
		case "f":
			m.searchBarFocus()
		}
	case "s":
		switch msg {
		case "s":
			m.focusOnProcessBar()
		case "m":
			m.focusOnMetadata()
		case "a":
			m.focusOnMainPanel()
		case "g":
			m.focusOnGit()
		case "p":
			m.jumpToSidebarSectionPinned()
		case "d":
			m.jumpToSidebarSectionDisks()
		case "f":
			m.jumpToSidebarSectionList()
		}
	case "m":
		if isMarkChar(msg) {
			m.rangerPrefix = "m" + strings.ToLower(msg)
			return nil
		}
	case "c":
		switch msg {
		case "z":
			cmd = m.getCompressSelectedFilesCmdWithFormat(compressFormatZip)
		case "t":
			cmd = m.getCompressSelectedFilesCmdWithFormat(compressFormatTarGz)
		case "j":
			cmd = m.getCompressSelectedFilesCmdWithFormat(compressFormatTarXz)
		case "l":
			cmd = m.getCompressSelectedFilesCmdWithFormat(compressFormatTarZs)
		case "E":
			m.openEncryptArchivePrompt()
		case "s":
			m.cycleCompressLevel()
		case "v":
			m.toggleCompressVerbose()
		case "e":
			m.toggleCompressExclude()
		}
	case "+":
		if isChmodClassChar(msg) {
			m.rangerPrefix = "+" + strings.ToLower(msg)
			return nil
		}
		if isChmodPermChar(msg) {
			m.applyChmodSymbol("+" + msg)
		}
	case "-":
		if isChmodClassChar(msg) {
			m.rangerPrefix = "-" + strings.ToLower(msg)
			return nil
		}
		if isChmodPermChar(msg) {
			m.applyChmodSymbol("-" + msg)
		}
	case "+u", "+g", "+o", "+a", "-u", "-g", "-o", "-a":
		if isChmodPermChar(msg) {
			if strings.HasPrefix(prefix, "+") {
				m.applyChmodSymbol(strings.TrimPrefix(prefix, "+") + "+" + msg)
			} else {
				m.applyChmodSymbol(strings.TrimPrefix(prefix, "-") + "-" + msg)
			}
		}
	case "ma", "mb", "mc", "md", "me", "mf", "mg", "mh", "mi", "mj", "mk", "ml", "mm", "mn", "mo", "mp", "mq", "mr", "ms", "mt", "mu", "mv", "mw", "mx", "my", "mz":
		if isMarkChar(msg) {
			key := strings.TrimPrefix(prefix, "m") + strings.ToLower(msg)
			m.rangerMarks[key] = m.getFocusedFilePanel().Location
			m.saveRangerMarks()
		}
	case ";":
		if msg == "?" {
			m.rangerPrefix = ";?"
			return nil
		}
		if msg == "D" {
			m.rangerMarks = map[string]string{}
			m.saveRangerMarks()
			break
		}
		if msg == "d" {
			m.rangerPrefix = ";d"
			return nil
		}
		if isMarkChar(msg) {
			m.rangerPrefix = ";" + strings.ToLower(msg)
			return nil
		}
	case ";?":
	case ";d":
		if isMarkChar(msg) {
			m.rangerPrefix = ";d" + strings.ToLower(msg)
			return nil
		}
	case ";da", ";db", ";dc", ";dd", ";de", ";df", ";dg", ";dh", ";di", ";dj", ";dk", ";dl", ";dm", ";dn", ";do", ";dp", ";dq", ";dr", ";ds", ";dt", ";du", ";dv", ";dw", ";dx", ";dy", ";dz":
		if isMarkChar(msg) {
			key := strings.TrimPrefix(prefix, ";d") + strings.ToLower(msg)
			delete(m.rangerMarks, key)
			m.saveRangerMarks()
		}
	case ";a", ";b", ";c", ";e", ";f", ";g", ";h", ";i", ";j", ";k", ";l", ";m", ";n", ";o", ";p", ";q", ";r", ";s", ";t", ";u", ";v", ";w", ";x", ";y", ";z":
		if isMarkChar(msg) {
			key := strings.TrimPrefix(prefix, ";") + strings.ToLower(msg)
			if path, ok := m.rangerMarks[key]; ok {
				cmd = m.changePanelDir(path)
			}
		}
	}

	m.rangerPrefix = ""
	return cmd
}

func isMarkChar(msg string) bool {
	if len(msg) != 1 {
		return false
	}
	c := msg[0]
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func isChmodClassChar(msg string) bool {
	if len(msg) != 1 {
		return false
	}
	switch msg[0] {
	case 'u', 'g', 'o', 'a', 'U', 'G', 'O', 'A':
		return true
	default:
		return false
	}
}

func isChmodPermChar(msg string) bool {
	if len(msg) != 1 {
		return false
	}
	switch msg[0] {
	case 'r', 'w', 'x', 'X', 's', 't':
		return true
	default:
		return false
	}
}

func (m *model) moveCursorToTop() {
	panel := m.getFocusedFilePanel()
	if panel.Empty() {
		return
	}
	panel.SetCursorPosition(0)
	if panel.IsVisualSelectMode() {
		panel.EnsureVisualSelection()
	}
}

func (m *model) moveCursorToBottom() {
	panel := m.getFocusedFilePanel()
	if panel.Empty() {
		return
	}
	panel.SetCursorPosition(panel.ElemCount() - 1)
	if panel.IsVisualSelectMode() {
		panel.EnsureVisualSelection()
	}
}

func (m *model) handleGoPrefix(msg string) tea.Cmd {
	switch msg {
	case "g":
		m.moveCursorToTop()
		return nil
	case "h":
		return m.changePanelDir(variable.HomeDir)
	case "l":
		m.parentDirectoryFromSymlinkSource()
		return nil
	case "/", "r":
		return m.changePanelDir(string(filepath.Separator))
	case "p":
		return m.changePanelDir(filepath.Join(string(filepath.Separator), "tmp"))
	case "e":
		return m.changePanelDir(filepath.Join(string(filepath.Separator), "etc"))
	case "u":
		return m.changePanelDir(filepath.Join(string(filepath.Separator), "usr"))
	case "d":
		return m.changePanelDir(filepath.Join(string(filepath.Separator), "dev"))
	case "o":
		return m.changePanelDir(filepath.Join(string(filepath.Separator), "opt"))
	case "v":
		return m.changePanelDir(filepath.Join(string(filepath.Separator), "var"))
	case "m":
		return m.changePanelDir(filepath.Join(string(filepath.Separator), "media"))
	case "M":
		return m.changePanelDir(filepath.Join(string(filepath.Separator), "mnt"))
	case "s":
		return m.changePanelDir(filepath.Join(string(filepath.Separator), "srv"))
	default:
		return nil
	}
}

func (m *model) changePanelDir(path string) tea.Cmd {
	if _, err := os.Stat(path); err != nil {
		slog.Debug("Skipping unavailable ranger-compatible path", "path", path, "error", err)
		return nil
	}
	if err := m.updateCurrentFilePanelDir(path); err != nil {
		slog.Error("Failed to change directory from ranger prefix", "path", path, "error", err)
	}
	return nil
}

func (m *model) handleSortPrefix(msg string) {
	panel := m.getFocusedFilePanel()
	setKind := func(kind sortmodel.SortKind, reversed bool) {
		panel.SortKind = kind
		panel.SortReversed = reversed
	}

	switch msg {
	case "r":
		panel.ToggleReverseSort()
	case "n":
		setKind(sortmodel.SortByName, false)
	case "s":
		setKind(sortmodel.SortBySize, false)
	case "m":
		setKind(sortmodel.SortByDate, false)
	case "t":
		setKind(sortmodel.SortByType, false)
	case "N":
		setKind(sortmodel.SortByName, true)
	case "S":
		setKind(sortmodel.SortBySize, true)
	case "M":
		setKind(sortmodel.SortByDate, true)
	case "T":
		setKind(sortmodel.SortByType, true)
	case "o":
		m.sortModal.Open(panel.SortKind)
	}
}

// Check the hotkey to cancel operation or create file
func (m *model) typingModalOpenKey(msg string) tea.Cmd {
	switch {
	case slices.Contains(common.Hotkeys.CancelTyping, msg):
		m.typingModal.errorMesssage = ""
		m.cancelTypingModal()
	case slices.Contains(common.Hotkeys.ConfirmTyping, msg):
		if m.typingModal.mode == typingModalOpenWith {
			return m.confirmOpenWith()
		}
		if m.typingModal.mode == typingModalEncryptArchive {
			return m.confirmEncryptArchive()
		}
		if m.typingModal.mode == typingModalDecryptArchive {
			return m.confirmDecryptArchive()
		}
		m.createItem()
	}
	return nil
}

func (m *model) notifyModelOpenKey(msg string) tea.Cmd {
	isCancel := slices.Contains(common.Hotkeys.CancelTyping, msg) || slices.Contains(common.Hotkeys.Quit, msg)
	isConfirm := slices.Contains(common.Hotkeys.ConfirmTyping, msg)

	if !isCancel && !isConfirm {
		slog.Warn("Invalid keypress in notifyModel", "msg", msg)
		return nil
	}
	m.notifyModel.Close()
	action := m.notifyModel.GetConfirmAction()
	if isCancel {
		return m.handleNotifyModelCancel(action)
	}
	return m.handleNotifyModelConfirm(action)
}

func (m *model) handleNotifyModelCancel(action notify.ConfirmActionType) tea.Cmd {
	switch action {
	case notify.RenameAction:
		m.cancelRename()
	case notify.QuitAction:
		m.modelQuitState = notQuitting
	case notify.ExtractAction:
		m.pendingExtractPath = ""
	case notify.DeleteAction, notify.NoAction, notify.PermanentDeleteAction:
		// Do nothing
	default:
		slog.Error("Unknown type of action", "action", action)
	}
	return nil
}

func (m *model) handleNotifyModelConfirm(action notify.ConfirmActionType) tea.Cmd {
	switch action {
	case notify.DeleteAction:
		return m.getDeleteCmd(false)
	case notify.PermanentDeleteAction:
		return m.getDeleteCmd(true)
	case notify.RenameAction:
		m.confirmRename()
	case notify.QuitAction:
		m.modelQuitState = quitConfirmationReceived
	case notify.ExtractAction:
		path := m.pendingExtractPath
		if shouldPromptDecryptArchive(path) {
			m.openDecryptArchivePrompt(path)
			return nil
		}
		m.pendingExtractPath = ""
		return m.getExtractFileCmdForPath(path)
	case notify.NoAction:
		// Ignore
	default:
		slog.Error("Unknown type of action", "action", action)
	}
	return nil
}

// Handles key inputs inside sort options menu
func (m *model) sortOptionsKey(msg string) {
	switch {
	case slices.Contains(common.Hotkeys.OpenSortOptionsMenu, msg):
		m.sortModal.Close()
	case slices.Contains(common.Hotkeys.Quit, msg):
		m.sortModal.Close()
	case slices.Contains(common.Hotkeys.Confirm, msg):
		m.confirmSortOptions()
	case slices.Contains(common.Hotkeys.ListUp, msg):
		m.sortModal.ListUp()
	case slices.Contains(common.Hotkeys.ListDown, msg):
		m.sortModal.ListDown()
	}
}

func (m *model) renamingKey(msg string) tea.Cmd {
	switch {
	case slices.Contains(common.Hotkeys.CancelTyping, msg):
		m.cancelRename()
	case slices.Contains(common.Hotkeys.ConfirmTyping, msg):
		if m.IsRenamingConflicting() {
			return m.warnModalForRenaming()
		}
		m.confirmRename()
	}

	return nil
}

func (m *model) sidebarRenamingKey(msg string) {
	switch {
	case slices.Contains(common.Hotkeys.CancelTyping, msg):
		m.sidebarModel.CancelSidebarRename()
	case slices.Contains(common.Hotkeys.ConfirmTyping, msg):
		m.sidebarModel.ConfirmSidebarRename()
	}
}

// Check the key input and cancel or confirms the search
func (m *model) focusOnSearchbarKey(msg string) tea.Cmd {
	switch {
	case msg == "esc" || msg == "ctrl+[":
		m.confirmSearch()
	case slices.Contains(common.Hotkeys.CancelTyping, msg):
		m.cancelSearch()
	case msg == "down":
		m.confirmSearch()
	case msg == "up":
		m.confirmSearch()
	case msg == "pgdown" || msg == "ctrl+f":
		m.confirmSearch()
		m.getFocusedFilePanel().PgDown()
	case msg == "pgup" || msg == "ctrl+b":
		m.confirmSearch()
		m.getFocusedFilePanel().PgUp()
	case slices.Contains(common.Hotkeys.ConfirmTyping, msg):
		m.confirmSearch()
	}
	return nil
}
