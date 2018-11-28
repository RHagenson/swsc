package uce_test

import (
	"testing"

	"bitbucket.org/rhagenson/swsc/internal/metrics"
	"bitbucket.org/rhagenson/swsc/internal/nexus"
	"bitbucket.org/rhagenson/swsc/internal/uce"
	"bitbucket.org/rhagenson/swsc/internal/windows"
	"gonum.org/v1/gonum/floats"
)

func TestProcessUce(t *testing.T) {
	tt := []struct {
		aln       nexus.Alignment
		metrics   []metrics.Metric
		minWin    int
		chars     []byte
		expWins   map[metrics.Metric]windows.Window
		expVals   map[metrics.Metric][]float64
		largeCore bool
	}{
		{ // All the same seq
			nexus.Alignment([]string{
				"ATGCAT",
				"ATGCAT",
				"ATGCAT",
			}),
			[]metrics.Metric{metrics.Entropy, metrics.GC},
			2,
			[]byte("ATGC"),
			map[metrics.Metric]windows.Window{
				metrics.Entropy: windows.Window{2, 5},
				metrics.GC:      windows.Window{2, 5},
			},
			map[metrics.Metric][]float64{
				metrics.Entropy: []float64{
					0.0,
					0.0,
					0.0,
					0.0,
					0.0,
					0.0,
				},
				metrics.GC: []float64{
					0,
					0,
					1,
					1,
					0,
					0,
				},
			},
			false,
		},
		{ // One seq reversed
			nexus.Alignment([]string{
				"ATGCAT",
				"ATGCAT",
				"TACGTA",
			}),
			[]metrics.Metric{metrics.Entropy, metrics.GC},
			2,
			[]byte("ATGC"),
			map[metrics.Metric]windows.Window{
				metrics.Entropy: windows.Window{2, 5},
				metrics.GC:      windows.Window{2, 5},
			},
			map[metrics.Metric][]float64{
				metrics.Entropy: []float64{
					0.6365141682948128,
					0.6365141682948128,
					0.6365141682948128,
					0.6365141682948128,
					0.6365141682948128,
					0.6365141682948128,
				},
				metrics.GC: []float64{
					0 / 3,
					0 / 3,
					3.0 / 3.0,
					3.0 / 3.0,
					0 / 3,
					0 / 3,
				},
			},
			false,
		},
		{ // No positional matches
			nexus.Alignment([]string{
				"ATGCAT",
				"TGCTGA",
				"GCTACC",
			}),
			[]metrics.Metric{metrics.Entropy, metrics.GC},
			2,
			[]byte("ATGC"),
			map[metrics.Metric]windows.Window{
				metrics.Entropy: windows.Window{2, 5},
				metrics.GC:      windows.Window{2, 5},
			},
			map[metrics.Metric][]float64{
				metrics.Entropy: []float64{
					1.0986122886681096,
					1.0986122886681096,
					1.0986122886681096,
					1.0986122886681096,
					1.0986122886681096,
					1.0986122886681096,
				},
				metrics.GC: []float64{
					1.0 / 3.0,
					2.0 / 3.0,
					2.0 / 3.0,
					1.0 / 3.0,
					2.0 / 3.0,
					1.0 / 3.0,
				},
			},
			false,
		},
		{ // Not all bases present
			nexus.Alignment([]string{
				"CCCCCC",
				"TTTTTT",
				"GGGGGG",
			}),
			[]metrics.Metric{metrics.Entropy, metrics.GC},
			2,
			[]byte("ATGC"),
			map[metrics.Metric]windows.Window{
				metrics.Entropy: windows.Window{2, 5},
				metrics.GC:      windows.Window{2, 5},
			},
			map[metrics.Metric][]float64{
				metrics.Entropy: []float64{
					1.098612288668110,
					1.098612288668110,
					1.098612288668110,
					1.098612288668110,
					1.098612288668110,
					1.098612288668110,
				},
				metrics.GC: []float64{
					2.0 / 3.0,
					2.0 / 3.0,
					2.0 / 3.0,
					2.0 / 3.0,
					2.0 / 3.0,
					2.0 / 3.0,
				},
			},
			false,
		},
		{ // All the same base
			nexus.Alignment([]string{
				"CCCCCC",
				"CCCCCC",
				"CCCCCC",
			}),
			[]metrics.Metric{metrics.Entropy, metrics.GC},
			2,
			[]byte("ATGC"),
			map[metrics.Metric]windows.Window{
				metrics.Entropy: windows.Window{2, 5},
				metrics.GC:      windows.Window{2, 5},
			},
			map[metrics.Metric][]float64{
				metrics.Entropy: []float64{
					0.0,
					0.0,
					0.0,
					0.0,
					0.0,
					0.0,
				},
				metrics.GC: []float64{
					3.0 / 3.0,
					3.0 / 3.0,
					3.0 / 3.0,
					3.0 / 3.0,
					3.0 / 3.0,
					3.0 / 3.0,
				},
			},
			false,
		},
		{ // More than one possible window
			nexus.Alignment([]string{
				"CCCCCCCC",
				"CCCCCCCC",
				"CCCCCCCC",
			}),
			[]metrics.Metric{metrics.Entropy, metrics.GC},
			2,
			[]byte("ATGC"),
			map[metrics.Metric]windows.Window{
				metrics.Entropy: windows.Window{2, 5},
				metrics.GC:      windows.Window{2, 5},
			},
			map[metrics.Metric][]float64{
				metrics.Entropy: []float64{
					0.0,
					0.0,
					0.0,
					0.0,
					0.0,
					0.0,
					0.0,
					0.0,
				},
				metrics.GC: []float64{
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
			false,
		},
	}
	for _, tc := range tt {
		gotWins, gotVals := uce.ProcessUce(tc.aln, tc.metrics, tc.minWin, tc.chars, tc.largeCore)
		t.Run("Windows", func(t *testing.T) {
			for m, got := range gotWins {
				exp := tc.expWins[m]
				if got != exp {
					t.Errorf("\nGot:\n%v\nExpected:\n%v\nFor:\n%v", got, exp, tc.aln)
				}
			}
		})
		t.Run("Values", func(t *testing.T) {
			for m, got := range gotVals {
				exp := tc.expVals[m]
				if !floats.EqualApprox(got, exp, 1e-15) {
					t.Errorf("\nGot:\n%v\nExpected:\n%v\nFor:\n%v", got, exp, tc.aln)
				}
			}
		})
	}
}
