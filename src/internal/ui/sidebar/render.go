package sidebar

import (
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/t4Linux/t4gfm/src/internal/common"
	"github.com/t4Linux/t4gfm/src/internal/ui"
	"github.com/t4Linux/t4gfm/src/internal/ui/rendering"
)

type sectionKind int

const (
	listSection sectionKind = iota
	pinnedSection
	disksSection
)

type sidebarSections struct {
	list      []int
	pinned    []int
	disks     []int
	indexKind map[int]sectionKind
}

func (s *Model) Render(sidebarFocused bool, currentFilePanelLocation string) string {
	if s.Disabled() {
		return ""
	}

	sections := s.sectionBuckets()
	listFocused, pinnedFocused, disksFocused := s.sectionFocus(sidebarFocused, sections)
	listHeight, pinnedHeight, disksHeight := s.sectionHeights(sections, sidebarFocused)
	listHeaderHeight, filesHeight := splitListPanelHeights(listHeight)

	listHeaderPanel := s.renderProgramBarPanel(listHeaderHeight, false)
	listPanel := s.renderSectionPanel("Files", filesHeight, listFocused, sections.list,
		currentFilePanelLocation, true, sidebarFocused)
	pinnedPanel := s.renderSectionPanel("Pinned", pinnedHeight, pinnedFocused, sections.pinned,
		currentFilePanelLocation, false, sidebarFocused)
	disksPanel := s.renderSectionPanel("Disks", disksHeight, disksFocused, sections.disks,
		currentFilePanelLocation, false, sidebarFocused)

	return lipgloss.JoinVertical(0, listHeaderPanel, listPanel, pinnedPanel, disksPanel)
}

func splitListPanelHeights(listHeight int) (int, int) {
	const headerHeight = 3
	const minFilesHeight = 3

	if listHeight >= headerHeight+minFilesHeight {
		return headerHeight, listHeight - headerHeight
	}
	return 0, listHeight
}

func (s *Model) renderProgramBarPanel(height int, focused bool) string {
	if height <= 0 {
		return ""
	}
	r := ui.SidebarSectionRenderer(height, s.width, focused, "")
	contentWidth := max(0, s.width-common.BorderPadding)
	programName := "t4gfm"
	timestamp := time.Now().Format("15:04 02.01.06")
	line := programName + " " + timestamp
	padding := contentWidth - lipgloss.Width(programName) - lipgloss.Width(timestamp)
	if padding > 0 {
		line = programName + strings.Repeat(" ", padding) + timestamp
	}
	r.AddLines(common.FilePanelTopPathStyle.Render(line))
	return r.Render()
}

func (s *Model) sectionBuckets() sidebarSections {
	res := sidebarSections{indexKind: make(map[int]sectionKind)}

	cur := listSection
	for i := range s.directories {
		switch s.directories[i] {
		case homeDividerDir:
			cur = listSection
			continue
		case pinnedDividerDir:
			cur = pinnedSection
			continue
		case diskDividerDir:
			cur = disksSection
			continue
		}

		switch cur {
		case pinnedSection:
			res.pinned = append(res.pinned, i)
		case disksSection:
			res.disks = append(res.disks, i)
		default:
			res.list = append(res.list, i)
		}
		res.indexKind[i] = cur
	}

	return res
}

func (s *Model) sectionFocus(sidebarFocused bool, sections sidebarSections) (bool, bool, bool) {
	if !sidebarFocused {
		return false, false, false
	}
	if s.searchBar.Focused() {
		return true, false, false
	}
	kind, ok := sections.indexKind[s.cursor]
	if !ok {
		return true, false, false
	}
	return kind == listSection, kind == pinnedSection, kind == disksSection
}

func (s *Model) sectionHeights(sections sidebarSections, sidebarFocused bool) (int, int, int) {
	if s.height < 9 {
		return s.height, 0, 0
	}

	content := s.height - 6
	listMin := 3
	if sidebarFocused || s.searchBar.Focused() || s.searchBar.Value() != "" {
		listMin = 4
	}
	pinnedMin := 1
	disksMin := 1

	if content <= listMin+pinnedMin+disksMin {
		listContent := content - pinnedMin - disksMin
		if listContent < 1 {
			listContent = 1
		}
		return listContent + 2, pinnedMin + 2, disksMin + 2
	}

	pinnedNeed := max(1, len(sections.pinned))
	disksNeed := max(1, len(sections.disks))
	pinnedContent := pinnedMin
	disksContent := disksMin

	extraCapacity := content - listMin - pinnedMin - disksMin
	pinnedExtraNeed := max(0, pinnedNeed-pinnedMin)
	disksExtraNeed := max(0, disksNeed-disksMin)

	allocatedPinned := min(extraCapacity, pinnedExtraNeed)
	pinnedContent += allocatedPinned
	extraCapacity -= allocatedPinned

	allocatedDisks := min(extraCapacity, disksExtraNeed)
	disksContent += allocatedDisks
	extraCapacity -= allocatedDisks

	listContent := content - pinnedContent - disksContent
	if listContent < listMin {
		deficit := listMin - listContent
		reducePinned := min(deficit, max(0, pinnedContent-pinnedMin))
		pinnedContent -= reducePinned
		deficit -= reducePinned
		reduceDisks := min(deficit, max(0, disksContent-disksMin))
		disksContent -= reduceDisks
		listContent = content - pinnedContent - disksContent
	}

	return listContent + 2, pinnedContent + 2, disksContent + 2
}

func (s *Model) renderSectionPanel(title string, height int, focused bool, indexes []int,
	curFilePanelFileLocation string, includeSearchBar bool, sidebarFocused bool) string {
	if height <= 0 {
		return ""
	}

	r := ui.SidebarSectionRenderer(height, s.width, focused, title)

	maxContentLines := height - common.BorderPadding
	if maxContentLines <= 0 {
		return r.Render()
	}

	usedLines := 0
	if usedLines >= maxContentLines {
		return r.Render()
	}

	visibleLimit := maxContentLines - usedLines
	if len(indexes) == 0 {
		r.AddLines(common.SideBarNoneText)
		return r.Render()
	}

	start := 0
	if includeSearchBar {
		for start < len(indexes) && indexes[start] < s.renderIndex {
			start++
		}
	}
	for i := start; i < len(indexes) && visibleLimit > 0; i++ {
		s.renderDirectoryItem(indexes[i], curFilePanelFileLocation, sidebarFocused, r)
		visibleLimit--
	}

	return r.Render()
}

func (s *Model) renderDirectoryItem(index int, curFilePanelFileLocation string,
	sideBarFocused bool, r *rendering.Renderer) {
	if index < 0 || index >= len(s.directories) {
		return
	}

	cursor := " "
	isCursorRow := s.cursor == index && sideBarFocused && !s.searchBar.Focused()
	isCurrentLocation := s.directories[index].Location == curFilePanelFileLocation
	if s.renaming && s.cursor == index {
		r.AddLines(s.rename.View())
		return
	}

	renderStyle := common.SidebarStyle
	if isCurrentLocation || isCursorRow {
		renderStyle = common.SidebarSelectedStyle
	}
	line := common.FilePanelCursorStyle.Render(cursor+" ") + renderStyle.Render(s.directories[index].Name)
	r.AddLineWithCustomTruncate(line, rendering.TailsTruncateRight)
}
