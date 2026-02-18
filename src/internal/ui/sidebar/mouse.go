package sidebar

import "github.com/t4Linux/t4gfm/src/internal/common"

func (s *Model) ClickedDirectoryIndex(localY int, sidebarFocused bool) int {
	if localY < 0 || s.Disabled() {
		return -1
	}

	sections := s.sectionBuckets()
	listHeight, pinnedHeight, disksHeight := s.sectionHeights(sections, sidebarFocused)

	row := localY
	if row < listHeight {
		return s.clickedIndexInSection(row, listHeight, sections.list, true, sidebarFocused)
	}
	row -= listHeight
	if row < pinnedHeight {
		return s.clickedIndexInSection(row, pinnedHeight, sections.pinned, false, sidebarFocused)
	}
	row -= pinnedHeight
	if row < disksHeight {
		return s.clickedIndexInSection(row, disksHeight, sections.disks, false, sidebarFocused)
	}

	return -1
}

func (s *Model) clickedIndexInSection(localY int, sectionHeight int, indexes []int, includeSearchBar bool, sidebarFocused bool) int {
	if localY <= 0 {
		return -1
	}

	maxContentLines := sectionHeight - common.BorderPadding
	if maxContentLines <= 0 {
		return -1
	}

	usedLines := 0
	if includeSearchBar {
		usedLines += 2
		if s.searchBar.Focused() || s.searchBar.Value() != "" || sidebarFocused {
			usedLines++
		}
	}

	if len(indexes) == 0 || usedLines >= maxContentLines {
		return -1
	}

	start := 0
	if includeSearchBar {
		for start < len(indexes) && indexes[start] < s.renderIndex {
			start++
		}
	}

	visibleLimit := maxContentLines - usedLines
	if visibleLimit <= 0 {
		return -1
	}

	contentRow := localY - 1
	if contentRow < usedLines {
		return -1
	}

	offset := contentRow - usedLines
	if offset >= visibleLimit {
		return -1
	}

	idx := start + offset
	if idx < 0 || idx >= len(indexes) {
		return -1
	}

	return indexes[idx]
}
