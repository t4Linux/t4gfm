package internal

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	variable "github.com/t4Linux/t4gfm/src/config"
	"github.com/t4Linux/t4gfm/src/internal/ui/filepanel"

	"github.com/t4Linux/t4gfm/src/internal/ui/notify"
	"github.com/t4Linux/t4gfm/src/internal/ui/processbar"
	"github.com/t4Linux/t4gfm/src/internal/utils"

	"github.com/t4Linux/t4gfm/src/internal/common"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
)

// Create a file in the currently focus file panel
// TODO: Fix it. It doesn't creates a new file. It just opens a file model,
// that allows you to create a file. Actual creation happens here - createItem() in handle_modal.go
func (m *model) panelCreateNewFile() {
	panel := m.getFocusedFilePanel()

	m.typingModal.location = panel.Location
	m.typingModal.targetPath = ""
	m.typingModal.mode = typingModalCreate
	m.typingModal.open = true
	m.typingModal.textInput = common.GenerateNewFileTextInput()
	m.firstTextInput = true
}

func (m *model) panelOpenWithPrompt() {
	panel := m.getFocusedFilePanel()
	if panel.Empty() || panel.GetFocusedItem().Directory {
		return
	}

	targetPath := panel.GetFocusedItem().Location
	m.typingModal.location = panel.Location
	m.typingModal.targetPath = targetPath
	m.typingModal.mode = typingModalOpenWith
	m.typingModal.open = true
	m.typingModal.errorMesssage = ""
	m.typingModal.textInput = common.GeneratePromptTextInput()
	_ = m.typingModal.textInput.Focus()
	m.typingModal.textInput.Width = common.ModalWidth - 10
	m.typingModal.textInput.SetValue(common.ResolveEnterOpenProgram(targetPath))
	m.firstTextInput = false
}

func (m *model) applyChmodSymbol(symbol string) {
	panel := m.getFocusedFilePanel()
	if panel.Empty() {
		return
	}

	items := []string{}
	if panel.PanelMode == filepanel.SelectMode {
		panel.EnsureVisualSelection()
		items = panel.GetSelectedLocations()
	}
	if len(items) == 0 {
		items = []string{panel.GetFocusedItem().Location}
	}

	if _, err := exec.LookPath("chmod"); err != nil {
		m.notifyModel = notify.New(true, "chmod unavailable", "chmod command not found in PATH", notify.NoAction)
		return
	}

	args := append([]string{symbol}, items...)
	cmd := exec.Command("chmod", args...)
	if err := cmd.Run(); err != nil {
		m.notifyModel = notify.New(true, "chmod failed", err.Error(), notify.NoAction)
		return
	}

	for _, item := range items {
		m.fileMetaData.InvalidatePath(item)
	}

	m.fileModel.UpdateFilePanelsIfNeeded(true)
}

// TODO : This function does not needs the entire model. Only pass the panel object
func (m *model) IsRenamingConflicting() bool {
	// TODO : Replace this with m.getCurrentFilePanel() everywhere
	panel := m.getFocusedFilePanel()
	if panel.ElemCount() == 0 {
		slog.Error("IsRenamingConflicting() being called on empty panel")
		return false
	}
	oldPath := panel.GetFocusedItem().Location
	newPath := filepath.Join(panel.Location, panel.Rename.Value())

	if oldPath == newPath {
		return false
	}

	_, err := os.Stat(newPath)
	return err == nil
}

// TODO: Remove channel messaging and use tea.Cmd
func (m *model) warnModalForRenaming() tea.Cmd {
	reqID := m.ioReqCnt
	m.ioReqCnt++
	slog.Debug("Submitting rename notify model request", "reqID", reqID)
	res := func() tea.Msg {
		notifyModel := notify.New(true,
			common.SameRenameWarnTitle,
			common.SameRenameWarnContent,
			notify.RenameAction)
		return NewNotifyModalMsg(notifyModel, reqID)
	}
	return res
}

