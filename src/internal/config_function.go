package internal

import (
	"errors"
	"log/slog"
	"os"
	"reflect"
	"runtime"

	zoxidelib "github.com/lazysegtree/go-zoxide"

	"github.com/t4Linux/t4gfm/src/internal/ui/filepanel"

	"github.com/t4Linux/t4gfm/src/internal/ui/processbar"
	"github.com/t4Linux/t4gfm/src/internal/ui/rendering"
	"github.com/t4Linux/t4gfm/src/internal/ui/sidebar"
	"github.com/t4Linux/t4gfm/src/internal/utils"

	"github.com/barasher/go-exiftool"

	variable "github.com/t4Linux/t4gfm/src/config"
	"github.com/t4Linux/t4gfm/src/config/icon"
	"github.com/t4Linux/t4gfm/src/internal/common"
)

// This is the only usecase of named returns, distinguish between multiple return values
func initialConfig(firstPanelPaths []string) (resolvedPanelPaths []string, //nolint: nonamedreturns // See above
	toggleDotFile bool, toggleFooter bool, compactFooter bool, previewOpen bool,
	focusedPanelIndex int, zClient *zoxidelib.Client) {
	// Open log stream
	file, err := os.OpenFile(variable.LogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, utils.LogFilePerm)

	// For example if the log file directories have access issues.
	// we could pass a dummy object to log.SetOutput() and the app would still function.
	if err != nil {
		utils.PrintfAndExitf("Error while opening t4gfm.log file : %v", err)
	}
	common.LoadConfigFile()

	logLevel := slog.LevelInfo
	if common.Config.Debug {
		logLevel = slog.LevelDebug
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(
		file, &slog.HandlerOptions{Level: logLevel})))

	printRuntimeInfo()

	common.LoadHotkeysFile(common.Config.IgnoreMissingFields)

	common.LoadThemeFile()

	icon.InitIcon(common.Config.Nerdfont, common.Theme.DirectoryIconColor)

	common.LoadThemeConfig()
	common.LoadPrerenderedVariables()

	// TODO: Make sure to clean it up. Via et.Close()
	// Note: All the tool we use to interact with OS, should be abstracted behind a struc
	// Have exiftool manager, Zoxide Manager, OS Manager, Xtractor, Zipper, Command Executor
	if common.Config.Metadata {
		et, err = exiftool.NewExiftool()
		if err != nil {
			slog.Error("Error while initial model function init exiftool error", "error", err)
		}
	}

	cwd, err := os.Getwd()
	if err != nil {
		slog.Error("cannot get current working directory", "error", err)
		cwd = variable.HomeDir
	}

	if common.Config.ZoxideSupport {
		zClient, err = zoxidelib.New()
		if err != nil {
			slog.Error("Error initializing zoxide client", "error", err)
		}
	}

	resolvedPanelPaths = append([]string{}, firstPanelPaths...)
	shouldRestorePanelSession := len(firstPanelPaths) == 1 && firstPanelPaths[0] == ""
	if shouldRestorePanelSession {
		sessionState, sessionErr := loadPanelSessionState()
		if sessionErr != nil {
			slog.Error("Failed loading panel session state", "error", sessionErr)
		} else if sessionState != nil && len(sessionState.PanelPaths) > 0 {
			resolvedPanelPaths = append([]string{}, sessionState.PanelPaths...)
			previewOpen = sessionState.FilePreviewVisible
			focusedPanelIndex = sessionState.FocusedPanelIndex
		}
	}

	updateFirstFilePanelPaths(resolvedPanelPaths, cwd, zClient)

	slog.Debug("Directory configuration", "cwd", cwd, "start_paths", resolvedPanelPaths)
	printRuntimeInfo()

	toggleDotFile = utils.ReadBoolFile(variable.ToggleDotFile, false)
	toggleFooter = utils.ReadBoolFile(variable.ToggleFooter, true)
	compactFooter = utils.ReadBoolFile(variable.ToggleCompactFooter, false)
	previewOpen = utils.ReadBoolFile(variable.ToggleFilePreview, common.Config.DefaultOpenFilePreview)
	sidebarVisible := utils.ReadBoolFile(variable.ToggleSidebar, common.Config.SidebarWidth != 0)
	if !toggleFooter {
		compactFooter = false
	}
	if !sidebarVisible {
		common.Config.SidebarWidth = 0
	}
	if focusedPanelIndex < 0 || focusedPanelIndex >= len(resolvedPanelPaths) {
		focusedPanelIndex = 0
	}

	return resolvedPanelPaths, toggleDotFile, toggleFooter, compactFooter, previewOpen,
		focusedPanelIndex, zClient
}

func updateFirstFilePanelPaths(firstPanelPaths []string, cwd string, zClient *zoxidelib.Client) {
	for i := range firstPanelPaths {
		if firstPanelPaths[i] == "" {
			firstPanelPaths[i] = common.Config.DefaultDirectory
		}
		originalPath := firstPanelPaths[i]
		firstPanelPaths[i] = utils.ResolveAbsPath(cwd, firstPanelPaths[i])
		if _, err := os.Stat(firstPanelPaths[i]); err != nil {
			slog.Error("cannot get stats", "path", firstPanelPaths[i], "error", err)
			// In case the path provided did not exist, use zoxide query
			// else, fallback to home dir
			if common.Config.ZoxideSupport && zClient != nil {
				path, err := attemptZoxideForInitPath(originalPath, zClient)
				if err != nil {
					slog.Error("Zoxide query error", "originalPath", originalPath, "error", err)
					firstPanelPaths[i] = variable.HomeDir
				} else {
					firstPanelPaths[i] = path
				}
			} else {
				firstPanelPaths[i] = variable.HomeDir
			}
		}
	}
}

func attemptZoxideForInitPath(originalPath string, zClient *zoxidelib.Client) (string, error) {
	path, err := zClient.Query(originalPath)

	if err != nil {
		return "", err
	}
	if path == "" {
		return "", errors.New("zoxide returned empty path")
	}
	if stat, statErr := os.Stat(path); statErr != nil || !stat.IsDir() {
		return "", errors.New("zoxide returned invalid path")
	}
	return path, nil
}

func printRuntimeInfo() {
	slog.Debug("Runtime information", "runtime.GOOS", runtime.GOOS)
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	slog.Debug("Memory usage",
		"alloc_bytes", memStats.Alloc,
		"total_alloc_bytes", memStats.TotalAlloc,
		"heap_objects", memStats.HeapObjects,
		"sys_bytes", memStats.Sys)
	slog.Debug("Object sizes",
		"model_size_bytes", reflect.TypeOf(model{}).Size(),
		"filePanel_size_bytes", reflect.TypeOf(filepanel.Model{}).Size(),
		"sidebarModel_size_bytes", reflect.TypeOf(sidebar.Model{}).Size(),
		"renderer_size_bytes", reflect.TypeOf(rendering.Renderer{}).Size(),
		"borderConfig_size_bytes", reflect.TypeOf(rendering.BorderConfig{}).Size(),
		"process_size_bytes", reflect.TypeOf(processbar.Process{}).Size())
}
