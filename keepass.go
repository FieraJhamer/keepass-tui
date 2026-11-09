package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tobischo/gokeepasslib/v3"
	w "github.com/tobischo/gokeepasslib/v3/wrappers"
)

func openDatabase(path, password string) (*gokeepasslib.Database, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	db := gokeepasslib.NewDatabase()
	db.Credentials = gokeepasslib.NewPasswordCredentials(password)

	err = gokeepasslib.NewDecoder(f).Decode(db)
	if err != nil {
		return nil, err
	}

	if err := db.UnlockProtectedEntries(); err != nil {
		return nil, err
	}
	return db, nil
}

// saveDatabase writes to a temp file and renames it over the target, so a failed
// or interrupted encode never truncates the existing database.
func saveDatabase(db *gokeepasslib.Database, path string) error {
	db.LockProtectedEntries()
	defer db.UnlockProtectedEntries()

	tmp := path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}

	err = gokeepasslib.NewEncoder(f).Encode(db)
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		os.Remove(tmp)
		return err
	}
	if err = os.Rename(tmp, path); err != nil {
		os.Remove(tmp)
		return err
	}
	return nil
}

// createDatabase writes a fresh KDBX 4 database with a single empty group and
// returns it opened and unlocked.
func createDatabase(path, masterPassword string) (*gokeepasslib.Database, error) {
	db := gokeepasslib.NewDatabase(gokeepasslib.WithDatabaseKDBXVersion4())
	db.Credentials = gokeepasslib.NewPasswordCredentials(masterPassword)
	db.Content.Meta.DatabaseName = strings.TrimSuffix(filepath.Base(path), ".kdbx")
	db.Content.Root.Groups = []gokeepasslib.Group{newGroup("General", 5)}

	if err := saveDatabase(db, path); err != nil {
		return nil, err
	}
	return openDatabase(path, masterPassword)
}

// deleteGroupPrompt warns how much is inside the group before deleting it.
func deleteGroupPrompt(name string, g *gokeepasslib.Group) string {
	msg := "Delete group '" + name + "'?"
	if g != nil {
		groups, entries := countContents(g)
		if groups > 0 || entries > 0 {
			msg += fmt.Sprintf(" It contains %s and %s.",
				plural(groups, "group", "groups"), plural(entries, "entry", "entries"))
		}
	}
	return msg + " (Y)es / (N)o"
}

// plural writes a count with its noun in the right number.
func plural(n int, singular, many string) string {
	if n == 1 {
		return fmt.Sprintf("%d %s", n, singular)
	}
	return fmt.Sprintf("%d %s", n, many)
}

// countContents returns how many groups and entries a group holds, including
// the ones in its subgroups.
func countContents(g *gokeepasslib.Group) (groups, entries int) {
	entries = len(g.Entries)
	for i := range g.Groups {
		subGroups, subEntries := countContents(&g.Groups[i])
		groups += 1 + subGroups
		entries += subEntries
	}
	return groups, entries
}

// shortcut: group paths are "/"-joined, so a group whose name contains "/" breaks
// lookup; switch to UUID-based paths if group names ever need slashes.
func flattenGroups(groups []gokeepasslib.Group, prefix string, depth int) []groupInfo {
	var result []groupInfo
	for i := range groups {
		g := &groups[i]
		path := g.Name
		if prefix != "" {
			path = prefix + "/" + g.Name
		}
		result = append(result, groupInfo{
			name:   g.Name,
			path:   path,
			depth:  depth,
			iconID: g.IconID,
			last:   i == len(groups)-1,
		})
		result = append(result, flattenGroups(g.Groups, path, depth+1)...)
	}
	return result
}

// groupTree returns the sidebar list: the database root followed by every
// group. The root is synthetic, it only exists to draw the tree and to create
// top level groups.
func (m model) groupTree() []groupInfo {
	root := groupInfo{name: m.rootName(), path: "", depth: 0, last: true}
	if m.db == nil {
		return []groupInfo{root}
	}
	tree := flattenGroups(m.db.Content.Root.Groups, "", 1)
	return append([]groupInfo{root}, tree...)
}

// rootName is the label of the synthetic root node.
func (m model) rootName() string {
	if m.db != nil && m.db.Content.Meta.DatabaseName != "" {
		return m.db.Content.Meta.DatabaseName
	}
	if m.filePath != "" {
		return strings.TrimSuffix(filepath.Base(m.filePath), ".kdbx")
	}
	return "Database"
}

// atRoot tells if the selected node is the synthetic database root.
func (m model) atRoot() bool {
	return m.groupCursor >= 0 && m.groupCursor < len(m.groups) && m.groups[m.groupCursor].path == ""
}

