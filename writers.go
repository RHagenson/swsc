package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
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
func writeOutput(f *os.File, bestWindows map[Metric]Window, metricArray map[Metric][]float64, alnSites []int, name string) {
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
				name,                                  // UCE name
				strconv.Itoa(uceSites[i]),             // UCE site position relative to center of alignment
				strconv.Itoa(i),                       // UCE site position absolute
				strconv.Itoa(window.Start()),          // Best window for metric, start
				strconv.Itoa(window.Stop()),           // Best window for metric, stop
				m.String(),                            // Metric under analysis
				strconv.FormatFloat(v[i], 'e', 5, 64), // Metric value at site position
				strconv.Itoa(int(matrixVals[i])),      // Prior to best window (-1), in best window (0), after window (+1)
			})
		}
	}

	if err := file.WriteAll(d); err != nil {
		log.Printf("Encountered %s in writing to file %s", err, f.Name())
	}
}
