package entropy_test

import (
	"testing"

	"bitbucket.org/rhagenson/swsc/internal/entropy"
	"gonum.org/v1/gonum/floats"
)

func TestAlignmentEntropy(t *testing.T) {
	tt := []struct {
		seqs []string
		exp  float64
	}{
		{[]string{"ATGC", "ATGC", "ATGC"}, 1.386294}, // All the same
		{[]string{"ATGC", "CGTA", "ATGC"}, 1.386294}, // One reversed
		{[]string{"ATAT", "GCGC", "TGTG"}, 1.329661}, // No positional matches
		{[]string{"GGGG", "TTTT", "CCCC"}, 1.098612}, // Not all bases present
		{[]string{"CCCC", "CCCC", "CCCC"}, 0.000000}, // All same base
	}

	for _, tc := range tt {
		got := entropy.AlignmentEntropy(tc.seqs, []byte("ATGC"))
		if !floats.EqualWithinAbs(got, tc.exp, 0.000001) {
			t.Errorf("Got %.6f, expected %.6f", got, tc.exp)
		}
	}
}
