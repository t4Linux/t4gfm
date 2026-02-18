package common

import (
	"os"
	"path/filepath"
	"strings"
)

const MaxSafeOpenFileSize = int64(300 * 1024 * 1024)

var BlockedOpenFileSuffixes = []string{
	".vmdk",
	".vdi",
	".img",
	".iso",
	".raw",
}

func ResolveEditorCommand() string {
	editor := Config.Editor
	if editor == "" {
		editor = os.Getenv("EDITOR")
	}
	if editor == "" {
		editor = "nvim"
	}
	return editor
}

func ResolveEnterOpenProgram(filePath string) string {
	if OpenBlockReason(filePath) != "" {
		return "blocked"
	}
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(filePath), "."))
	if extEditor, ok := Config.OpenWith[ext]; ok {
		return extEditor
	}
	return ResolveEditorCommand()
}

func OpenBlockReason(filePath string) string {
	resolvedPath := filePath
	if target, err := filepath.EvalSymlinks(filePath); err == nil {
		resolvedPath = target
	}

	lower := strings.ToLower(resolvedPath)
	baseLower := strings.ToLower(filepath.Base(resolvedPath))
	qcow2Idx := strings.LastIndex(baseLower, ".qcow2")
	if qcow2Idx != -1 {
		tail := baseLower[qcow2Idx+len(".qcow2"):]
		if tail == "" || strings.HasPrefix(tail, ".") {
			return "blocked file type: .qcow2*"
		}
	}
	for _, suffix := range BlockedOpenFileSuffixes {
		if strings.HasSuffix(lower, suffix) {
			return "blocked file type: " + suffix
		}
	}

	info, err := os.Stat(resolvedPath)
	if err != nil || info.IsDir() {
		return ""
	}

	if info.Size() > MaxSafeOpenFileSize {
		return "file is larger than 300M"
	}

	return ""
}