// Rename file where the cusror is located
// TODO: Fix this. It doesn't do any rename, just opens the rename text input
// Actual rename happens at confirmRename() in handle_modal.go
func (m *model) panelItemRename() {
	panel := m.getFocusedFilePanel()
	if panel.Empty() {
		return
	}

	cursorPos := -1
	nameRunes := []rune(panel.GetFocusedItem().Name)
	nameLen := len(nameRunes)
	for i := nameLen - 1; i >= 0; i-- {
		if nameRunes[i] == '.' {
			cursorPos = i
			break
		}
	}
	if cursorPos == -1 || cursorPos == 0 && nameLen > 0 || panel.GetFocusedItem().Directory {
		cursorPos = nameLen
	}

	m.fileModel.Renaming = true
	panel.Renaming = true
	m.firstTextInput = true
	// TODO: Don't re-create a new model on each rename. Don't create
	// unnecessary gargage for collection. Reuse the existing model.
	// Maintain its state, dimensions. Update its cursor and text when needed
	panel.Rename = common.GenerateRenameTextInput(
		m.fileModel.SinglePanelWidth-common.InnerPadding,
		cursorPos,
		panel.GetFocusedItem().Name)
}

func (m *model) getDeleteCmd(permDelete bool) tea.Cmd {
	panel := m.getFocusedFilePanel()
	if panel.Empty() {
		return nil
	}

	var items []string
	if panel.PanelMode == filepanel.SelectMode {
		items = panel.GetSelectedLocations()
	} else {
		items = []string{panel.GetFocusedItem().Location}
	}

	useTrash := m.hasTrash && !isExternalDiskPath(panel.Location) && !permDelete

	reqID := m.ioReqCnt
	m.ioReqCnt++
	slog.Debug("Submitting delete request", "id", reqID, "items cnt", len(items))
	return func() tea.Msg {
		state := deleteOperation(&m.processBarModel, items, useTrash)
		return NewDeleteOperationMsg(state, reqID)
	}
}

func deleteOperation(processBarModel *processbar.Model, items []string, useTrash bool) processbar.ProcessState {
	if len(items) == 0 {
		return processbar.Cancelled
	}
	p, err := processBarModel.SendAddProcessMsg(filepath.Base(items[0]), processbar.OpDelete, len(items), true)
	if err != nil {
		slog.Error("Cannot spawn a new process", "error", err)
		return processbar.Failed
	}

	deleteFunc := os.RemoveAll
	if useTrash {
		deleteFunc = moveToTrash
	}
	for _, item := range items {
		err = deleteFunc(item)
		if err != nil {
			p.State = processbar.Failed
			slog.Error("Error in delete operation", "item", item, "useTrash", useTrash, "error", err)
			break
		}
		p.CurrentFile = filepath.Base(item)
		p.Done++
		processBarModel.TrySendingUpdateProcessMsg(p)
	}

	if p.State != processbar.Failed {
		p.State = processbar.Successful
	}
	p.DoneTime = time.Now()
	err = processBarModel.SendUpdateProcessMsg(p, true)
	if err != nil {
		slog.Error("Failed to send final delete operation update", "error", err)
	}
	return p.State
}

func (m *model) getDeleteTriggerCmd(deletePermanent bool) tea.Cmd {
	panel := m.getFocusedFilePanel()
	if (panel.PanelMode == filepanel.SelectMode && panel.SelectedCount() == 0) ||
		(panel.PanelMode == filepanel.BrowserMode && panel.Empty()) {
		return nil
	}

	reqID := m.ioReqCnt
	m.ioReqCnt++

	return func() tea.Msg {
		title := common.TrashWarnTitle
		content := common.TrashWarnContent
		action := notify.DeleteAction

		if !m.hasTrash || isExternalDiskPath(panel.Location) || deletePermanent {
			title = common.PermanentDeleteWarnTitle
			content = common.PermanentDeleteWarnContent
			action = notify.PermanentDeleteAction
		}
		return NewNotifyModalMsg(notify.New(true, title, content, action), reqID)
	}
}

func (m *model) copySingleItem(cut bool) {
	panel := m.getFocusedFilePanel()
	m.clipboard.Reset(cut)
	if panel.Empty() {
		return
	}
	slog.Debug("handle_file_operations.copySingleItem", "cut", cut,
		"panel location", panel.GetFocusedItem().Location)
	m.clipboard.Add(panel.GetFocusedItem().Location)
}

// Copy all selected file or directory's paths to the clipboard
func (m *model) copyMultipleItem(cut bool) {
	panel := m.getFocusedFilePanel()
	panel.EnsureVisualSelection()
	m.clipboard.Reset(cut)
	items := panel.SelectedLocationsForCopy()
	if len(items) == 0 {
		return
	}
	slog.Debug("handle_file_operations.copyMultipleItem", "cut", cut,
		"panel selected files", items)
	m.clipboard.SetItems(items)
}

