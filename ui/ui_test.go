package ui_test

import (
	"fmt"
	"path"
	"testing"

	"bitbucket.org/rhagenson/swsc/ui"
)

var weirdUTF = []struct {
	f string
}{
	{"testdataset"},
	{"田中さんにあげて下さい"},
	{",。・:*:・゜’( ☻ ω ☻ )。・:*:・゜’"},
	{"Ω≈ç√∫˜µ≤≥÷"},
}

// TestHeader validates the Header has not changed along with strange UTF-8
func TestHeader(t *testing.T) {
	tt := weirdUTF
	for _, tc := range tt {
		exp := fmt.Sprintf("\nAnalysing %s\n", path.Base(tc.f))
		got := ui.Header(tc.f)
		if got != exp {
			t.Errorf("Expected: %s, got %s", exp, got)
		}
	}
}

// TestHeader validates the Header has not changed along with strange UTF-8
func TestFooter(t *testing.T) {
	tt := weirdUTF
	for _, tc := range tt {
		exp := fmt.Sprintf("\nWrote partitions to %s\n", tc.f)
		got := ui.Footer(tc.f)
		if got != exp {
			t.Errorf("Expected: %s, got %s", exp, got)
		}
	}
}
