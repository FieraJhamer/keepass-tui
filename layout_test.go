package main

import (
	"strings"
	"testing"
)

// The detail screen must not resize the boxes: the sidebar keeps the width it
// has in the browse screen and both panels end at the same height.
func TestDetailKeepsTheSameBoxes(t *testing.T) {
	m := searchModel(t)
	m.width = 120
	m.height = 30
	m.screen = screenDetail
	m.detailEntry = &m.db.Content.Root.Groups[0].Entries[0]

	browse := m.browseView()
	detail := m.detailView()

	// The top border of the boxes has to be the same in both screens.
	browseTop := boxTop(browse)
	detailTop := boxTop(detail)
	if browseTop != detailTop {
		t.Fatalf("the boxes changed size: browse %q, detail %q", browseTop, detailTop)
	}

	// Both panels have to close on the same line, inside the same view.
	if n := strings.Count(boxBottom(detail), "└"); n != 2 {
		t.Fatalf("the detail panels do not end together: %q", boxBottom(detail))
	}
	if n := strings.Count(boxBottom(browse), "└"); n != 2 {
		t.Fatalf("the browse panels do not end together: %q", boxBottom(browse))
	}
}

// boxTop returns the line that opens the two panels.
func boxTop(view string) string {
	for _, l := range strings.Split(view, "\n") {
		if strings.HasPrefix(l, "┌") {
			return l
		}
	}
	return ""
}

// boxBottom returns the line that closes the two panels.
func boxBottom(view string) string {
	for _, l := range strings.Split(view, "\n") {
		if strings.HasPrefix(l, "└") {
			return l
		}
	}
	return ""
}