func (m *model) getPasteItemCmd() tea.Cmd {
	copyItems := m.clipboard.PruneInaccessibleItemsAndGet()
	cut := m.clipboard.IsCut()
	if len(copyItems) == 0 {
		return nil
	}

	// TODO: Do it via m.getNewReqID()
	// TODO: Have an IO Req Management, collecting info about pending IO Req too
	reqID := m.ioReqCnt
	m.ioReqCnt++
	panelLocation := m.getFocusedFilePanel().Location

	slog.Debug("Submitting pasteItems request", "id", reqID, "items cnt", len(copyItems), "dest", panelLocation)
	return func() tea.Msg {
		err := validatePasteOperation(panelLocation, copyItems, cut)
		if err != nil {
			return NewNotifyModalMsg(notify.New(true, "Invalid paste location", err.Error(), notify.NoAction),
				reqID)
		}
		state := executePasteOperation(&m.processBarModel, panelLocation, copyItems, cut)
		return NewPasteOperationMsg(state, reqID, panelLocation)
	}
}

func (m *model) getTransferToOtherPanelCmd(cut bool) tea.Cmd {
	if !m.getFocusedFilePanel().IsFocused || m.fileModel.PanelCount() != 2 {
		return nil
	}

	srcPanel := m.getFocusedFilePanel()
	if srcPanel.Empty() {
		return nil
	}

	var items []string
	if srcPanel.PanelMode == filepanel.SelectMode {
		srcPanel.EnsureVisualSelection()
		items = srcPanel.SelectedLocationsForCopy()
	} else {
		items = []string{srcPanel.GetFocusedItem().Location}
	}
	if len(items) == 0 {
		return nil
	}

	destIndex := 1 - m.fileModel.FocusedPanelIndex
	if destIndex < 0 || destIndex >= len(m.fileModel.FilePanels) {
		return nil
	}
	destPath := m.fileModel.FilePanels[destIndex].Location

	reqID := m.ioReqCnt
	m.ioReqCnt++

	slog.Debug("Submitting transfer-to-other-panel request", "id", reqID,
		"items", len(items), "cut", cut, "dest", destPath)

	return func() tea.Msg {
		err := validatePasteOperation(destPath, items, cut)
		if err != nil {
			return NewNotifyModalMsg(notify.New(true, "Invalid transfer destination", err.Error(), notify.NoAction), reqID)
		}
		state := executePasteOperation(&m.processBarModel, destPath, items, cut)
		return NewPasteOperationMsg(state, reqID, destPath)
	}
}

func (m *model) getPasteVariantCmd(variant string) tea.Cmd {
	switch variant {
	case "pp":
		return m.getPasteItemCmd()
	case "po":
		return m.getPasteItemCmdWithOptions(true, false)
	case "pP":
		return m.getPasteItemCmdWithOptions(false, true)
	case "pO":
		return m.getPasteItemCmdWithOptions(true, true)
	case "pl":
		return m.getPasteLinkCmd(false, false)
	case "pL":
		return m.getPasteLinkCmd(true, false)
	case "phl":
		return m.getPasteLinkCmd(false, true)
	case "pht":
		return m.getPasteHardlinkedSubtreeCmd()
	default:
		return nil
	}
}

func (m *model) getPasteItemCmdWithOptions(overwrite bool, appendMode bool) tea.Cmd {
	copyItems := m.clipboard.PruneInaccessibleItemsAndGet()
	cut := m.clipboard.IsCut()
	if len(copyItems) == 0 {
		return nil
	}

	reqID := m.ioReqCnt
	m.ioReqCnt++
	panelLocation := m.getFocusedFilePanel().Location

	return func() tea.Msg {
		err := validatePasteOperation(panelLocation, copyItems, cut)
		if err != nil {
			return NewNotifyModalMsg(notify.New(true, "Invalid paste location", err.Error(), notify.NoAction), reqID)
		}
		state := executePasteOperationWithOptions(&m.processBarModel, panelLocation, copyItems, cut, overwrite, appendMode)
		return NewPasteOperationMsg(state, reqID, panelLocation)
	}
}

