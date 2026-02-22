package preview

import (
	"errors"
	"fmt"
	"image"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/charmbracelet/lipgloss"
	"github.com/yorukot/ansichroma"

	"github.com/t4Linux/t4gfm/src/internal/common"
	"github.com/t4Linux/t4gfm/src/internal/ui"
	"github.com/t4Linux/t4gfm/src/internal/ui/rendering"
	"github.com/t4Linux/t4gfm/src/internal/utils"
)

func applyPreviewPositionInfo(r *rendering.Renderer, pos int, total int) {
	if total <= 0 {
		return
	}
	if pos < 1 {
		pos = 1
	}
	if pos > total {
		pos = total
	}
	r.SetBorderInfoItems(fmt.Sprintf("preview %d/%d", pos, total))
}

func renderDirectoryPreview(r *rendering.Renderer, itemPath string, previewHeight int, offset int) (string, bool, int, int) {
	files, err := os.ReadDir(itemPath)
	if err != nil {
		slog.Error("Error render directory preview", "error", err)
		r.AddLines(common.FilePreviewDirectoryUnreadableText)
		return r.Render(), false, 0, 0
	}

	if len(files) == 0 {
		r.AddLines(common.FilePreviewEmptyText)
		return r.Render(), false, 0, 0
	}

	sort.Slice(files, func(i, j int) bool {
		if files[i].IsDir() && !files[j].IsDir() {
			return true
		}
		if !files[i].IsDir() && files[j].IsDir() {
			return false
		}
		return files[i].Name() < files[j].Name()
	})

	contentOffset := 0
	if previewHeight > 4 {
		contentOffset = 3
		r.AddLines("", "", "")
	}

	maxRows := previewHeight - contentOffset - 2
	if maxRows < 0 {
		maxRows = 0
	}
	if offset < 0 {
		offset = 0
	}
	start := min(offset, len(files))
	end := min(start+maxRows, len(files))
	current := min(start+1, len(files))
	applyPreviewPositionInfo(r, current, len(files))
	for i := start; i < end; i++ {
		file := files[i]
		isLink := false
		if info, err := file.Info(); err == nil {
			isLink = info.Mode()&os.ModeSymlink != 0
		}
		style := common.GetElementIcon(file.Name(), file.IsDir(), isLink, common.Config.Nerdfont)
		res := lipgloss.NewStyle().Foreground(lipgloss.Color(style.Color)).Background(common.FilePanelBGColor).
			Render(style.Icon+" ") + common.FilePanelStyle.Render(file.Name())
		r.AddLines(res)
	}
	return r.Render(), end < len(files), current, len(files)
}

func (m *Model) renderImagePreview(r *rendering.Renderer, itemPath string, previewWidth,
	previewHeight int, sideAreaWidth int, clearCmd string,
) string {
	if !m.open {
		return r.AddLines(common.FilePreviewPanelClosedText).Render() + clearCmd
	}

	if !common.Config.ShowImagePreview {
		return r.AddLines(common.FilePreviewImagePreviewDisabledText).Render() + clearCmd
	}

	// Use the new auto-detection function to choose the best renderer
	imageRender, err := m.imagePreviewer.ImagePreview(itemPath, previewWidth, previewHeight,
		common.Theme.FilePanelBG, sideAreaWidth)
	if errors.Is(err, image.ErrFormat) {
		return r.AddLines(common.FilePreviewUnsupportedImageFormatsText).Render() + clearCmd
	}

	if err != nil {
		slog.Error("Error convert image to ansi", "error", err)
		return r.AddLines(common.FilePreviewImageConversionErrorText).Render() + clearCmd
	}

	// Check if this looks like Kitty protocol output (starts with escape sequences)
	// For Kitty protocol, avoid using lipgloss alignment to prevent layout drift
	if strings.HasPrefix(imageRender, "\x1b_G") {
		r.AddLines(imageRender)
		return r.Render()
	}

	// For ANSI output, we can safely use vertical alignment
	return r.AddStyleModifier(func(s lipgloss.Style) lipgloss.Style {
		return s.AlignHorizontal(lipgloss.Center).AlignVertical(lipgloss.Center)
	}).AddLines(imageRender).Render() + clearCmd
}

