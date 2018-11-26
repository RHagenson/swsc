package nexus

import (
	"fmt"
)

// Alignment is a collection of equal length sequences
// Appends missing characters (see Nexus.Missing()) to shorter sequences
type Alignment []string

// Column is the letters from each internal sequence at position p
func (aln Alignment) Column(p uint) []byte {
	pos := make([]byte, aln.NSeq())
	for i := uint(0); i < aln.NSeq(); i++ {
		pos[i] = aln.Seq(i)[p]
	}
	return pos
}

// NSeq is the number of sequences in the alignment
// Note: len(alignment) == alignment.NSeq()
func (aln Alignment) NSeq() uint {
	return uint(len(aln))
}

// Seq returns the i-th sequence in the alignment
func (aln Alignment) Seq(i uint) string {
	return aln[i]
}

// Subseq creates a slice from the original alignment
// A negative argument is interpreted as ultimate start or end of alignment
func (aln Alignment) Subseq(s, e int) Alignment {
	subseqs := make(Alignment, 0)
	for _, seq := range aln {
		switch {
		case s < 0 && e < 0: // Whole alignment
			subseqs = append(subseqs, seq[:])
		case s < 0 && e <= aln.Len(): // Start to defined end
			subseqs = append(subseqs, seq[:e])
		case s < aln.Len() && e < 0: // Defined start to end
			subseqs = append(subseqs, seq[s:])
		case e <= aln.Len() && s < e: // Defined start to defined end
			subseqs = append(subseqs, seq[s:e])
		default:
			msg := fmt.Sprintf("Requested out of bounds slice, "+
				"bounds [%d:%d], requested [%d:%d]",
				0, len(aln.Seq(0)), s, e)
			panic(msg)
		}
	}
	return subseqs
}

// String is al sequences with a newline after each sequence
func (aln Alignment) String() string {
	str := ""
	for i := range aln {
		str += aln[i] + "\n"
	}
	return str
}

// Len is the length of the alignment
// Note that len(alignment) != alignment.Len(), the former equals alignment.NSeq()
func (aln Alignment) Len() (length int) {
	for i := range aln {
		if length < len(aln[i]) {
			length = len(aln[i])
		}
	}
	return
}

func (aln Alignment) BasesCount(bases []byte) map[byte]int {
	counts := make(map[byte]int)
	allSeqs := aln.String()
	for _, char := range allSeqs {
		counts[byte(char)]++
	}
	return counts
}

func (aln Alignment) BasesFrequency(bases []byte) map[byte]float64 {
	freqs := make(map[byte]float64)
	baseCounts := aln.BasesCount(bases)
	sumCounts := 0.0
	for _, count := range baseCounts {
		sumCounts += float64(count)
	}
	if sumCounts == 0 {
		sumCounts = 1.0
	}
	for char, count := range baseCounts {
		freqs[char] = float64(count) / sumCounts
	}
	return freqs
}