func executePasteOperationWithOptions(processBarModel *processbar.Model,
	panelLocation string, copyItems []string, cut bool, overwrite bool, appendMode bool,
) processbar.ProcessState {
	var operation processbar.OperationType
	if cut {
		operation = processbar.OpCut
	} else {
		operation = processbar.OpCopy
	}

	p, err := processBarModel.SendAddProcessMsg(filepath.Base(copyItems[0]), operation, len(copyItems), true)
	if err != nil {
		slog.Error("Cannot spawn a new process", "error", err)
		return processbar.Failed
	}

	for _, src := range copyItems {
		dst := filepath.Join(panelLocation, filepath.Base(src))
		p.CurrentFile = filepath.Base(src)
		err = pasteWithOptions(src, dst, cut, overwrite, appendMode)
		if err != nil {
			p.State = processbar.Failed
			if p.ErrorMsg == "" {
				p.ErrorMsg = err.Error()
			}
			slog.Error("paste item error", "error", err)
			continue
		}
		p.Done++
		processBarModel.TrySendingUpdateProcessMsg(p)
	}

	if p.State != processbar.Failed {
		p.State = processbar.Successful
		p.Done = p.Total
	}
	p.DoneTime = time.Now()
	_ = processBarModel.SendUpdateProcessMsg(p, true)
	return p.State
}

func pasteWithOptions(src, dst string, cut bool, overwrite bool, appendMode bool) error {
	srcInfo, err := os.Lstat(src)
	if err != nil {
		return err
	}

	if overwrite && !appendMode {
		if _, err := os.Lstat(dst); err == nil {
			if err := os.RemoveAll(dst); err != nil {
				return err
			}
		}
	}

	if appendMode {
		if dstInfo, err := os.Lstat(dst); err == nil {
			if srcInfo.IsDir() && dstInfo.IsDir() {
				return mergeDir(src, dst, cut)
			}
			if !srcInfo.IsDir() && !dstInfo.IsDir() {
				if cut {
					if err := appendFile(src, dst, srcInfo); err != nil {
						return err
					}
					return os.Remove(src)
				}
				return appendFile(src, dst, srcInfo)
			}
			if overwrite {
				if err := os.RemoveAll(dst); err != nil {
					return err
				}
			}
		}
	}

	if srcInfo.Mode()&os.ModeSymlink != 0 {
		if cut {
			return moveElement(src, dst)
		}
		return copySymlink(src, dst)
	}

	if srcInfo.IsDir() {
		if cut {
			return moveElement(src, dst)
		}
		return copyDirNoRename(src, dst)
	}

	if cut {
		return moveElement(src, dst)
	}
	return copyFile(src, dst, srcInfo)
}

func appendFile(src, dst string, srcInfo os.FileInfo) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_APPEND, srcInfo.Mode())
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

func copyDirNoRename(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		newPath := filepath.Join(dst, relPath)
		if info.IsDir() {
			return os.MkdirAll(newPath, info.Mode())
		}
		return copyFile(path, newPath, info)
	})
}

func mergeDir(src, dst string, cut bool) error {
	err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if relPath == "." {
			return nil
		}
		target := filepath.Join(dst, relPath)
		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}
		if _, err := os.Lstat(target); err == nil {
			return appendFile(path, target, info)
		}
		return copyFile(path, target, info)
	})
	if err != nil {
		return err
	}
	if cut {
		return os.RemoveAll(src)
	}
	return nil
}

func (m *model) getPasteLinkCmd(relative bool, hardlink bool) tea.Cmd {
	items := m.clipboard.PruneInaccessibleItemsAndGet()
	if len(items) == 0 {
		return nil
	}

	reqID := m.ioReqCnt
	m.ioReqCnt++
	destDir := m.getFocusedFilePanel().Location

	return func() tea.Msg {
		state := processbar.Successful
		for _, src := range items {
			dst := filepath.Join(destDir, filepath.Base(src))
			dst, _ = renameIfDuplicate(dst)
			if hardlink {
				if err := os.Link(src, dst); err != nil {
					state = processbar.Failed
					break
				}
				continue
			}
			target := src
			if relative {
				rel, err := filepath.Rel(destDir, src)
				if err != nil {
					state = processbar.Failed
					break
				}
				target = rel
			}
			if err := os.Symlink(target, dst); err != nil {
				state = processbar.Failed
				break
			}
		}
		return NewPasteOperationMsg(state, reqID, destDir)
	}
}