func (m *Model) renderTextPreview(r *rendering.Renderer, itemPath string,
	previewWidth, previewHeight int,
) string {
	if isArchiveLikePath(itemPath) {
		return r.AddLines(common.FilePreviewBinaryArchiveDisabledText).Render()
	}

	format := lexers.Match(filepath.Base(itemPath))
	if format == nil {
		isText, err := common.IsTextFile(itemPath)
		if err != nil {
			slog.Error("Error while checking text file", "error", err)
			return r.AddLines(common.FilePreviewError).Render()
		}
		if !isText && !isTextConfigPath(itemPath) {
			return r.AddLines(common.FilePreviewUnsupportedFormatText).Render()
		}
	}

	fileContent, hasMore, err := utils.ReadFileContentWithOffset(itemPath, previewWidth, previewHeight, m.textScroll)
	if err != nil {
		slog.Error("Error open file", "error", err)
		m.hasMoreText = false
		m.previewPos = 0
		m.previewTotal = 0
		return r.AddLines(common.FilePreviewError).Render()
	}
	m.hasMoreText = hasMore
	totalLines, lineErr := utils.CountFileLines(itemPath)
	if lineErr != nil {
		totalLines = 0
	}
	current := 0
	if totalLines > 0 {
		current = min(m.textScroll+1, totalLines)
	}
	m.previewPos = current
	m.previewTotal = totalLines
	applyPreviewPositionInfo(r, current, totalLines)

	if fileContent == "" {
		if m.textScroll > 0 {
			m.textScroll--
			fileContent, hasMore, err = utils.ReadFileContentWithOffset(itemPath, previewWidth, previewHeight, m.textScroll)
			if err == nil && fileContent != "" {
				m.hasMoreText = hasMore
				if totalLines > 0 {
					current = min(m.textScroll+1, totalLines)
					m.previewPos = current
				}
			} else if err == nil {
				m.hasMoreText = false
			}
		}
		if fileContent != "" {
			r.AddLines(fileContent)
			return r.Render()
		}
		m.hasMoreText = false
		return r.AddLines(common.FilePreviewEmptyText).Render()
	}

	if format != nil {
		rawContent := fileContent
		background := ""
		if !common.Config.TransparentBackground {
			background = common.Theme.FilePanelBG
		}
		useBat := m.batCmd != "" && common.Config.CodePreviewer == "bat"
		if useBat {
			fileContent, err = getBatSyntaxHighlightedContent(itemPath, previewHeight, background, m.batCmd)
			if err != nil {
				fileContent, err = ansichroma.HightlightString(rawContent, format.Config().Name,
					common.Theme.CodeSyntaxHighlightTheme, background)
			}
		} else {
			fileContent, err = ansichroma.HightlightString(rawContent, format.Config().Name,
				common.Theme.CodeSyntaxHighlightTheme, background)
			if err != nil && m.batCmd != "" && common.Config.CodePreviewer == "" {
				fileContent, err = getBatSyntaxHighlightedContent(itemPath, previewHeight, background, m.batCmd)
			}
		}
		if err != nil {
			slog.Warn("Falling back to plain text preview due highlighting error", "error", err)
			fileContent = rawContent
		}
	}

	r.AddLines(fileContent)
	return r.Render()
}

func isTextConfigPath(itemPath string) bool {
	ext := strings.ToLower(filepath.Ext(itemPath))
	switch ext {
	case ".conf", ".cfg", ".ini", ".toml", ".yaml", ".yml", ".json", ".env", ".rc":
		return true
	default:
		return false
	}
}

func isArchiveLikePath(itemPath string) bool {
	lowerPath := strings.ToLower(itemPath)
	archiveLikeSuffixes := []string{
		".tar.gz", ".tar.bz2", ".tar.xz", ".tar.zst",
		".tgz", ".tbz", ".tbz2", ".txz", ".tzst",
		".deb", ".rpm", ".apk", ".pkg", ".msi", ".cab",
		".zst", ".xz", ".lz4", ".lzma", ".bz2",
	}
	for _, suffix := range archiveLikeSuffixes {
		if strings.HasSuffix(lowerPath, suffix) {
			return true
		}
	}
	if common.IsExtensionExtractable(filepath.Ext(lowerPath)) {
		return true
	}
	return false
}

