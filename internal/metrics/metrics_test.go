package metrics_test

import (
	"testing"

	"github.com/rhagenson/swsc/internal/metrics"
	"github.com/rhagenson/swsc/internal/nexus"
	"gonum.org/v1/gonum/floats"
)

func TestMetric(t *testing.T) {
	tt := []struct {
		met  metrics.Metric
		name string
	}{
		{metrics.Entropy, "Entropy"},
		{metrics.GC, "GC"},
		{metrics.Multi, "Multinomial"},
		{metrics.Entropy ^ metrics.GC ^ metrics.Multi, ""},
	}

	for _, tc := range tt {
		if tc.met.String() != tc.name {
			t.Errorf("Got: %s, expected: %s", tc.met.String(), tc.name)
		}
	}
}

func TestSitewiseEntropy(t *testing.T) {
	tt := []struct {
		aln   nexus.Alignment
		chars []byte
		exp   []float64
	}{
		{ // All the same seq
			nexus.Alignment([]string{
				"ATGCAT",
				"ATGCAT",
				"ATGCAT",
			}),
			[]byte("ATGC"),
			[]float64{
				0,
				0,
				0,
				0,
				0,
				0,
			},
		},
		{ // One seq reversed
			nexus.Alignment([]string{
				"ATGCAT",
				"ATGCAT",
				"TACGTA",
			}),
			[]byte("ATGC"),
			[]float64{
				0.6365141682948128,
				0.6365141682948128,
				0.6365141682948128,
				0.6365141682948128,
				0.6365141682948128,
				1.3296613488547580,
			},
		},
		{ // No positional matches
			nexus.Alignment([]string{
				"ATGCAT",
				"TGCTGA",
				"GCTACC",
			}),
			[]byte("ATGC"),
			[]float64{
				1.0986122886681096,
				1.0986122886681096,
				1.0986122886681096,
				1.0986122886681096,
				1.0986122886681096,
				1.3801087571572686,
			},
		},
		{ // Not all bases present
			nexus.Alignment([]string{
				"CCCCCC",
				"TTTTTT",
				"GGGGGG",
			}),
			[]byte("ATGC"),
			[]float64{
				1.098612288668110,
				1.098612288668110,
				1.098612288668110,
				1.098612288668110,
				1.098612288668110,
				1.098612288668110,
			},
		},
		{ // All the same base
			nexus.Alignment([]string{
				"CCCCCC",
				"CCCCCC",
				"CCCCCC",
			}),

			[]byte("ATGC"),
			[]float64{
				0.0,
				0.0,
				0.0,
				0.0,
				0.0,
				0.0,
			},
		},
		{ // More than one possible window
			nexus.Alignment([]string{
				"CCCCCCCC",
				"CCCCCCCC",
				"CCCCCCCC",
			}),
			[]byte("ATGC"),
			[]float64{
				0.0,
				0.0,
				0.0,
				0.0,
				0.0,
				0.0,
				0.0,
				0.0,
			},
		},
	}

	for _, tc := range tt {
		got := metrics.SitewiseEntropy(&tc.aln, tc.chars)
		t.Run("Length", func(t *testing.T) {
			if len(got) != len(tc.exp) {
				t.Errorf("Lengths do not match. Got %d, expected %d",
					len(got), len(tc.exp),
				)
			}
		})
		t.Run("Match", func(t *testing.T) {
			for i := range got {
				if !floats.EqualWithinAbs(got[i], tc.exp[i], 1e10-5) {
					t.Errorf("Got %.5f, expected %.5f", got[i], tc.exp[i])
				}
			}
		})
	}
}

