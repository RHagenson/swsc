package main

import (
	"fmt"
	"io"
	"math"
	"os"
	"path"
	"sort"
	"strings"

	"bitbucket.org/rhagenson/swsc/internal/metric"
	"bitbucket.org/rhagenson/swsc/internal/nexus"
	"bitbucket.org/rhagenson/swsc/internal/pfinder"
	"bitbucket.org/rhagenson/swsc/internal/uce"
	"bitbucket.org/rhagenson/swsc/internal/ui"
	"bitbucket.org/rhagenson/swsc/internal/utils"
	"bitbucket.org/rhagenson/swsc/internal/windows"
	"bitbucket.org/rhagenson/swsc/internal/writers"
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
	cfg    = pflag.Bool("cfg", false, "Write config file for PartionFinder2 (same dir as --output)")
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
	metrics         = make([]metric.Metric, 0)
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
		ui.Errorf("Must provide input and output names")
	case !strings.HasSuffix(*read, ".nex"):
		ui.Errorf("Input expected to end in .nex, got %s", path.Ext(*read))
	case !strings.HasSuffix(*write, ".csv"):
		ui.Errorf("Output expected to end in .csv, got %s", path.Ext(*write))
	case *minWin < 0:
		ui.Errorf("Window size must be positive, got %d", *minWin)
	case !(*entropy || *gc):
		ui.Errorf("At least one metric is required")
	default:
	}

	// Set global vars
	datasetName = strings.TrimRight(path.Base(*read), ".nex")
	if *cfg {
		pFinderFileName = path.Join(path.Dir(*write), datasetName) + ".cfg"
	}
	if *entropy {
		metrics = append(metrics, metric.Entropy)
	}
	if *gc {
		metrics = append(metrics, metric.GC)
	}
}

func main() {
	// Parse CLI arguments
	setup()

	// Inform user on STDOUT what is being done
	fmt.Println(ui.Header(*read))

	// Open input file
	in, err := os.Open(*read)
	defer in.Close()
	if err != nil {
		ui.Errorf("Could not read input file: %s", err)
	}

	out, err := os.Create(*write)
	defer out.Close()
	if err != nil {
		ui.Errorf("Could not create output file: %s", err)
	}

	// Write the header to the output file
	writers.WriteOutputHeader(out)

	// If PartitionFinder2 config file is desired, write its header/starting block
	if *cfg {
		pfinderFile, err = os.Create(pFinderFileName)
		if err != nil {
			ui.Errorf("Could not read PartitionFinder2 file: %s", err)
		}
		block := pfinder.StartBlock(datasetName)
		if _, err := io.WriteString(pfinderFile, block); err != nil {
			ui.Errorf("Failed to write .cfg start block: %s", err)
		}

	}

	// Read in the input Nexus file
	nex := nexus.Read(in)

	// Early panic if minWin has been set too large to create flanks and core of that length
	if err := utils.ValidateMinWin(nex.Alignment().Len(), *minWin); err != nil {
		ui.Errorf("Early exit: %v", err)
	}

	var (
		aln  = nex.Alignment()        // Sequence alignment
		uces = nex.Charsets()         // UCE set
		bar  = pb.StartNew(len(uces)) // Progress bar
	)

	// Order UCEs
	// Create reverse lookup to maintain order
	revUCEs := make(map[int]string)
	keys := make([]int, 0)
	for name, sites := range uces {
		var (
			start = math.MaxInt16 // Minimum position in UCE
		)
		// Get the inclusive window for the UCE if multiple windows exist (which they should not, but can in the Nexus format)
		for _, pair := range sites {
			if pair.First() < start {
				start = pair.First()
			}
		}
		revUCEs[start] = name
		keys = append(keys, start)
	}
	sort.Ints(keys) // Sort done in place

	// Process each UCE in turn
	for _, key := range keys {
		name := revUCEs[key]
		sites := uces[name]
		var (
			start = math.MaxInt16 // Minimum position in UCE
			stop  = math.MinInt16 // Maximum position in UCE
		)
		// Get the inclusive window for the UCE if multiple windows exist (which they should not, but can in the Nexus format)
		for _, pair := range sites {
			if pair.First() < start {
				start = pair.First()
			}
			if stop < pair.Second() {
				stop = pair.Second()
			}
		}

		uceAln := aln.Subseq(start, stop)
		bestWindows, metricArray := uce.ProcessUce(uceAln, metrics, *minWin, nex.Letters())

		if *cfg {
			for _, bestWindow := range bestWindows {
				block := pfinder.ConfigBlock(
					name, bestWindow, start, stop,
					windows.UseFullRange(bestWindow, aln, nex.Letters()),
				)
				if _, err := io.WriteString(pfinderFile, block); err != nil {
					ui.Errorf("Failed to write .cfg config block: %s", err)
				}

			}
		}
		alnSites := make([]int, stop-start)
		for i := range alnSites {
			alnSites[i] = i + start
		}
		writers.WriteOutput(out, bestWindows, metricArray, alnSites, name)
		bar.Increment()
	}
	bar.FinishPrint("Finished processing UCEs")
	if *cfg {
		block := pfinder.EndBlock()
		if _, err := io.WriteString(pfinderFile, block); err != nil {
			ui.Errorf("Failed to write .cfg end block: %s", err)
		}
	}

	// Inform user of where output was written
	fmt.Println(ui.Footer(*write))

	// Close the config file if it was opened
	if *cfg {
		pfinderFile.Close()
	}
}
