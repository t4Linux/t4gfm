package internal

import (
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	filepreview "github.com/t4Linux/t4gfm/src/pkg/file_preview"

	"github.com/t4Linux/t4gfm/src/internal/common"

	"github.com/charmbracelet/lipgloss"

	"github.com/t4Linux/t4gfm/src/config/icon"
)

func (m *model) sidebarRender() string {
	if common.Config.SidebarWidth == 0 {
		return ""
	}
	return m.sidebarModel.Render(m.focusPanel == sidebarFocus,
		m.getFocusedFilePanel().Location)
}

func (m *model) systemPanelRender() string {
	return m.systemPanel.Render(m.focusPanel == processBarFocus)
}

func (m *model) terminalSizeWarnRender() string {
	fullWidthString := strconv.Itoa(m.fullWidth)
	fullHeightString := strconv.Itoa(m.fullHeight)
	minimumWidthString := strconv.Itoa(common.MinimumWidth)
	minimumHeight := m.minimumRenderableHeight()
	minimumHeightString := strconv.Itoa(minimumHeight)
	if m.fullHeight < minimumHeight {
		fullHeightString = common.TerminalTooSmall.Render(fullHeightString)
	}
	if m.fullWidth < common.MinimumWidth {
		fullWidthString = common.TerminalTooSmall.Render(fullWidthString)
	}
	fullHeightString = common.TerminalCorrectSize.Render(fullHeightString)
	fullWidthString = common.TerminalCorrectSize.Render(fullWidthString)

	heightString := common.MainStyle.Render(" Height = ")
	return common.FullScreenStyle(m.fullHeight, m.fullWidth).Render(`Terminal size too small:`+"\n"+
		"Width = "+fullWidthString+
		heightString+fullHeightString+"\n\n"+
		"Needed for current config:"+"\n"+
		"Width = "+common.TerminalCorrectSize.Render(minimumWidthString)+
		heightString+common.TerminalCorrectSize.Render(minimumHeightString)) + filepreview.ClearKittyImages()
}

func (m *model) terminalSizeWarnAfterFirstRender() string {
	minimumWidthInt := common.Config.SidebarWidth + common.FilePanelWidthUnit*len(
		m.fileModel.FilePanels,
	) + common.FilePanelWidthUnit - 1
	minimumWidthString := strconv.Itoa(minimumWidthInt)
	fullWidthString := strconv.Itoa(m.fullWidth)
	fullHeightString := strconv.Itoa(m.fullHeight)
	minimumHeight := m.minimumRenderableHeight()
	minimumHeightString := strconv.Itoa(minimumHeight)

	if m.fullHeight < minimumHeight {
		fullHeightString = common.TerminalTooSmall.Render(fullHeightString)
	}
	if m.fullWidth < minimumWidthInt {
		fullWidthString = common.TerminalTooSmall.Render(fullWidthString)
	}
	fullHeightString = common.TerminalCorrectSize.Render(fullHeightString)
	fullWidthString = common.TerminalCorrectSize.Render(fullWidthString)

	heightString := common.MainStyle.Render(" Height = ")
	return common.FullScreenStyle(m.fullHeight, m.fullWidth).Render(`You change your terminal size too small:`+"\n"+
		"Width = "+fullWidthString+
		heightString+fullHeightString+"\n\n"+
		"Needed for current config:"+"\n"+
		"Width = "+common.TerminalCorrectSize.Render(minimumWidthString)+
		heightString+common.TerminalCorrectSize.Render(minimumHeightString)) + filepreview.ClearKittyImages()
}

func (m *model) typineModalRender() string {
	previewPath := filepath.Join(m.typingModal.location, m.typingModal.textInput.Value())
	confirmLabel := "Create"
	if m.typingModal.mode == typingModalOpenWith {
		previewPath = m.typingModal.targetPath
		confirmLabel = "Open"
	}

	fileLocation := common.FilePanelTopDirectoryIconStyle.Render(" "+icon.Directory+icon.Space) +
		common.FilePanelTopPathStyle.Render(
			common.TruncateTextBeginning(previewPath, common.ModalWidth-common.InnerPadding, "..."),
		) + "\n"

	confirm := common.ModalConfirm.Render(" (" + common.Hotkeys.ConfirmTyping[0] + ") " + confirmLabel + " ")
	cancel := common.ModalCancel.Render(" (" + common.Hotkeys.CancelTyping[0] + ") Cancel ")

	tip := confirm +
		lipgloss.NewStyle().Background(common.ModalBGColor).Render("           ") +
		cancel

	var err string
	if m.typingModal.errorMesssage != "" {
		err = "\n\n" + common.ModalErrorStyle.Render(m.typingModal.errorMesssage)
	}
	// TODO : Move this all to rendering package to avoid specifying newlines manually
	return common.ModalBorderStyle(common.ModalHeight, common.ModalWidth).
		Render(fileLocation + "\n" + m.typingModal.textInput.View() + "\n\n" + tip + err)
}

