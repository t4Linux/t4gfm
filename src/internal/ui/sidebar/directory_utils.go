package sidebar

import (
	"os"
	"path/filepath"
	"sort"

	"github.com/adrg/xdg"

	"github.com/t4Linux/t4gfm/src/config/icon"
	"github.com/t4Linux/t4gfm/src/internal/common"
	"github.com/t4Linux/t4gfm/src/internal/utils"
)

// Fuzzy search function for a list of directories.
func fuzzySearch(query string, dirs []directory) []directory {
	if len(dirs) == 0 {
		return []directory{}
	}

	var filteredDirs []directory

	// Optimization - This haystack can be kept precomputed based on directories
	// instead of re computing it in each call
	haystack := make([]string, len(dirs))
	dirMap := make(map[string]directory, len(dirs))
	for i, dir := range dirs {
		haystack[i] = dir.Name
		dirMap[dir.Name] = dir
	}

	for _, match := range utils.FzfSearch(query, haystack) {
		if d, ok := dirMap[match.Key]; ok {
			filteredDirs = append(filteredDirs, d)
		}
	}

	return filteredDirs
}

// getDirectories returns the list of directories to display in the sidebar.
func getDirectories(pinnedMgr *PinnedManager, sections []string, currentLocation string) []directory {
	return formDirctorySlice(
		getParentDirectoryEntries(currentLocation),
		getPinnedDirectoriesWithIcon(pinnedMgr),
		getExternalMediaFolders(),
		sections,
	)
}

func getParentDirectoryEntries(currentLocation string) []directory {
	if currentLocation == "" {
		currentLocation = xdg.Home
	}

	parent := filepath.Dir(currentLocation)
	if parent == "." || parent == "" {
		parent = currentLocation
	}

	entries, err := os.ReadDir(parent)
	if err != nil {
		return nil
	}

	type parentEntry struct {
		name  string
		isDir bool
	}

	items := make([]parentEntry, 0, len(entries))
	for _, entry := range entries {
		isDir := entry.IsDir()
		if info, infoErr := entry.Info(); infoErr == nil && info.Mode()&os.ModeSymlink != 0 {
			if targetInfo, statErr := os.Stat(filepath.Join(parent, entry.Name())); statErr == nil {
				isDir = targetInfo.IsDir()
			}
		}
		items = append(items, parentEntry{name: entry.Name(), isDir: isDir})
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].isDir != items[j].isDir {
			return items[i].isDir
		}
		return items[i].name < items[j].name
	})

	dirs := make([]directory, 0, len(items)+1)
	dirs = append(dirs, directory{Location: parent, Name: icon.Directory + icon.Space + ".."})
	for _, entry := range items {
		iconInfo := common.GetElementIcon(entry.name, entry.isDir, false, common.Config.Nerdfont)
		dirs = append(dirs, directory{
			Location: filepath.Join(parent, entry.name),
			Name:     iconInfo.Icon + icon.Space + entry.name,
		})
	}

	return dirs
}

func getPinnedDirectoriesWithIcon(pinnedMgr *PinnedManager) []directory {
	dirs := pinnedMgr.Load()
	for i := range dirs {
		iconInfo := common.GetElementIcon(dirs[i].Name, true, false, common.Config.Nerdfont)
		dirs[i].Name = iconInfo.Icon + icon.Space + dirs[i].Name
	}
	return dirs
}

// Get filtered directories using fuzzy search logic with three haystacks.
func getFilteredDirectories(query string, pinnedMgr *PinnedManager, sections []string, currentLocation string) []directory {
	return formDirctorySlice(
		fuzzySearch(query, getParentDirectoryEntries(currentLocation)),
		fuzzySearch(query, getPinnedDirectoriesWithIcon(pinnedMgr)),
		fuzzySearch(query, getExternalMediaFolders()),
		sections,
	)
}

func formDirctorySlice(homeDirectories []directory, pinnedDirectories []directory,
	diskDirectories []directory, sections []string) []directory {
	// Preallocation for efficiency
	totalCapacity := len(homeDirectories) + len(pinnedDirectories) + len(diskDirectories) + directoryCapacityForDividers
	directories := make([]directory, 0, totalCapacity)

	for _, section := range sections {
		switch section {
		case utils.SidebarSectionHome:
			if len(directories) > 0 {
				directories = append(directories, homeDividerDir)
			}
			directories = append(directories, homeDirectories...)
		case utils.SidebarSectionPinned:
			directories = append(directories, pinnedDividerDir)
			directories = append(directories, pinnedDirectories...)
		case utils.SidebarSectionDisks:
			directories = append(directories, diskDividerDir)
			directories = append(directories, diskDirectories...)
		}
	}

	return directories
}
