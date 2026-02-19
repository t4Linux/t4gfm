package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/rkoesters/xdg/trash"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	variable "github.com/t4Linux/t4gfm/src/config"
	"github.com/t4Linux/t4gfm/src/internal/common"
	"github.com/t4Linux/t4gfm/src/internal/ui/filepanel"
	"github.com/t4Linux/t4gfm/src/internal/ui/notify"
	"github.com/t4Linux/t4gfm/src/internal/utils"
)

// TODO : Add test for model initialized with multiple directories
// TODO : Add test for clipboard different variations, cut paste
// TODO : Add test for tea resizing
// TODO : Add test for quitting

func TestCopy(t *testing.T) {
	curTestDir := filepath.Join(testDir, "TestCopy")
	dir1 := filepath.Join(curTestDir, "dir1")
	dir2 := filepath.Join(curTestDir, "dir2")
	file1 := filepath.Join(dir1, "file1.txt")
	t.Run("Basic Copy", func(t *testing.T) {
		utils.SetupDirectories(t, curTestDir, dir1, dir2)
		utils.SetupFiles(t, file1)
		t.Cleanup(func() {
			os.RemoveAll(curTestDir)
		})

		p := NewTestTeaProgWithEventLoop(t, defaultTestModel(dir1))

		require.Equal(t, "file1.txt",
			p.getModel().getFocusedFilePanel().GetFocusedItem().Name)
		p.SendKeyDirectly(common.Hotkeys.CopyItems[0])
		if common.Hotkeys.CopyItems[0] == "y" {
			p.SendKeyDirectly("y")
		}
		assert.False(t, p.getModel().clipboard.IsCut())
		assert.Equal(t, file1, p.getModel().clipboard.GetFirstItem())

		p.getModel().updateCurrentFilePanelDir("../dir2")
		p.SendKey(common.Hotkeys.PasteItems[0])

		assert.Eventually(t, func() bool {
			_, err := os.Lstat(filepath.Join(dir2, "file1.txt"))
			return err == nil
		}, DefaultTestTimeout, DefaultTestTick)

		assert.False(t, p.getModel().clipboard.IsCut())
		assert.Equal(t, file1, p.getModel().clipboard.GetFirstItem())

		p.SendKey(common.Hotkeys.PasteItems[0])
		assert.Eventually(t, func() bool {
			_, err := os.Lstat(filepath.Join(dir2, "file1(1).txt"))
			return err == nil
		}, DefaultTestTimeout, DefaultTestTick)
		assert.FileExists(t, filepath.Join(dir2, "file1(1).txt"))
		//TODO: Also verify if there are only 2 items in process bar
	})
}

func TestTwoPanelTransferShortcuts(t *testing.T) {
	curTestDir := t.TempDir()
	srcDir := filepath.Join(curTestDir, "src")
	dstDir := filepath.Join(curTestDir, "dst")
	fileCopy := filepath.Join(srcDir, "copy.txt")
	fileMove := filepath.Join(srcDir, "move.txt")

	utils.SetupDirectories(t, srcDir, dstDir)
	utils.SetupFilesWithData(t, []byte("copy"), fileCopy, fileMove)

	m := defaultTestModel(srcDir)
	p := NewTestTeaProgWithEventLoop(t, m)

	_, err := m.fileModel.CreateNewFilePanel(dstDir)
	require.NoError(t, err)
	m.fileModel.PreviousFilePanel()
	TeaUpdate(m, nil)

	setFilePanelSelectedItemByLocation(t, m.getFocusedFilePanel(), fileCopy)
	p.SendKey("C")

	assert.Eventually(t, func() bool {
		_, err := os.Stat(filepath.Join(dstDir, "copy.txt"))
		return err == nil
	}, DefaultTestTimeout, DefaultTestTick, "File should be copied to second panel location")
	assert.Eventually(t, func() bool {
		return m.fileModel.FilePanels[1].FindElementIndexByLocation(filepath.Join(dstDir, "copy.txt")) != -1
	}, DefaultTestTimeout, DefaultTestTick, "Destination panel should refresh and show copied file")
	assert.FileExists(t, fileCopy)

	TeaUpdate(m, nil)
	setFilePanelSelectedItemByLocation(t, m.getFocusedFilePanel(), fileMove)
	p.SendKey("M")

	assert.Eventually(t, func() bool {
		_, err := os.Stat(filepath.Join(dstDir, "move.txt"))
		return err == nil
	}, DefaultTestTimeout, DefaultTestTick, "File should be moved to second panel location")
	assert.Eventually(t, func() bool {
		return m.fileModel.FilePanels[1].FindElementIndexByLocation(filepath.Join(dstDir, "move.txt")) != -1
	}, DefaultTestTimeout, DefaultTestTick, "Destination panel should refresh and show moved file")
	assert.Eventually(t, func() bool {
		_, err := os.Stat(fileMove)
		return os.IsNotExist(err)
	}, DefaultTestTimeout, DefaultTestTick, "Source file should be moved away")
}

