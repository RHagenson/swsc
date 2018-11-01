package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"path"
	"strings"

	"bitbucket.org/rhagenson/swsc/nexus"
	"bitbucket.org/rhagenson/swsc/pfinder"
	"bitbucket.org/rhagenson/swsc/ui"
	"github.com/spf13/pflag"
	pb "gopkg.in/cheggaaa/pb.v1"
)

// Required flags
var (
	read  = pflag.String("input", "", "Nexus (.nex) file to process")
	write = pflag.String("output", "", "Partition file to write")
)

// General use flags
var (
	minWin = pflag.Int("minWin", 50, "Minimum window size")
	cfg    = pflag.Bool("cfg", false, "Write config file for PartionFinder2")
	help   = pflag.Bool("help", false, "Print help and exit")
)

// Metric flags
var (
	entropy = pflag.Bool("entropy", false, "Calculate Shannon's entropy metric")
	gc      = pflag.Bool("gc", false, "Calculate GC content metric")
	// multi = pflag.Bool("multi", false, "Calculate multinomial distribution metric")
)

// Global reference vars
var (
	pFinderFileName = ""
	pfinderFile     = new(os.File)
	datasetName     = ""
	metrics         = make([]string, 0)
)

func setup() {
	pflag.Parse()
	if *help {
		pflag.Usage()
		os.Exit(1)
	}

	// Failure states
	switch {
	case *read == "" && *write == "":
		pflag.Usage()
		log.Fatalf("Must provide input and output names")
	case !strings.HasSuffix(*read, ".nex"):
		log.Fatalf("Input expected in .nex format, got %s format", path.Ext(*read))
	case !strings.HasSuffix(*write, ".csv"):
		log.Fatalf("Output written in .csv format, got %s format", path.Ext(*write))
	case *minWin < 0:
		log.Fatalf("Window size must be positive, got %d", *minWin)
	case !(*entropy || *gc):
		log.Fatalf("At least one metric is required")
	default:
		fmt.Printf("Arguments are reasonable")
	}

	// Set global vars
	datasetName = strings.TrimRight(path.Base(*read), ".nex")
	if *cfg {
		pFinderFileName = path.Join(path.Dir(*read), datasetName) + ".cfg"
	}
	if *entropy {
		metrics = append(metrics, "entropy")
	}
	if *gc {
		metrics = append(metrics, "gc")
	}
}

func main() {
	// Parse CLI arguments
	setup()

	// Inform user on STDOUT what is being done
	ui.Header(*read)

	// Open input file
	in, err := os.Open(*read)
	defer in.Close()
	if err != nil {
		log.Fatalf("Could not read input file: %s", err)
	}

	out, err := os.Create(*write)
	defer out.Close()
	if err != nil {
		log.Fatalf("Could not create output file: %s", err)
	}

	// Write the header to the output file (clears output file if present)
	writeOutputHeader(out)

	// If PartitionFinder2 config file is desired, write its header/starting block
	if *cfg {
		pfinderFile, err = os.Create(pFinderFileName)
		if err != nil {
			log.Fatalf("Could not read file: %s", err)
		}
		pfinder.WriteStartBlock(pfinderFile, datasetName)
	}

	// Read in the input Nexus file
	nex := nexus.Read(in)

	// Early panic if minWin has been set too large to create flanks and core of that length
	if err := validateMinWin(nex.Alignment().Len(), *minWin); err != nil {
		log.Fatalln(err)
	}

	// Process the input with selected metrics and minimum window size, internally writing output files
	// Initialize useful variables
	var (
		start = math.MaxInt16 // Minimum position in UCE
		stop  = math.MinInt16 // Maximum position in UCE, inclusive
		aln   = nex.Alignment()
		uces  = nex.Charsets()
		bar   = pb.StartNew(len(uces))
	)

	// Process each UCE in turn
	for name, sites := range uces {
		// Get the widest window for the UCE if multiple windows exist (which they should not, but can in the Nexus format)
		for _, pair := range sites {
			if pair.First() < start {
				start = pair.First()
			}
			if stop < pair.Second() {
				stop = pair.Second()
			}
		}

		uceAln := aln.Subseq(start, stop)
		bestWindows, metricArray := processUce(uceAln, metrics, *minWin)

		if *cfg {
			for _, bestWindow := range bestWindows {
				pfinder.WriteConfigBlock(
					pfinderFile, name, bestWindow, start, stop,
					anyBlocksWoAllSites(bestWindow, aln) || anyUndeterminedBlocks(bestWindow, aln), // Write full range if either
				)
			}
		}
		alnSites := make([]int, stop-start)
		for i := range alnSites {
			alnSites[i] = i + start
		}
		writeOutput(out, bestWindows, metricArray, alnSites, name)
		bar.Increment()
	}
	bar.FinishPrint("Finished processing UCEs")
	if *cfg {
		for range metrics {
			writeCfgEndBlock(pfinderFile, datasetName)
		}
	}

	// Inform user of where output was written
	ui.Footer(*write)

	// Close the config file if it was opened
	if *cfg {
		pfinderFile.Close()
	}
}

// processUce computes the corresponding metrics within the minimum window size,
// returning the best window and list of values for each metric
func processUce(uceAln nexus.Alignment, metrics []string, minWin int) (map[string]window, map[string][]float64) {
	metricBestWindow := make(map[string]window, len(metrics))
	metricBestVals := make(map[string][]float64, len(metrics))

	windows := getAllWindows(uceAln, minWin)
	inVarSites := invariantSites(uceAln)

	for _, m := range metrics {
		switch m {
		case "entropy":
			metricBestVals["entropy"] = sitewiseEntropy(uceAln)
		case "gc":
			metricBestVals["gc"] = sitewiseGc(uceAln)
			// case "multi":
			// 	metricBestVals["multi"] = sitewiseMulti(uceAln)
		}
	}
	if len(windows) > 1 {
		metricBestWindow = getBestWindows(metricBestVals,
			windows, uceAln.Len(), inVarSites,
		)
	} else {
		for _, k := range metrics {
			metricBestWindow[k] = windows[0]
		}
	}
	return metricBestWindow, metricBestVals
}
