package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/rhagenson/swsc/internal/metrics"
	"github.com/rhagenson/swsc/internal/nexus"
	"github.com/rhagenson/swsc/internal/pfinder"
	"github.com/rhagenson/swsc/internal/uce"
	"github.com/rhagenson/swsc/internal/ui"
	"github.com/rhagenson/swsc/internal/utils"
	"github.com/rhagenson/swsc/internal/windows"
	"github.com/rhagenson/swsc/internal/writers"
	"github.com/shenwei356/bio/seq"
	"github.com/shenwei356/bio/seqio/fastx"
	"github.com/spf13/pflag"
	pb "gopkg.in/cheggaaa/pb.v1"
)

// Required flags
var (
	fNex    = pflag.String("nexus", "", "Nexus file to process (.nex)")
	fFasta  = pflag.String("fasta", "", "Multi-FASTA file to process (.fna/fasta)")
	fUces   = pflag.String("uces", "", "CSV file with UCE ranges, format: Name,Start,Stop (inclusive)")
	fOutput = pflag.String("output", "", "Partition file to write (.csv)")
	fCfg    = pflag.String("cfg", "", "Config file for PartionFinder2 (.cfg)")
)

// General use flags
var (
	fMinWin      = pflag.Uint("minWin", 50, "Minimum window size")
	fLargeCore   = pflag.Bool("largeCore", false, "When a small and large core have equivalent metrics, choose the large core")
	fNCandidates = pflag.Uint("candidates", 3, "Number of best candidates to search with")
	fHelp        = pflag.Bool("help", false, "Print help and exit")
)

// Metric flags
var (
	fEntropy = pflag.Bool("entropy", false, "Calculate Shannon's entropy metric")
	fGc      = pflag.Bool("gc", false, "Calculate GC content metric")
	// multi = pflag.Bool("multi", false, "Calculate multinomial distribution metric")
)

func setup() {
	pflag.Parse()
	if *fHelp {
		pflag.Usage()
		os.Exit(1)
	}

	// Failure states
	switch {
	case (*fNex == "") != (*fFasta == "" && *fUces == ""):  // != used as XOR
		pflag.Usage()
		ui.Errorf("Must provide either nexus, or fasta and uces\n")
	case *fOutput == "":
		pflag.Usage()
		ui.Errorf("Must provide output\n")
	case *fNex != "" && !strings.HasSuffix(*fNex, ".nex"):
		ui.Errorf("Input expected to end in .nex, got %s\n", path.Ext(*fNex))
	case *fFasta != "" && !(strings.HasSuffix(*fFasta, ".fna") || strings.HasSuffix(*fFasta, ".fasta")):
		ui.Errorf("FASTA expected to end in .fna, got %s\n", path.Ext(*fFasta))
	case *fUces != "" && !strings.HasSuffix(*fUces, ".csv"):
		ui.Errorf("UCEs expected to end in .csv, got %s\n", path.Ext(*fUces))
	case *fOutput != "" && !strings.HasSuffix(*fOutput, ".csv"):
		ui.Errorf("Output expected to end in .csv, got %s\n", path.Ext(*fOutput))
	case *fCfg != "" && !strings.HasSuffix(*fCfg, ".cfg"):
		ui.Errorf("Config expected to end in .cfg, got %s\n", path.Ext(*fCfg))
	case *fEntropy == *fGc && (*fEntropy || *fGc):
		ui.Errorf("Only one metric is allowed\n")
	case !(*fEntropy || *fGc):
		ui.Errorf("At least one metric is needed\n")
	}
}

