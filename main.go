package main

import (
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"path"
	"strings"

	"bitbucket.org/rhagenson/swsc/nexus"
	"github.com/spf13/pflag"
	pb "gopkg.in/cheggaaa/pb.v1"
)

// General use flags
var (
	read   = pflag.String("input", "", "Nexus (.nex) file to process")
	write  = pflag.String("output", "", "Partition file to write")
	minWin = pflag.Int("minwindow", 50, "Minimum window size")
	cfg    = pflag.Bool("cfg", false, "Write file for PartionFinder2")
	help   = pflag.Bool("help", false, "Print help and exit")
)

// Global reference vars
var (
	pFinderFileName = ""
	datasetName     = ""
	metrics         = []string{"entropy"}
)

func setup() {
	pflag.Parse()
	if *help {
		pflag.Usage()
		os.Exit(1)
	}
	switch {
	case *read == "" && *write == "":
		pflag.Usage()
		log.Fatalf("Must provide input and output names")
	case !strings.HasSuffix(*read, ".nex"):
		log.Fatalf("Input expected .nex format, got %s format", path.Ext(*read))
	case !strings.HasSuffix(*write, ".csv"):
		log.Fatalf("Output expected .csv format, got %s format", path.Ext(*write))
	case *minWin > 0:
		log.Fatalf("Window size must be positive, got %d", *minWin)
	default:
		fmt.Printf("Arguments are reasonable")
	}
	datasetName = strings.TrimRight(path.Base(*read), ".nex")
	if *cfg {
		pFinderFileName = path.Join(path.Dir(*read), datasetName) + ".cfg"
	}
}

func main() {
	setup()
	printHeader(*read)
	file, err := os.Open(*read)
	defer file.Close()
	if err != nil {
		log.Fatalf("Could not read file: %s", err)
	}
	if *cfg {
		for _, m := range metrics {
			writeCfgStartBlock(pFinderFileName, datasetName, m)
		}
	}
	nex := nexus.New()
	nex.Read(file)
	partitions := processDatasetMetrics(nex, []string{"entropy"}, *minWin)
	writeOutput(*write, partitions)
	printFooter(*write, len(partitions))
}

func writeCfgStartBlock(f, name string) {
	err := ioutil.WriteFile(f, pFinderStartBlock(name), 0644)
	if err != nil {
		log.Fatalf("Could not write PartionFinder2 file: %s", err)
	}
}

// writeOutput handles writing data to a specified file
func writeOutput(f string, d [][]string) {
	header := []string{
		"name",
		"uce_site", "aln_site",
		"window_start", "window_stop",
		"type", "value",
		"plot_mtx",
	}
	if w, err := os.Open(f); err != nil {
		file := csv.NewWriter(w)
		if err := file.Write(header); err != nil {
			log.Printf("Encountered %s in writing to file %s", err, f)
		}
		if err := file.WriteAll(d); err != nil {
			log.Printf("Encountered %s in writing to file %s", err, f)
		}
	} else {
		log.Fatal(err)
	}
}

// processDatasetMetrics read in a Nexus file, calculating defined properties
// using a minimum sliding window size
func processDatasetMetrics(nex *nexus.Nexus, props []int8, win int) {
	var (
		start    = math.MaxInt16 // Minimum position in UCE
		stop     = math.MinInt16 // Maximum position in UCE, inclusive
		aln      = nex.Alignment()
		charsets = nex.Charsets()
		bar      = pb.StartNew(len(charsets))
	)

	for name, sites := range charsets {
		start = sites[0].Start()
		stop = sites[0].Stop()
		for _, pair := range sites {
			if pair.Start() < start {
				start = pair.Start()
			}
			if stop < pair.Stop() {
				stop = pair.Stop()
			}
		}
		uceAln, _ := aln.Subseq(start, stop+1)
		bestWindows, metricArray := process_uce(uceAln, props, minWin)
		if *cfg {
			pfinderConfig, err := os.OpenFile(
				pFinderFileName,
				os.O_APPEND|os.O_WRONLY, 0644,
			)
			defer pfinderConfig.Close()
			if err != nil {
				log.Fatalf("Could not append to PartitionFinder2 file: %s", err)
			}
			for i, bestWindow := range bestWindows {
				pfinderConfig.Write(pFinderConfigBlock(bestWindow, name, start, stop, uceAln))
			}
		}
		writeOutput(*write, [][]string{bestWindows, metricArray, sites, name})
		bar.Increment()
	}
	bar.FinishPrint("Finished processing UCEs")
	if *cfg {
		for _, m := range metrics {
			writeCfgEndBlock(pFinderFileName, datasetName, m)
		}
	}
	return
}
