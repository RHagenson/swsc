package nexus

import (
	"bufio"
	"strings"
)

// splitOnChar is meant to be used with strings.FieldsFunc(str, splitOnChar(...))
// It defines fields based on splitting a string on a given rune
func splitOnChar(c rune) func(rune) bool {
	return func(d rune) bool {
		return d == c
	}
}

// blockByName is whether a given line contains the start of a named block
func blockByName(line, name string) bool {
	return strings.HasSuffix(strings.ToUpper(line), strings.ToUpper(" "+name+";"))
}

// copyLines extracts the lines between "BEGIN <block name>;" and "END;", trimming whitespace in the process
// The scanner should be at the BEGIN line and will be on the END line upon return
func copyLines(s *bufio.Scanner) []string {
	lines := make([]string, 0)
	for s.Scan() {
		if strings.HasPrefix(strings.ToUpper(s.Text()), "END;") {
			return lines
		}
		lines = append(lines, strings.TrimSpace(s.Text()))
	}
	return lines
}
