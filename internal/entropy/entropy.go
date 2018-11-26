package entropy

import (
	"bitbucket.org/rhagenson/swsc/internal/nexus"
	"gonum.org/v1/gonum/stat"
)

// AlignmentEntropy calculates entropies of characters
func AlignmentEntropy(aln nexus.Alignment, chars []byte) float64 {
	bpFreq := aln.Frequency(chars)
	entropy := entropyCalc(bpFreq)
	return entropy
}

// entropyCalc computes Shannon's entropy after removing elements equal to zero as Ln(0) == -Inf
func entropyCalc(bpFreqs map[byte]float64) float64 {
	freqs := make([]float64, 0)
	for _, val := range bpFreqs {
		// Ln(0) == -Inf, Shannon's entropy uses Ln()
		if val != 0 {
			freqs = append(freqs, float64(val))
		}
	}
	return stat.Entropy(freqs)
}
