package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"math"
	"os"
	"path"
	"strings"

	"bitbucket.org/rhagenson/swsc/nexus"
	"github.com/biogo/biogo/seq/linear"
	"github.com/biogo/biogo/seq/multi"
	"github.com/spf13/pflag"
	"gonum.org/v1/gonum/mat"
	"gonum.org/v1/gonum/stat"
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
	metrics         = []string{"entropy"} // Possible entries: entropy, gc, or multi
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

func main() {
	setup()
	printHeader(*read)
	file, err := os.Open(*read)
	defer file.Close()
	if err != nil {
		log.Fatalf("Could not read file: %s", err)
	}
	writeOutputHeader()
	if *cfg {
		for _, m := range metrics {
			writeCfgStartBlock(pFinderFileName, datasetName)
		}
	}
	nex := nexus.New()
	nex.Read(file)
	partitions := processDatasetMetrics(nex, []string{"entropy"}, *minWin)
	writeOutput(*write, partitions)
	printFooter(*write, len(partitions))
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

// processDatasetMetrics calculates defined metrics from a *nexus.Nexus
// using a minimum sliding window size
func processDatasetMetrics(nex *nexus.Nexus, metrics []string, win int) {
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
		bestWindows, metricArray := processUce(uceAln, metrics, minWin)
		if *cfg {
			pFinderConfig, err := os.OpenFile(
				pFinderFileName,
				os.O_APPEND|os.O_WRONLY, 0644,
			)
			defer pFinderConfig.Close()
			if err != nil {
				log.Fatalf("Could not append to PartitionFinder2 file: %s", err)
			}
			for i, bestWindow := range bestWindows {
				pFinderConfig.Write(pFinderConfigBlock(bestWindow, name, start, stop, uceAln))
			}
		}
		writeOutput(*write, [][]string{bestWindows, metricArray, sites, name})
		bar.Increment()
	}
	bar.FinishPrint("Finished processing UCEs")
	if *cfg {
		for _, m := range metrics {
			writeCfgEndBlock(pFinderFileName, datasetName)
		}
	}
	return
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
func pFinderConfigBlock(pFinderFileName, name string, bestWindow [2]int, start, stop int, uceAln multi.Multi) {
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

// anyUndeterminedBlocks checks if any blocks are only undetermined/ambiguous characters
// Not the same as anyBlocksWoAllSites()
func anyUndeterminedBlocks(bestWindow [2]int, uceAln multi.Multi) bool {
	leftAln, _ := uceAln.Subseq(0, bestWindow[0])
	coreAln, _ := uceAln.Subseq(bestWindow[0], bestWindow[1])
	rightAln, _ := uceAln.Subseq(bestWindow[1], uceAln.Len())

	leftFreq := bpFreqCalc(leftAln)
	coreFreq := bpFreqCalc(coreAln)
	rightFreq := bpFreqCalc(rightAln)

	// If any frequency is NaN
	// TODO: Likely better with bpFreqCalc returning an error value
	if maxInFreqMap(leftFreq) == 0 || maxInFreqMap(coreFreq) == 0 || maxInFreqMap(rightFreq) == 0 {
		return true
	}
	return false
}

func bpFreqCalc(aln *multi.Multi) map[byte]float32 {
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

// anyBlocksWoAllSites checks for blocks with only undetermined/ambiguous characters
// Not the same as anyUndeterminedBlocks()
func anyBlocksWoAllSites(bestWindow [2]int, uceAln *multi.Multi) bool {
	leftAln, _ := uceAln.Subseq(0, bestWindow[0])
	coreAln, _ := uceAln.Subseq(bestWindow[0], bestWindow[1])
	rightAln, _ := uceAln.Subseq(bestWindow[1], uceAln.Len())

	leftCounts := countBases(leftAln)
	coreCounts := countBases(coreAln)
	rightCounts := countBases(rightAln)

	if minInCountsMap(leftCounts) == 0 || minInCountsMap(coreCounts) == 0 || minInCountsMap(rightCounts) == 0 {
		return true
	}
	return false
}

func countBases(aln *multi.Multi) map[byte]int {
	counts := map[byte]int{
		'A': 0,
		'T': 0,
		'G': 0,
		'C': 0,
	}
	allSeqs := ""
	for _, seq := range aln.Seq {
		rSeq := seq.(*linear.Seq)
		allSeqs += rSeq.String()
	}
	for _, char := range allSeqs {
		counts[byte(char)]++
	}
	return counts
}

func minInCountsMap(counts map[byte]int) int {
	min := math.MaxInt16
	for _, val := range counts {
		if val < min {
			min = val
		}
	}
	return min
}

func maxInFreqMap(freqs map[byte]float32) float32 {
	max := float32(math.SmallestNonzeroFloat32)
	for _, val := range freqs {
		if max < val {
			max = val
		}
	}
	return max
}

func getMinVarWindow(windows [][2]int, alnLength int) [2]int {
	best := float64(math.MaxInt16)
	bestWindow := windows[0]

	for _, w := range windows {
		l1 := float64(w[0])
		l2 := float64(w[1] - w[0])
		l3 := float64(alnLength - w[0])
		variance := stat.Variance([]float64{l1, l2, l3}, nil)
		if variance < best {
			best = variance
			bestWindow = w
		}
	}
	return bestWindow
}

// alignmentEntropy calculates entropies of characters
func alignmentEntropy(aln *multi.Multi) map[byte]float32 {
	bpFreq := bpFreqCalc(aln)
	entropy := entropyCalc(bpFreq)
	return entropy
}

func entropyCalc(bpFreqs map[byte]float32) map[byte]float32 {
	freqs := make([]float64, len(bpFreqs))
	i := 0
	for _, val := range bpFreqs {
		freqs[i] = float64(val)
		i++
	}

	// Should be equivalent to:
	// p = p[p!=0]
	// numpy.dot(-p,np.log2(p))
	return mat.Dot(mat.NewVecDense(len(bpFreqs), freqs))
}
