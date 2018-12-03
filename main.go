package main

import (
	"fmt"
	"io"
	"math"
	"os"
	"path"
	"sort"
	"strings"

	"bitbucket.org/rhagenson/swsc/internal/metrics"
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
	read  = pflag.String("input", "", "Nexus file to process (.nex)")
	write = pflag.String("output", "", "Partition file to write (.csv)")
	cfg   = pflag.String("cfg", "", "Config file for PartionFinder2 (.cfg)")
)

// General use flags
var (
	minWin      = pflag.Uint("minWin", 50, "Minimum window size")
	largeCore   = pflag.Bool("largeCore", false, "When a small and large core have equivalent metrics, choose the large core")
	nCandidates = pflag.Uint("candidates", 3, "Number of best candidates to search with")
	help        = pflag.Bool("help", false, "Print help and exit")
)

// Metric flags
var (
	entropy = pflag.Bool("entropy", false, "Calculate Shannon's entropy metric")
	gc      = pflag.Bool("gc", false, "Calculate GC content metric")
	// multi = pflag.Bool("multi", false, "Calculate multinomial distribution metric")
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
		ui.Errorf("Must provide input and output names\n")
	case !strings.HasSuffix(*read, ".nex"):
		ui.Errorf("Input expected to end in .nex, got %s\n", path.Ext(*read))
	case !strings.HasSuffix(*write, ".csv"):
		ui.Errorf("Output expected to end in .csv, got %s\n", path.Ext(*write))
	case !strings.HasSuffix(*cfg, ".cfg"):
		ui.Errorf("Config file expected to end in .cfg, got %s\n", path.Ext(*cfg))
	case *minWin < 0:
		ui.Errorf("Window size must be positive, got %d\n", *minWin)
	case *entropy == *gc && (*entropy || *gc):
		ui.Errorf("Only one metric is allowed\n")
	case !(*entropy || *gc):
		ui.Errorf("At least one metric is needed\n")
	}
}

func main() {
	// Parse CLI arguments
	setup()

	// Inform user on STDOUT what is being done
	fmt.Println(ui.Header(*read))

	in, err := os.Open(*read)
	defer in.Close()
	if err != nil {
		ui.Errorf("Could not read input file: %s", err)
	}

	// Read in the input Nexus file
	nex := nexus.Read(in)

	// Early panic if minWin has been set too large to create flanks and core of that length
	if err := utils.ValidateMinWin(nex.Alignment().Len(), int(*minWin)); err != nil {
		ui.Errorf("Failed due to: %v", err)
	}

	out, err := os.Create(*write)
	defer out.Close()
	if err != nil {
		ui.Errorf("Could not create output file: %s", err)
	}

	writers.WriteOutputHeader(out)

	var (
		aln     = nex.Alignment()        // Sequence alignment
		uces    = nex.Charsets()         // UCE set
		bar     = pb.StartNew(len(uces)) // Progress bar
		metVals = make(map[metrics.Metric][]float64, 3)
	)

	if *entropy {
		metVals[metrics.Entropy] = metrics.SitewiseEntropy(aln, nex.Letters())
	}
	if *gc {
		metVals[metrics.GC] = metrics.SitewiseGc(aln)
	}

	// Sort UCEs
	// Create reverse lookup to maintain order
	revUCEs := make(map[int]string, len(uces))
	keys := make([]int, 0, len(uces))
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
	pFinderConfigBlocks := make([]string, 0, len(uces))
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

		// Currently uceAln is the subsequence while inVarSites and metVals the entire sequence
		// It should be the case that processing a UCE considers where the start and stop of the UCE are
		// finding the best Window within that range
		bestWindows := uce.ProcessUce(start, stop, metVals, *minWin, nex.Letters(), *largeCore, *nCandidates)
		if *cfg != "" {
			for _, bestWindow := range bestWindows {
				block := pfinder.ConfigBlock(
					name, bestWindow, start, stop-1,
					windows.UseFullRange(bestWindow, aln, nex.Letters()),
				)
				pFinderConfigBlocks = append(pFinderConfigBlocks, block)
			}
		}
		alnSites := make([]int, stop-start)
		for i := range alnSites {
			alnSites[i] = i + start
		}
		writers.WriteOutput(out, bestWindows, metVals, alnSites, name)
		bar.Increment()
	}
	bar.FinishPrint("Finished processing UCEs")

	if *cfg != "" {
		pfinderFile, err := os.Create(*cfg)
		defer pfinderFile.Close()
		if err != nil {
			ui.Errorf("Could not create PartitionFinder2 file: %s", err)
		}
		block := pfinder.StartBlock(strings.TrimRight(path.Base(*read), ".nex"))
		if _, err := io.WriteString(pfinderFile, block); err != nil {
			ui.Errorf("Failed to write PartitionFinder2 start block: %s", err)
		}
		for _, b := range pFinderConfigBlocks {
			if _, err := io.WriteString(pfinderFile, b); err != nil {
				ui.Errorf("Failed to write PartitionFinder2 config block: %s", err)
			}
		}
		block = pfinder.EndBlock()
		if _, err := io.WriteString(pfinderFile, block); err != nil {
			ui.Errorf("Failed to write PartitionFinder2 end block: %s", err)
		}
	}

	// Inform user of where output was written
	fmt.Println(ui.Footer(*write))
}