func findGroup(groups []gokeepasslib.Group, path string) *gokeepasslib.Group {
	if path == "" {
		return nil
	}
	parts := strings.Split(path, "/")
	currentGroups := groups
	var current *gokeepasslib.Group
	for _, part := range parts {
		found := false
		for i := range currentGroups {
			if currentGroups[i].Name == part {
				current = &currentGroups[i]
				currentGroups = current.Groups
				found = true
				break
			}
		}
		if !found {
			return nil
		}
	}
	return current
}

func getEntryValue(entry *gokeepasslib.Entry, key string) string {
	for _, v := range entry.Values {
		if v.Key == key {
			return v.Value.Content
		}
	}
	return ""
}

// searchItems returns every entry in the database whose title, username or URL
// contains the query, in group order.
func searchItems(groups []gokeepasslib.Group, query string) []entryItem {
	q := strings.ToLower(query)
	var items []entryItem

	var walk func(gs []gokeepasslib.Group, prefix string)
	walk = func(gs []gokeepasslib.Group, prefix string) {
		for i := range gs {
			g := &gs[i]
			path := g.Name
			if prefix != "" {
				path = prefix + "/" + g.Name
			}
			for j := range g.Entries {
				e := &g.Entries[j]
				if entryMatches(e, q) {
					items = append(items, entryItem{groupPath: path, index: j, entry: e})
				}
			}
			walk(g.Groups, path)
		}
	}
	walk(groups, "")
	return items
}

func entryMatches(entry *gokeepasslib.Entry, lowerQuery string) bool {
	for _, key := range []string{"Title", "UserName", "URL"} {
		if strings.Contains(strings.ToLower(getEntryValue(entry, key)), lowerQuery) {
			return true
		}
	}
	return false
}

func setEntryValue(entry *gokeepasslib.Entry, key, value string) {
	for i := range entry.Values {
		if entry.Values[i].Key == key {
			entry.Values[i].Value.Content = value
			if key == "Password" {
				entry.Values[i].Value.Protected = w.NewBoolWrapper(true)
			}
			return
		}
	}
	vd := gokeepasslib.ValueData{
		Key:   key,
		Value: gokeepasslib.V{Content: value},
	}
	if key == "Password" {
		vd.Value.Protected = w.NewBoolWrapper(true)
	}
	entry.Values = append(entry.Values, vd)
}

func newEntry(title, userName, password, url, notes string, iconID int64) gokeepasslib.Entry {
	e := gokeepasslib.NewEntry()
	e.IconID = iconID
	e.Values = []gokeepasslib.ValueData{
		{Key: "Title", Value: gokeepasslib.V{Content: title}},
		{Key: "UserName", Value: gokeepasslib.V{Content: userName}},
		{Key: "Password", Value: gokeepasslib.V{Content: password, Protected: w.NewBoolWrapper(true)}},
		{Key: "URL", Value: gokeepasslib.V{Content: url}},
		{Key: "Notes", Value: gokeepasslib.V{Content: notes}},
	}
	return e
}

func newGroup(name string, iconID int64) gokeepasslib.Group {
	g := gokeepasslib.NewGroup()
	g.Name = name
	g.IconID = iconID
	return g
}

func removeEntry(group *gokeepasslib.Group, idx int) {
	if idx >= 0 && idx < len(group.Entries) {
		group.Entries = append(group.Entries[:idx], group.Entries[idx+1:]...)
	}
}

func addGroupByPath(groups *[]gokeepasslib.Group, path string, ng gokeepasslib.Group) bool {
	if path == "" {
		*groups = append(*groups, ng)
		return true
	}
	parts := strings.Split(path, "/")
	for i := range *groups {
		if (*groups)[i].Name == parts[0] {
			if len(parts) == 1 {
				(*groups)[i].Groups = append((*groups)[i].Groups, ng)
				return true
			}
			return addGroupByPath(&(*groups)[i].Groups, strings.Join(parts[1:], "/"), ng)
		}
	}
	return false
}

func removeGroupByPath(groups *[]gokeepasslib.Group, path string) bool {
	if path == "" {
		return false
	}
	parts := strings.Split(path, "/")
	for i := range *groups {
		if (*groups)[i].Name == parts[0] {
			if len(parts) == 1 {
				*groups = append((*groups)[:i], (*groups)[i+1:]...)
				return true
			}
			return removeGroupByPath(&(*groups)[i].Groups, strings.Join(parts[1:], "/"))
		}
	}
	return false
}
