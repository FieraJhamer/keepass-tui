package main

import (
	"strings"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if !m.ready {
			m.ready = true
		}
		return m, nil

	case tea.KeyMsg:
		if m.quit {
			return m, tea.Quit
		}

		if m.confirmMode != "" {
			return m.handleConfirmKey(msg)
		}
		if m.iconPickMode {
			return m.handleIconPickKey(msg)
		}
		if m.editMode != editNone {
			return m.handleFormKey(msg)
		}

		switch m.screen {
		case screenPicker:
			return m.handlePickerKey(msg)
		case screenUnlock:
			return m.handleUnlockKey(msg)
		case screenBrowse:
			return m.handleBrowseKey(msg)
		case screenDetail:
			return m.handleDetailKey(msg)
		}
	}

	return m, nil
}

func (m model) handleConfirmKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.confirmMode {
	case "quit":
		switch msg.String() {
		case "s", "S":
			m.confirmMode = ""
			m.confirmMsg = ""
			return m.executeCommand("/save")
		case "d", "D":
			m.confirmMode = ""
			m.confirmMsg = ""
			m.quit = true
		case "c", "C", "esc":
			m.confirmMode = ""
			m.confirmMsg = ""
		}
	case "deleteEntry":
		switch msg.String() {
		case "y", "Y":
			if len(m.entries) > 0 && m.entryCursor < len(m.entries) {
				group := findGroup(m.db.Content.Root.Groups, m.groups[m.groupCursor].path)
				if group != nil {
					removeEntry(group, m.entryCursor)
					m.dirty = true
					m.statusMsg = "Entry deleted"
					m = m.loadEntries()
				}
			}
			m.confirmMode = ""
			m.confirmMsg = ""
		case "n", "N", "esc":
			m.confirmMode = ""
			m.confirmMsg = ""
		}
	case "deleteGroup":
		switch msg.String() {
		case "y", "Y":
			if m.groupCursor >= 0 && m.groupCursor < len(m.groups) {
				path := m.groups[m.groupCursor].path
				if removeGroupByPath(&m.db.Content.Root.Groups, path) {
					m.dirty = true
					m.statusMsg = "Group deleted"
					m.groups = flattenGroups(m.db.Content.Root.Groups, "", 0)
					if m.groupCursor >= len(m.groups) {
						m.groupCursor = len(m.groups) - 1
					}
					if m.groupCursor < 0 {
						m.groupCursor = 0
					}
					m = m.loadEntries()
				}
			}
			m.confirmMode = ""
			m.confirmMsg = ""
		case "n", "N", "esc":
			m.confirmMode = ""
			m.confirmMsg = ""
		}
	}
	return m, nil
}

func (m model) handleIconPickKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	cols := 8
	totalIcons := len(icons)
	switch msg.String() {
	case "esc":
		m.iconPickMode = false
	case "enter", "\r", "\n":
		id := int64(m.iconPickCursor)
		if m.iconPickTarget == "entry" {
			if m.screen == screenDetail && m.detailEntry != nil {
				m.detailEntry.IconID = id
				m.dirty = true
				m.statusMsg = iconEmoji(id) + " Icon updated"
			} else if m.focus == focusEntries && len(m.entries) > 0 {
				m.entries[m.entryCursor].IconID = id
				m.dirty = true
				m.statusMsg = iconEmoji(id) + " Icon updated"
			}
		} else if m.iconPickTarget == "group" && m.groupCursor < len(m.groups) {
			path := m.groups[m.groupCursor].path
			g := findGroup(m.db.Content.Root.Groups, path)
			if g != nil {
				g.IconID = id
				m.groups[m.groupCursor].iconID = id
				m.dirty = true
				m.statusMsg = iconEmoji(id) + " Icon updated"
			}
		}
		m.iconPickMode = false
	case "up", "k", "K":
		if m.iconPickCursor-cols >= 0 {
			m.iconPickCursor -= cols
		}
	case "down", "j", "J":
		if m.iconPickCursor+cols < totalIcons {
			m.iconPickCursor += cols
		}
	case "left", "h", "H":
		if m.iconPickCursor > 0 {
			m.iconPickCursor--
		}
	case "right", "l", "L":
		if m.iconPickCursor < totalIcons-1 {
			m.iconPickCursor++
		}
	case "tab":
		m.iconPickCursor = (m.iconPickCursor + 1) % totalIcons
	}
	return m, nil
}

