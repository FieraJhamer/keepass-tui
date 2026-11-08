package main

import (
	"path/filepath"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/tobischo/gokeepasslib/v3"
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
	editNewDB
)

type groupInfo struct {
	name   string
	path   string
	depth  int
	iconID int64
}

// entryItem points at an entry inside the database. It remembers the group and
// the index so edits, icon changes and deletions reach the right entry even
// when the list is a search result from another group.
type entryItem struct {
	groupPath string
	index     int
	entry     *gokeepasslib.Entry
}

type model struct {
	width  int
	height int

	screen     screen
	ready      bool
	filePath   string
	files      []string
	fileCursor int

	password      string
	showPassword  bool
	showEntryPass bool
	unlockErr     string

	db     *gokeepasslib.Database
	groups []groupInfo

	groupCursor int
	focus       focus

	entries     []entryItem
	entryCursor int
	searchQuery string

	detailEntry *gokeepasslib.Entry

	mode    string // "normal" or "command"
	command string

	statusMsg string

	quit  bool
	dirty bool

	editMode     editMode
	formTitle    textinput.Model
	formUserName textinput.Model
	formPassword textinput.Model
	formURL      textinput.Model
	formNotes    textinput.Model
	formGroup    textinput.Model
	formField    int
	formIconID   int64

	newDBInputs   []textinput.Model
	newDBCursor   int
	newDBShowPass bool

	iconPickMode   bool
	iconPickTarget string
	iconPickCursor int

	confirmMode string
	confirmMsg  string
}

func initialModel() model {
	files := kdbxFiles()

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
	if len(files) > 0 {
		m.fileCursor = 1
	}
	return m
}

func kdbxFiles() []string {
	files, _ := filepath.Glob("*.kdbx")
	if files == nil {
		return []string{}
	}
	return files
}

func (m model) Init() tea.Cmd {
	return nil
}
