package uce_test

import (
	"testing"

	"bitbucket.org/rhagenson/swsc/internal/metrics"
	"bitbucket.org/rhagenson/swsc/internal/nexus"
	"bitbucket.org/rhagenson/swsc/internal/uce"
	"bitbucket.org/rhagenson/swsc/internal/windows"
)

func TestProcessUce(t *testing.T) {
	tt := []struct {
		aln       nexus.Alignment
		minWin    int
		chars     []byte
		expWins   map[metrics.Metric]windows.Window
		metVals   map[metrics.Metric][]float64
		largeCore bool
	}{
		{ // All the same seq
			nexus.Alignment([]string{
				"ATGCAT",
				"ATGCAT",
				"ATGCAT",
			}),
			2,
			[]byte("ATGC"),
			map[metrics.Metric]windows.Window{
				metrics.Entropy: windows.Window{2, 4},
				metrics.GC:      windows.Window{2, 4},
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
			2,
			[]byte("ATGC"),
			map[metrics.Metric]windows.Window{
				metrics.Entropy: windows.Window{2, 4},
				metrics.GC:      windows.Window{2, 4},
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
			2,
			[]byte("ATGC"),
			map[metrics.Metric]windows.Window{
				metrics.Entropy: windows.Window{2, 4},
				metrics.GC:      windows.Window{2, 4},
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
			2,
			[]byte("ATGC"),
			map[metrics.Metric]windows.Window{
				metrics.Entropy: windows.Window{2, 4},
				metrics.GC:      windows.Window{2, 4},
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
			2,
			[]byte("ATGC"),
			map[metrics.Metric]windows.Window{
				metrics.Entropy: windows.Window{2, 4},
				metrics.GC:      windows.Window{2, 4},
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
			2,
			[]byte("ATGC"),
			map[metrics.Metric]windows.Window{
				metrics.Entropy: windows.Window{3, 5},
				metrics.GC:      windows.Window{3, 5},
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
		inVars := make([]bool, tc.aln.Len())
		gotWins := uce.ProcessUce(0, tc.aln.Len(), inVars, tc.metVals, tc.minWin, tc.chars, tc.largeCore)
		t.Run("Windows", func(t *testing.T) {
			for m, got := range gotWins {
				exp := tc.expWins[m]
				if got != exp {
					t.Errorf("\nGot:\n%v\nExpected:\n%v\nFor:\n%v", got, exp, tc.aln)
				}
			}
		})
	}
}
