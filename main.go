package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
)

// Encodings
const (
	entropy int8 = iota
)

// General use flags
var (
	read   = flag.String("input", "", "Nexus (.nex) file to process")
	write  = flag.String("output", "", "Partition file to write")
	minWin = flag.Int("minwindow", 50, "Minimum window size")
	help   = flag.Bool("h", false, "Print help and exit")
)

func setup() {
	flag.Parse()
	if *help {
		flag.Usage()
		os.Exit(1)
	}
	confirmArgs(*read, *write, *minWin)
}

func main() {
	setup()

	printHeader(*read)
	partitions := processDatasetMetrics(*read, []int8{entropy}, *minWin)
	writeOutput(*write, partitions)
	printFooter(*write, len(partitions))
}

func printHeader(f string) {
	fmt.Println()
	fmt.Printf("Analysing %s\n", path.Base(f))
}

func writeOutput(f string, d []byte) {
	if stat, err := os.Stat(f); err != nil && !stat.IsDir() {
		ioutil.WriteFile(f, d, os.ModeType)
	} else {
		log.Fatal(err)
	}
}

func confirmArgs(in, out string, win int) {
	switch {
	case in == "" && out == "":
		flag.Usage()
		log.Fatalf("Must provide input and output names")
	case !strings.HasSuffix(in, ".nex"):
		log.Fatalf("Input expected .nex format, got %s format", path.Ext(in))
	case !strings.HasSuffix(out, ".csv"):
		log.Fatalf("Output expected .csv format, got %s format", path.Ext(out))
	case win > 0:
		log.Fatalf("Window size must be positive, got %d", win)
	default:
		fmt.Printf("Arguments are reasonable")
	}
	return
}

func processDatasetMetrics(nex string, props []int8, win int) []byte {
	return make([]byte, 0)
}

func printFooter(f string, l int) {
	fmt.Println()
	fmt.Printf("Wrote %d paritions to %s\n", l, f)
}
