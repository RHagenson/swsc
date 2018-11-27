package uce_test

import (
	"testing"

	"bitbucket.org/rhagenson/swsc/internal/metric"
	"bitbucket.org/rhagenson/swsc/internal/nexus"
	"bitbucket.org/rhagenson/swsc/internal/uce"
	"bitbucket.org/rhagenson/swsc/internal/windows"
	"gonum.org/v1/gonum/floats"
)

func TestProcessUce(t *testing.T) {
	tt := []struct {
		aln     nexus.Alignment
		metrics []metric.Metric
		minWin  int
		chars   []byte
		expWins map[metric.Metric]windows.Window
		expVals map[metric.Metric][]float64
	}{
		{ // All the same seq
			nexus.Alignment([]string{
				"ATGCAT",
				"ATGCAT",
				"ATGCAT",
			}),
			[]metric.Metric{metric.Entropy, metric.GC},
			2,
			[]byte("ATGC"),
			map[metric.Metric]windows.Window{
				metric.Entropy: windows.Window{2, 5},
				metric.GC:      windows.Window{2, 5},
			},
			map[metric.Metric][]float64{
				metric.Entropy: []float64{
					0.0,
					0.0,
					0.0,
					0.0,
					0.0,
					0.0,
				},
				metric.GC: []float64{
					0,
					0,
					1,
					1,
					0,
					0,
				},
			},
		},
		// { // One seq reversed
		// 	nexus.Alignment([]string{
		// 		"ATGCAT",
		// 		"ATGCAT",
		// 		"TACGTA",
		// 	}),
		// 	[]metric.Metric{metric.Entropy, metric.GC},
		// 	2,
		// 	[]byte("ATGC"),
		// 	map[metric.Metric]windows.Window{
		// 		metric.Entropy: windows.Window{2, 5},
		// 		metric.GC:      windows.Window{2, 5},
		// 	},
		// 	map[metric.Metric][]float64{
		// 		metric.Entropy: []float64{
		// 			0.6365141682948128,
		// 			0.6365141682948128,
		// 			0.6365141682948128,
		// 			0.6365141682948128,
		// 			0.6365141682948128,
		// 			1.3296613488547580,
		// 		},
		// 		metric.GC: []float64{
		// 			0 / 3,
		// 			0 / 3,
		// 			3.0 / 3.0,
		// 			3.0 / 3.0,
		// 			0 / 3,
		// 			0 / 3,
		// 		},
		// 	},
		// },
		// { // No positional matches
		// 	nexus.Alignment([]string{
		// 		"ATGCAT",
		// 		"TGCTGA",
		// 		"GCTACC",
		// 	}),
		// 	[]metric.Metric{metric.Entropy, metric.GC},
		// 	2,
		// 	[]byte("ATGC"),
		// 	map[metric.Metric]windows.Window{
		// 		metric.Entropy: windows.Window{2, 5},
		// 		metric.GC:      windows.Window{2, 5},
		// 	},
		// 	map[metric.Metric][]float64{
		// 		metric.Entropy: []float64{
		// 			1.0986122886681096,
		// 			1.0986122886681096,
		// 			1.0986122886681096,
		// 			1.0986122886681096,
		// 			1.0986122886681096,
		// 			1.3801087571572686,
		// 		},
		// 		metric.GC: []float64{
		// 			1.0 / 3.0,
		// 			2.0 / 3.0,
		// 			2.0 / 3.0,
		// 			1.0 / 3.0,
		// 			2.0 / 3.0,
		// 			1.0 / 3.0,
		// 		},
		// 	},
		// },
		// { // Not all bases present
		// 	nexus.Alignment([]string{
		// 		"CCCCCC",
		// 		"TTTTTT",
		// 		"GGGGGG",
		// 	}),
		// 	[]metric.Metric{metric.Entropy, metric.GC},
		// 	2,
		// 	[]byte("ATGC"),
		// 	map[metric.Metric]windows.Window{
		// 		metric.Entropy: windows.Window{2, 5},
		// 		metric.GC:      windows.Window{2, 5},
		// 	},
		// 	map[metric.Metric][]float64{
		// 		metric.Entropy: []float64{
		// 			1.098612288668110,
		// 			1.098612288668110,
		// 			1.098612288668110,
		// 			1.098612288668110,
		// 			1.098612288668110,
		// 			1.098612288668110,
		// 		},
		// 		metric.GC: []float64{
		// 			2.0 / 3.0,
		// 			2.0 / 3.0,
		// 			2.0 / 3.0,
		// 			2.0 / 3.0,
		// 			2.0 / 3.0,
		// 			2.0 / 3.0,
		// 		},
		// 	},
		// },
		// { // All the same base
		// 	nexus.Alignment([]string{
		// 		"CCCCCC",
		// 		"CCCCCC",
		// 		"CCCCCC",
		// 	}),
		// 	[]metric.Metric{metric.Entropy, metric.GC},
		// 	2,
		// 	[]byte("ATGC"),
		// 	map[metric.Metric]windows.Window{
		// 		metric.Entropy: windows.Window{2, 5},
		// 		metric.GC:      windows.Window{2, 5},
		// 	},
		// 	map[metric.Metric][]float64{
		// 		metric.Entropy: []float64{
		// 			0.0,
		// 			0.0,
		// 			0.0,
		// 			0.0,
		// 			0.0,
		// 			0.0,
		// 		},
		// 		metric.GC: []float64{
		// 			3.0 / 3.0,
		// 			3.0 / 3.0,
		// 			3.0 / 3.0,
		// 			3.0 / 3.0,
		// 			3.0 / 3.0,
		// 			3.0 / 3.0,
		// 		},
		// 	},
		// },
		// { // More than one possible window
		// 	nexus.Alignment([]string{
		// 		"CCCCCCCC",
		// 		"CCCCCCCC",
		// 		"CCCCCCCC",
		// 	}),
		// 	[]metric.Metric{metric.Entropy, metric.GC},
		// 	2,
		// 	[]byte("ATGC"),
		// 	map[metric.Metric]windows.Window{
		// 		metric.Entropy: windows.Window{2, 5},
		// 		metric.GC:      windows.Window{2, 5},
		// 	},
		// 	map[metric.Metric][]float64{
		// 		metric.Entropy: []float64{
		// 			0.0,
		// 			0.0,
		// 			0.0,
		// 			0.0,
		// 			0.0,
		// 			0.0,
		// 			0.0,
		// 			0.0,
		// 		},
		// 		metric.GC: []float64{
		// 			3.0 / 3.0,
		// 			3.0 / 3.0,
		// 			3.0 / 3.0,
		// 			3.0 / 3.0,
		// 			3.0 / 3.0,
		// 			3.0 / 3.0,
		// 			3.0 / 3.0,
		// 			3.0 / 3.0,
		// 		},
		// 	},
		// },
	}
	for _, tc := range tt {
		gotWins, gotVals := uce.ProcessUce(tc.aln, tc.metrics, tc.minWin, tc.chars)
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