func (m *model) getPasteHardlinkedSubtreeCmd() tea.Cmd {
	items := m.clipboard.PruneInaccessibleItemsAndGet()
	if len(items) == 0 {
		return nil
	}
	reqID := m.ioReqCnt
	m.ioReqCnt++
	destDir := m.getFocusedFilePanel().Location

	return func() tea.Msg {
		state := processbar.Successful
		for _, src := range items {
			dst := filepath.Join(destDir, filepath.Base(src))
			dst, _ = renameIfDuplicate(dst)
			if err := hardlinkTree(src, dst); err != nil {
				state = processbar.Failed
				break
			}
		}
		return NewPasteOperationMsg(state, reqID, destDir)
	}
}

func hardlinkTree(src, dst string) error {
	info, err := os.Lstat(src)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return os.Link(src, dst)
	}
	if err := os.MkdirAll(dst, info.Mode()); err != nil {
		return err
	}
	return filepath.Walk(src, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		target := filepath.Join(dst, rel)
		if fi.IsDir() {
			return os.MkdirAll(target, fi.Mode())
		}
		return os.Link(path, target)
	})
}

func validatePasteOperation(panelLocation string, copyItems []string, cut bool) error {
	// Check if trying to paste into source or subdirectory for both cut and copy operations
	for _, srcPath := range copyItems {
		// Check if trying to cut and paste into the same directory - this would be a no-op
		// and could potentially cause issues, so we prevent it
		if filepath.Dir(srcPath) == panelLocation && cut {
			return fmt.Errorf("cannot paste into parent directory of source, srcPath : %v, panelLocation : %v",
				srcPath, panelLocation)
		}
		if cut && srcPath == panelLocation {
			return errors.New("cannot paste a directory into itself")
		}

		if isAncestor(srcPath, panelLocation) {
			return fmt.Errorf("cannot %s and paste a directory into itself or its subdirectory",
				getCopyOrCutOperationName(cut))
		}
	}

	return nil
}

// new func to check and return an error that will go in m.content
// create a new error type

// Paste all clipboard items
func executePasteOperation(processBarModel *processbar.Model,
	panelLocation string, copyItems []string, cut bool,
) processbar.ProcessState {
	slog.Debug("executePasteOperation", "items", copyItems, "cut", cut, "panel location", panelLocation)

	var operation processbar.OperationType
	if cut {
		operation = processbar.OpCut
	} else {
		operation = processbar.OpCopy
	}

	p, err := processBarModel.SendAddProcessMsg(
		filepath.Base(copyItems[0]),
		operation,
		getTotalProgressUnits(copyItems), true)
	if err != nil {
		slog.Error("Cannot spawn a new process", "error", err)
		return processbar.Failed
	}
	startCancelGeneration := processBarModel.CancelGeneration()
	isCancelled := func() bool {
		return processBarModel.IsCancelRequested(startCancelGeneration)
	}

	for _, filePath := range copyItems {
		errMessage := "cut item error"
		if isCancelled() {
			err = errOperationCancelled
		} else if cut && !isExternalDiskPath(filePath) {
			err = moveElement(filePath, filepath.Join(panelLocation, filepath.Base(filePath)))
		} else {
			// TODO : These error cases are hard to test. We have to somehow make the paste operations fail,
			// which is time consuming and manual. We should test these with automated testcases
			err = pasteDir(filePath, filepath.Join(panelLocation, filepath.Base(filePath)), &p, cut, processBarModel, isCancelled)
			if err != nil {
				errMessage = "paste item error"
			}
		}

		p.CurrentFile = filepath.Base(filePath)
		if err != nil {
			if errors.Is(err, errOperationCancelled) {
				p.State = processbar.Cancelled
				p.ErrorMsg = "cancelled by user"
				break
			}
			slog.Debug("model.pasteItem - paste failure", "error", err,
				"current item", filePath, "errMessage", errMessage)
			p.State = processbar.Failed
			if p.ErrorMsg == "" {
				p.ErrorMsg = err.Error()
			}
			slog.Error(errMessage, "error", err)
			continue
		}
		processBarModel.TrySendingUpdateProcessMsg(p)
	}

	if p.State != processbar.Failed && p.State != processbar.Cancelled {
		p.State = processbar.Successful
		p.Done = p.Total
	}
	p.DoneTime = time.Now()
	err = processBarModel.SendUpdateProcessMsg(p, true)
	if err != nil {
		slog.Error("Could not send final update for process Bar", "error", err)
	}

	return p.State
}

