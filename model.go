package main

import (
	"path/filepath"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/tobischo/gokeepasslib/v3"
	tea "github.com/charmbracelet/bubbletea"
)

type screen int

const (
	screenPicker screen = iota
	screenUnlock
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
	filePath   string
	files      []string
	fileCursor int

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
	files, _ := filepath.Glob("*.kdbx")
	if len(files) == 0 {
		files = []string{}
	}

	m := model{
		screen:      screenPicker,
		files:       files,
		fileCursor:  0,
		mode:        "normal",
		groupCursor: 0,
		entryCursor: 0,
		focus:       focusGroups,
		ready:       true,
	}
	return m
}



func (m model) Init() tea.Cmd {
	return nil
}