func TestYankMenu(t *testing.T) {
	curTestDir := t.TempDir()
	file1 := filepath.Join(curTestDir, "file1.txt")
	utils.SetupFilesWithData(t, []byte("f1"), file1)

	m := defaultTestModel(curTestDir)

	TeaUpdate(m, nil)
	TeaUpdate(m, utils.TeaRuneKeyMsg("y"))
	assert.Equal(t, "y", m.rangerPrefix)

	TeaUpdate(m, utils.TeaRuneKeyMsg("y"))
	assert.Empty(t, m.rangerPrefix)
	assert.Equal(t, file1, m.clipboard.GetFirstItem())
	assert.False(t, m.clipboard.IsCut())

	TeaUpdate(m, utils.TeaRuneKeyMsg("y"))
	assert.Equal(t, "y", m.rangerPrefix)
	TeaUpdate(m, utils.TeaRuneKeyMsg("esc"))
	assert.Empty(t, m.rangerPrefix)
}

func TestCopySelectedItemsExitsSelectMode(t *testing.T) {
	curTestDir := t.TempDir()
	file1 := filepath.Join(curTestDir, "file1.txt")
	utils.SetupFilesWithData(t, []byte("f1"), file1)

	m := defaultTestModel(curTestDir)
	TeaUpdate(m, nil)

	panel := m.getFocusedFilePanel()
	panel.ChangeFilePanelMode()
	require.Equal(t, filepanel.SelectMode, panel.PanelMode)
	panel.SingleItemSelect()
	require.EqualValues(t, 1, panel.SelectedCount())

	TeaUpdate(m, utils.TeaRuneKeyMsg("y"))
	TeaUpdate(m, utils.TeaRuneKeyMsg("y"))

	assert.Equal(t, filepanel.BrowserMode, panel.PanelMode)
	assert.Equal(t, file1, m.clipboard.GetFirstItem())
	assert.False(t, m.clipboard.IsCut())
}

func TestRangerPrefixOperations(t *testing.T) {
	curTestDir := t.TempDir()
	dir2 := filepath.Join(curTestDir, "dir2")
	file1 := filepath.Join(curTestDir, "file1.txt")
	utils.SetupDirectories(t, dir2)
	utils.SetupFilesWithData(t, []byte("f1"), file1)

	m := defaultTestModel(curTestDir)
	TeaUpdate(m, nil)
	setFilePanelSelectedItemByLocation(t, m.getFocusedFilePanel(), file1)

	TeaUpdate(m, utils.TeaRuneKeyMsg("d"))
	assert.Equal(t, "d", m.rangerPrefix)
	TeaUpdate(m, utils.TeaRuneKeyMsg("d"))
	assert.True(t, m.clipboard.IsCut())
	assert.Equal(t, file1, m.clipboard.GetFirstItem())

	TeaUpdate(m, utils.TeaRuneKeyMsg("g"))
	TeaUpdate(m, utils.TeaRuneKeyMsg("h"))
	assert.Equal(t, variable.HomeDir, m.getFocusedFilePanel().Location)
}

