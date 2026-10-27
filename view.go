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

	var pwDisplay string
	if m.showPassword {
		pwDisplay = m.password
	} else {
		pwDisplay = strings.Repeat("*", len(m.password))
	}
	if len(m.password) == 0 {
		pwDisplay = ""
	}

	content := fmt.Sprintf(
		"%s\n\nPassword: %s",
		lipgloss.NewStyle().Foreground(accentColor).Bold(true).Render("KeePass TUI"),
		lipgloss.NewStyle().Foreground(textColor).Render(pwDisplay),
	)

	if m.unlockErr != "" {
		content += "\n\n" + lipgloss.NewStyle().Foreground(redColor).Render(m.unlockErr)
	}

	box := unlockStyle.Width(width).Render(content)

	help := lipgloss.NewStyle().Foreground(grayColor).Render(" [Enter] Unlock  │  [Ctrl+P] Toggle  │  [Esc] Quit ")
	box = lipgloss.JoinVertical(lipgloss.Center, box, help)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

func (m model) browseView() string {
	sidebarWidth := 40
	contentWidth := m.width - sidebarWidth - 6
	if contentWidth < 20 {
		contentWidth = 20
	}

	sidebarContent := m.renderSidebar()
	sidebarView := sidebarStyle.Width(sidebarWidth).Render(sidebarContent)

	var contentView string
	if m.focus == focusEntries && len(m.entries) > 0 {
		contentView = m.renderEntryList()
	} else {
		if len(m.groups) > 0 && m.groupCursor < len(m.groups) {
			group := m.groups[m.groupCursor]
			info := fmt.Sprintf("%s  %s\n\nEntries: %d\n\n[d] Delete group  │  [i] Icon",
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

	footer := m.renderFooter()

	return top + "\n" + footer + "\n" + m.statusMsg
}

func (m model) renderSidebar() string {
	var b strings.Builder
	for i, g := range m.groups {
		prefix := strings.Repeat("  ", g.depth)
		emoji := iconEmoji(g.iconID)
		if i == m.groupCursor && m.focus == focusGroups {
			b.WriteString(lipgloss.NewStyle().Foreground(accentColor).Bold(true).Render("> "+emoji+" "+prefix+g.name) + "\n")
		} else {
			b.WriteString(lipgloss.NewStyle().Foreground(textColor).Render("  "+emoji+" "+prefix+g.name) + "\n")
		}
	}
	return b.String()
}

func (m model) renderEntryList() string {
	var b strings.Builder
	for i, entry := range m.entries {
		title := getEntryValue(&entry, "Title")
		if title == "" {
			title = "(untitled)"
		}
		emoji := iconEmoji(entry.IconID)
		if i == m.entryCursor {
			b.WriteString(lipgloss.NewStyle().Foreground(accentColor).Bold(true).Render("> "+emoji+" "+title) + "\n")
		} else {
			b.WriteString(fmt.Sprintf("  %s %s\n", emoji, title))
		}
	}
	return b.String()
}

func (m model) renderFooter() string {
	if m.mode == "command" {
		return footerStyle.Width(m.width - 2).Render(
			lipgloss.NewStyle().Foreground(accentColor).Render(m.command),
		)
	}

	var hints string
	switch {
	case m.screen == screenBrowse && m.focus == focusGroups:
		hints = "[/] Command  │  [Enter] Open  │  [n] New  │  [r] Rename  │  [d] Del  │  [i] Icon  │  [q] Quit"
	case m.screen == screenBrowse:
		hints = "[/] Command  │  [Enter] Details  │  [Esc] Back  │  [d] Del  │  [i] Icon  │  [q] Quit"
	default:
		toggle := "[p] Show pass"
		if m.showEntryPass {
			toggle = "[p] Hide pass"
		}
		hints = "[/] Command  │  [e] Edit  │  [i] Icon  │  " + toggle + "  │  [Esc] Back  │  [q] Quit"
	}

	if m.dirty {
		hints += "  ◆ unsaved"
		return footerStyle.Width(m.width - 2).Render(
			lipgloss.NewStyle().Foreground(yellowColor).Render(hints),
		)
	}
	return footerStyle.Width(m.width - 2).Render(hints)
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

	footer := m.renderFooter()

	return top + "\n" + footer + "\n" + m.statusMsg
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

	b.WriteString(lipgloss.NewStyle().Foreground(grayColor).Render("/copy-pass  /copy-user  /copy-url  /copy-all"))

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
		}
		content = lipgloss.NewStyle().Foreground(accentColor).Bold(true).Render(label) + "\n\n" +
			strings.Join(fields, "\n") + "\n\n" +
			helpStyle.Render(" [Tab/Shift+Tab or Arrows] Field  │  [Enter] Save  │  [Esc] Cancel ")

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

func (m model) iconPickView() string {
	cols := 11
	var b strings.Builder

	b.WriteString(lipgloss.NewStyle().Foreground(accentColor).Bold(true).Render("Select Icon") + "\n\n")

	for i, icon := range icons {
		if i > 0 && i%cols == 0 {
			b.WriteString("\n")
		}
		cell := fmt.Sprintf(" %s%02d ", icon.Emoji, icon.ID)
		if i == m.iconPickCursor {
			b.WriteString(lipgloss.NewStyle().Foreground(accentColor).Bold(true).Render(">" + cell))
		} else {
			b.WriteString(" " + cell)
		}
	}

	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render(" [Arrow keys] Navigate  │  [Enter] Select  │  [Esc] Cancel "))

	box := formStyle.Width(100).Render(b.String())
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

func (m model) confirmView() string {
	box := confirmStyle.Width(50).Render(
		lipgloss.NewStyle().Foreground(yellowColor).Bold(true).Render(m.confirmMsg),
	)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}
