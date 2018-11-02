package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"

	"bitbucket.org/rhagenson/swsc/ui"
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
		ui.Errorf("Problem writing %s, encountered %s.", f.Name(), err)
	}
	return
}

// writeOutput appends partitioning data to output
func writeOutput(f *os.File, bestWindows map[Metric]Window, metricArray map[Metric][]float64, alnSites []int, name string) {
	// Validate input
	for k, v := range metricArray {
		if len(v) != len(alnSites) {
			msg := fmt.Sprintf("Not enough alignment sites "+
				"(%d) produced to match metric %q of len %d",
				len(alnSites), k, len(v))
			ui.Errorf(msg)
		}
	}
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

	file := csv.NewWriter(f)
	str := make([]string, 8)
	for m, v := range metricArray {
		window := bestWindows[m]
		for i := range alnSites {
			str = []string{
				name,                                  // 1) UCE name
				strconv.Itoa(uceSites[i]),             // 2) UCE site position relative to center of alignment
				strconv.Itoa(i),                       // 3) UCE site position absolute
				strconv.Itoa(window.Start()),          // 4) Best window for metric, start
				strconv.Itoa(window.Stop()),           // 5) Best window for metric, stop
				m.String(),                            // 6) Metric under analysis
				strconv.FormatFloat(v[i], 'e', 5, 64), // 7) Metric value at site position
				strconv.Itoa(relToWindow(window.Start(), i, window.Stop())), // 8) -1 if before window, 0 if in window, 1 if after window
			}
			if err := file.Write(str); err != nil {
				ui.Errorf("Encountered %s in writing to file %s", err, f.Name())
			}
		}
	}
}

// relToWindow is a codified function for use in later tools of whether the current alignment position
// is before (-1), in (0), or after (1) the window
func relToWindow(start, cur, stop int) int {
	if cur < start {
		return -1 // Before window
	} else if stop < cur {
		return 1 // After window
	}
	return 0 // In Window
}
