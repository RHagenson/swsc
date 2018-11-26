package ui

import (
	"fmt"
	"path"
)

// Header informs the user what work is being performed
func Header(f string) string {
	return fmt.Sprintf("\nAnalysing %s\n", path.Base(f))
}

// Footer informs the user of what work was just performed
func Footer(f string) string {
	return fmt.Sprintf("\nWrote partitions to %s\n", f)
}
