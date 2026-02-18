package internal

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type gitInfo struct {
	branch  string
	subject string
	date    string
	author  string
	status  string
	noRepo  bool
}

func (m *model) getGitInfoCmd() tea.Cmd {
	path := m.getFocusedFilePanel().Location
	if path == "" || path == m.gitPanel.GetPath() {
		return nil
	}

	reqCnt := m.ioReqCnt
	m.ioReqCnt++

	return func() tea.Msg {
		info := fetchGitInfo(path)
		return NewGitInfoMsg(path, info, reqCnt)
	}
}

func fetchGitInfo(path string) gitInfo {
	branchCmd := exec.Command("git", "-C", path, "branch", "--show-current")
	branchOut, err := branchCmd.Output()
	if err != nil {
		return gitInfo{noRepo: true}
	}

	branch := strings.TrimSpace(string(branchOut))
	if branch == "" {
		headCmd := exec.Command("git", "-C", path, "rev-parse", "--short", "HEAD")
		headOut, headErr := headCmd.Output()
		if headErr == nil {
			branch = "detached@" + strings.TrimSpace(string(headOut))
		}
	}

	logCmd := exec.Command("git", "-C", path, "log", "-1", "--pretty=format:%s%n%ad%n%an", "--date=iso")
	logOut, logErr := logCmd.Output()
	if logErr != nil {
		return gitInfo{branch: branch, status: fetchGitStatusSummary(path)}
	}

	lines := strings.Split(strings.TrimSpace(string(logOut)), "\n")
	info := gitInfo{branch: branch}
	if len(lines) > 0 {
		info.subject = lines[0]
	}
	if len(lines) > 1 {
		info.date = lines[1]
	}
	if len(lines) > 2 {
		info.author = lines[2]
	}
	info.status = fetchGitStatusSummary(path)
	return info
}

func fetchGitStatusSummary(path string) string {
	cmd := exec.Command("git", "-C", path, "status", "--porcelain", "--branch")
	out, err := cmd.Output()
	if err != nil {
		return "unknown"
	}

	var ahead, behind int
	staged := 0
	unstaged := 0
	untracked := 0
	conflicts := 0

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for i, line := range lines {
		if line == "" {
			continue
		}
		if i == 0 && strings.HasPrefix(line, "## ") {
			ahead, behind = parseAheadBehind(line)
			continue
		}
		if strings.HasPrefix(line, "??") {
			untracked++
			continue
		}
		if len(line) < 2 {
			continue
		}
		x := line[0]
		y := line[1]
		if x != ' ' {
			staged++
		}
		if y != ' ' {
			unstaged++
		}
		if x == 'U' || y == 'U' {
			conflicts++
		}
	}

	parts := make([]string, 0, 6)
	if ahead > 0 {
		parts = append(parts, fmt.Sprintf("ahead %d", ahead))
	}
	if behind > 0 {
		parts = append(parts, fmt.Sprintf("behind %d", behind))
	}
	if staged > 0 {
		parts = append(parts, fmt.Sprintf("staged %d", staged))
	}
	if unstaged > 0 {
		parts = append(parts, fmt.Sprintf("unstaged %d", unstaged))
	}
	if untracked > 0 {
		parts = append(parts, fmt.Sprintf("untracked %d", untracked))
	}
	if conflicts > 0 {
		parts = append(parts, fmt.Sprintf("conflicts %d", conflicts))
	}
	if len(parts) == 0 {
		return "clean ✓"
	}
	return strings.Join(parts, ", ")
}

func parseAheadBehind(branchLine string) (int, int) {
	ahead := 0
	behind := 0
	openIdx := strings.Index(branchLine, "[")
	closeIdx := strings.Index(branchLine, "]")
	if openIdx == -1 || closeIdx == -1 || closeIdx <= openIdx+1 {
		return ahead, behind
	}
	parts := strings.Split(branchLine[openIdx+1:closeIdx], ",")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if strings.HasPrefix(p, "ahead ") {
			if v, err := strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(p, "ahead "))); err == nil {
				ahead = v
			}
		}
		if strings.HasPrefix(p, "behind ") {
			if v, err := strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(p, "behind "))); err == nil {
				behind = v
			}
		}
	}
	return ahead, behind
}