func (m model) handleFormKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch m.editMode {
	case editNewEntry, editEntry:
		switch msg.String() {
		case "esc":
			m.editMode = editNone
			m.statusMsg = "Cancelled"
		case "tab", "down":
			m.formField = (m.formField + 1) % 5
		case "shift+tab", "up":
			m.formField = (m.formField - 1 + 5) % 5
		case "enter", "\r", "\n":
			title := m.formTitle.Value()
			userName := m.formUserName.Value()
			password := m.formPassword.Value()
			url := m.formURL.Value()
			notes := m.formNotes.Value()

			if title == "" {
				m.statusMsg = "Title cannot be empty"
				return m, nil
			}

			if m.editMode == editNewEntry {
				e := newEntry(title, userName, password, url, notes, 0)
				group := findGroup(m.db.Content.Root.Groups, m.groups[m.groupCursor].path)
				if group != nil {
					group.Entries = append(group.Entries, e)
					m.dirty = true
					m.statusMsg = "Entry created"
					m = m.loadEntries()
				}
			} else {
				if len(m.entries) > 0 && m.entryCursor < len(m.entries) {
					e := &m.entries[m.entryCursor]
					setEntryValue(e, "Title", title)
					setEntryValue(e, "UserName", userName)
					setEntryValue(e, "Password", password)
					setEntryValue(e, "URL", url)
					setEntryValue(e, "Notes", notes)
					m.dirty = true
					m.statusMsg = "Entry saved"
				}
			}
			m.editMode = editNone

		default:
			inputs := []*textinput.Model{&m.formTitle, &m.formUserName, &m.formPassword, &m.formURL, &m.formNotes}
			for i := range inputs {
				if i == m.formField {
					inputs[i].Focus()
				} else {
					inputs[i].Blur()
				}
			}
			m.formTitle, cmd = m.formTitle.Update(msg)
			m.formUserName, cmd = m.formUserName.Update(msg)
			m.formPassword, cmd = m.formPassword.Update(msg)
			m.formURL, cmd = m.formURL.Update(msg)
			m.formNotes, cmd = m.formNotes.Update(msg)
		}

	case editNewGroup, editGroup:
		switch msg.String() {
		case "esc":
			m.editMode = editNone
			m.statusMsg = "Cancelled"
		case "enter", "\r", "\n":
			name := m.formGroup.Value()
			if name == "" {
				m.statusMsg = "Name cannot be empty"
				return m, nil
			}
			if m.editMode == editNewGroup {
				ng := newGroup(name, 0)
				path := ""
				if len(m.groups) > 0 {
					path = m.groups[m.groupCursor].path
				}
				if addGroupByPath(&m.db.Content.Root.Groups, path, ng) {
					m.dirty = true
					m.groups = flattenGroups(m.db.Content.Root.Groups, "", 0)
					m.statusMsg = "Group created"
					m = m.loadEntries()
				}
			} else {
				if m.groupCursor < len(m.groups) {
					g := findGroup(m.db.Content.Root.Groups, m.groups[m.groupCursor].path)
					if g != nil {
						g.Name = name
						m.groups[m.groupCursor].name = name
						m.dirty = true
						m.statusMsg = "Group renamed"
					}
				}
			}
			m.editMode = editNone
		default:
			m.formGroup.Focus()
			m.formGroup, cmd = m.formGroup.Update(msg)
		}
	}

	return m, cmd
}

func (m model) handlePickerKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q", "Q", "esc":
		return m, tea.Quit
	case "up", "k", "K":
		if m.fileCursor > 0 {
			m.fileCursor--
		}
	case "down", "j", "J":
		if m.fileCursor < len(m.files)-1 {
			m.fileCursor++
		}
	case "enter", "\r", "\n":
		if len(m.files) > 0 && m.fileCursor >= 0 && m.fileCursor < len(m.files) {
			m.filePath = m.files[m.fileCursor]
			m.screen = screenUnlock
			m.password = ""
			m.showPassword = false
			m.unlockErr = ""
		}
	}
	return m, nil
}

