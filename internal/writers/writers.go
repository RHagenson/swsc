package writers

import (
	"encoding/csv"
	"io"
	"math"
	"strconv"

	"bitbucket.org/rhagenson/swsc/internal/metrics"
	"bitbucket.org/rhagenson/swsc/internal/ui"
	"bitbucket.org/rhagenson/swsc/internal/windows"
)

// WriteOutputHeader truncates the *write file to only the header row
func WriteOutputHeader(f io.Writer) {
	header := []string{
		"name",
		"uce_site", "aln_site",
		"window_start", "window_stop",
		"type", "value",
		"plot_mtx",
	}
	file := csv.NewWriter(f)
	if err := file.Write(header); err != nil {
		ui.Errorf("Problem writing output header: %s.", err)
	}
	file.Flush()
	return
}

// WriteOutput appends partitioning data to output
func WriteOutput(f io.Writer, bestWindows map[metrics.Metric]windows.Window, metricArray map[metrics.Metric][]float64, alnSites []int, name string) {
	d := make([][]string, len(metricArray)*len(alnSites))
	N := len(alnSites)
	middle := int(math.Ceil(float64(N) / 2.0))
	uceSites := make([]int, N)
	for i := range uceSites {
		uceSites[i] = i - middle
	}

	file := csv.NewWriter(f)
	mNum := 0
	for m, v := range metricArray {
		window := bestWindows[m]
		for i := range v {
			d[mNum+i] = []string{
				name,                                  // 1) UCE name
				strconv.Itoa(uceSites[i]),             // 2) UCE site position relative to center of alignment
				strconv.Itoa(alnSites[i]),             // 3) UCE site position absolute
				strconv.Itoa(window.Start()),          // 4) Best window for metric, start
				strconv.Itoa(window.Stop() + 1),       // 5) Best window for metric, stop
				m.String(),                            // 6) Metric under analysis
				strconv.FormatFloat(v[i], 'e', 5, 64), // 7) Metric value at site position
				strconv.Itoa(relToWindow(window.Start(), i, window.Stop())), // 8) -1 if before window, 0 if in window, 1 if after window
			}
		}
		mNum++
	}
	if err := file.WriteAll(d); err != nil {
		ui.Errorf("Problem writing to output: %s", err)
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
