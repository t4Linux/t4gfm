package internal

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/rkoesters/xdg/trash"
	"github.com/t4Linux/t4gfm/src/internal/ui/processbar"

	variable "github.com/t4Linux/t4gfm/src/config"
)

var errOperationCancelled = errors.New("operation cancelled")

// isSamePartition checks if two paths are on the same filesystem partition
func isSamePartition(path1, path2 string) (bool, error) {
	// Get the absolute path to handle relative paths
	absPath1, err := filepath.Abs(path1)
	if err != nil {
		return false, fmt.Errorf("failed to get absolute path of the first path: %w", err)
	}

	absPath2, err := filepath.Abs(path2)
	if err != nil {
		return false, fmt.Errorf("failed to get absolute path of the second path: %w", err)
	}

	return filepath.VolumeName(absPath1) == filepath.VolumeName(absPath2), nil
}

// moveElement moves a file or directory efficiently
func moveElement(src, dst string) error {
	// Check if source and destination are on the same partition
	sameDev, err := isSamePartition(src, dst)
	if err != nil {
		return fmt.Errorf("failed to check partitions: %w", err)
	}

	// If on the same partition, attempt to rename (which will use the same inode)
	if sameDev {
		if err = os.Rename(src, dst); err == nil {
			return nil
		}
		// If rename fails, fall back to copy+delete
	}

	// If on different partitions or rename failed, fall back to copy+delete
	err = copyElement(src, dst)
	if err != nil {
		return fmt.Errorf("failed to copy: %w", err)
	}

	err = os.RemoveAll(src)
	if err != nil {
		return fmt.Errorf("failed to remove source after copy: %w", err)
	}

	return nil
}

// copyElement handles copying of both files and directories
func copyElement(src, dst string) error {
	srcInfo, err := os.Lstat(src)
	if err != nil {
		return fmt.Errorf("failed to stat source: %w", err)
	}
	if srcInfo.Mode()&os.ModeSymlink != 0 {
		return copySymlink(src, dst)
	}

	if srcInfo.IsDir() {
		return copyDir(src, dst, srcInfo)
	}
	return copyFile(src, dst, srcInfo)
}

// copyDir recursively copies a directory
func copyDir(src, dst string, srcInfo os.FileInfo) error {
	err := os.MkdirAll(dst, srcInfo.Mode())
	if err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("failed to read source directory: %w", err)
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		entryInfo, err := os.Lstat(srcPath)
		if err != nil {
			return fmt.Errorf("failed to get entry info: %w", err)
		}
		if entryInfo.Mode()&os.ModeSymlink != 0 {
			err = copySymlink(srcPath, dstPath)
			if err != nil {
				return err
			}
			continue
		}

		if entryInfo.IsDir() {
			err = copyDir(srcPath, dstPath, entryInfo)
		} else {
			err = copyFile(srcPath, dstPath, entryInfo)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// copyFile copies a single file
func copyFile(src, dst string, srcInfo os.FileInfo) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy file contents: %w", err)
	}
	return nil
}

func copySymlink(src, dst string) error {
	target, err := os.Readlink(src)
	if err != nil {
		return fmt.Errorf("failed to read symlink target: %w", err)
	}
	if _, err = os.Lstat(dst); err == nil {
		if err = os.RemoveAll(dst); err != nil {
			return fmt.Errorf("failed to replace existing destination while copying symlink: %w", err)
		}
	}
	if err = os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return fmt.Errorf("failed to create destination directory for symlink: %w", err)
	}
	if err = os.Symlink(target, dst); err != nil {
		return fmt.Errorf("failed to create symlink: %w", err)
	}
	return nil
}

func copyFileWithProgress(src, dst string, srcInfo os.FileInfo,
	onProgress func(written int64), shouldCancel func() bool,
) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dstFile.Close()

	buf := make([]byte, 256*1024)
	for {
		if shouldCancel != nil && shouldCancel() {
			return errOperationCancelled
		}
		n, readErr := srcFile.Read(buf)
		if n > 0 {
			written, writeErr := dstFile.Write(buf[:n])
			if writeErr != nil {
				return fmt.Errorf("failed to write destination file: %w", writeErr)
			}
			if written != n {
				return fmt.Errorf("short write while copying file")
			}
			if onProgress != nil {
				onProgress(int64(written))
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return fmt.Errorf("failed to copy file contents: %w", readErr)
		}
	}

	return nil
}

func progressUnitsForFile(info os.FileInfo) int {
	if info == nil {
		return 1
	}
	if info.Size() <= 0 {
		return 1
	}
	maxInt := int64(^uint(0) >> 1)
	if info.Size() > maxInt {
		return int(maxInt)
	}
	return int(info.Size())
}

