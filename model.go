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
	name string
	// path is only used to show where the group is, never to find it: two
	// groups can have the same name and a name can hold a slash.
	path   string
	uuid   gokeepasslib.UUID
	depth  int
	iconID int64
	last   bool // whether it is the last child of its parent
}

// entryItem points at an entry inside the database. It remembers the group and
// the index so edits, icon changes and deletions reach the right entry even
// when the list is a search result from another group.
type entryItem struct {
	groupPath string            // only to show where the entry comes from
	groupUUID gokeepasslib.UUID // the group the entry belongs to
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

	showPassword  bool
	showEntryPass bool
	unlockErr     string
	unlockInput   textinput.Model

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

	quit  bool // close the program after the next successful save
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
		unlockInput: newPasswordInput(),
	}
	if len(files) > 0 {
		m.fileCursor = 1
	}
	return m
}

// newPasswordInput builds the masked master password field.
func newPasswordInput() textinput.Model {
	ti := textinput.New()
	ti.Placeholder = "Master password"
	ti.CharLimit = 256
	ti.Width = 40
	ti.EchoMode = textinput.EchoPassword
	ti.Focus()
	return ti
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