func getTotalFilesCnt(copyItems []string) int {
	totalFiles := 0
	for _, folderPath := range copyItems {
		// TODO : Fix this. This is inefficient
		// In case of a cut operations for a directory with a lot of files
		// we are unnecessarily walking the whole directory recursively
		// while os will just perform a rename
		// So instead of few operations this will cause the cut paste
		// to read the whole directory recursively
		// we should avoid doing this.
		// Although this allows us a more detailed progress tracking
		// this make the copy/cut more inefficient
		// instead, we could just track progress based on total items in
		// copyItems
		// efficiency should be prioritized over more detailed feedback.
		count, err := countFiles(folderPath)
		if err != nil {
			slog.Error("Error in countFiles", "error", err)
			continue
		}
		totalFiles += count
	}
	return totalFiles
}

func getTotalProgressUnits(copyItems []string) int {
	totalUnits := 0
	for _, itemPath := range copyItems {
		err := filepath.Walk(itemPath, func(_ string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if info.IsDir() {
				return nil
			}
			totalUnits += progressUnitsForFile(info)
			return nil
		})
		if err != nil {
			slog.Debug("Skipping progress units for inaccessible path", "path", itemPath, "error", err)
		}
	}

	if totalUnits <= 0 {
		return max(getTotalFilesCnt(copyItems), 1)
	}
	return totalUnits
}

// Extract compressed file
// TODO : err should be returned and properly handled by the caller
func (m *model) getExtractFileCmd() tea.Cmd {
	panel := m.getFocusedFilePanel()
	if panel.Empty() {
		return nil
	}

	item := panel.GetFocusedItem().Location

	ext := strings.ToLower(filepath.Ext(item))
	if !common.IsExtensionExtractable(ext) {
		slog.Error("Error unexpected file", "extension type", ext, "item", item, "error", errors.ErrUnsupported)
		return nil
	}
	reqID := m.ioReqCnt
	m.ioReqCnt++

	slog.Debug("Submitting Extract file request", "reqID", reqID, "item", item)

	return func() tea.Msg {
		outputDir := common.FileNameWithoutExtension(item)
		outputDir, err := renameIfDuplicate(outputDir)
		if err != nil {
			slog.Error("Error while renaming for duplicates", "error", err)
			return NewExtractOperationMsg(processbar.Failed, reqID)
		}

		err = os.MkdirAll(
			outputDir,
			utils.ExtractedDirMode,
		)
		if err != nil {
			slog.Error("Error while making directory for extracting files", "error", err)
			return NewExtractOperationMsg(processbar.Failed, reqID)
		}
		err = extractCompressFile(item, outputDir, &m.processBarModel)
		if err != nil {
			slog.Error("Error extract file", "error", err)
			return NewExtractOperationMsg(processbar.Failed, reqID)
		}
		return NewExtractOperationMsg(processbar.Successful, reqID)
	}
}

func (m *model) getCompressSelectedFilesCmd() tea.Cmd {
	panel := m.getFocusedFilePanel()

	if panel.Empty() {
		return nil
	}
	var filesToCompress []string
	var firstFile string

	if panel.SelectedCount() == 0 {
		firstFile = panel.GetFocusedItem().Location
		filesToCompress = append(filesToCompress, firstFile)
	} else {
		firstFile = panel.GetFirstSelectedLocation()
		filesToCompress = panel.GetSelectedLocations()
	}

	reqID := m.ioReqCnt
	m.ioReqCnt++

	return func() tea.Msg {
		zipName, err := getZipArchiveName(filepath.Base(firstFile))
		if err != nil {
			slog.Error("Error in getZipArchiveName", "error", err)
			return NewCompressOperationMsg(processbar.Failed, reqID)
		}
		zipPath := filepath.Join(panel.Location, zipName)
		if err := zipSources(filesToCompress, zipPath, &m.processBarModel); err != nil {
			slog.Error("Error in zipping files", "error", err)
			return NewCompressOperationMsg(processbar.Failed, reqID)
		}
		return NewCompressOperationMsg(processbar.Successful, reqID)
	}
}

