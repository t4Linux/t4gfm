package filepanel

func (m *Model) ClickedItemIndex(localY int) int {
	if m.Empty() {
		return -1
	}

	firstFileRow := 1 + (contentPadding - 1)
	if m.SearchBar.Focused() || m.SearchBar.Value() != "" {
		firstFileRow++
	}
	if m.NeedRenderHeaders() {
		firstFileRow += ColumnHeaderHeight
	}

	row := localY - firstFileRow
	if row < 0 {
		return -1
	}

	itemIndex := m.renderIndex + row
	if itemIndex < 0 || itemIndex >= m.ElemCount() {
		return -1
	}

	return itemIndex
}
