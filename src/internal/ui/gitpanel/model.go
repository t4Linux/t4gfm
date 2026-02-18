package gitpanel

import (
	"strings"

	"github.com/t4Linux/t4gfm/src/internal/common"
	"github.com/t4Linux/t4gfm/src/internal/ui"
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
}

func New() Model {
	return Model{noRepo: true}
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

func (m *Model) Date() string {
	return m.date
}

func (m *Model) Subject() string {
	return m.subject
}

func (m *Model) Render(focused bool) string {
	r := ui.GitRenderer(m.height, m.width, focused)
	if m.noRepo {
		r.AddLines("", " Not a git repository")
		return r.Render()
	}

	viewWidth := m.width - common.InnerPadding
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
