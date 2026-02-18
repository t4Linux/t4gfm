package internal

import (
	"time"

	zoxidelib "github.com/lazysegtree/go-zoxide"

	"github.com/t4Linux/t4gfm/src/internal/ui/helpmenu"

	"github.com/t4Linux/t4gfm/src/internal/ui/clipboard"
	"github.com/t4Linux/t4gfm/src/internal/ui/gitpanel"
	"github.com/t4Linux/t4gfm/src/internal/ui/sortmodel"
	"github.com/t4Linux/t4gfm/src/internal/ui/systempanel"

	"github.com/t4Linux/t4gfm/src/internal/ui/filemodel"

	"github.com/t4Linux/t4gfm/src/internal/ui/metadata"
	"github.com/t4Linux/t4gfm/src/internal/ui/notify"
	"github.com/t4Linux/t4gfm/src/internal/ui/processbar"
	"github.com/t4Linux/t4gfm/src/internal/ui/sidebar"

	"github.com/charmbracelet/bubbles/textinput"

	"github.com/t4Linux/t4gfm/src/internal/ui/prompt"
	zoxideui "github.com/t4Linux/t4gfm/src/internal/ui/zoxide"
)

// Type representing the type of focused panel
type focusPanelType int

type modelQuitStateType int

// Constants for panel with no focus
const (
	nonePanelFocus focusPanelType = iota
	processBarFocus
	sidebarFocus
	metadataFocus
	gitPanelFocus
)

const (
	notQuitting modelQuitStateType = iota
	quitInitiated
	quitConfirmationInitiated
	quitConfirmationReceived
	quitDone
)

// Main model
// TODO : We could consider using *model as tea.Model, instead of model.
// for reducing re-allocations. The struct is 20K bytes. But this could lead to
// issues like race conditions and whatnot, which are hidden since we are creating
// new model in each tea update.
type model struct {
	// Main Panels
	fileModel       filemodel.Model
	sidebarModel    sidebar.Model
	processBarModel processbar.Model
	systemPanel     systempanel.Model
	clipboard       clipboard.Model
	gitPanel        gitpanel.Model
	focusPanel      focusPanelType

	// Modals
	notifyModel notify.Model
	typingModal typingModal
	helpMenu    helpmenu.Model
	promptModal prompt.Model
	zoxideModal zoxideui.Model
	sortModal   sortmodel.Model

	// Zoxide client for directory tracking
	zClient *zoxidelib.Client

	fileMetaData         metadata.Model
	ioReqCnt             int
	modelQuitState       modelQuitStateType
	firstTextInput       bool
	rangerPrefix         string
	rangerMarks          map[string]string
	lastPaneTitle        string
	compactFooter        bool
	metadataDebouncePath string
	metadataDebounceMode bool
	metadataDebounceAt   time.Time
	sidebarVisibleWidth  int
	toggleFooter         bool
	firstLoadingComplete bool
	firstUse             bool

	// This entirely disables metadata fetching. Used in test model
	disableMetadata bool

	// Height in number of lines of actual viewport of
	// main panel and sidebar excluding border
	mainPanelHeight int

	// Height in number of lines of actual viewport of
	// footer panels - process/metadata/clipboard - excluding border
	footerHeight int
	fullWidth    int
	fullHeight   int

	// whether usable trash directory exists or not
	hasTrash bool
}

type typingModal struct {
	location      string
	targetPath    string
	mode          typingModalMode
	open          bool
	textInput     textinput.Model
	errorMesssage string
}

type typingModalMode int

const (
	typingModalCreate typingModalMode = iota
	typingModalOpenWith
)

type editorFinishedMsg struct{ err error }

type metadataDebounceMsg struct{}
