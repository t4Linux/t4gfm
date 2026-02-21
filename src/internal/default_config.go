package internal

import (
	zoxidelib "github.com/lazysegtree/go-zoxide"

	"github.com/t4Linux/t4gfm/src/internal/ui/helpmenu"

	"github.com/t4Linux/t4gfm/src/internal/ui/filemodel"
	"github.com/t4Linux/t4gfm/src/internal/ui/gitpanel"
	"github.com/t4Linux/t4gfm/src/internal/ui/sortmodel"
	"github.com/t4Linux/t4gfm/src/internal/ui/systempanel"

	"github.com/t4Linux/t4gfm/src/internal/ui/metadata"
	"github.com/t4Linux/t4gfm/src/internal/ui/processbar"
	"github.com/t4Linux/t4gfm/src/internal/ui/sidebar"

	"github.com/t4Linux/t4gfm/src/internal/common"
	"github.com/t4Linux/t4gfm/src/internal/ui/prompt"
	zoxideui "github.com/t4Linux/t4gfm/src/internal/ui/zoxide"
)

// Generate and return model containing default configurations for interface
// Maybe we can replace slice of strings with var args - Should we ?
// TODO: Move the configuration parameters to a ModelConfig struct.
// Something like `RendererConfig` struct for `Renderer` struct in ui/renderer package
// Or even better API like varargs lambda function opts
// which can be WithFooter(), WithXYZ()
// Lots of improvements are waiting on it
//   - Allow Sending thumbnailGeneratorNeeded as false to preview.New()
//     to prevent noise in test logs. Same with imagePreviewer
func defaultModelConfig(toggleDotFile, toggleFooter, compactFooter, previewOpen, firstUse bool,
	firstPanelPaths []string, focusedPanelIndex int, zClient *zoxidelib.Client) *model {
	sidebarVisibleWidth := common.Config.SidebarWidth
	if sidebarVisibleWidth == 0 {
		sidebarVisibleWidth = 20
	}

	fileModelState := filemodel.New(firstPanelPaths, toggleDotFile)
	if previewOpen {
		fileModelState.FilePreview.Open()
	} else {
		fileModelState.FilePreview.Close()
	}

	if focusedPanelIndex > 0 && focusedPanelIndex < len(fileModelState.FilePanels) {
		for i := range fileModelState.FilePanels {
			fileModelState.FilePanels[i].IsFocused = i == focusedPanelIndex
		}
		fileModelState.FocusedPanelIndex = focusedPanelIndex
	}

	return &model{
		focusPanel:          nonePanelFocus,
		processBarModel:     processbar.New(),
		systemPanel:         systempanel.New(),
		sidebarModel:        sidebar.New(),
		fileMetaData:        metadata.New(),
		gitPanel:            gitpanel.New(),
		fileModel:           fileModelState,
		helpMenu:            helpmenu.New(),
		promptModal:         prompt.DefaultModel(prompt.PromptMinHeight, prompt.PromptMinWidth),
		zoxideModal:         zoxideui.DefaultModel(zoxideui.ZoxideMinHeight, zoxideui.ZoxideMinWidth, zClient),
		sortModal:           sortmodel.New(),
		zClient:             zClient,
		modelQuitState:      notQuitting,
		rangerMarks:         loadRangerMarks(),
		sidebarVisibleWidth: sidebarVisibleWidth,
		toggleFooter:        toggleFooter,
		compactFooter:       compactFooter,
		compressLevel:       compressLevelBalanced,
		compressVerbose:     false,
		compressExclude:     false,
		firstUse:            firstUse,
		hasTrash:            common.InitTrash(),
	}
}
