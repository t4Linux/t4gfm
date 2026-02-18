package gitpanel

import (
	"os"
	"strconv"
	"strings"

	"github.com/t4Linux/t4gfm/src/internal/common"
	"github.com/t4Linux/t4gfm/src/internal/ui"
	"github.com/t4Linux/t4gfm/src/internal/ui/rendering"
)

const (
	gitTab = iota
	clipboardTab
)

type Model struct {
	width   int
	height  int
	path    string
	branch  string
	subject string
	date    string
	author  string
	status  string
	noRepo  bool
	tab     int
	cbItems []string
	cbCut   bool
	cbIndex int
}

func New() Model {
	return Model{noRepo: true, tab: gitTab}
}

func (m *Model) SetDimensions(width int, height int) {
	m.width = width
	m.height = height
}

func (m *Model) SetNoRepo(path string) {
	m.path = path
	m.noRepo = true
	m.branch = ""
	m.subject = ""
	m.date = ""
	m.author = ""
	m.status = ""
}

func (m *Model) SetData(path string, branch string, subject string, date string, author string, status string) {
	m.path = path
	m.noRepo = false
	m.branch = branch
	m.subject = subject
	m.date = date
	m.author = author
	m.status = status
}

func (m *Model) GetPath() string {
	return m.path
}

func (m *Model) GetWidth() int {
	return m.width
}

func (m *Model) GetHeight() int {
	return m.height
}

func (m *Model) IsNoRepo() bool {
	return m.noRepo
}

func (m *Model) Branch() string {
	return m.branch
}

func (m *Model) Status() string {
	return m.status
}

func (m *Model) SetClipboard(items []string, cut bool) {
	m.cbItems = make([]string, len(items))
	copy(m.cbItems, items)
	m.cbCut = cut
	if m.cbIndex >= len(m.cbItems) {
		m.cbIndex = 0
	}
}

func (m *Model) NextTab() {
	if m.tab == clipboardTab {
		m.tab = gitTab
		return
	}
	m.tab = clipboardTab
}

func (m *Model) PrevTab() {
	m.NextTab()
}

func (m *Model) ActiveTab() int {
	return m.tab
}

func (m *Model) ActiveTabName() string {
	if m.tab == clipboardTab {
		return "clipboard"
	}
	return "git"
}

func (m *Model) ListUp() {
	if m.tab != clipboardTab || len(m.cbItems) == 0 {
		return
	}
	if m.cbIndex > 0 {
		m.cbIndex--
		return
	}
	m.cbIndex = len(m.cbItems) - 1
}

func (m *Model) ListDown() {
	if m.tab != clipboardTab || len(m.cbItems) == 0 {
		return
	}
	if m.cbIndex < len(m.cbItems)-1 {
		m.cbIndex++
		return
	}
	m.cbIndex = 0
}

func (m *Model) Date() string {
	return m.date
}

func (m *Model) Subject() string {
	return m.subject
}

func (m *Model) Render(focused bool) string {
	r := ui.GitRenderer(m.height, m.width, focused)
	r.SetBorderTitle(renderTabTitle(m.tab))

	viewWidth := m.width - common.InnerPadding
	if m.tab == clipboardTab {
		m.renderClipboardTab(r, viewWidth)
		return r.Render()
	}

	if m.noRepo {
		r.AddLines("", " Not a git repository")
		return r.Render()
	}

	r.AddLines(" branch: " + common.TruncateText(m.branch, viewWidth-9, "..."))
	r.AddLines(" commit: " + common.TruncateText(m.subject, viewWidth-9, "..."))
	r.AddLines(" date: " + common.TruncateText(m.date, viewWidth-7, "..."))
	r.AddLines(" author: " + common.TruncateText(m.author, viewWidth-9, "..."))
	r.AddLines(" status: " + common.TruncateText(m.status, viewWidth-9, "..."))

	if strings.TrimSpace(m.subject) == "" {
		r.AddLines("", " No commit history")
	}

	return r.Render()
}

func renderTabTitle(activeTab int) string {
	if activeTab == clipboardTab {
		return "G.. | Clipboard"
	}
	return "Git | C.."
}

func (m *Model) renderClipboardTab(r *rendering.Renderer, viewWidth int) {
	mode := "copy"
	if m.cbCut {
		mode = "cut"
	}
	r.AddLines(" mode: " + mode)

	contentRows := m.height - common.BorderPadding
	if contentRows <= 1 {
		return
	}
	listRows := contentRows - 1

	if len(m.cbItems) == 0 {
		r.AddLines(common.ClipboardNoneText)
		return
	}
	r.SetBorderInfoItems(strconv.Itoa(m.cbIndex+1) + "/" + strconv.Itoa(len(m.cbItems)))

	for i := m.cbIndex; i < len(m.cbItems) && i < m.cbIndex+listRows; i++ {
		if i == m.cbIndex+listRows-1 && i != len(m.cbItems)-1 {
			r.AddLines(strconv.Itoa(len(m.cbItems)-i) + " items left....")
			continue
		}
		fileInfo, err := os.Lstat(m.cbItems[i])
		if err != nil {
			continue
		}
		isLink := fileInfo.Mode()&os.ModeSymlink != 0
		r.AddLines(common.ClipboardPrettierName(m.cbItems[i], viewWidth, fileInfo.IsDir(), isLink, false))
	}
}