func (m *model) chooserFileWriteAndQuit(path string) error {
	// Attempt to write to the file
	err := os.WriteFile(variable.ChooserFile, []byte(path), utils.ConfigFilePerm)
	if err != nil {
		return err
	}
	m.modelQuitState = quitInitiated
	return nil
}

// Open file with default editor
func (m *model) openFileWithEditor() tea.Cmd {
	panel := m.getFocusedFilePanel()
	// Check if panel is empty
	if panel.Empty() {
		return nil
	}

	if variable.ChooserFile != "" {
		err := m.chooserFileWriteAndQuit(panel.GetFocusedItem().Location)
		if err == nil {
			return nil
		}
		// Continue with preview if file is not writable
		slog.Error("Error while writing to chooser file, continuing with open via file editor", "error", err)
	}

	if m.blockUnsafeOpenPath(panel.GetFocusedItem().Location) {
		return nil
	}

	return m.openFileWithCommand(common.ResolveEditorCommand(), panel.GetFocusedItem().Location)
}

func (m *model) openFileWithCommand(command string, targetPath string) tea.Cmd {
	if m.blockUnsafeOpenPath(targetPath) {
		return nil
	}

	parts := strings.Fields(strings.TrimSpace(command))
	if len(parts) == 0 {
		return nil
	}

	cmd := parts[0]
	args := append(parts[1:], targetPath)
	c := exec.Command(cmd, args...)

	return tea.ExecProcess(c, func(err error) tea.Msg {
		return editorFinishedMsg{err}
	})
}

func (m *model) blockUnsafeOpenPath(path string) bool {
	reason := common.OpenBlockReason(path)
	if reason == "" {
		return false
	}
	m.notifyModel = notify.New(true, "Open blocked", reason, notify.NoAction)
	return true
}

// Open directory with default editor
func (m *model) openDirectoryWithEditor() tea.Cmd {
	if variable.ChooserFile != "" {
		err := m.chooserFileWriteAndQuit(m.getFocusedFilePanel().Location)
		if err == nil {
			return nil
		}
		// Continue with preview if file is not writable
		slog.Error("Error while writing to chooser file, continuing with open via directory editor", "error", err)
	}

	editor := common.Config.DirEditor

	if editor == "" {
		editor = "nvim"
	}

	// Split the editor command into command and arguments
	parts := strings.Fields(editor)
	cmd := parts[0]
	//nolint:gocritic // appendAssign: intentionally creating a new slice
	args := append(parts[1:], m.getFocusedFilePanel().Location)

	c := exec.Command(cmd, args...)
	return tea.ExecProcess(c, func(err error) tea.Msg {
		return editorFinishedMsg{err}
	})
}

// Copy file path
// TODO: This is also an IO operations, do it via tea.Cmd
func (m *model) copyPath() {
	panel := m.getFocusedFilePanel()

	if panel.Empty() {
		return
	}

	if err := clipboard.WriteAll(panel.GetFocusedItem().Location); err != nil {
		slog.Error("Error while copy path", "error", err)
	}
}

// TODO: This is also an IO operations, do it via tea.Cmd
func (m *model) copyName() {
	panel := m.getFocusedFilePanel()

	if panel.Empty() {
		return
	}

	if err := clipboard.WriteAll(filepath.Base(panel.GetFocusedItem().Location)); err != nil {
		slog.Error("Error while copy file name", "error", err)
	}
}

// TODO: This is also an IO operations, do it via tea.Cmd
func (m *model) copyNameWithoutExtension() {
	panel := m.getFocusedFilePanel()

	if panel.Empty() {
		return
	}

	name := panel.GetFocusedItem().Name
	ext := filepath.Ext(name)
	nameWithoutExt := strings.TrimSuffix(name, ext)
	if err := clipboard.WriteAll(nameWithoutExt); err != nil {
		slog.Error("Error while copy file name without extension", "error", err)
	}
}

// TODO: This is also an IO operations, do it via tea.Cmd
func (m *model) copyPWD() {
	panel := m.getFocusedFilePanel()
	if err := clipboard.WriteAll(panel.Location); err != nil {
		slog.Error("Error while copy present working directory", "error", err)
	}
}
