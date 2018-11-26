package windows_test

import (
	"testing"

	"bitbucket.org/rhagenson/swsc/internal/metric"
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
		got, _ := windows.GenerateWindows(tc.length, tc.minWin)
		if len(got) != tc.expected {
			t.Errorf("Given len:%d, win:%d, expected %d, got %d\n",
				tc.length, tc.minWin, tc.expected, len(got))
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
		exp, _ := windows.GenerateWindows(tc.aln.Len(), 2)
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
		metrics    map[metric.Metric][]float64
		windows    []windows.Window
		alnLen     int
		inVarSites []bool
		exp        map[metric.Metric]windows.Window
	}{
		{
			map[metric.Metric][]float64{
				metric.Entropy: []float64{
					1.694,
					1.438,
					1.610,
					0,
					1.608,
					0,
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
			[]bool{false, false, false, false, false, false},
			map[metric.Metric]windows.Window{
				metric.Entropy: windows.New(2, 5),
			},
		},
	}

	for _, tc := range tt {
		got := windows.GetBest(tc.metrics, tc.windows, tc.alnLen, tc.inVarSites)
		for m, v := range tc.exp {
			if got[m] != v {
				t.Errorf("Got: %v, Expected: %v", got[m], v)
			}
		}
	}
}
