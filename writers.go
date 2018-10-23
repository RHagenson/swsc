package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"

	"github.com/biogo/biogo/seq/multi"
)

// writeOutputHeader truncates the *write file to only the header row
func writeOutputHeader() {
	header := []string{
		"name",
		"uce_site", "aln_site",
		"window_start", "window_stop",
		"type", "value",
		"plot_mtx",
	}
	if w, err := os.Open(*write); err != nil {
		file := csv.NewWriter(w)
		if err := file.Write(header); err != nil {
			log.Printf("Encountered %s in writing to file %s", err, *write)
		}
	} else {
		w.Close()
		log.Fatal(err)
	}
	return
}

// writeOutput appends partitioning data to *write
func writeOutput(f string, d [][]string) {
	if w, err := os.OpenFile(f, os.O_APPEND|os.O_WRONLY, 0644); err != nil {
		file := csv.NewWriter(w)
		if err := file.WriteAll(d); err != nil {
			log.Printf("Encountered %s in writing to file %s", err, f)
		}
	} else {
		w.Close()
		log.Fatal(err)
	}
}

// writeCfgEndBlock appends the end block to the specified .cfg file
func writeCfgEndBlock(pFinderFileName, datasetName string) {
	search := "rclusterf"
	block := "\n" +
		"## SCHEMES, search: all | user | greedy | rcluster | hcluster | kmeans ##\n" +
		"[schemes]\n" +
		fmt.Sprintf("search = %s;\n\n", search)
	if file, err := os.OpenFile(pFinderFileName, os.O_APPEND|os.O_WRONLY, 0644); err != nil {
		file.WriteString(block)
	} else {
		log.Fatalf("Failed to write .cfg end block: %s", err)
	}
}

// writeCfgStartBlock truncates the file to only the start block
func writeCfgStartBlock(pFinderFileName, datasetName string) {
	branchLengths := "linked"
	models := "GTR+G"
	modelSelection := "aicc"

	block := "## ALIGNMENT FILE ##\n" +
		fmt.Sprint("alignment = %s.phy;\n\n", datasetName) +
		"## BRANCHLENGTHS: linked | unlinked ##\n" +
		fmt.Sprintf("branchlengths = %s;\n\n", branchLengths) +
		"MODELS OF EVOLUTION: all | allx | mybayes | beast | gamma | gammai <list> ##\n" +
		fmt.Sprintf("models = %s;\n\n", models) +
		"# MODEL SELECTION: AIC | AICc | BIC #\n" +
		fmt.Sprintf("model_selection = %s;\n\n", modelSelection) +
		"## DATA BLOCKS: see manual for how to define ##\n" +
		"[data_blocks]\n"
	if file, err := os.Open(pFinderFileName); err != nil {
		file.WriteString(block)
	} else {
		log.Fatalf("Could not write PartionFinder2 file: %s", err)
	}
}

// pFinderConfigBlock appends the proper window size for the UCE
func pFinderConfigBlock(pFinderFileName, name string, bestWindow window, start, stop int, uceAln *multi.Multi) {
	block := ""
	// anyUndeterminedBlocks and anyBlocksWoAllSites are the frequency and absolute ATGC counts
	// indetermination is by zero frequency or zero count of any letter
	// Likely not the desired effect.
	// What is likely supposed to occur is any UCE that is composed of ambiguous(N), missing (?), or gap (-) should be consider indeterminate
	if bestWindow[1]-bestWindow[0] == stop-start || anyUndeterminedBlocks(bestWindow, uceAln) || anyBlocksWoAllSites(bestWindow, uceAln) {
		block = fmt.Sprintf("%s_all = %d-%d;\n", name, start+1, stop)
	} else {
		// left UCE
		leftStart := start + 1
		leftEnd := start + bestWindow[0]
		// core UCE
		coreStart := leftEnd + 1
		coreEnd := start + bestWindow[1]
		// right UCE
		rightStart := coreEnd + 1
		rightEnd := stop
		block = fmt.Sprintf("%s_core = %d-%d;\n", name, coreStart, coreEnd) +
			fmt.Sprintf("%s_left = %d-%d;\n", name, leftStart, leftEnd) +
			fmt.Sprintf("%s_right = %d-%d;\n", name, rightStart, rightEnd)
	}

	if file, err := os.OpenFile(pFinderFileName, os.O_APPEND|os.O_WRONLY, 0644); err != nil {
		file.WriteString(block)
	} else {
		log.Fatalf("Failed to write .cfg config block: %s", err)
	}
}
