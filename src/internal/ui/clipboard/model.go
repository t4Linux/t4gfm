package clipboard

import (
	"log/slog"
	"os"
	"slices"
	"strconv"

	"github.com/t4Linux/t4gfm/src/internal/common"
	"github.com/t4Linux/t4gfm/src/internal/ui"
)

// The fact that its visible in UI or not, is controlled by the main model
type Model struct {
	width  int
	height int
	items  copyItems
	meta   map[string]clipboardItemMeta
	dirty  bool
}

type clipboardItemMeta struct {
	isDir  bool
	isLink bool
	ok     bool
}

// Copied items
type copyItems struct {
	items []string
	cut   bool
}

func (m *Model) SetDimensions(width int, height int) {
	m.width = width
	m.height = height
}

func (m *Model) Render() string {
	r := ui.ClipboardRenderer(m.height, m.width)
	viewHeight := m.height - common.BorderPadding
	viewWidth := m.width - common.InnerPadding
	m.ensureItemMetaCache()
	if len(m.items.items) == 0 {
		// TODO move this to a string
		r.AddLines("", common.ClipboardNoneText)
	} else {
		for i := 0; i < len(m.items.items) && i < viewHeight; i++ {
			if i == viewHeight-1 && i != len(m.items.items)-1 {
				// Last Entry we can render, but there are more that one left
				r.AddLines(strconv.Itoa(len(m.items.items)-i) + " items left....")
			} else {
				meta, ok := m.meta[m.items.items[i]]
				if !ok || !meta.ok {
					continue
				}
				r.AddLines(common.ClipboardPrettierName(m.items.items[i],
					viewWidth, meta.isDir, meta.isLink, false))
			}
		}
	}
	return r.Render()
}

func (m *Model) ensureItemMetaCache() {
	if !m.dirty {
		return
	}
	if m.meta == nil {
		m.meta = make(map[string]clipboardItemMeta, len(m.items.items))
	} else {
		for k := range m.meta {
			delete(m.meta, k)
		}
	}
	for _, item := range m.items.items {
		fileInfo, err := os.Lstat(item)
		if err != nil {
			slog.Debug("Clipboard item metadata skipped", "path", item, "error", err)
			m.meta[item] = clipboardItemMeta{ok: false}
			continue
		}
		m.meta[item] = clipboardItemMeta{
			isDir:  fileInfo.IsDir(),
			isLink: fileInfo.Mode()&os.ModeSymlink != 0,
			ok:     true,
		}
	}
	m.dirty = false
}

func (m *Model) IsCut() bool {
	return m.items.cut
}

func (m *Model) Reset(cut bool) {
	m.items.cut = cut
	m.items.items = m.items.items[:0]
	m.dirty = true
}

func (m *Model) Add(location string) {
	m.items.items = append(m.items.items, location)
	m.dirty = true
}

func (m *Model) SetItems(items []string) {
	m.items.items = make([]string, len(items))
	copy(m.items.items, items)
	m.dirty = true
}

func (m *Model) pruneInaccessibleItems() {
	m.items.items = slices.DeleteFunc(m.items.items, func(item string) bool {
		_, err := os.Lstat(item)
		return err != nil
	})
	m.dirty = true
}

func (m *Model) GetItems() []string {
	// return a copy to prevent external mutation
	items := make([]string, len(m.items.items))
	copy(items, m.items.items)
	return items
}

// Use this to use a copy that is in sync with current state of filesystem
func (m *Model) PruneInaccessibleItemsAndGet() []string {
	// Clipboard items might becomes outdated with
	// externally/interally triggered changes
	m.pruneInaccessibleItems()
	return m.GetItems()
}

func (m *Model) Len() int {
	return len(m.items.items)
}

func (m *Model) GetWidth() int {
	return m.width
}

func (m *Model) GetHeight() int {
	return m.height
}

func (m *Model) GetFirstItem() string {
	if len(m.items.items) == 0 {
		return ""
	}
	return m.items.items[0]
}
