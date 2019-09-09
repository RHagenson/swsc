package pfinder_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/rhagenson/swsc/internal/pfinder"
)

func TestStartBlock(t *testing.T) {
	branchLengths := "linked"
	models := "mrbayes"
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
		got := pfinder.StartBlock(tc.datasetName)
		exp := "## ALIGNMENT FILE ##\n" +
			fmt.Sprintf("alignment = %s.nex;\n\n", tc.datasetName) +
			"## BRANCHLENGTHS: linked | unlinked ##\n" +
			fmt.Sprintf("branchlengths = %s;\n\n", branchLengths) +
			"## MODELS OF EVOLUTION: all | allx | mrbayes | beast | gamma | gammai <list> ##\n" +
			fmt.Sprintf("models = %s;\n\n", models) +
			"# MODEL SELECTION: AIC | AICc | BIC #\n" +
			fmt.Sprintf("model_selection = %s;\n\n", modelSelection) +
			"## DATA BLOCKS: see manual for how to define ##\n" +
			"[data_blocks]\n"
		if got != exp {
			diff := make([]string, 0)
			gotLines := strings.Split(got, "\n")
			expLines := strings.Split(exp, "\n")
			for i := range gotLines {
				if gotLines[i] != expLines[i] {
					diff = append(diff, fmt.Sprintf("Got:\n%s\nExpected:\n%s\n", gotLines[i], expLines[i]))
				}
			}
			t.Errorf("Expected block and received block did not match. Differences:\n%s", diff)
		}
	}
}

func TestConfigBlock(t *testing.T) {
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
		got := pfinder.ConfigBlock(tc.name, tc.bestWindow, tc.start, tc.stop, tc.fullRange)
		t.Run("Correct number of lines", func(t *testing.T) {
			nLines := len(strings.Split(strings.TrimSpace(got), "\n"))
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

func TestEndBlock(t *testing.T) {
	got := pfinder.EndBlock()
	search := "rclusterf"
	exp := "\n" +
		"## SCHEMES, search: all | user | greedy | rcluster | hcluster | kmeans ##\n" +
		"[schemes]\n" +
		fmt.Sprintf("search = %s;\n\n", search)
	if got != exp {
		diff := make([]string, 0)
		gotLines := strings.Split(got, "\n")
		expLines := strings.Split(exp, "\n")
		for i := range gotLines {
			if gotLines[i] != expLines[i] {
				diff = append(diff, fmt.Sprintf("Got:\n%s\nExpected:\n%s\n", gotLines[i], expLines[i]))
			}
		}
		t.Errorf("Expected block and received block did not match. Differences:\n%s", diff)
	}
}
