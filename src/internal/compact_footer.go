package internal

import (
	"os"
	"os/user"
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
	line := mergeFooterLine(m.fullWidth, left, center, right)
	return common.FooterStyle.Width(m.fullWidth).Render(line)
}

func (m *model) compactSelectionInfo() string {
	panel := m.getFocusedFilePanel()
	if panel.EmptyOrInvalid() {
		return "---------- - - -"
	}
	item := panel.GetFocusedItem()
	perm := item.Info.Mode().String()
	owner, group := ownerGroup(item.Info)
	sizeOrCount := common.FormatFileSize(item.Info.Size())
	if item.Directory {
		if entries, err := os.ReadDir(item.Location); err == nil {
			sizeOrCount = strconv.Itoa(len(entries))
		}
	}
	date := item.Info.ModTime().Format("06-01-02 15:04")
	return perm + " " + owner + " " + group + " " + sizeOrCount + " " + date
}

func ownerGroup(info os.FileInfo) (string, string) {
	owner := "-"
	group := "-"
	if stat, ok := info.Sys().(*syscall.Stat_t); ok {
		uid := strconv.FormatUint(uint64(stat.Uid), 10)
		gid := strconv.FormatUint(uint64(stat.Gid), 10)
		if userData, err := user.LookupId(uid); err == nil {
			owner = userData.Username
		}
		if groupData, err := user.LookupGroupId(gid); err == nil {
			group = groupData.Name
		}
	}
	return owner, group
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
	date := strings.TrimSpace(m.gitPanel.Date())
	if len(date) >= 10 {
		date = date[:10]
		if len(date) == 10 {
			date = date[2:]
		}
	}
	commit := strings.TrimSpace(m.gitPanel.Subject())
	if commit == "" {
		commit = "-"
	}
	commit = ansi.Truncate(commit, 36, "...")
	parts := []string{"(git: " + branch + ")", status}
	if date != "" {
		parts = append(parts, date)
	}
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
	for i, r := range []rune(left) {
		if i >= len(line) {
			break
		}
		line[i] = r
	}
	right = string([]rune(ansi.Truncate(right, width, "...")))
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
	center = string([]rune(ansi.Truncate(center, width, "...")))
	centerRunes := []rune(center)
	start := (width - len(centerRunes)) / 2
	if start < len([]rune(left))+1 {
		start = len([]rune(left)) + 1
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
