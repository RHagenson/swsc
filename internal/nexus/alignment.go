package nexus

// Alignment is a collection of equal length sequences
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
// An argument out of bounds is interpreted as ultimate start or end of alignment, relatively
func (aln Alignment) Subseq(s, e int) Alignment {
	subseqs := make(Alignment, aln.NSeq())
	for i, seq := range aln {
		switch {
		case 0 <= s && s < aln.Len() && 0 <= e && e < aln.Len(): // Defined start to defined end
			subseqs[i] = seq[s:e]
		case s < aln.Len() && (e < 0 || e > aln.Len()): // Defined start to ultimate end
			subseqs[i] = seq[s:]
		case (s < 0 || s > aln.Len()) && e < aln.Len(): // Ultimate start to defined end
			subseqs[i] = seq[:e]
		default:
			subseqs[i] = seq[:] // Whole alignment
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

// Count returns the number of times each base in a set is found
func (aln Alignment) Count(bases []byte) map[byte]int {
	counts := make(map[byte]int)
	allSeqs := aln.String()
	for _, char := range allSeqs {
		counts[byte(char)]++
	}
	return counts
}

// Frequency returns the normalized frequency of each base in a set
// Note that bases that exist in the Alignment, but not in the set are ignored
func (aln Alignment) Frequency(bases []byte) map[byte]float64 {
	freqs := make(map[byte]float64, len(bases))
	baseCounts := aln.Count(bases)
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
