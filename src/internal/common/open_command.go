package common

import (
	"os"
	"path/filepath"
	"strings"
)

const MaxSafeOpenFileSize = int64(300 * 1024 * 1024)

var BlockedOpenFileSuffixes = []string{
	".qcow2",
	".qcow2.bak",
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

	isText, textErr := IsTextFile(resolvedPath)
	if textErr == nil && !isText {
		return "binary file detected"
	}

	return ""
}
