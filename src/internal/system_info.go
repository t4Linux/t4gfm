package internal

import (
	"fmt"
	"net"
	"os"
	"sort"
	"strings"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/t4Linux/t4gfm/src/internal/common"
)

type systemInfo struct {
	hostname string
	localIP  string
	disk     string
}

func (m *model) getSystemInfoCmd() tea.Cmd {
	path := m.getFocusedFilePanel().Location
	if path == "" || path == m.systemPanel.GetPath() {
		return nil
	}
	reqCnt := m.ioReqCnt
	m.ioReqCnt++
	return func() tea.Msg {
		return NewSystemInfoMsg(path, fetchSystemInfo(path), reqCnt)
	}
}

func fetchSystemInfo(path string) systemInfo {
	hostname := "unknown"
	if h, err := os.Hostname(); err == nil && h != "" {
		hostname = h
	}

	ipSet := map[string]struct{}{}
	ifaces, err := net.Interfaces()
	if err == nil {
		for _, iface := range ifaces {
			if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
				continue
			}
			if !isPhysicalInterface(iface.Name) {
				continue
			}
			addrs, addrErr := iface.Addrs()
			if addrErr != nil {
				continue
			}
			for _, addr := range addrs {
				ipnet, ok := addr.(*net.IPNet)
				if !ok || ipnet.IP.IsLoopback() {
					continue
				}
				if ip4 := ipnet.IP.To4(); ip4 != nil {
					ipSet[ip4.String()] = struct{}{}
				}
			}
		}
	}

	localIP := "unavailable"
	if len(ipSet) > 0 {
		ips := make([]string, 0, len(ipSet))
		for ip := range ipSet {
			ips = append(ips, ip)
		}
		sort.Strings(ips)
		localIP = strings.Join(ips, ", ")
	}

	disk := "unavailable"
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err == nil {
		total := uint64(stat.Blocks) * uint64(stat.Bsize)
		free := uint64(stat.Bavail) * uint64(stat.Bsize)
		used := total - free
		freePct := 0
		if total > 0 {
			freePct = int((free * 100) / total)
		}
		disk = fmt.Sprintf("%s / %s [ %d%% free ]", common.FormatFileSize(int64(used)), common.FormatFileSize(int64(total)), freePct)
	}

	return systemInfo{hostname: hostname, localIP: localIP, disk: disk}
}

func isPhysicalInterface(name string) bool {
	return strings.HasPrefix(name, "eth") ||
		strings.HasPrefix(name, "en") ||
		strings.HasPrefix(name, "wl") ||
		strings.HasPrefix(name, "wlan")
}