func (m *model) introduceModalRender() string {
	title := common.SidebarTitleStyle.Render(" Thanks for using t4gfm!!") +
		common.ModalStyle.Render("\n You can read the following information before starting to use it!")
	vimUserWarn := common.ProcessErrorStyle.Render("  ** Very importantly ** If you are a Vim/Nvim user,\n" +
		"  tune your hotkeys in ~/.config/t4gfm/hotkeys.toml")
	subOne := common.SidebarTitleStyle.Render("  (1)") +
		common.ModalStyle.Render(" If this is your first time, read README.md in this repository.")
	subTwo := common.SidebarTitleStyle.Render("  (2)") +
		common.ModalStyle.Render(" If you forget the relevant keys during use,\n"+
			"      you can press \"?\" (shift+/) at any time to query the keys!")
	subThree := common.SidebarTitleStyle.Render("  (3)") +
		common.ModalStyle.Render(" For more customization, edit files in ~/.config/t4gfm/")
	subFour := common.SidebarTitleStyle.Render("  (4)") +
		common.ModalStyle.Render(" Thank you again for using t4gfm.\n"+
			"      If you have any questions, please feel free to ask at:\n"+
			"      https://github.com/t4Linux/t4gfm\n"+
			"      Of course, you can always open a new issue to share your idea \n"+
			"      or report a bug!")
	subFive := common.SidebarTitleStyle.Render("  (5)") +
		common.ModalStyle.Render(" t4gfm is built as a fork inspired by superfile.\n"+
			"      We sincerely thank @yorukot for the original project and for\n"+
			"      creating such a strong foundation for this work:\n"+
			"      https://github.com/yorukot/superfile")
	return common.FirstUseModal(m.helpMenu.GetHeight(), m.helpMenu.GetWidth()).
		Render(title + "\n\n" + vimUserWarn + "\n\n" + subOne + "\n\n" +
			subTwo + "\n\n" + subThree + "\n\n" + subFour + "\n\n" + subFive + "\n\n")
}

func (m *model) promptModalRender() string {
	return m.promptModal.Render()
}

func (m *model) zoxideModalRender() string {
	return m.zoxideModal.Render()
}

func (m *model) rangerPrefixMenuRender() string {
	title := common.SidebarTitleStyle.Render(" Key menu")
	var lines string

	switch m.rangerPrefix {
	case "y":
		lines = " yy - copy selected item(s)\n" +
			" yp - copy selected path\n" +
			" yd - copy current directory\n" +
			" yn - copy selected name\n" +
			" y. - copy name without extension"
	case "d":
		lines = " dd - cut selected item(s)\n" +
			" dT - delete (trash) selected item(s)\n" +
			" dD - permanently delete selected item(s)\n" +
			" du - clear clipboard"
	case "p":
		lines = " pp - paste\n" +
			" po - paste overwrite\n" +
			" pP - paste append\n" +
			" pO - paste overwrite+append\n" +
			" pl - paste symlink (absolute)\n" +
			" pL - paste symlink (relative)\n" +
			" ph - hardlink submenu"
	case "ph":
		lines = " phl - paste hardlink\n" +
			" pht - paste hardlinked subtree"
	case "g":
		lines = " gg - go to top\n" +
			" gh - go home\n" +
			" gl - go to symlink source parent\n" +
			" g/ - go root\n" +
			" gp - go /tmp\n" +
			" ge - go /etc\n" +
			" gu - go /usr\n" +
			" gd - go /dev\n" +
			" go - go /opt\n" +
			" gv - go /var\n" +
			" gm - go /media\n" +
			" gM - go /mnt\n" +
			" gs - go /srv"
	case "o":
		lines = " or - toggle reverse sort\n" +
			" on/os/om/ot - sort by name/size/date/type\n" +
			" oN/oS/oM/oT - same in reverse"
	case "z":
		lines = " zh - toggle hidden files\n" +
			" zf - search"
	case "s":
		lines = " ss - focus System panel\n" +
			" sm - focus Metadata panel\n" +
			" sg - focus main file panel\n" +
			" st - focus Git panel\n" +
			" sp - jump to Pinned section\n" +
			" sd - jump to Disks section\n" +
			" sl - jump to top list section"
	case "m":
		lines = " mm - save current location under two-letter key"
	case "mm":
		lines = " mm? - first letter for location key (a-z)"
	case "mma", "mmb", "mmc", "mmd", "mme", "mmf", "mmg", "mmh", "mmi", "mmj", "mmk", "mml", "mmm", "mmn", "mmo", "mmp", "mmq", "mmr", "mms", "mmt", "mmu", "mmv", "mmw", "mmx", "mmy", "mmz":
		lines = " mm" + m.rangerPrefix[2:] + "? - second letter for location key (a-z)"
	case ";":
		lines = " ;? - list saved marks\n" +
			" ;xy - jump to saved mark\n" +
			" ;dxy - delete saved mark\n" +
			" ;D - delete all saved marks"
	case ";?":
		if len(m.rangerMarks) == 0 {
			lines = " no saved marks"
			break
		}
		keys := make([]string, 0, len(m.rangerMarks))
		for key := range m.rangerMarks {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		var b strings.Builder
		for i, key := range keys {
			if i > 0 {
				b.WriteString("\n")
			}
			b.WriteString(" ")
			b.WriteString(key)
			b.WriteString(" -> ")
			b.WriteString(m.rangerMarks[key])
		}
		lines = b.String()
	case ";d":
		lines = " ;d? - first letter of mark to delete (a-z)"
	case ";da", ";db", ";dc", ";dd", ";de", ";df", ";dg", ";dh", ";di", ";dj", ";dk", ";dl", ";dm", ";dn", ";do", ";dp", ";dq", ";dr", ";ds", ";dt", ";du", ";dv", ";dw", ";dx", ";dy", ";dz":
		lines = " ;d" + m.rangerPrefix[2:] + "? - second letter of mark to delete (a-z)"
	case ";a", ";b", ";c", ";e", ";f", ";g", ";h", ";i", ";j", ";k", ";l", ";m", ";n", ";o", ";p", ";q", ";r", ";s", ";t", ";u", ";v", ";w", ";x", ";y", ";z":
		lines = " ;" + m.rangerPrefix[1:] + "? - second letter of saved location key (a-z)"
	default:
		lines = " esc/q - cancel"
	}

	lines += "\n\n esc/q - cancel"
	content := common.ModalStyle.Render(lines)
	return common.ModalBorderStyle(common.ModalHeight, common.ModalWidth).Render(title + "\n\n" + content)
}
