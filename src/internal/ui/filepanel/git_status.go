package filepanel

import (
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func (m *Model) updateGitMarks(nowTime time.Time) {
	if m.gitMarksLocation == m.Location && nowTime.Sub(m.gitLastUpdate) < gitStatusRefreshInterval {
		return
	}

	m.gitLastUpdate = nowTime
	m.gitMarksLocation = m.Location
	m.gitMarks = make(map[string]string)
	m.gitAware = false

	out, err := exec.Command("git", "-C", m.Location, "status", "--porcelain", "--untracked-files=normal").Output()
	if err != nil {
		return
	}
	m.gitAware = true

	for _, line := range strings.Split(string(out), "\n") {
		if len(line) < 4 {
			continue
		}
		status := line[:2]
		path := strings.TrimSpace(line[3:])
		if path == "" {
			continue
		}
		if strings.Contains(path, " -> ") {
			parts := strings.Split(path, " -> ")
			path = parts[len(parts)-1]
		}
		path = strings.Trim(path, "\"")
		name := path
		if strings.Contains(path, string(filepath.Separator)) {
			name = strings.Split(path, string(filepath.Separator))[0]
		}

		if name == "" {
			continue
		}
		m.gitMarks[name] = mapGitStatusToMark(status)
	}
}

func mapGitStatusToMark(status string) string {
	switch {
	case status == "??":
		return "?"
	case strings.Contains(status, "A"):
		return "+"
	case strings.Contains(status, "M"):
		return "M"
	case strings.Contains(status, "D"):
		return "D"
	case strings.Contains(status, "R"):
		return "R"
	case strings.Contains(status, "U"):
		return "U"
	default:
		return "*"
	}
}