func TestSitewiseBaseCounts(t *testing.T) {
	tt := []struct {
		aln   nexus.Alignment
		chars []byte
		exp   map[byte][]int
	}{
		{ // All the same seq
			nexus.Alignment([]string{
				"ATGCAT",
				"ATGCAT",
				"ATGCAT",
			}),
			[]byte("ATGC"),
			map[byte][]int{
				'A': []int{3, 0, 0, 0, 3, 0},
				'T': []int{0, 3, 0, 0, 0, 3},
				'G': []int{0, 0, 3, 0, 0, 0},
				'C': []int{0, 0, 0, 3, 0, 0},
			},
		},
		{ // One seq reversed
			nexus.Alignment([]string{
				"ATGCAT",
				"ATGCAT",
				"TACGTA",
			}),
			[]byte("ATGC"),
			map[byte][]int{
				'A': []int{2, 1, 0, 0, 2, 1},
				'T': []int{1, 2, 0, 0, 1, 2},
				'G': []int{0, 0, 2, 1, 0, 0},
				'C': []int{0, 0, 1, 2, 0, 0},
			},
		},
		{ // No positional matches
			nexus.Alignment([]string{
				"ATGCAT",
				"TGCTGA",
				"GCTACC",
			}),
			[]byte("ATGC"),
			map[byte][]int{
				'A': []int{1, 0, 0, 1, 1, 1},
				'T': []int{1, 1, 1, 1, 0, 1},
				'G': []int{1, 1, 1, 0, 1, 0},
				'C': []int{0, 1, 1, 1, 1, 1},
			},
		},
		{ // Not all bases present
			nexus.Alignment([]string{
				"CCCCCC",
				"TTTTTT",
				"GGGGGG",
			}),
			[]byte("ATGC"),
			map[byte][]int{
				'A': []int{0, 0, 0, 0, 0, 0},
				'T': []int{1, 1, 1, 1, 1, 1},
				'G': []int{1, 1, 1, 1, 1, 1},
				'C': []int{1, 1, 1, 1, 1, 1},
			},
		},
		{ // All the same base
			nexus.Alignment([]string{
				"CCCCCC",
				"CCCCCC",
				"CCCCCC",
			}),

			[]byte("ATGC"),
			map[byte][]int{
				'A': []int{0, 0, 0, 0, 0, 0},
				'T': []int{0, 0, 0, 0, 0, 0},
				'G': []int{0, 0, 0, 0, 0, 0},
				'C': []int{3, 3, 3, 3, 3, 3},
			},
		},
		{ // More than one possible window
			nexus.Alignment([]string{
				"CCCCCCCC",
				"CCCCCCCC",
				"CCCCCCCC",
			}),
			[]byte("ATGC"),
			map[byte][]int{
				'A': []int{0, 0, 0, 0, 0, 0, 0, 0},
				'T': []int{0, 0, 0, 0, 0, 0, 0, 0},
				'G': []int{0, 0, 0, 0, 0, 0, 0, 0},
				'C': []int{3, 3, 3, 3, 3, 3, 3, 3},
			},
		},
	}

	for _, tc := range tt {
		got := metrics.SitewiseBaseCounts(&tc.aln, tc.chars)
		t.Run("Length", func(t *testing.T) {
			if len(got) != len(tc.exp) {
				t.Errorf("Lengths do not match. Got %d, expected %d",
					len(got), len(tc.exp),
				)
			}
		})
		t.Run("Match", func(t *testing.T) {
			for b := range got {
				for i := range got[b] {
					if got[b][i] != tc.exp[b][i] {
						t.Errorf("Got %d, expected %d", got[b][i], tc.exp[b][i])
					}
				}
			}
		})
	}
}

func TestSitewiseGc(t *testing.T) {
	tt := []struct {
		aln nexus.Alignment
		exp []float64
	}{
		{ // All the same seq
			nexus.Alignment([]string{
				"ATGCAT",
				"ATGCAT",
				"ATGCAT",
			}),
			[]float64{
				0.0 / 3.0,
				0.0 / 3.0,
				3.0 / 3.0,
				3.0 / 3.0,
				0.0 / 3.0,
				0.0 / 3.0,
			},
		},
		{ // One seq reversed
			nexus.Alignment([]string{
				"ATGCAT",
				"ATGCAT",
				"TACGTA",
			}),
			[]float64{
				0.0 / 3.0,
				0.0 / 3.0,
				3.0 / 3.0,
				3.0 / 3.0,
				0.0 / 3.0,
				0.0 / 3.0,
			},
		},
		{ // No positional matches
			nexus.Alignment([]string{
				"ATGCAT",
				"TGCTGA",
				"GCTACC",
			}),
			[]float64{
				1.0 / 3.0,
				2.0 / 3.0,
				2.0 / 3.0,
				1.0 / 3.0,
				2.0 / 3.0,
				1.0 / 3.0,
			},
		},
		{ // Not all bases present
			nexus.Alignment([]string{
				"CCCCCC",
				"TTTTTT",
				"GGGGGG",
			}),
			[]float64{
				2.0 / 3.0,
				2.0 / 3.0,
				2.0 / 3.0,
				2.0 / 3.0,
				2.0 / 3.0,
				2.0 / 3.0,
			},
		},
		{ // All the same base
			nexus.Alignment([]string{
				"CCCCCC",
				"CCCCCC",
				"CCCCCC",
			}),
			[]float64{
				3.0 / 3.0,
				3.0 / 3.0,
				3.0 / 3.0,
				3.0 / 3.0,
				3.0 / 3.0,
				3.0 / 3.0,
			},
		},
		{ // More than one possible window
			nexus.Alignment([]string{
				"CCCCCCCC",
				"CCCCCCCC",
				"CCCCCCCC",
			}),
			[]float64{
				3.0 / 3.0,
				3.0 / 3.0,
				3.0 / 3.0,
				3.0 / 3.0,
				3.0 / 3.0,
				3.0 / 3.0,
				3.0 / 3.0,
				3.0 / 3.0,
			},
		},
	}

	for _, tc := range tt {
		got := metrics.SitewiseGc(&tc.aln)
		t.Run("Length", func(t *testing.T) {
			if len(got) != len(tc.exp) {
				t.Errorf("Lengths do not match. Got %d, expected %d",
					len(got), len(tc.exp),
				)
			}
		})
		t.Run("Match", func(t *testing.T) {
			for i := range got {
				if !floats.EqualWithinAbs(got[i], tc.exp[i], 1e10-5) {
					t.Errorf("Got %.5f, expected %.5f", got[i], tc.exp[i])
				}
			}
		})
	}
}