func TestRangerTwoLetterMarks(t *testing.T) {
	curTestDir := t.TempDir()
	oldMarksFile := variable.RangerMarksFile
	variable.RangerMarksFile = filepath.Join(curTestDir, "ranger_marks_test.json")
	t.Cleanup(func() {
		variable.RangerMarksFile = oldMarksFile
	})

	dirA := filepath.Join(curTestDir, "dir-a")
	dirB := filepath.Join(curTestDir, "dir-b")
	utils.SetupDirectories(t, dirA, dirB)

	m := defaultTestModel(dirA)
	TeaUpdate(m, nil)

	TeaUpdate(m, utils.TeaRuneKeyMsg("m"))
	TeaUpdate(m, utils.TeaRuneKeyMsg("c"))
	TeaUpdate(m, utils.TeaRuneKeyMsg("a"))
	assert.Equal(t, dirA, m.rangerMarks["ca"])

	require.NoError(t, m.updateCurrentFilePanelDir(dirB))
	TeaUpdate(m, nil)

	TeaUpdate(m, utils.TeaRuneKeyMsg(";"))
	TeaUpdate(m, utils.TeaRuneKeyMsg("c"))
	TeaUpdate(m, utils.TeaRuneKeyMsg("a"))
	assert.Equal(t, dirA, m.getFocusedFilePanel().Location)

	TeaUpdate(m, utils.TeaRuneKeyMsg(";"))
	TeaUpdate(m, utils.TeaRuneKeyMsg("?"))
	assert.Equal(t, ";?", m.rangerPrefix)
	TeaUpdate(m, utils.TeaRuneKeyMsg("esc"))
	assert.Empty(t, m.rangerPrefix)

	TeaUpdate(m, utils.TeaRuneKeyMsg(";"))
	TeaUpdate(m, utils.TeaRuneKeyMsg("d"))
	TeaUpdate(m, utils.TeaRuneKeyMsg("c"))
	TeaUpdate(m, utils.TeaRuneKeyMsg("a"))
	_, ok := m.rangerMarks["ca"]
	assert.False(t, ok)

	TeaUpdate(m, utils.TeaRuneKeyMsg("m"))
	TeaUpdate(m, utils.TeaRuneKeyMsg("c"))
	TeaUpdate(m, utils.TeaRuneKeyMsg("a"))
	assert.Equal(t, dirA, m.rangerMarks["ca"])

	TeaUpdate(m, utils.TeaRuneKeyMsg(";"))
	TeaUpdate(m, utils.TeaRuneKeyMsg("D"))
	assert.Empty(t, m.rangerMarks)
}

func TestFileCreation(t *testing.T) {
	// TODO Also add directory creation test to this
	curTestDir := filepath.Join(testDir, "TestNaming")
	testParentDir := filepath.Join(curTestDir, "parentDir")
	testChildDir := filepath.Join(testParentDir, "childDir")

	utils.SetupDirectories(t, curTestDir, testParentDir, testChildDir)

	t.Cleanup(func() {
		os.RemoveAll(curTestDir)
	})

	testdata := []struct {
		name          string
		fileName      string
		expectedError bool
	}{
		{"valid name", "file.txt", false},
		{"invalid single dot", ".", true},
		{"invalid double dot", "..", true},
		{"invalid trailing slash-dot", fmt.Sprintf("test%c.", filepath.Separator), true},
		{"invalid trailing slash-dot-dot", fmt.Sprintf("test%c..", filepath.Separator), true},
		{"valid name with trailing .", "abc.", false},
	}

	for _, tt := range testdata {
		m := defaultTestModel(testChildDir)

		TeaUpdate(m, nil)
		TeaUpdate(m, utils.TeaRuneKeyMsg(common.Hotkeys.FilePanelItemCreate[0]))

		assert.Empty(t, m.typingModal.errorMesssage)

		m.typingModal.textInput.SetValue(tt.fileName)

		TeaUpdate(m, utils.TeaRuneKeyMsg(common.Hotkeys.ConfirmTyping[0]))

		if tt.expectedError {
			assert.NotEmpty(t, m.typingModal.errorMesssage, "expected an error for input: %q", tt.fileName)
		} else {
			assert.Empty(t, m.typingModal.errorMesssage, "expected an error for input: %q", tt.fileName)
			assert.FileExists(
				t,
				filepath.Join(testChildDir, tt.fileName),
				"expected file to be created: %q",
				tt.fileName,
			)
		}
	}
}

