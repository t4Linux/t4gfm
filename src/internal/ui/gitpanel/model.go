package gitpanel

import (
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/x/ansi"

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
	cbMeta  map[string]gitClipboardItemMeta
	cbDirty bool
}

type gitClipboardItemMeta struct {
	isDir  bool
	isLink bool
	ok     bool
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
	if m.cbCut == cut && equalStringSlices(m.cbItems, items) {
		return
	}
	m.cbItems = make([]string, len(items))
	copy(m.cbItems, items)
	m.cbCut = cut
	m.cbDirty = true
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

	r.AddLines(oneLineField("branch", m.branch, viewWidth))
	r.AddLines(multiLineFieldLines("commit", m.subject, viewWidth, 2)...)
	r.AddLines(oneLineField("date", m.date, viewWidth))
	r.AddLines(oneLineField("author", m.author, viewWidth))
	r.AddLines(multiLineFieldLines("status", m.status, viewWidth, 3)...)

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
	m.ensureClipboardMetaCache()
	r.SetBorderInfoItems(strconv.Itoa(m.cbIndex+1) + "/" + strconv.Itoa(len(m.cbItems)))

	for i := m.cbIndex; i < len(m.cbItems) && i < m.cbIndex+listRows; i++ {
		if i == m.cbIndex+listRows-1 && i != len(m.cbItems)-1 {
			r.AddLines(strconv.Itoa(len(m.cbItems)-i) + " items left....")
			continue
		}
		meta, ok := m.cbMeta[m.cbItems[i]]
		if !ok || !meta.ok {
			continue
		}
		r.AddLines(common.ClipboardPrettierName(m.cbItems[i], viewWidth, meta.isDir, meta.isLink, false))
	}
}

func (m *Model) ensureClipboardMetaCache() {
	if !m.cbDirty {
		return
	}
	if m.cbMeta == nil {
		m.cbMeta = make(map[string]gitClipboardItemMeta, len(m.cbItems))
	} else {
		for k := range m.cbMeta {
			delete(m.cbMeta, k)
		}
	}
	for _, item := range m.cbItems {
		fileInfo, err := os.Lstat(item)
		if err != nil {
			m.cbMeta[item] = gitClipboardItemMeta{ok: false}
			continue
		}
		m.cbMeta[item] = gitClipboardItemMeta{
			isDir:  fileInfo.IsDir(),
			isLink: fileInfo.Mode()&os.ModeSymlink != 0,
			ok:     true,
		}
	}
	m.cbDirty = false
}

func equalStringSlices(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func oneLineField(label string, value string, viewWidth int) string {
	prefix := " " + label + ": "
	avail := max(0, viewWidth-ansi.StringWidth(prefix))
	return prefix + truncateWithEllipsis(value, avail)
}

func multiLineFieldLines(label string, value string, viewWidth int, maxLines int) []string {
	if maxLines < 1 {
		maxLines = 1
	}
	prefix := " " + label + ": "
	availFirst := max(0, viewWidth-ansi.StringWidth(prefix))
	first, rest := splitByDisplayWidth(value, availFirst)
	continuationPrefix := strings.Repeat(" ", ansi.StringWidth(prefix))
	availContinuation := max(0, viewWidth-ansi.StringWidth(continuationPrefix))

	lines := []string{prefix + first}
	remaining := rest
	for i := 1; i < maxLines; i++ {
		if remaining == "" {
			break
		}
		if i == maxLines-1 {
			lines = append(lines, continuationPrefix+truncateWithEllipsis(remaining, availContinuation))
			break
		}
		part, next := splitByDisplayWidth(remaining, availContinuation)
		lines = append(lines, continuationPrefix+part)
		remaining = next
	}

	return lines
}

func truncateWithEllipsis(text string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	if ansi.StringWidth(text) <= maxWidth {
		return text
	}
	if maxWidth <= 3 {
		return strings.Repeat(".", maxWidth)
	}
	head, _ := splitByDisplayWidth(text, maxWidth-3)
	return head + "..."
}

func splitByDisplayWidth(text string, maxWidth int) (string, string) {
	if maxWidth <= 0 || text == "" {
		return "", text
	}
	curWidth := 0
	splitIdx := len(text)
	for idx, r := range text {
		rWidth := ansi.StringWidth(string(r))
		if curWidth+rWidth > maxWidth {
			splitIdx = idx
			break
		}
		curWidth += rWidth
	}
	if splitIdx == len(text) {
		return text, ""
	}
	return text[:splitIdx], text[splitIdx:]
}