// Only use this when height and width are synced with filemodel's expectations
func (m *Model) RenderText(text string) string {
	return m.RenderTextWithDimension(text, m.contentHeight, m.contentWidth)
}

func (m *Model) RenderTextWithDimension(text string, height int, width int) string {
	// For zero size, don't need to render anything. Its kinda hack, but
	// its to prevent error logs
	clearCmd := m.imagePreviewer.ClearKittyImages()
	if width == 0 && height == 0 {
		return clearCmd
	}
	return ui.FilePreviewPanelRenderer(height, width).
		AddLines(text).
		Render() + clearCmd
}

func (m *Model) RenderWithPath(itemPath string, previewWidth int, previewHeight int, fullModelWidth int) string {
	r := ui.FilePreviewPanelRenderer(previewHeight, previewWidth)
	clearCmd := m.imagePreviewer.ClearKittyImages()
	m.hasMoreText = false
	m.previewPos = 0
	m.previewTotal = 0

	// Adjust dimensions if border is enabled
	contentWidth := previewWidth
	contentHeight := previewHeight
	if common.Config.EnableFilePreviewBorder &&
		previewWidth >= rendering.MinWidthForBorder && previewHeight >= rendering.MinHeightForBorder {
		contentWidth = previewWidth - common.BorderPadding
		contentHeight = previewHeight - common.BorderPadding
	}

	lstatInfo, lstatErr := os.Lstat(itemPath)
	if lstatErr != nil {
		slog.Error("Error get file info", "error", lstatErr)
		return r.AddLines(common.FilePreviewNoFileInfoText).Render() + clearCmd
	}

	resolvedPath := itemPath
	if lstatInfo.Mode()&os.ModeSymlink != 0 {
		if targetPath, err := filepath.EvalSymlinks(itemPath); err == nil {
			resolvedPath = targetPath
		}
	}

	fileInfo, infoErr := os.Stat(resolvedPath)
	if infoErr != nil {
		slog.Error("Error get file info", "error", infoErr)
		return r.AddLines(common.FilePreviewNoFileInfoText).Render() + clearCmd
	}
	slog.Debug("Attempting to render preview", "itemPath", itemPath, "resolvedPath", resolvedPath,
		"mode", fileInfo.Mode().String(), "isRegular", fileInfo.Mode().IsRegular())

	// For non regular files which are not directories Dont try to read them
	// See Issue #876
	if !fileInfo.Mode().IsRegular() && (fileInfo.Mode()&fs.ModeDir) == 0 {
		return r.AddLines(common.FilePreviewUnsupportedFileMode).Render() + clearCmd
	}

	ext := filepath.Ext(resolvedPath)
	if slices.Contains(common.UnsupportedPreviewFormats, ext) {
		return r.AddLines(common.FilePreviewUnsupportedFormatText).Render() + clearCmd
	}

	if fileInfo.IsDir() {
		rendered, hasMore, current, total := renderDirectoryPreview(r, resolvedPath, contentHeight, m.textScroll)
		m.hasMoreText = hasMore
		m.previewPos = current
		m.previewTotal = total
		return rendered + clearCmd
	}

	if m.thumbnailGenerator != nil && m.thumbnailGenerator.SupportsExt(ext) {
		thumbnailPath, err := m.thumbnailGenerator.GetThumbnailOrGenerate(resolvedPath)
		if err != nil {
			slog.Error("Error generating thumbnail", "error", err)
			return r.AddLines(common.FilePreviewThumbnailGenerationErrorText).Render() + clearCmd
		}
		// Notes : If renderImagePreview fails, and return some error message
		// render, then we dont apply clearCmd. This might cause issues.
		// same for below usage of renderImagePreview
		return m.renderImagePreview(
			r, thumbnailPath, contentWidth, contentHeight,
			fullModelWidth-previewWidth, clearCmd)
	}

	if isImageFile(resolvedPath) {
		return m.renderImagePreview(
			r, resolvedPath, contentWidth, contentHeight,
			fullModelWidth-previewWidth, clearCmd)
	}

	return m.renderTextPreview(r, resolvedPath, contentWidth, contentHeight) + clearCmd
}