func TestFileRename(t *testing.T) {
	curTestDir, err := os.MkdirTemp(variable.HomeDir, "t4gfm-delete-test-")
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = os.RemoveAll(curTestDir)
	})
	file1 := filepath.Join(curTestDir, "file1.txt")
	file2 := filepath.Join(curTestDir, "file2.txt")
	file3 := filepath.Join(curTestDir, "file3.txt")

	utils.SetupFilesWithData(t, []byte("f1"), file1)
	utils.SetupFilesWithData(t, []byte("f2"), file2)
	utils.SetupFilesWithData(t, []byte("f3"), file3)

	file1New := filepath.Join(curTestDir, "file1_new.txt")

	t.Run("Basic rename", func(t *testing.T) {
		m := defaultTestModel(curTestDir)
		p := NewTestTeaProgWithEventLoop(t, m)
		setFilePanelSelectedItemByLocation(t, m.getFocusedFilePanel(), file1)

		p.SendKey(common.Hotkeys.FilePanelItemRename[0])
		p.SendKey("_new")
		p.Send(tea.KeyMsg{Type: tea.KeyEnter})

		assert.Eventually(t, func() bool {
			_, err1 := os.Stat(file1)
			_, err1New := os.Stat(file1New)
			return err1New == nil && os.IsNotExist(err1)
		}, DefaultTestTimeout, DefaultTestTick, "File never got renamed")
	})

	t.Run("Rename confirmation for same name", func(t *testing.T) {
		actualTest := func(doRename bool) {
			m := defaultTestModel(curTestDir)
			p := NewTestTeaProgWithEventLoop(t, m)
			setFilePanelSelectedItemByLocation(t, m.getFocusedFilePanel(), file3)

			p.SendKey(common.Hotkeys.FilePanelItemRename[0])
			p.Send(tea.KeyMsg{Type: tea.KeyBackspace})
			p.SendKey("2")
			p.Send(tea.KeyMsg{Type: tea.KeyEnter})

			require.Eventually(t, func() bool {
				return m.notifyModel.IsOpen()
			}, DefaultTestTimeout, DefaultTestTick,
				"Notify modal never opened, renaming text : %v", m.getFocusedFilePanel().Rename.Value())

			assert.Equal(t, notify.New(true,
				common.SameRenameWarnTitle,
				common.SameRenameWarnContent,
				notify.RenameAction), m.notifyModel, "Notify model should be as expected")

			if doRename {
				p.Send(tea.KeyMsg{Type: tea.KeyEnter})
			} else {
				p.SendKey(common.Hotkeys.CancelTyping[0])
			}

			assert.Eventually(t, func() bool {
				_, err2 := os.Stat(file2)
				_, err3 := os.Stat(file3)
				f2Data, err := os.ReadFile(file2)
				require.NoError(t, err)
				if doRename {
					// f3 should be gone. f2 should have content of f3
					return os.IsNotExist(err3) && err2 == nil &&
						string(f2Data) == "f3"
				}
				return err2 == nil && err3 == nil
			}, DefaultTestTimeout, DefaultTestTick,
				"Rename could not be done/not done appropriately")
		}

		actualTest(false)
		actualTest(true)
	})
}

func isTrashed(fileAbsPath string) bool {
	fileName := filepath.Base(fileAbsPath)
	switch runtime.GOOS {
	case utils.OsDarwin:
		_, err := os.Stat(filepath.Join(variable.DarwinTrashDirectory, fileName))
		return err == nil
	case utils.OsLinux:
		_, err := trash.Stat(fileAbsPath)
		return err == nil
	default:
		return false
	}
}