func main() {
	// Parse CLI arguments
	setup()

	var (
		aln     = new(nexus.Alignment)             // Sequence alignment
		uces    = make(map[string][]nexus.Pair, 0) // UCE set
		letters []byte                             // Valid letters in Alignment
	)

	switch {
	case *fNex != "": // Nexus input, all in one input
		in, err := os.Open(*fNex)
		defer in.Close()
		if err != nil {
			ui.Errorf("Could not read input file: %s", err)
		}

		// Read in the input Nexus file
		nex := nexus.Read(in)
		*aln = nex.Alignment()
		uces = nex.Charsets()
		letters = nex.Letters()
	case *fFasta != "" && *fUces != "": // FASTA and UCE input,
		fna, err := fastx.NewDefaultReader(*fFasta)
		defer fna.Close()
		if err != nil {
			ui.Errorf("Could not read input file: %s", err)
		}
		seqs := make([]string, 0)
		for {
			record, err := fna.Read()
			if err != nil {
				if err == io.EOF {
					break
				}
				ui.Errorf("Failed parsing FASTA: %v", err)
			}
			seqs = append(seqs, record.Seq.String())
		}
		*aln = nexus.Alignment(seqs)
		letters = seq.DNA.Letters()

		inUce, err := os.Open(*fUces)
		defer inUce.Close()
		if err != nil {
			ui.Errorf("Could not read input file: %s", err)
		}
		uceCsv := csv.NewReader(inUce)
		colMap := make(map[string]int, 3)
		header, err := uceCsv.Read()
		for i, col := range header {
			switch col {
			case "Name":
				colMap["Name"] = i
			case "Start":
				colMap["Start"] = i
			case "Stop":
				colMap["Stop"] = i
			default:
				ui.Errorf("Did not understand column %q", col)
			}
		}
		rows, err := uceCsv.ReadAll()
		for _, row := range rows {
			start, err := strconv.Atoi(row[colMap["Start"]])
			if err != nil {
				ui.Errorf("Failed to read Start in UCE row: %q", row)
			}
			stop, err := strconv.Atoi(row[colMap["Stop"]])
			if err != nil {
				ui.Errorf("Failed to read Stop in UCE row: %q", row)
			}
			uces[row[colMap["Name"]]] = append(uces[row[colMap["Name"]]],
				nexus.NewPair(
					start,
					stop+1, // Inclusive range in file, while exclusive range used internally
				),
			)
		}
	default:
		ui.Errorf("Did not understand how to read input")
	}

	out, err := os.Create(*fOutput)
	defer out.Close()
	if err != nil {
		ui.Errorf("Could not create output file: %s", err)
	}

	writers.WriteOutputHeader(out)

	var (
		bar     = pb.StartNew(len(uces)) // Progress bar
		metVals = make(map[metrics.Metric][]float64, 3)
	)

	// Early panic if minWin has been set too large to create flanks and core of that length
	if err := utils.ValidateMinWin(aln.Len(), int(*fMinWin)); err != nil {
		ui.Errorf("Failed due to: %v", err)
	}

	if *fEntropy {
		metVals[metrics.Entropy] = metrics.SitewiseEntropy(aln, letters)
	}
	if *fGc {
		metVals[metrics.GC] = metrics.SitewiseGc(aln)
	}

	// Sort UCEs
	// Create reverse lookup to maintain order
	revUCEs := make(map[int]string, len(uces))
	keys := make([]int, 0, len(uces))
	for name, sites := range uces {
		var (
			start = math.MaxInt64 // Minimum position in UCE
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
	pFinderConfigBlocks := make([]string, len(uces))
	outputFrames := make([][][]string, len(uces))
	sem := make(chan struct{}, len(uces))
	uceNum := 0
	for _, key := range keys {
		go func(key, uceNum int) {
			name := revUCEs[key]
			sites := uces[name]
			var (
				start = sites[0].First()  // Minimum position in UCE
				stop  = sites[0].Second() // Maximum position in UCE
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
			bestWindows := uce.ProcessUce(start, stop, metVals, *fMinWin, letters, *fLargeCore, *fNCandidates)
			if *fCfg != "" {
				for _, bestWindow := range bestWindows {
					block := pfinder.ConfigBlock(
						name, bestWindow, start, stop-1,
						windows.UseFullRange(bestWindow, aln, letters),
					)
					pFinderConfigBlocks[uceNum] = block
				}
			}
			alnSites := make([]int, stop-start)
			for i := range alnSites {
				alnSites[i] = i + start
			}
			frame := writers.Output(bestWindows, metVals, alnSites, name)
			outputFrames[uceNum] = frame
			sem <- struct{}{}

			bar.Increment()
		}(key, uceNum)
		uceNum++
	}
	for i := 0; i < len(uces); i++ {
		<-sem
	}
	bar.FinishPrint("Finished processing UCEs")

	if *fCfg != "" {
		pfinderFile, err := os.Create(*fCfg)
		defer pfinderFile.Close()
		if err != nil {
			ui.Errorf("Could not create PartitionFinder2 file: %s", err)
		}
		block := pfinder.StartBlock(strings.TrimRight(path.Base(*fNex), ".nex"))
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

	outCsv := csv.NewWriter(out)
	for _, s := range outputFrames {
		if err := outCsv.WriteAll(s); err != nil {
			ui.Errorf("Failed to write output: %v", err)
		}
	}

	// Inform user of where output was written
	fmt.Println(ui.Footer(*fOutput))
}
