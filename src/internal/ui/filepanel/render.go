package filepanel

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/t4Linux/t4gfm/src/config/icon"
	"github.com/t4Linux/t4gfm/src/internal/common"
	"github.com/t4Linux/t4gfm/src/internal/ui"
	"github.com/t4Linux/t4gfm/src/internal/ui/rendering"
	"github.com/t4Linux/t4gfm/src/internal/ui/sortmodel"
)

/*
- TODO: Write File Panel Specific unit test
  - Individual panel resizes
  - Footer content of filepanel changes due to resizing
  - i Only mode icons remains on smaller
  - ii Other things that change too
  - Other panels like clipboard and metadata's content changes too on resize
*/
func (m *Model) Render(focused bool) string {
	r := ui.FilePanelRenderer(m.height, m.width, focused)

	m.renderTopBar(r)
	m.renderSearchBar(r)
	m.renderFooter(r, m.SelectedCount())
	if m.NeedRenderHeaders() {
		m.renderColumnHeaders(r)
	}
	m.renderFileEntries(r)
	return r.Render()
}

func (m *Model) renderTopBar(r *rendering.Renderer) {
	const maxPathSegments = 6
	displayPath := compactTopBarPath(m.Location, maxPathSegments)
	// TODO - Add ansitruncate left in renderer and remove truncation here
	truncatedPath := common.TruncateTextBeginning(displayPath, m.GetContentWidth()-common.InnerPadding, "...")
	r.AddLines(common.FilePanelTopDirectoryIcon + common.FilePanelTopPathStyle.Render(truncatedPath))
	r.AddSection()
}

func compactTopBarPath(path string, maxSegments int) string {
	cleanPath := filepath.Clean(path)
	if cleanPath == string(filepath.Separator) || cleanPath == "." || maxSegments <= 0 {
		return cleanPath
	}

	segments := strings.Split(cleanPath, string(filepath.Separator))
	filtered := make([]string, 0, len(segments))
	for _, segment := range segments {
		if segment == "" || segment == "." {
			continue
		}
		filtered = append(filtered, segment)
	}

	if len(filtered) <= maxSegments {
		if filepath.IsAbs(cleanPath) {
			return string(filepath.Separator) + strings.Join(filtered, string(filepath.Separator))
		}
		return strings.Join(filtered, string(filepath.Separator))
	}

	lastSegments := filtered[len(filtered)-maxSegments:]
	return "..." + string(filepath.Separator) + strings.Join(lastSegments, string(filepath.Separator))
}

func (m *Model) renderSearchBar(r *rendering.Renderer) {
	if m.SearchBar.Focused() || m.SearchBar.Value() != "" {
		r.AddLines(" " + m.SearchBar.View())
	}
}

// TODO : Unit test this
func (m *Model) renderFooter(r *rendering.Renderer, selectedCount uint) {
	sortLabel, sortIcon := m.getSortInfo()
	modeLabel, modeIcon := m.getPanelModeInfo(selectedCount)
	cursorStr := m.getCursorString()

	if common.Config.Nerdfont {
		sortLabel = sortIcon + icon.Space + sortLabel
		modeLabel = modeIcon + icon.Space + modeLabel
	} else {
		// TODO : Figure out if we can set icon.Space to " " if nerdfont is false
		// That would simplify code
		sortLabel = sortIcon + " " + sortLabel
	}

	if common.Config.ShowPanelFooterInfo {
		r.SetBorderInfoItems(sortLabel, modeLabel, cursorStr)
		if r.AreInfoItemsTruncated() {
			r.SetBorderInfoItems(sortIcon, modeIcon, cursorStr)
		}
	} else {
		r.SetBorderInfoItems(cursorStr)
	}
}

func (m *Model) renderColumnHeaders(r *rendering.Renderer) {
	var builder strings.Builder
	for _, column := range m.columns {
		builder.WriteString(column.RenderHeader())
	}
	r.AddLines(builder.String())
}

func (m *Model) renderFileEntries(r *rendering.Renderer) {
	if m.Empty() {
		r.AddLines(common.FilePanelNoneText)
		return
	}
	end := min(m.renderIndex+m.PanelElementHeight(), m.ElemCount())

	for itemIndex := m.renderIndex; itemIndex < end; itemIndex++ {
		if m.Renaming && itemIndex == m.GetCursor() {
			r.AddLines(m.Rename.View())
			continue
		}
		var builder strings.Builder
		for _, column := range m.columns {
			colData := column.Render(itemIndex)
			builder.WriteString(colData)
		}
		r.AddLines(builder.String())
	}
}

func (m *Model) getSortInfo() (string, string) {
	iconStr := icon.SortAsc
	if m.SortReversed {
		iconStr = icon.SortDesc
	}
	return sortmodel.SortOptionsShortStr[m.SortKind], iconStr
}

func (m *Model) getPanelModeInfo(selectedCount uint) (string, string) {
	switch m.PanelMode {
	case BrowserMode:
		return "Browser", icon.Browser
	case SelectMode:
		return "Select" + icon.Space + fmt.Sprintf("(%d)", selectedCount), icon.Select
	default:
		return "", ""
	}
}

func (m *Model) getCursorString() string {
	cursor := m.GetCursor()
	if !m.Empty() {
		cursor++ // Convert to 1-based
	}
	return fmt.Sprintf("%d/%d", cursor, m.ElemCount())
}

func (m *Model) renderSelectBox(isSelected bool) string {
	if !common.Config.ShowSelectIcons || !common.Config.Nerdfont || m.PanelMode != SelectMode {
		return ""
	}

	if m.IsFocused {
		if isSelected {
			return common.CheckboxCheckedFocused
		}
		return common.CheckboxEmptyFocused
	}
	if isSelected {
		return common.CheckboxChecked
	}
	return common.CheckboxEmpty
}

// Checks whether a panel needs re-render due to being invalid or due to directory change
func (m *Model) NeedsReRender() bool {
	if !m.EmptyOrInvalid() {
		return filepath.Dir(m.GetFirstElement().Location) != m.Location
	}
	return true
}
