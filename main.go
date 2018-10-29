package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"path"
	"strings"

	"bitbucket.org/rhagenson/swsc/nexus"
	"github.com/spf13/pflag"
	"gonum.org/v1/gonum/stat"
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
	cfg    = pflag.Bool("cfg", false, "Write file for PartionFinder2")
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
	pfinder         = new(os.File)
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
	printHeader(*read)

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
		pfinder, err = os.Create(pFinderFileName)
		if err != nil {
			log.Fatalf("Could not read file: %s", err)
		}
		writeCfgStartBlock(pfinder, datasetName)
	}

	// Read in the input Nexus file
	nex := nexus.New()
	nex.Read(in)

	// Early panic if minWin has been set too large to create flanks and core of that length
	length := nex.Alignment().Len()
	if length/3 <= *minWin {
		message := fmt.Sprintf(
			"minWin is too large, maximum allowed value is %d\n",
			length/3,
		)
		panic(message)
	}

	// Process the input with selected metrics and minimum window size, internally writing output files
	processDatasetMetrics(nex, metrics, *minWin, pfinder, out)

	// Inform user of where output was written
	printFooter(*write)
	if *cfg {
		pfinder.Close()
	}
}

// processDatasetMetrics calculates defined metrics from a *nexus.Nexus
// using a minimum sliding window size
func processDatasetMetrics(nex *nexus.Nexus, metrics []string, win int, pfinder, out *os.File) {
	// Initialize local variables
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
			if pair.Start() < start {
				start = pair.Start()
			}
			if stop < pair.Stop() {
				stop = pair.Stop()
			}
		}

		uceAln := aln.Subseq(start, stop+1) // Nexus UCE ranges are inclusive so a +1 adjustment is needed
		bestWindows, metricArray := processUce(uceAln, metrics, *minWin)
		if *cfg {
			for _, bestWindow := range bestWindows {
				pFinderConfigBlock(pfinder, name, bestWindow, start, stop, uceAln)
			}
		}
		writeOutput(out, bestWindows, metricArray, sites, name)
		bar.Increment()
	}
	bar.FinishPrint("Finished processing UCEs")
	if *cfg {
		for range metrics {
			writeCfgEndBlock(pfinder, datasetName)
		}
	}
	return
}

func factorialMatrix(vs map[byte][]int) []float64 {
	product := make([]float64, len(vs[0])) // vs['A'][i] * vs['T'][i] * vs['G'][i] * vs['C'][i]
	for i := range product {
		product[i] = 1.0
	}
	for i := range product {
		for nuc := range vs {
			product[i] *= factorial(vs[nuc][i])
		}
	}
	return product
}

// invariantSites streams across an alignment and calls sites invariant by their entropy
func invariantSites(aln nexus.Alignment) []bool {
	entropies := sitewiseEntropy(aln)
	calls := make([]bool, len(entropies))
	for i, v := range entropies {
		if v > 0 {
			calls[i] = true
		}
	}
	return calls
}

func allInvariantSites(vs []bool) bool {
	for _, v := range vs {
		if v {
			return false
		}
	}
	return true
}

// Compute Sum of Square Errors
func sse(vs []float64) float64 {
	mean := stat.Mean(vs, nil)
	total := 0.0
	for _, v := range vs {
		total += math.Pow((v / mean), 2)
	}
	return total
}

func getSse(metric []float64, win window, siteVar []bool) float64 {
	leftAln := allInvariantSites(siteVar[:win.Start()])
	coreAln := allInvariantSites(siteVar[win.Start():win.Stop()])
	rightAln := allInvariantSites(siteVar[win.Stop():])

	if leftAln || coreAln || rightAln {
		return math.MaxFloat64
	}

	left := sse(metric[:win.Start()])
	core := sse(metric[win.Start():win.Stop()])
	right := sse(metric[win.Stop():])
	return left + core + right
}

// getSses generalized getSse over each site window.
// metrics is [metric name][value index] arranged
func getSses(metrics map[string][]float64, win window, siteVar []bool) map[string][]float64 {
	// TODO: Preallocate array
	sses := make(map[string][]float64, len(metrics))
	for m := range metrics {
		sses[m] = append(sses[m], getSse(metrics[m], win, siteVar))
	}
	return sses
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

func bpFreqCalc(aln []string) map[byte]float32 {
	freqs := map[byte]float32{
		'A': 0.0,
		'T': 0.0,
		'C': 0.0,
		'G': 0.0,
	}
	baseCounts := countBases(aln)
	sumCounts := 0
	for _, count := range baseCounts {
		sumCounts += count
	}
	if sumCounts == 0 {
		sumCounts = 1
	}
	for char, count := range baseCounts {
		freqs[char] = float32(count / sumCounts)
	}
	return freqs
}

func countBases(aln nexus.Alignment) map[byte]int {
	counts := map[byte]int{
		'A': 0,
		'T': 0,
		'G': 0,
		'C': 0,
	}
	allSeqs := ""
	for _, seq := range aln {
		allSeqs += seq
	}
	for _, char := range allSeqs {
		counts[byte(char)]++
	}
	return counts
}
