package pfinder_test

import (
	"fmt"
	"strings"
	"testing"

	"bitbucket.org/rhagenson/swsc/internal/pfinder"
)

func TestWriteStartBlock(t *testing.T) {
	branchLengths := "linked"
	models := "GTR+G"
	modelSelection := "aicc"
	tt := []struct {
		datasetName string
	}{
		{"testdataset"},
		{"田中さんにあげて下さい"},
		{",。・:*:・゜’( ☻ ω ☻ )。・:*:・゜’"},
		{"Ω≈ç√∫˜µ≤≥÷"},
	}
	for _, tc := range tt {
		w := new(strings.Builder)
		pfinder.WriteStartBlock(w, tc.datasetName)
		block := "## ALIGNMENT FILE ##\n" +
			fmt.Sprintf("alignment = %s.nex;\n\n", tc.datasetName) +
			"## BRANCHLENGTHS: linked | unlinked ##\n" +
			fmt.Sprintf("branchlengths = %s;\n\n", branchLengths) +
			"MODELS OF EVOLUTION: all | allx | mybayes | beast | gamma | gammai <list> ##\n" +
			fmt.Sprintf("models = %s;\n\n", models) +
			"# MODEL SELECTION: AIC | AICc | BIC #\n" +
			fmt.Sprintf("model_selection = %s;\n\n", modelSelection) +
			"## DATA BLOCKS: see manual for how to define ##\n" +
			"[data_blocks]\n"
		if w.String() != block {
			diff := make([]string, 0)
			gotLines := strings.Split(w.String(), "\n")
			expLines := strings.Split(block, "\n")
			for i := range gotLines {
				if gotLines[i] != expLines[i] {
					diff = append(diff, fmt.Sprintf("Got:\n%s\nExpected:\n%s\n", gotLines[i], expLines[i]))
				}
			}
			t.Errorf("Expected block and received block did not match. Differences:\n%s", diff)
		}
	}
}

func TestWriteConfigBlock(t *testing.T) {
	tt := []struct {
		name        string
		bestWindow  [2]int
		start, stop int
		fullRange   bool
	}{
		{"UCE01-partial", [2]int{10, 60}, 5, 100, false},
		{"UCE01-full", [2]int{10, 60}, 5, 100, true},
	}
	for _, tc := range tt {
		w := new(strings.Builder)
		pfinder.WriteConfigBlock(w, tc.name, tc.bestWindow, tc.start, tc.stop, tc.fullRange)
		t.Run("Correct number of lines", func(t *testing.T) {
			nLines := len(strings.Split(strings.TrimSpace(w.String()), "\n"))
			if tc.fullRange || tc.bestWindow[1]-tc.bestWindow[0] == tc.stop-tc.start {
				if nLines != 1 {
					t.Errorf("Expected 1 output line (full range), got %d", nLines)
				}
			} else {
				if nLines != 3 {
					t.Errorf("Expected 3 output line (left, core, right flank), got %d", nLines)
				}
			}
		})
	}
}

func TestWriteEndBlock(t *testing.T) {
	w := new(strings.Builder)
	pfinder.WriteEndBlock(w)
	search := "rclusterf"
	block := "\n" +
		"## SCHEMES, search: all | user | greedy | rcluster | hcluster | kmeans ##\n" +
		"[schemes]\n" +
		fmt.Sprintf("search = %s;\n\n", search)
	if w.String() != block {
		diff := make([]string, 0)
		gotLines := strings.Split(w.String(), "\n")
		expLines := strings.Split(block, "\n")
		for i := range gotLines {
			if gotLines[i] != expLines[i] {
				diff = append(diff, fmt.Sprintf("Got:\n%s\nExpected:\n%s\n", gotLines[i], expLines[i]))
			}
		}
		t.Errorf("Expected block and received block did not match. Differences:\n%s", diff)
	}
}