func (m model) handleUnlockKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q", "Q":
		return m, tea.Quit
	case "esc":
		m.password = ""
		m.showPassword = false
		m.unlockErr = ""
		m.screen = screenPicker
	case "enter", "\r", "\n":
		db, err := openDatabase(m.filePath, m.password)
		if err != nil {
			m.unlockErr = err.Error()
			m.password = ""
			return m, nil
		}
		m.db = db
		m.groups = flattenGroups(m.db.Content.Root.Groups, "", 0)
		m.screen = screenBrowse
		m.unlockErr = ""
		m.password = ""
		m.groupCursor = 0
		m.entryCursor = 0
		m.focus = focusGroups
		m.mode = "normal"
		m = m.loadEntries()
		return m, nil
	case "backspace":
		if len(m.password) > 0 {
			m.password = m.password[:len(m.password)-1]
		}
	case "ctrl+p":
		m.showPassword = !m.showPassword
	default:
		if msg.Type == tea.KeyRunes && len(msg.Runes) > 0 {
			for _, r := range msg.Runes {
				if r >= 32 && r <= 126 {
					m.password += string(r)
				}
			}
		}
	}
	return m, nil
}

func (m model) handleBrowseKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.mode == "command" {
		return m.handleCommandKey(msg)
	}

	k := msg.String()
	switch k {
	case "ctrl+c", "q", "Q":
		if m.dirty {
			m.confirmMode = "quit"
			m.confirmMsg = "Unsaved changes. (S)ave & quit / (D)iscard / (C)ancel"
			return m, nil
		}
		return m, tea.Quit
	case "up", "k", "K":
		if m.focus == focusGroups {
			if m.groupCursor > 0 {
				m.groupCursor--
				m = m.loadEntries()
			}
		} else {
			if m.entryCursor > 0 {
				m.entryCursor--
			}
		}
	case "down", "j", "J":
		if m.focus == focusGroups {
			if m.groupCursor < len(m.groups)-1 {
				m.groupCursor++
				m = m.loadEntries()
			}
		} else {
			if m.entryCursor < len(m.entries)-1 {
				m.entryCursor++
			}
		}
	case "tab":
		if m.focus == focusGroups && len(m.entries) > 0 {
			m.focus = focusEntries
		} else {
			m.focus = focusGroups
		}
	case "enter", "\r", "\n":
		if m.focus == focusEntries && len(m.entries) > 0 {
			m.detailEntry = &m.entries[m.entryCursor]
			m.screen = screenDetail
		}
	case "/":
		m.mode = "command"
		m.command = "/"
		m.statusMsg = ""
	case "n", "N":
		if m.focus == focusEntries || (m.focus == focusGroups && len(m.groups) > 0) {
			m = m.initEntryForm(editNewEntry, "", "", "", "", "")
			m.editMode = editNewEntry
		}
	case "d", "D":
		if m.focus == focusEntries && len(m.entries) > 0 {
			title := getEntryValue(&m.entries[m.entryCursor], "Title")
			if title == "" {
				title = "untitled"
			}
			m.confirmMode = "deleteEntry"
			m.confirmMsg = "Delete entry '" + title + "'? (Y)es / (N)o"
		} else if m.focus == focusGroups && len(m.groups) > 0 {
			name := m.groups[m.groupCursor].name
			m.confirmMode = "deleteGroup"
			m.confirmMsg = "Delete group '" + name + "'? (Y)es / (N)o"
		}
	case "i", "I":
		m = m.initIconPicker()
	}
	return m, nil
}

func (m model) handleCommandKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = "normal"
		m.command = ""
	case "enter", "\r", "\n":
		cmd := m.command
		m.mode = "normal"
		m.command = ""
		return m.executeCommand(cmd)
	case "backspace":
		if len(m.command) > 1 {
			m.command = m.command[:len(m.command)-1]
		} else {
			m.mode = "normal"
			m.command = ""
		}
	default:
		if msg.Type == tea.KeyRunes && len(msg.Runes) > 0 {
			for _, r := range msg.Runes {
				if r >= 32 && r <= 126 {
					m.command += string(r)
				}
			}
		}
	}
	return m, nil
}

