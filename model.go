package main

import (
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/tobischo/gokeepasslib/v3"
	tea "github.com/charmbracelet/bubbletea"
)

type screen int

const (
	screenUnlock screen = iota
	screenBrowse
	screenDetail
)

type focus int

const (
	focusGroups focus = iota
	focusEntries
)

type editMode int

const (
	editNone editMode = iota
	editNewEntry
	editEntry
	editNewGroup
	editGroup
)

type groupInfo struct {
	name   string
	path   string
	depth  int
	iconID int64
}

type model struct {
	width  int
	height int

	screen   screen
	ready    bool
	filePath string

	password     string
	showPassword bool
	unlockErr    string

	db     *gokeepasslib.Database
	groups []groupInfo

	groupCursor int
	focus       focus

	entries     []gokeepasslib.Entry
	entryCursor int

	detailEntry *gokeepasslib.Entry

	mode    string // "normal" or "command"
	command string

	statusMsg string

	quit bool
	dirty bool

	editMode    editMode
	formTitle    textinput.Model
	formUserName textinput.Model
	formPassword textinput.Model
	formURL      textinput.Model
	formNotes    textinput.Model
	formGroup    textinput.Model
	formField    int

	iconPickMode   bool
	iconPickTarget string
	iconPickCursor int

	confirmMode string
	confirmMsg  string
}

func initialModel() model {
	return model{
		screen:      screenUnlock,
		filePath:    "Database.kdbx",
		mode:        "normal",
		groupCursor: 0,
		entryCursor: 0,
		focus:       focusGroups,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}
