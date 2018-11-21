package ui

import (
	"fmt"
	"path"
)

// Header informs the user what work is being performed
func Header(f string) {
	fmt.Printf("\nAnalysing %s\n", path.Base(f))
}

// Footer informs the user of what work was just performed
func Footer(f string) {
	fmt.Printf("\nWrote partitions to %s\n", f)
}