func (m model) handleDetailKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.mode == "command" {
		return m.handleCommandKey(msg)
	}

	k := msg.String()
	switch k {
	case "ctrl+c", "q", "Q":
		if m.dirty {
			m.confirmMode = "quit"
			m.confirmMsg = "Unsaved changes. (S)ave & quit / (D)iscard / (C)ancel"
			return m, nil
		}
		return m, tea.Quit
	case "esc":
		m.screen = screenBrowse
		m.detailEntry = nil
	case "/":
		m.mode = "command"
		m.command = "/"
		m.statusMsg = ""
	case "e", "E":
		if m.detailEntry != nil {
			m = m.initEntryForm(editEntry,
				getEntryValue(m.detailEntry, "Title"),
				getEntryValue(m.detailEntry, "UserName"),
				getEntryValue(m.detailEntry, "Password"),
				getEntryValue(m.detailEntry, "URL"),
				getEntryValue(m.detailEntry, "Notes"),
			)
			m.editMode = editEntry
		}
	case "i", "I":
		m = m.initIconPicker()
	}
	return m, nil
}

func (m model) loadEntries() model {
	if m.groupCursor >= len(m.groups) {
		m.groupCursor = len(m.groups) - 1
	}
	if m.groupCursor < 0 || len(m.groups) == 0 {
		m.entries = nil
		return m
	}
	group := findGroup(m.db.Content.Root.Groups, m.groups[m.groupCursor].path)
	if group != nil {
		m.entries = group.Entries
	} else {
		m.entries = nil
	}
	m.entryCursor = 0
	return m
}

func (m model) initEntryForm(mode editMode, title, userName, password, url, notes string) model {
	ti := textinput.New()
	ti.Placeholder = "Title"
	ti.Focus()
	ti.SetValue(title)
	ti.CharLimit = 256
	ti.Width = 40

	ui := textinput.New()
	ui.Placeholder = "Username"
	ui.SetValue(userName)
	ui.CharLimit = 256
	ui.Width = 40

	pi := textinput.New()
	pi.Placeholder = "Password"
	pi.SetValue(password)
	pi.CharLimit = 256
	pi.Width = 40
	pi.EchoMode = textinput.EchoPassword

	uri := textinput.New()
	uri.Placeholder = "URL"
	uri.SetValue(url)
	uri.CharLimit = 512
	uri.Width = 40

	ni := textinput.New()
	ni.Placeholder = "Notes"
	ni.SetValue(notes)
	ni.CharLimit = 1024
	ni.Width = 40

	m.formTitle = ti
	m.formUserName = ui
	m.formPassword = pi
	m.formURL = uri
	m.formNotes = ni
	m.formField = 0
	m.formTitle.Focus()
	m.formUserName.Blur()
	m.formPassword.Blur()
	m.formURL.Blur()
	m.formNotes.Blur()
	return m
}

func (m model) initIconPicker() model {
	m.iconPickMode = true
	m.iconPickCursor = 0
	if m.screen == screenDetail && m.detailEntry != nil {
		m.iconPickTarget = "entry"
		m.iconPickCursor = int(m.detailEntry.IconID)
	} else if m.focus == focusEntries && len(m.entries) > 0 {
		m.iconPickTarget = "entry"
		m.iconPickCursor = int(m.entries[m.entryCursor].IconID)
	} else if m.focus == focusGroups && m.groupCursor < len(m.groups) {
		m.iconPickTarget = "group"
		m.iconPickCursor = int(m.groups[m.groupCursor].iconID)
	}
	return m
}

