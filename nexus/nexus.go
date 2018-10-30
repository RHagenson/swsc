package nexus

import (
	"bufio"
	"io"
	"log"
)

// Nexus only understands two blocks: DATA and SETS
// Note: meant for exclusive use in swsc
type Nexus struct {
	handlers map[string]func([]string, *Nexus)
	data     *dataBlock
	sets     *setsBlock
}

// New creates a new empty Nexus with registered handlers and deferred block creation
func New() *Nexus {
	nex := &Nexus{
		handlers: map[string]func([]string, *Nexus){
			"DATA": processDataBlock,
			"SETS": processSetsBlock,
		},
		data: new(dataBlock),
		sets: new(setsBlock),
	}
	return nex
}

// Read reads a Nexus file from a reader returning the filled Nexus
func Read(file io.Reader) *Nexus {
	nex := New()
	nex.FillFrom(file)
	return nex
}

// FillFrom fills in the Nexus with data from a file
// It overwrites existing values, but does not clear all values
func (nex *Nexus) FillFrom(file io.Reader) {
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		handled := false
		for k, f := range nex.handlers {
			if blockByName(scanner.Text(), k) {
				lines := copyLines(scanner)
				f(lines, nex)
				handled = true
			}
		}
		if !handled && scanner.Text() != "" {
			log.Printf("Scanner ignored line:\n%q\n", scanner.Text())
		}
	}
	return
}

// NTax is the number of taxa
func (nex *Nexus) NTax() int {
	return nex.data.ntax
}

// NChar is the number of characters
func (nex *Nexus) NChar() int {
	return nex.data.nchar
}

// DataType is the type of data (e.g. DNA, RNA, Nucleotide, Protein)
func (nex *Nexus) DataType() string {
	return nex.data.dataType
}

// Gap is the character used to represent a gap
func (nex *Nexus) Gap() byte {
	return nex.data.gap
}

// Missing is the character used to represent a missing element
func (nex *Nexus) Missing() byte {
	return nex.data.missing
}

// Charsets returns a copy of the internal character sets
func (nex *Nexus) Charsets() map[string][]Pair {
	copy := make(map[string][]Pair)
	for k, v := range nex.sets.charSets {
		copy[k] = v
	}
	return copy
}

// Alignment returns a copy of the internal alignment
func (nex *Nexus) Alignment() Alignment {
	return nex.data.alignment
}