func TestFileDelete(t *testing.T) {
	if runtime.GOOS != utils.OsLinux && runtime.GOOS != utils.OsDarwin {
		t.Skip("Skipping unsupported platform")
	}
	curTestDir := t.TempDir()
	file1 := filepath.Join(curTestDir, "file1.txt")
	file2 := filepath.Join(curTestDir, "file2.txt")

	utils.SetupFilesWithData(t, []byte("f1"), file1)
	utils.SetupFilesWithData(t, []byte("f2"), file2)

	testdata := []struct {
		name            string
		filePath        string
		permanentDelete bool
	}{
		{
			name:            "Move to trash",
			filePath:        file1,
			permanentDelete: false,
		},
		{
			name:            "Permanently delete",
			filePath:        file2,
			permanentDelete: true,
		},
	}

	for _, tt := range testdata {
		t.Run(tt.name, func(t *testing.T) {
			m := defaultTestModel(curTestDir)
			m.hasTrash = common.InitTrash()
			p := NewTestTeaProgWithEventLoop(t, m)
			setFilePanelSelectedItemByLocation(t, m.getFocusedFilePanel(), tt.filePath)
			if tt.permanentDelete {
				p.SendKey(common.Hotkeys.PermanentlyDeleteItems[0])
			} else {
				p.SendKey(common.Hotkeys.DeleteItems[0])
			}
			assert.Eventually(t, m.notifyModel.IsOpen, DefaultTestTimeout,
				DefaultTestTick, "Notify model never opened")
			expectedTitle := common.TrashWarnTitle
			expectedAction := notify.DeleteAction
			if tt.permanentDelete {
				expectedTitle = common.PermanentDeleteWarnTitle
				expectedAction = notify.PermanentDeleteAction
			}
			assert.Equal(t, expectedTitle, m.notifyModel.GetTitle())
			assert.Equal(t, expectedAction, m.notifyModel.GetConfirmAction())

			p.Send(tea.KeyMsg{Type: tea.KeyEnter})

			assert.Eventually(t, func() bool {
				_, err := os.Stat(tt.filePath)
				return err != nil && os.IsNotExist(err)
			}, DefaultTestTimeout, DefaultTestTick, "File never removed from original location")

			if runtime.GOOS == utils.OsDarwin || runtime.GOOS == utils.OsLinux {
				assert.Equal(t, tt.permanentDelete, !isTrashed(filepath.Base(tt.filePath)),
					"Existence in trash status should be expected only of not permanently deleted")
			}
		})
	}
}

func TestDeleteFromSelectModeExitsToBrowserMode(t *testing.T) {
	curTestDir := t.TempDir()
	filePath := filepath.Join(curTestDir, "file1.txt")
	utils.SetupFilesWithData(t, []byte("f1"), filePath)

	m := defaultTestModel(curTestDir)
	p := NewTestTeaProgWithEventLoop(t, m)
	setFilePanelSelectedItemByLocation(t, m.getFocusedFilePanel(), filePath)

	panel := m.getFocusedFilePanel()
	panel.ChangeFilePanelMode()
	require.Equal(t, filepanel.SelectMode, panel.PanelMode)
	panel.SetSelectedAll([]string{filePath})
	require.EqualValues(t, 1, panel.SelectedCount())

	p.SendKey(common.Hotkeys.PermanentlyDeleteItems[0])
	assert.Eventually(t, m.notifyModel.IsOpen, DefaultTestTimeout,
		DefaultTestTick, "Notify model never opened")
	p.Send(tea.KeyMsg{Type: tea.KeyEnter})

	assert.Eventually(t, func() bool {
		_, err := os.Stat(filePath)
		return err != nil && os.IsNotExist(err)
	}, DefaultTestTimeout, DefaultTestTick, "File never removed from original location")

	assert.Eventually(t, func() bool {
		return panel.PanelMode == filepanel.BrowserMode
	}, DefaultTestTimeout, DefaultTestTick, "Panel mode did not switch back to browser mode")
}
