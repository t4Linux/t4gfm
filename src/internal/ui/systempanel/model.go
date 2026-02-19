package systempanel

import (
	"strings"

	"github.com/t4Linux/t4gfm/src/internal/common"
	"github.com/t4Linux/t4gfm/src/internal/ui"
)

type Model struct {
	width    int
	height   int
	path     string
	hostname string
	localIP  string
	disk     string
	procs    []string
}

func New() Model {
	return Model{}
}

func (m *Model) SetDimensions(width int, height int) {
	m.width = width
	m.height = height
}

func (m *Model) SetData(path string, hostname string, localIP string, disk string) {
	m.path = path
	m.hostname = hostname
	m.localIP = localIP
	m.disk = disk
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

func (m *Model) Render(focused bool) string {
	r := ui.SystemRenderer(m.height, m.width, focused)
	r.AddLines(" Hostname: " + m.hostname)
	ipList := strings.Split(m.localIP, ", ")
	if len(ipList) == 0 {
		r.AddLines(" IP: unavailable")
	} else {
		r.AddLines(" IP: " + ipList[0])
		for i := 1; i < len(ipList); i++ {
			r.AddLines("     " + ipList[i])
		}
	}
	r.AddLines(" Disk: " + m.disk)
	if len(m.procs) == 0 {
		r.AddLines("", " Processes: none")
	} else {
		r.AddLines("", " Processes:")
		viewWidth := m.width - common.InnerPadding
		for i := range m.procs {
			r.AddLines("  - " + common.TruncateText(m.procs[i], viewWidth-4, "..."))
		}
	}
	return r.Render()
}

func (m *Model) SetProcesses(processes []string) {
	m.procs = processes
}