func moveToTrash(src string) error {
	var err error
	switch runtime.GOOS {
	case "darwin":
		err = moveElement(src, filepath.Join(variable.DarwinTrashDirectory, filepath.Base(src)))
	default:
		// TODO: We should consider moving away from this package. Its not well written.
		// It uses package globals, It doesn't initializes trash directory, and we have to do it
		// separately outside of the this package. There is not documentation about this
		// It also uses deprecated libraries, and isn't well maintained.
		err = trash.Trash(src)
	}
	if err != nil {
		slog.Error("Error while deleting single item, in function to move file to trash can", "error", err)
	}
	return err
}

// pasteDir handles directory copying with progress tracking
func pasteDir(src, dst string, p *processbar.Process, cut bool,
	processBarModel *processbar.Model, shouldCancel func() bool,
) error {
	dst, err := renameIfDuplicate(dst)
	if err != nil {
		return err
	}

	// Check if we can do a fast move within the same partition
	sameDev, err := isSamePartition(src, dst)
	if err == nil && sameDev && cut {
		// For cut operations on same partition, try fast rename first
		err = os.Rename(src, dst)
		if err == nil {
			return nil
		}
		// If rename fails, fall back to manual copy
	}

	err = filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if shouldCancel != nil && shouldCancel() {
			return errOperationCancelled
		}
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		newPath := filepath.Join(dst, relPath)
		return actualPasteOperation(info, path, newPath, cut, sameDev, p, processBarModel, shouldCancel)
	})

	if err != nil {
		return err
	}

	// If this was a cut operation and we had to do a manual copy, remove the source
	if cut && !sameDev {
		err = os.RemoveAll(src)
		if err != nil {
			return fmt.Errorf("failed to remove source after move: %w", err)
		}
	}

	return nil
}

func actualPasteOperation(info os.FileInfo, path string, newPath string, cut bool, sameDev bool,
	p *processbar.Process, processBarModel *processbar.Model, shouldCancel func() bool,
) error {
	var err error
	if info.Mode()&os.ModeSymlink != 0 {
		p.CurrentFile = filepath.Base(path)
		if cut && sameDev {
			err = os.Rename(path, newPath)
		} else {
			err = copySymlink(path, newPath)
			if err == nil && cut {
				err = os.Remove(path)
			}
		}
		if err != nil {
			p.State = processbar.Failed
			pSendErr := processBarModel.SendUpdateProcessMsg(*p, true)
			if pSendErr != nil {
				slog.Error("Error sending process update", "error", pSendErr)
			}
			return err
		}
		p.Done += progressUnitsForFile(info)
		if p.Done > p.Total {
			p.Done = p.Total
		}
		processBarModel.TrySendingUpdateProcessMsg(*p)
		return nil
	}

	if info.IsDir() {
		// TODO - this is likely not needed because we did
		// dst, err := renameIfDuplicate(dst) above
		newPath, err = renameIfDuplicate(newPath)
		if err != nil {
			return err
		}
		err = os.MkdirAll(newPath, info.Mode())
		return err
	}

	// File
	p.CurrentFile = filepath.Base(path)
	if cut && sameDev {
		err = os.Rename(path, newPath)
		if err == nil {
			p.Done += progressUnitsForFile(info)
		}
	} else {
		err = copyFileWithProgress(path, newPath, info, func(written int64) {
			p.Done += int(written)
			if p.Done > p.Total {
				p.Done = p.Total
			}
			processBarModel.TrySendingUpdateProcessMsg(*p)
		}, shouldCancel)
	}

	if err != nil {
		p.State = processbar.Failed
		pSendErr := processBarModel.SendUpdateProcessMsg(*p, true)
		if pSendErr != nil {
			slog.Error("Error sending process update", "error", pSendErr)
		}
		return err
	}

	if p.Done > p.Total {
		p.Done = p.Total
	}
	processBarModel.TrySendingUpdateProcessMsg(*p)
	return nil
}

// isAncestor checks if dst is the same as src or a subdirectory of src.
func isAncestor(src, dst string) bool {
	// Resolve symlinks for both paths
	srcResolved, err := filepath.EvalSymlinks(src)
	if err != nil {
		// If we can't resolve symlinks, fall back to original path
		srcResolved = src
	}

	dstResolved, err := filepath.EvalSymlinks(dst)
	if err != nil {
		// If we can't resolve symlinks, fall back to original path
		dstResolved = dst
	}

	// Get absolute paths. Abs() also Cleans paths to normalize separators and resolve . and ..
	srcAbs, err := filepath.Abs(srcResolved)
	if err != nil {
		return false
	}

	dstAbs, err := filepath.Abs(dstResolved)
	if err != nil {
		return false
	}

	// Check if dst is the same as src
	if srcAbs == dstAbs {
		return true
	}

	// Check if dst is a subdirectory of src
	// Use filepath.Rel to check the relationship
	rel, err := filepath.Rel(srcAbs, dstAbs)
	if err != nil {
		return false
	}

	// If rel is "." or doesn't start with "..", then dst is inside src
	return rel == "." || !strings.HasPrefix(rel, "..")
}
