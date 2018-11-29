package windows_test

import (
	"testing"

	"bitbucket.org/rhagenson/swsc/internal/metrics"
	"bitbucket.org/rhagenson/swsc/internal/nexus"
	"bitbucket.org/rhagenson/swsc/internal/windows"
)

func TestGenerateWindows(t *testing.T) {
	tt := []struct {
		length   int
		minWin   int
		expected int
	}{
		{320, 100, 231},
		{321, 100, 253},
		{322, 100, 276},
		{323, 100, 300},
		{324, 100, 325},
		{325, 100, 351},

		{325, 101, 276},
		{325, 102, 210},
		{325, 103, 153},
		{325, 104, 105},
		{325, 105, 66},

		{5786, 50, 15890703},
		{5786, 100, 15056328},
	}
	for _, tc := range tt {
		got := windows.GenerateWindows(tc.length, tc.minWin)
		if len(got) != tc.expected {
			t.Errorf("Given len:%d, win:%d, expected %d, got %d\n",
				tc.length, tc.minWin, tc.expected, len(got))
		}
	}
}

func TestGenerateCandidates(t *testing.T) {
	tt := []struct {
		length   int
		minWin   int
		expected int
	}{
		{300, 50, 4}, // [[50, 100], [100, 150], [150, 200], [150, 200]]

		{320, 100, 2}, // [[100, 200], [120, 220]]
		{321, 100, 2}, // [[100, 200], [121, 221]]
		{322, 100, 2}, // [[100, 200], [122, 222]]
		{323, 100, 2}, // [[100, 200], [123, 221]]
		{324, 100, 2}, // [[100, 200], [124, 224]]
		{325, 100, 2}, // [[100, 200], [125, 225]]

		{325, 101, 2},
		{325, 102, 2},
		{325, 103, 2},
		{325, 104, 2},
		{325, 105, 2},

		{5786, 50, 226},
		{5786, 100, 110},
	}
	for _, tc := range tt {
		got := windows.GenerateCandidates(tc.length, tc.minWin)
		if len(got) != tc.expected {
			t.Errorf("Given len:%d, win:%d, expected %d, got %d\n",
				tc.length, tc.minWin, tc.expected, len(got),
			)
		}
	}
}

func TestExtendCandidate(t *testing.T) {
	tt := []struct {
		win      windows.Window
		length   int
		minWin   int
		expected int
	}{
		{windows.Window{100, 200}, 301, 100, 3},
		{windows.Window{100, 200}, 302, 100, 6},
		{windows.Window{100, 200}, 303, 100, 10},

		{windows.Window{50, 100}, 300, 50, 1326},

		{windows.Window{100, 200}, 321, 100, 253},
		{windows.Window{100, 200}, 322, 100, 276},
		{windows.Window{100, 200}, 323, 100, 300},
		{windows.Window{100, 200}, 324, 100, 325},
		{windows.Window{100, 200}, 325, 100, 351},

		{windows.Window{101, 202}, 325, 101, 276},
		{windows.Window{102, 204}, 325, 102, 210},
		{windows.Window{103, 206}, 325, 103, 153},
		{windows.Window{104, 208}, 325, 104, 105},
		{windows.Window{105, 210}, 325, 105, 66},

		{windows.Window{50, 100}, 5786, 50, 1326},
		{windows.Window{100, 200}, 5786, 100, 5151},
	}
	for _, tc := range tt {
		got := windows.ExtendCandidate(tc.win, tc.length, tc.minWin)
		if len(got) != tc.expected {
			t.Errorf("Given len:%d, win:%d, expected %d, got %d\n",
				tc.length, tc.minWin, tc.expected, len(got),
			)
		}
	}
}

func TestWindow(t *testing.T) {
	tt := []struct {
		win windows.Window
	}{
		{windows.New(0, 0)},    // Start == Stop
		{windows.New(0, 50)},   // 0 <= Start < Stop
		{windows.New(50, 0)},   //  0 <= Stop < Start, orients values such as Start < Stop
		{windows.New(-50, 0)},  // Start < Stop <= 0
		{windows.New(0, -50)},  // Stop < Start <= 0, orients values such as Start < Stop
		{windows.New(-50, 50)}, // Start < 0 < Stop
		{windows.New(50, -50)}, // Stop < 0 < Start, orients values such as Start < Stop
	}

	for _, tc := range tt {
		t.Run("Start", func(t *testing.T) {
			if tc.win.Start() != tc.win[0] {
				t.Errorf("Got: %d, Expected %d", tc.win.Start(), tc.win[0])
			}
		})
		t.Run("Stop", func(t *testing.T) {
			if tc.win.Stop() != tc.win[1] {
				t.Errorf("Got: %d, Expected %d", tc.win.Start(), tc.win[0])
			}
		})
	}
}

func TestGetAll(t *testing.T) {
	tt := []struct {
		aln    nexus.Alignment
		minWin int
	}{
		{ // All the same seq
			nexus.Alignment([]string{
				"ATGCAT",
				"ATGCAT",
				"ATGCAT",
			}),
			2,
		},
		{ // One seq reversed
			nexus.Alignment([]string{
				"ATGCAT",
				"ATGCAT",
				"TACGTA",
			}),
			2,
		},
		{ // No positional matches
			nexus.Alignment([]string{
				"ATGCAT",
				"TGCTGA",
				"GCTACC",
			}),
			2,
		},
		{ // Not all bases present
			nexus.Alignment([]string{
				"CCCCCC",
				"TTTTTT",
				"GGGGGG",
			}),
			2,
		},
		{ // All the same base
			nexus.Alignment([]string{
				"CCCCCC",
				"CCCCCC",
				"CCCCCC",
			}),
			2,
		},
		{ // More than one possible window
			nexus.Alignment([]string{
				"CCCCCCCC",
				"CCCCCCCC",
				"CCCCCCCC",
			}),
			2,
		},
	}

	for _, tc := range tt {
		got := windows.GetAll(tc.aln, tc.minWin)
		exp := windows.GenerateWindows(tc.aln.Len(), 2)
		t.Run("Length", func(t *testing.T) {
			if len(got) != len(exp) {
				t.Errorf("Got %d, Expected %d", len(got), len(exp))
			}
		})

		for i := range got {
			if got[i] != exp[i] {
				t.Errorf("Got %v, Expected %v", got[i], exp[i])
			}
		}
	}
}

func TestGetBest(t *testing.T) {
	tt := []struct {
		metrics    map[metrics.Metric][]float64
		windows    []windows.Window
		alnLen     int
		inVarSites []bool
		exp        map[metrics.Metric]windows.Window
		largeCore  bool
	}{
		{
			map[metrics.Metric][]float64{
				metrics.Entropy: []float64{
					2.694,
					1.438,
					1.210,
					1.110,
					0.005,
					0.000,
				},
			},
			[]windows.Window{
				windows.New(2, 4),
				windows.New(2, 5),
				windows.New(2, 6),
				windows.New(3, 5),
				windows.New(4, 6),
				windows.New(3, 6),
			},
			8,
			[]bool{false, false, false, true, false, true},
			map[metrics.Metric]windows.Window{
				metrics.Entropy: windows.New(3, 5),
			},
			false,
		},
	}

	for _, tc := range tt {
		got := windows.GetBest(tc.metrics, tc.windows, tc.alnLen, tc.inVarSites, tc.largeCore)
		for m, v := range tc.exp {
			if got[m] != v {
				t.Errorf("Got: %v, Expected: %v", got[m], v)
			}
		}
	}
}
