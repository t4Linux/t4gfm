package internal

import (
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/charmbracelet/x/ansi"

	"github.com/t4Linux/t4gfm/src/internal/common"
)

func (m *model) compactFooterLine() string {
	left := m.compactSelectionInfo()
	center := m.compactGitInfo()
	right := m.compactRightInfo()
	if m.shouldHideCompactRightInfo() {
		right = center
		center = ""
	}
	line := mergeFooterLine(m.fullWidth, left, center, right)
	return common.FooterStyle.Width(m.fullWidth).Render(line)
}

func (m *model) shouldHideCompactRightInfo() bool {
	panel := m.getFocusedFilePanel()
	if panel.EmptyOrInvalid() {
		return false
	}
	return panel.IsFocusedNameTruncated()
}

func (m *model) compactSelectionInfo() string {
	panel := m.getFocusedFilePanel()
	if panel.EmptyOrInvalid() {
		return "- ---------- - -"
	}
	item := panel.GetFocusedItem()
	name := item.Name
	if strings.TrimSpace(name) == "" {
		name = filepath.Base(item.Location)
	}
	if strings.TrimSpace(name) == "" {
		name = "-"
	}
	perm := item.Info.Mode().String()
	return name + " " + perm
}

func (m *model) compactGitInfo() string {
	if m.gitPanel.IsNoRepo() {
		return ""
	}
	branch := strings.TrimSpace(m.gitPanel.Branch())
	if branch == "" {
		branch = "detached"
	}
	status := compactGitStatus(m.gitPanel.Status())
	commit := strings.TrimSpace(m.gitPanel.Subject())
	if commit == "" {
		commit = "-"
	}
	commit = ansi.Truncate(commit, 36, "...")
	parts := []string{"(git: " + branch + ")", status}
	parts = append(parts, commit)
	return strings.Join(parts, " ")
}

func (m *model) compactRightInfo() string {
	if m.processBarModel.HasRunningProcesses() {
		lines := m.processBarModel.SummaryLines(1)
		if len(lines) > 0 {
			return lines[0]
		}
	}

	path := m.getFocusedFilePanel().Location
	if path == "" {
		return "free: -"
	}
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return "free: -"
	}
	total := uint64(stat.Blocks) * uint64(stat.Bsize)
	free := uint64(stat.Bavail) * uint64(stat.Bsize)
	used := total - free
	freePct := 0
	if total > 0 {
		freePct = int((free * 100) / total)
	}
	return common.FormatFileSize(int64(used)) + "/" + common.FormatFileSize(int64(total)) + " " + strconv.Itoa(freePct) + "% free"
}

func compactGitStatus(status string) string {
	status = strings.TrimSpace(status)
	if status == "" || status == "unknown" {
		return "?"
	}
	if status == "clean ✓" {
		return "=0"
	}
	parts := strings.Split(status, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		split := strings.Split(part, " ")
		if len(split) != 2 {
			continue
		}
		n := split[1]
		switch split[0] {
		case "ahead":
			out = append(out, "↑"+n)
		case "behind":
			out = append(out, "↓"+n)
		case "staged":
			out = append(out, "+"+n)
		case "unstaged":
			out = append(out, "~"+n)
		case "untracked":
			out = append(out, "?"+n)
		case "conflicts":
			out = append(out, "!"+n)
		}
	}
	if len(out) == 0 {
		return "?"
	}
	return strings.Join(out, " ")
}

func mergeFooterLine(width int, left string, center string, right string) string {
	if width <= 0 {
		return ""
	}
	line := []rune(strings.Repeat(" ", width))
	left = string([]rune(ansi.Truncate(left, width, "...")))
	leftRunes := []rune(left)
	for i, r := range []rune(left) {
		if i >= len(line) {
			break
		}
		line[i] = r
	}

	rightFullRunes := []rune(right)
	maxRightWidth := width - len(leftRunes) - 1
	if maxRightWidth <= 0 || len(rightFullRunes) > maxRightWidth {
		right = ""
	} else {
		right = string([]rune(ansi.Truncate(right, maxRightWidth, "...")))
	}
	rightRunes := []rune(right)
	if len(rightRunes) > 0 {
		start := width - len(rightRunes)
		if start < 0 {
			start = 0
		}
		for i, r := range rightRunes {
			pos := start + i
			if pos >= len(line) {
				break
			}
			line[pos] = r
		}
	}

	if center == "" {
		return string(line)
	}

	availableCenterWidth := width - len(leftRunes) - len(rightRunes) - 2
	if availableCenterWidth <= 0 {
		return string(line)
	}
	center = string([]rune(ansi.Truncate(center, availableCenterWidth, "...")))
	centerRunes := []rune(center)
	start := (width - len(centerRunes)) / 2
	if start < len(leftRunes)+1 {
		start = len(leftRunes) + 1
	}
	rightStart := width - len(rightRunes)
	if rightStart < 0 {
		rightStart = 0
	}
	if start+len(centerRunes) >= rightStart {
		start = rightStart - len(centerRunes) - 1
	}
	if start < 0 {
		start = 0
	}
	for i, r := range centerRunes {
		pos := start + i
		if pos >= len(line) {
			break
		}
		line[pos] = r
	}
	return string(line)
}
