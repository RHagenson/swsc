package ui

import (
	"fmt"
	"path"
)

// Header informs the user what work is being performed
func Header(f string) {
	fmt.Println()
	fmt.Printf("Analysing %s\n", path.Base(f))
}

// Footer informs the user of what work was just performed
func Footer(f string) {
	fmt.Println()
	fmt.Printf("Wrote partitions to %s\n", f)
}
