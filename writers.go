package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"

	"bitbucket.org/rhagenson/swsc/nexus"
)

// writeOutputHeader truncates the *write file to only the header row
func writeOutputHeader(f *os.File) {
	header := []string{
		"name",
		"uce_site", "aln_site",
		"window_start", "window_stop",
		"type", "value",
		"plot_mtx",
	}
	file := csv.NewWriter(f)
	if err := file.Write(header); err != nil {
		file.Flush()
		log.Printf("Problem writing %s, encountered %s.", f.Name(), err)
	}
	return
}

// writeOutput appends partitioning data to output
func writeOutput(f *os.File, bestWindows map[string]window, metricArray map[string][]float64, alnSites []int, name string) {
	// Validate input
	for k, v := range metricArray {
		if len(v) != len(alnSites) {
			msg := fmt.Sprintf("Not enough alignment sites "+
				"(%d) produced to match metric %q of len %d",
				len(alnSites), k, len(v))
			log.Fatalf(msg)
		}
	}

	d := make([][]string, 0)
	N := len(alnSites)
	middle := int(float64(N) / 2.0)
	uceSites := make([]int, N)
	for i := range uceSites {
		uceSites[i] = i - middle
	}

	names := make([]string, N*3)
	for i := range names {
		names[i] = name
	}

	matrixVals := make([]int8, N)
	for _, w := range bestWindows {
		matrixVals = csvColToPlotMatrix(w, N)
	}

	file := csv.NewWriter(f)
	for m, v := range metricArray {
		window := bestWindows[m]
		for i := 0; i < N; i++ {
			file.Write([]string{
				name,                    // UCE name
				string(uceSites[i]),     // UCE site position relative to center of alignment
				string(i),               // UCE site position absolute
				string(window.Start()),  // Best window for metric, start
				string(window.Stop()),   // Best window for metric, stop
				m,                       // Metric under analysis
				fmt.Sprintf("%f", v[i]), // Metric value at site position
				string(matrixVals[i]),   // Prior to best window (-1), in best window (0), after window (+1)
			})
		}
	}

	if err := file.WriteAll(d); err != nil {
		log.Printf("Encountered %s in writing to file %s", err, f.Name())
	}
}

// writeCfgEndBlock appends the end block to the specified .cfg file
func writeCfgEndBlock(f *os.File, datasetName string) {
	search := "rclusterf"
	block := "\n" +
		"## SCHEMES, search: all | user | greedy | rcluster | hcluster | kmeans ##\n" +
		"[schemes]\n" +
		fmt.Sprintf("search = %s;\n\n", search)
	if _, err := f.WriteString(block); err != nil {
		log.Fatalf("Failed to write .cfg end block: %s", err)
	}
}

// writeCfgStartBlock truncates the file to only the start block
func writeCfgStartBlock(f *os.File, datasetName string) {
	branchLengths := "linked"
	models := "GTR+G"
	modelSelection := "aicc"

	block := "## ALIGNMENT FILE ##\n" +
		fmt.Sprintf("alignment = %s.nex;\n\n", datasetName) +
		"## BRANCHLENGTHS: linked | unlinked ##\n" +
		fmt.Sprintf("branchlengths = %s;\n\n", branchLengths) +
		"MODELS OF EVOLUTION: all | allx | mybayes | beast | gamma | gammai <list> ##\n" +
		fmt.Sprintf("models = %s;\n\n", models) +
		"# MODEL SELECTION: AIC | AICc | BIC #\n" +
		fmt.Sprintf("model_selection = %s;\n\n", modelSelection) +
		"## DATA BLOCKS: see manual for how to define ##\n" +
		"[data_blocks]\n"
	if _, err := f.WriteString(block); err != nil {
		log.Fatalf("Could not write PartionFinder2 file: %s", err)
	}
}

// pFinderConfigBlock appends the proper window size for the UCE
func pFinderConfigBlock(f *os.File, name string, bestWindow window, start, stop int, uceAln nexus.Alignment, chars []byte) {
	block := ""
	// anyUndeterminedBlocks and anyBlocksWoAllSites are the frequency and absolute ATGC counts
	// indetermination is by zero frequency or zero count of any letter
	// Likely not the desired effect.
	// What is likely supposed to occur is any UCE that is composed of ambiguous(N), missing (?), or gap (-) should be consider indeterminate
	if bestWindow[1]-bestWindow[0] == stop-start || anyUndeterminedBlocks(bestWindow, uceAln, chars) || anyBlocksWoAllSites(bestWindow, uceAln, chars) {
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

	if _, err := f.WriteString(block); err != nil {
		log.Fatalf("Failed to write .cfg config block: %s", err)
	}
}
