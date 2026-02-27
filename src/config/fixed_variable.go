package variable

import (
	"os"
	"path/filepath"

	"github.com/urfave/cli/v3"

	"github.com/t4Linux/t4gfm/src/internal/utils"

	"github.com/adrg/xdg"
)

const (
	AppName = "t4gfm"

	CurrentVersion = "v1.5.0"
	// Allowing pre-releases with non production version
	// Set this to "" for production releases
	PreReleaseSuffix = ""

	// This gives most recent non-prerelease, non-draft release
	LatestVersionURL    = "https://api.github.com/repos/t4Linux/t4gfm/releases/latest"
	LatestVersionGithub = "github.com/t4Linux/t4gfm/releases/latest"

	EmbedConfigDir           = "src/t4gfm_config"
	EmbedConfigFile          = EmbedConfigDir + "/config.toml"
	EmbedHotkeysFile         = EmbedConfigDir + "/hotkeys.toml"
	EmbedThemeDir            = EmbedConfigDir + "/theme"
	EmbedThemeCatppuccinFile = EmbedThemeDir + "/catppuccin-mocha.toml"
)

var (
	HomeDir     = xdg.Home
	AppMainDir  = filepath.Join(xdg.ConfigHome, AppName)
	AppCacheDir = filepath.Join(xdg.CacheHome, AppName)
	AppDataDir  = filepath.Join(xdg.DataHome, AppName)
	AppStateDir = filepath.Join(xdg.StateHome, AppName)

	// MainDir files
	ThemeFolder = filepath.Join(AppMainDir, "theme")

	// DataDir files
	LastCheckVersion    = filepath.Join(AppDataDir, "lastCheckVersion")
	ThemeFileVersion    = filepath.Join(AppDataDir, "themeFileVersion")
	FirstUseCheck       = filepath.Join(AppDataDir, "firstUseCheck")
	PinnedFile          = filepath.Join(AppDataDir, "pinned.json")
	RangerMarksFile     = filepath.Join(AppDataDir, "ranger_marks.json")
	ToggleDotFile       = filepath.Join(AppDataDir, "toggleDotFile")
	ToggleFooter        = filepath.Join(AppDataDir, "toggleFooter")
	ToggleCompactFooter = filepath.Join(AppDataDir, "toggleCompactFooter")
	ToggleSidebar       = filepath.Join(AppDataDir, "toggleSidebar")
	ToggleFilePreview   = filepath.Join(AppDataDir, "toggleFilePreview")
	PanelSessionFile    = filepath.Join(AppDataDir, "panelSession.json")

	// StateDir files
	LogFile     = filepath.Join(AppStateDir, "t4gfm.log")
	LastDirFile = filepath.Join(AppStateDir, "lastdir")

	// Trash Directories
	DarwinTrashDirectory = filepath.Join(HomeDir, ".Trash")

	// These are used by github.com/rkoesters/xdg/trash package
	// We need to make sure that these directories exist
	LinuxTrashDirectory      = filepath.Join(xdg.DataHome, "Trash")
	LinuxTrashDirectoryFiles = filepath.Join(xdg.DataHome, "Trash", "files")
	LinuxTrashDirectoryInfo  = filepath.Join(xdg.DataHome, "Trash", "info")
)

// These variables are actually not fixed, they are sometimes updated dynamically
var (
	ConfigFile  = filepath.Join(AppMainDir, "config.toml")
	HotkeysFile = filepath.Join(AppMainDir, "hotkeys.toml")

	ChooserFile = ""

	// Other state variables
	FixHotkeys    = false
	FixConfigFile = false
	LastDir       = ""
	PrintLastDir  = false
	ThemeOverride = ""
)

// Still we are preventing other packages to directly modify them via reassign linter

func SetLastDir(path string) {
	LastDir = path
}

func SetChooserFile(path string) {
	ChooserFile = path
}

func UpdateVarFromCliArgs(c *cli.Command) {
	// Setting the config file path
	configFileArg := c.String("config-file")

	// Validate the config file exists
	if configFileArg != "" {
		if _, err := os.Stat(configFileArg); err != nil {
			utils.PrintfAndExitf("Error: While reading config file '%s' from argument : %v", configFileArg, err)
		}
		ConfigFile = configFileArg
	}

	hotkeyFileArg := c.String("hotkey-file")

	if hotkeyFileArg != "" {
		if _, err := os.Stat(hotkeyFileArg); err != nil {
			utils.PrintfAndExitf("Error: While reading hotkey file '%s' from argument : %v", hotkeyFileArg, err)
		}
		HotkeysFile = hotkeyFileArg
	}

	// It could be non existent. We are writing to the file. If file doesn't exists, we would attempt to create it.
	SetChooserFile(c.String("chooser-file"))

	FixHotkeys = c.Bool("fix-hotkeys")
	FixConfigFile = c.Bool("fix-config-file")
	PrintLastDir = c.Bool("print-last-dir")
	ThemeOverride = c.String("theme")
}
