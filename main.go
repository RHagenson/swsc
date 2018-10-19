package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"bitbucket.org/rhagenson/swsc/nexus"
	"github.com/spf13/pflag"
)

// Encodings
const (
	entropy int8 = iota
)

// General use flags
var (
	read   = pflag.String("input", "", "Nexus (.nex) file to process")
	write  = pflag.String("output", "", "Partition file to write")
	minWin = pflag.Int("minwindow", 50, "Minimum window size")
	cfg    = pflag.String("cfg", "", "cfg file for use by PartionFinder2")
	help   = pflag.Bool("help", false, "Print help and exit")
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
}

func main() {
	setup()
	printHeader(*read)
	partitions := processDatasetMetrics(*read, []int8{entropy}, *minWin)
	writeOutput(*write, partitions)
	printFooter(*write, len(partitions))
}

// printHeader informs the user what work is being performed
func printHeader(f string) {
	fmt.Println()
	fmt.Printf("Analysing %s\n", path.Base(f))
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
func processDatasetMetrics(f string, props []int8, win int) [][]string {
	file, err := os.Open(f)
	if err != nil {
		log.Fatalf("Could not read file: %s", err)
	}

	data := nexus.New()
	data.Read(f)
	aln := new(alignio).Read(f, alignIONexus)

	return make([][]string, 0)
}

// printFooter informs the user of what work was just performed with helpful metrics
func printFooter(f string, l int) {
	fmt.Println()
	fmt.Printf("Wrote %d partitions to %s\n", l, f)
}
