package main

import (
	"fmt"
	"path"
)

// printHeader informs the user what work is being performed
func printHeader(f string) {
	fmt.Println()
	fmt.Printf("Analysing %s\n", path.Base(f))
}

// printFooter informs the user of what work was just performed with helpful metrics
func printFooter(f string, l int) {
	fmt.Println()
	fmt.Printf("Wrote %d partitions to %s\n", l, f)
}