func (m model) executeCommand(cmd string) (tea.Model, tea.Cmd) {
	cmd = strings.TrimSpace(cmd)
	switch {
	case cmd == "/copy-pass":
		if len(m.entries) > 0 && m.entryCursor < len(m.entries) {
			val := getEntryValue(&m.entries[m.entryCursor], "Password")
			if err := clipboard.WriteAll(val); err != nil {
				m.statusMsg = "Clipboard error: " + err.Error()
			} else {
				m.statusMsg = "✓ Password copied to clipboard"
			}
		}
	case cmd == "/copy-user":
		if len(m.entries) > 0 && m.entryCursor < len(m.entries) {
			val := getEntryValue(&m.entries[m.entryCursor], "UserName")
			if err := clipboard.WriteAll(val); err != nil {
				m.statusMsg = "Clipboard error: " + err.Error()
			} else {
				m.statusMsg = "✓ Username copied to clipboard"
			}
		}
	case cmd == "/copy-url":
		if len(m.entries) > 0 && m.entryCursor < len(m.entries) {
			val := getEntryValue(&m.entries[m.entryCursor], "URL")
			if err := clipboard.WriteAll(val); err != nil {
				m.statusMsg = "Clipboard error: " + err.Error()
			} else {
				m.statusMsg = "✓ URL copied to clipboard"
			}
		}
	case cmd == "/copy-all":
		var sb strings.Builder
		if len(m.entries) > 0 && m.entryCursor < len(m.entries) {
			e := &m.entries[m.entryCursor]
			sb.WriteString("Title: ")
			sb.WriteString(getEntryValue(e, "Title"))
			sb.WriteString("\nUsername: ")
			sb.WriteString(getEntryValue(e, "UserName"))
			sb.WriteString("\nPassword: ")
			sb.WriteString(getEntryValue(e, "Password"))
			sb.WriteString("\nURL: ")
			sb.WriteString(getEntryValue(e, "URL"))
			sb.WriteString("\nNotes: ")
			sb.WriteString(getEntryValue(e, "Notes"))
			if err := clipboard.WriteAll(sb.String()); err != nil {
				m.statusMsg = "Clipboard error: " + err.Error()
			} else {
				m.statusMsg = "✓ Entry copied to clipboard"
			}
		}

	case cmd == "/save":
		err := saveDatabase(m.db, m.filePath)
		if err != nil {
			m.statusMsg = "Save error: " + err.Error()
		} else {
			m.dirty = false
			m.statusMsg = "✓ Database saved"
			if m.quit {
				return m, tea.Quit
			}
		}

	case cmd == "/new-entry":
		if len(m.groups) > 0 {
			m = m.initEntryForm(editNewEntry, "", "", "", "", "")
			m.editMode = editNewEntry
		} else {
			m.statusMsg = "No groups available"
		}

	case cmd == "/new-group":
		m.formGroup = textinput.New()
		m.formGroup.Placeholder = "Group name"
		m.formGroup.CharLimit = 128
		m.formGroup.Width = 30
		m.formGroup.Focus()
		m.editMode = editNewGroup

	case cmd == "/edit":
		if m.screen == screenDetail && m.detailEntry != nil {
			m = m.initEntryForm(editEntry,
				getEntryValue(m.detailEntry, "Title"),
				getEntryValue(m.detailEntry, "UserName"),
				getEntryValue(m.detailEntry, "Password"),
				getEntryValue(m.detailEntry, "URL"),
				getEntryValue(m.detailEntry, "Notes"),
			)
			m.editMode = editEntry
		} else if m.focus == focusEntries && len(m.entries) > 0 {
			m = m.initEntryForm(editEntry,
				getEntryValue(&m.entries[m.entryCursor], "Title"),
				getEntryValue(&m.entries[m.entryCursor], "UserName"),
				getEntryValue(&m.entries[m.entryCursor], "Password"),
				getEntryValue(&m.entries[m.entryCursor], "URL"),
				getEntryValue(&m.entries[m.entryCursor], "Notes"),
			)
			m.editMode = editEntry
		} else {
			m.statusMsg = "No entry selected"
		}

	case cmd == "/delete":
		if m.focus == focusEntries && len(m.entries) > 0 {
			title := getEntryValue(&m.entries[m.entryCursor], "Title")
			if title == "" {
				title = "untitled"
			}
			m.confirmMode = "deleteEntry"
			m.confirmMsg = "Delete entry '" + title + "'? (Y)es / (N)o"
		} else if m.focus == focusGroups && len(m.groups) > 0 {
			name := m.groups[m.groupCursor].name
			m.confirmMode = "deleteGroup"
			m.confirmMsg = "Delete group '" + name + "'? (Y)es / (N)o"
		} else {
			m.statusMsg = "Nothing to delete"
		}

	case cmd == "/icon":
		m = m.initIconPicker()

	case cmd == "/lock":
		m.db = nil
		m.groups = nil
		m.entries = nil
		m.detailEntry = nil
		m.password = ""
		m.unlockErr = ""
		m.screen = screenUnlock
		m.mode = "normal"
		m.command = ""
		m.dirty = false

	case cmd == "/quit":
		if m.dirty {
			m.confirmMode = "quit"
			m.confirmMsg = "Unsaved changes. (S)ave & quit / (D)iscard / (C)ancel"
			return m, nil
		}
		m.quit = true

	default:
		m.statusMsg = "Unknown command: " + cmd
	}
	return m, nil
}
