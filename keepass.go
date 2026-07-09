package main

import (
	"os"
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

	db.UnlockProtectedEntries()
	return db, nil
}

func saveDatabase(db *gokeepasslib.Database, path string) error {
	db.LockProtectedEntries()

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	encoder := gokeepasslib.NewEncoder(f)
	return encoder.Encode(db)
}

func flattenGroups(groups []gokeepasslib.Group, prefix string, depth int) []groupInfo {
	var result []groupInfo
	for i := range groups {
		g := &groups[i]
		path := g.Name
		if prefix != "" {
			path = prefix + "/" + g.Name
		}
		result = append(result, groupInfo{name: g.Name, path: path, depth: depth, iconID: g.IconID})
		result = append(result, flattenGroups(g.Groups, path, depth+1)...)
	}
	return result
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

func getGroupIcon(groups []gokeepasslib.Group, path string) int64 {
	g := findGroup(groups, path)
	if g != nil {
		return g.IconID
	}
	return 0
}
