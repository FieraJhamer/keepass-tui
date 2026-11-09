package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
	"github.com/tobischo/gokeepasslib/v3"
)

func (m model) View() string {
	if !m.ready {
		return "\n  Loading..."
	}

	if m.confirmMode != "" {
		return m.confirmView()
	}
	if m.iconPickMode {
		return m.iconPickView()
	}
	if m.editMode != editNone {
		return m.formView()
	}

	switch m.screen {
	case screenPicker:
		return m.pickerView()
	case screenUnlock:
		return m.unlockView()
	case screenBrowse:
		return m.browseView()
	case screenDetail:
		return m.detailView()
	}
	return ""
}

func (m model) pickerView() string {
	width := 80
	var b strings.Builder
	b.WriteString(lipgloss.NewStyle().Foreground(accentColor).Bold(true).Render("Select Database") + "\n\n")

	// Index 0 is the "create new" entry, the .kdbx files follow.
	items := []string{iconEmoji(4) + " Create new database"}
	for _, f := range m.files {
		items = append(items, iconEmoji(0)+" "+f)
	}

	for i, item := range items {
		if i == m.fileCursor {
			b.WriteString(lipgloss.NewStyle().Foreground(accentColor).Bold(true).Render("> "+item) + "\n")
		} else {
			b.WriteString("  " + item + "\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render(" [Arrow/j/k] Navigate  │  [Enter] Select  │  [Ctrl+N] New database  │  [q] Quit "))

	box := unlockStyle.Width(width).Render(b.String())
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

func (m model) unlockView() string {
	width := 50

	content := fmt.Sprintf(
		"%s\n\nPassword: %s",
		lipgloss.NewStyle().Foreground(accentColor).Bold(true).Render("KeePass TUI"),
		m.unlockInput.View(),
	)

	if m.unlockErr != "" {
		content += "\n\n" + lipgloss.NewStyle().Foreground(redColor).Render(m.unlockErr)
	}

	box := unlockStyle.Width(width).Render(content)

	help := lipgloss.NewStyle().Foreground(grayColor).Render(" [Enter] Unlock  │  [Ctrl+P] Toggle  │  [Esc] Back  │  [Ctrl+C] Quit ")
	box = lipgloss.JoinVertical(lipgloss.Center, box, help)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

func (m model) browseView() string {
	contentWidth := m.width - sidebarWidth - 6
	if contentWidth < 20 {
		contentWidth = 20
	}

	sidebarContent := m.renderSidebar()
	sidebarView := sidebarStyle.Width(sidebarWidth).Render(sidebarContent)

	var contentView string
	switch {
	case m.searchQuery != "":
		header := lipgloss.NewStyle().Foreground(accentColor).Bold(true).
			Render(fmt.Sprintf("🔍 Results for %q", m.searchQuery))
		if len(m.entries) > 0 {
			// The header takes two lines of the panel.
			contentView = header + "\n\n" + m.renderEntryList(m.listHeight()-2)
		} else {
			contentView = header + "\n\n" + lipgloss.NewStyle().Foreground(grayColor).Render("No matching entries")
		}
	case m.focus == focusEntries:
		if m.atRoot() {
			contentView = lipgloss.NewStyle().Foreground(grayColor).Render(
				"The database root holds no entries.\n\n[Tab] Back to groups")
		} else if len(m.entries) > 0 {
			// The path of the group tells where the entries belong.
			header := lipgloss.NewStyle().Foreground(accentColor).Bold(true).Render(m.currentPath())
			contentView = header + "\n\n" + m.renderEntryList(m.listHeight()-2)
		} else {
			contentView = lipgloss.NewStyle().Foreground(grayColor).Render(
				"No entries in this group.\n\n[n] New entry  │  [Tab] Back to groups")
		}
	default:
		if m.atRoot() {
			contentView = fmt.Sprintf("%s\n\nRoot of the database. %s.\n\n[n] New top level group  │  [/] Search in all groups",
				lipgloss.NewStyle().Foreground(accentColor).Bold(true).Render(m.rootName()),
				plural(len(m.db.Content.Root.Groups), "top level group", "top level groups"),
			)
		} else if len(m.groups) > 0 && m.groupCursor < len(m.groups) {
			group := m.groups[m.groupCursor]
			info := fmt.Sprintf("%s  %s\n\nEntries: %d\n\n[n] New  │  [r] Rename  │  [d] Del  │  [i] Icon\n[/] Search in all groups",
				iconEmoji(group.iconID),
				lipgloss.NewStyle().Foreground(accentColor).Bold(true).Render(group.path),
				len(m.entries),
			)
			contentView = info
		} else {
			contentView = lipgloss.NewStyle().Foreground(grayColor).Render("No groups found")
		}
	}

	panel := contentStyle.Width(contentWidth).Render(contentView)

	top := lipgloss.JoinHorizontal(lipgloss.Top, sidebarView, panel)

	return m.composeScreen(top)
}

// currentPath returns the path of the selected group.
func (m model) currentPath() string {
	if m.groupCursor < 0 || m.groupCursor >= len(m.groups) {
		return ""
	}
	return m.groups[m.groupCursor].path
}

// listHeight returns how many content lines fit inside the panels: the screen
// minus the top bar, the panel borders, the key hints and the status line.
func (m model) listHeight() int {
	h := m.height - 5
	if h < 1 {
		h = 1
	}
	return h
}

// visibleRange returns the window of items to draw so the cursor stays on
// screen. The window follows the cursor and never scrolls past the end.
func visibleRange(cursor, count, size int) (int, int) {
	if size < 1 {
		size = 1
	}
	if count <= size {
		return 0, count
	}
	start := cursor - size + 1
	if start > count-size {
		start = count - size
	}
	if start < 0 {
		start = 0
	}
	return start, start + size
}

// treePrefix returns the branch characters that precede a node, built from
// its ancestors, like the tree command of the terminal.
func treePrefix(groups []groupInfo, index int) string {
	var segs []string
	level := groups[index].depth
	for i := index - 1; i >= 0 && level > 1; i-- {
		if groups[i].depth == level-1 {
			if groups[i].last {
				segs = append(segs, "    ")
			} else {
				segs = append(segs, "│   ")
			}
			level--
		}
	}
	for i, j := 0, len(segs)-1; i < j; i, j = i+1, j-1 {
		segs[i], segs[j] = segs[j], segs[i]
	}
	return strings.Join(segs, "")
}

func (m model) renderSidebar() string {
	var lines []string
	// Every group takes two lines: its name and the blank separator.
	visible := m.listHeight() / 2
	start, end := visibleRange(m.groupCursor, len(m.groups), visible)
	for i := start; i < end; i++ {
		g := m.groups[i]
		// A blank line separates the top level branches, the subtrees stay
		// compact so the tree reads as a tree.
		if i > start && g.depth <= 1 {
			lines = append(lines, "")
		}

		label := g.name
		if g.depth > 0 {
			branch := treePrefix(m.groups, i)
			if g.last {
				branch += "└───"
			} else {
				branch += "├───"
			}
			label = branch + label
		}
		if g.path != "" {
			// The icon goes on the right of the title, in its own column.
			emoji := iconEmoji(g.iconID)
			pad := sidebarTextWidth - lipgloss.Width(label) - lipgloss.Width(emoji)
			if pad < 1 {
				pad = 1
			}
			label += strings.Repeat(" ", pad) + emoji
		}

		if i == m.groupCursor && m.focus == focusGroups {
			lines = append(lines, lipgloss.NewStyle().Foreground(accentColor).Bold(true).Render("▸ "+label))
		} else {
			lines = append(lines, "  "+label)
		}
	}
	return strings.Join(lines, "\n")
}

func (m model) renderEntryList(height int) string {
	var lines []string
	// Every entry takes two lines: its name and the blank separator.
	start, end := visibleRange(m.entryCursor, len(m.entries), height/2)
	for i := start; i < end; i++ {
		item := m.entries[i]
		if i > start {
			lines = append(lines, "")
		}
		title := getEntryValue(item.entry, "Title")
		if title == "" {
			title = "(untitled)"
		}
		emoji := iconEmoji(item.entry.IconID)
		label := emoji + " " + title
		if m.searchQuery != "" {
			label += "  " + lipgloss.NewStyle().Foreground(grayColor).Render(item.groupPath)
		}
		if i == m.entryCursor {
			lines = append(lines, lipgloss.NewStyle().Foreground(accentColor).Bold(true).Render("> "+label))
		} else {
			lines = append(lines, "  "+label)
		}
	}
	return strings.Join(lines, "\n")
}

// fitHints joins the hints that fit on a single line. The quit hint is
// always kept so the user can always leave the app.
func fitHints(width int, parts ...string) string {
	const sep = " │ "
	quit := "[q] Quit"
	limit := width - lipgloss.Width(sep+quit)

	line := ""
	for _, p := range parts {
		candidate := p
		if line != "" {
			candidate = line + sep + p
		}
		if lipgloss.Width(candidate) > limit {
			continue
		}
		line = candidate
	}
	if line == "" {
		return quit
	}
	return line + sep + quit
}

func (m model) renderFooter() string {
	var hints []string
	switch {
	case m.screen == screenBrowse && m.searchQuery != "":
		hints = []string{"[/] Command", "[↵] Details", "[Esc] Clear search", "[d] Del", "[i] Icon"}
	case m.screen == screenBrowse && m.focus == focusGroups:
		hints = []string{"[/] Command", "[Tab] Entries", "[n] New group", "[r] Rename", "[d] Del", "[i] Icon"}
	case m.screen == screenBrowse:
		hints = []string{"[/] Command", "[↵] Details", "[n] New entry", "[Tab] Groups", "[d] Del", "[i] Icon"}
	default:
		toggle := "[p] Show pass"
		if m.showEntryPass {
			toggle = "[p] Hide pass"
		}
		// The order matters: when the terminal is narrow, fitHints drops the
		// last ones, so the way out and the copy come first.
		hints = []string{"[/] Command", "[Esc] Back", "[c] Copy", "[e] Edit", "[i] Icon", toggle}
	}

	// The save hint shows up only when there is something to save.
	marker := ""
	if m.dirty {
		marker = "  ◆ unsaved"
		hints = append([]string{"[Ctrl+S] Save"}, hints...)
	}
	line := fitHints(m.width-4-lipgloss.Width(marker), hints...) + marker

	if m.dirty {
		return footerStyle.Width(m.width - 2).Render(
			lipgloss.NewStyle().Foreground(yellowColor).Render(line),
		)
	}
	return footerStyle.Width(m.width - 2).Render(line)
}

// renderCommandBar draws the command line at the top of the screen. It is
// empty when the command mode is not active.
func (m model) renderCommandBar() string {
	if m.mode != "command" {
		return ""
	}
	return footerStyle.Width(m.width - 2).Render(
		lipgloss.NewStyle().Foreground(accentColor).Render(m.command),
	)
}

// composeScreen puts the top bar on top and the content below it. The bar
// shows the command line while typing a command and the key hints otherwise.
// The status message stays at the bottom, where transient feedback belongs.
func (m model) composeScreen(content string) string {
	var b strings.Builder
	bar := m.renderCommandBar()
	if bar == "" {
		bar = m.renderFooter()
	}
	b.WriteString(bar + "\n")
	b.WriteString(content + "\n")
	if m.statusMsg != "" {
		b.WriteString(m.statusMsg)
	}
	return b.String()
}

func (m model) detailView() string {
	sidebarWidth := 28
	contentWidth := m.width - sidebarWidth - 6
	if contentWidth < 20 {
		contentWidth = 20
	}

	sidebarContent := m.renderSidebar()
	sidebarView := sidebarStyle.Width(sidebarWidth).Render(sidebarContent)

	var contentView string
	if m.detailEntry != nil {
		contentView = m.renderEntryDetail(m.detailEntry)
	} else {
		contentView = "No entry selected"
	}

	panel := contentStyle.Width(contentWidth).Render(contentView)

	top := lipgloss.JoinHorizontal(lipgloss.Top, sidebarView, panel)

	return m.composeScreen(top)
}

func (m model) renderEntryDetail(entry *gokeepasslib.Entry) string {
	if entry == nil {
		return ""
	}

	title := getEntryValue(entry, "Title")
	userName := getEntryValue(entry, "UserName")
	password := getEntryValue(entry, "Password")
	url := getEntryValue(entry, "URL")
	notes := getEntryValue(entry, "Notes")

	var passDisplay string
	if m.showEntryPass {
		passDisplay = password
	} else {
		passDisplay = strings.Repeat("*", len(password))
	}

	emoji := iconEmoji(entry.IconID)
	iconLabel := iconName(entry.IconID)

	var b strings.Builder
	b.WriteString(fmt.Sprintf("%s  %s\n\n", emoji, lipgloss.NewStyle().Foreground(accentColor).Bold(true).Render(title)))

	b.WriteString(lipgloss.NewStyle().Foreground(grayColor).Render("Username") + "\n")
	b.WriteString(userName + "\n\n")

	b.WriteString(lipgloss.NewStyle().Foreground(grayColor).Render("Password") + "\n")
	b.WriteString(passDisplay + "\n\n")

	b.WriteString(lipgloss.NewStyle().Foreground(grayColor).Render("URL") + "\n")
	b.WriteString(url + "\n\n")

	if notes != "" {
		b.WriteString(lipgloss.NewStyle().Foreground(grayColor).Render("Notes") + "\n")
		b.WriteString(notes + "\n\n")
	}

	b.WriteString(lipgloss.NewStyle().Foreground(grayColor).Render(fmt.Sprintf("Icon: %s %s (#%d)", emoji, iconLabel, entry.IconID)) + "\n\n")

	b.WriteString(lipgloss.NewStyle().Foreground(grayColor).Render("/copy"))

	return b.String()
}

func (m model) formView() string {
	var content string
	switch m.editMode {
	case editNewEntry, editEntry:
		label := "Edit Entry"
		if m.editMode == editNewEntry {
			label = "New Entry"
		}
		fields := []string{
			m.renderField("Title", &m.formTitle, 0),
			m.renderField("Username", &m.formUserName, 1),
			m.renderField("Password", &m.formPassword, 2),
			m.renderField("URL", &m.formURL, 3),
			m.renderField("Notes", &m.formNotes, 4),
			m.renderIconField(),
		}
		help := " [Tab/Shift+Tab or Arrows] Field  │  [Enter] Save  │  [Esc] Cancel "
		if m.formField == 5 {
			help = " [Enter] Pick icon  │  [Tab] Next field  │  [Esc] Cancel "
		}
		content = lipgloss.NewStyle().Foreground(accentColor).Bold(true).Render(label) + "\n\n" +
			strings.Join(fields, "\n") + "\n\n" +
			helpStyle.Render(help)

	case editNewGroup, editGroup:
		label := "Edit Group"
		if m.editMode == editNewGroup {
			label = "New Group"
		}
		content = lipgloss.NewStyle().Foreground(accentColor).Bold(true).Render(label) + "\n\n" +
			m.renderField("Name", &m.formGroup, 0) + "\n\n" +
			helpStyle.Render(" [Enter] Save  │  [Esc] Cancel ")

	case editNewDB:
		fields := []string{
			m.renderNewDBField("File", 0),
			m.renderNewDBField("Password", 1),
			m.renderNewDBField("Confirm", 2),
		}
		toggle := "[Ctrl+P] Show pass"
		if m.newDBShowPass {
			toggle = "[Ctrl+P] Hide pass"
		}
		content = lipgloss.NewStyle().Foreground(accentColor).Bold(true).Render("New Database") + "\n\n" +
			strings.Join(fields, "\n\n") + "\n\n" +
			helpStyle.Render(" [Tab/Shift+Tab or Arrows] Field  │  "+toggle+"  │  [Enter] Create  │  [Esc] Cancel ")
	}

	box := formStyle.Width(80).Render(content)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

func (m model) renderNewDBField(label string, idx int) string {
	l := labelStyle.Render(label + ":")
	input := &m.newDBInputs[idx]
	inp := input.View()
	if idx == m.newDBCursor {
		inp = lipgloss.NewStyle().Foreground(accentColor).Render(inp)
	}
	return l + " " + inp
}

func (m model) renderField(label string, input *textinput.Model, idx int) string {
	l := labelStyle.Render(label + ":")
	var inp string
	if m.editMode == editEntry || m.editMode == editNewEntry {
		if idx == m.formField {
			inp = lipgloss.NewStyle().Foreground(accentColor).Render(input.View())
		} else {
			inp = input.View()
		}
	} else {
		inp = input.View()
	}
	return l + " " + inp
}

func (m model) renderIconField() string {
	l := labelStyle.Render("Icon:")
	inp := iconEmoji(m.formIconID) + " " + iconName(m.formIconID)
	if m.formField == 5 {
		inp = lipgloss.NewStyle().Foreground(accentColor).Render(inp)
	}
	return l + " " + inp
}

const (
	sidebarWidth = 40
	// The room left for the group text: the sidebar width without its
	// border, its padding and the cursor marker.
	sidebarTextWidth = sidebarWidth - 6
)

// iconPickerCellWidth is the width of every icon cell in the picker.
const iconPickerCellWidth = 8

// iconPickerBoxWidth returns the width of the picker frame, capped so it does
// not get wider than the terminal.
func iconPickerBoxWidth(m model) int {
	w := m.width - 2
	if w > 100 {
		w = 100
	}
	if w < 20 {
		w = 20
	}
	return w
}

// iconPickerCols returns how many icons fit on a row. The view and the key
// handler must agree on it or the cursor jumps off the grid.
func (m model) iconPickerCols() int {
	// The frame, its padding and the row marker eat eight columns.
	cols := (iconPickerBoxWidth(m) - 8) / iconPickerCellWidth
	if cols < 1 {
		cols = 1
	}
	return cols
}

// selectedIcon returns the icon under the cursor.
func (m model) selectedIcon() iconInfo {
	if m.iconPickCursor < 0 || m.iconPickCursor >= len(icons) {
		return icons[0]
	}
	return icons[m.iconPickCursor]
}

func (m model) iconPickView() string {
	var b strings.Builder

	title := "Select Icon"
	switch m.iconPickTarget {
	case "group":
		title = "Select Icon for Group"
	case "form":
		title = "Select Icon for Entry"
	}
	b.WriteString(lipgloss.NewStyle().Foreground(accentColor).Bold(true).Render(title) + "\n\n")

	cols := m.iconPickerCols()

	// The grid scrolls when the terminal is too short for every row. The box
	// spends 10 lines on the title, the selected icon, the help and the frame.
	rows := (len(icons) + cols - 1) / cols
	avail := m.height - 10
	if avail < 1 {
		avail = 1
	}
	if avail > rows {
		avail = rows
	}
	firstRow := 0
	if row := m.iconPickCursor / cols; row >= avail {
		firstRow = row - avail + 1
	}
	if firstRow > rows-avail {
		firstRow = rows - avail
	}
	if firstRow < 0 {
		firstRow = 0
	}

	cell := lipgloss.NewStyle().Width(iconPickerCellWidth).Align(lipgloss.Center)
	for row := firstRow; row < firstRow+avail && row < rows; row++ {
		if m.iconPickCursor/cols == row {
			b.WriteString(lipgloss.NewStyle().Foreground(accentColor).Bold(true).Render("▸ "))
		} else {
			b.WriteString("  ")
		}
		for i := row * cols; i < (row+1)*cols && i < len(icons); i++ {
			if i == m.iconPickCursor {
				b.WriteString(cell.Background(accentColor).Foreground(textColor).Bold(true).Render(icons[i].Emoji))
			} else {
				b.WriteString(cell.Render(icons[i].Emoji))
			}
		}
		b.WriteString("\n")
	}

	selected := m.selectedIcon()
	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Foreground(grayColor).Render("Selected:  ") +
		selected.Emoji + " " +
		lipgloss.NewStyle().Foreground(accentColor).Bold(true).Render(selected.Name) + "\n\n")
	b.WriteString(helpStyle.Render(" [Arrow keys] Navigate  │  [Enter] Select  │  [Esc] Cancel "))

	box := formStyle.Width(iconPickerBoxWidth(m)).Render(b.String())
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

func (m model) confirmView() string {
	box := confirmStyle.Width(50).Render(
		lipgloss.NewStyle().Foreground(yellowColor).Bold(true).Render(m.confirmMsg),
	)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}
