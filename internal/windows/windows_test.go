package windows_test

import (
	"testing"

	"bitbucket.org/rhagenson/swsc/internal/metrics"
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
		start    int
		stop     int
		minWin   int
		expected int
	}{
		{0, 300, 50, 4}, // [[50, 100], [100, 150], [150, 200], [150, 200]]

		{0, 320, 100, 2}, // [[100, 200], [120, 220]]
		{0, 321, 100, 2}, // [[100, 200], [121, 221]]
		{0, 322, 100, 2}, // [[100, 200], [122, 222]]
		{0, 323, 100, 2}, // [[100, 200], [123, 221]]
		{0, 324, 100, 2}, // [[100, 200], [124, 224]]
		{0, 325, 100, 2}, // [[100, 200], [125, 225]]

		{0, 325, 101, 2},
		{0, 325, 102, 2},
		{0, 325, 103, 2},
		{0, 325, 104, 2},
		{0, 325, 105, 2},

		{1, 325, 101, 2},
		{1, 325, 102, 2},
		{1, 325, 103, 2},
		{1, 325, 104, 2},
		{1, 325, 105, 2},

		{0, 5786, 50, 226},
		{0, 5786, 100, 110},

		{1, 376, 50, 10},
	}
	for _, tc := range tt {
		got := windows.GenerateCandidates(tc.start, tc.stop, tc.minWin)
		if len(got) != tc.expected {
			t.Errorf("Given start:%d, stop:%d, min: %d expected %d, got %d, got %v\n",
				tc.start, tc.stop, tc.minWin, tc.expected, len(got), got,
			)
		}
	}
}

func TestExtendCandidate(t *testing.T) {
	tt := []struct {
		win      windows.Window
		start    int
		stop     int
		minWin   int
		expected int
	}{
		{windows.Window{100, 200}, 0, 301, 100, 3},
		{windows.Window{100, 200}, 0, 302, 100, 6},
		{windows.Window{100, 200}, 0, 303, 100, 10},

		{windows.Window{50, 100}, 0, 300, 50, 1326},

		{windows.Window{100, 200}, 0, 321, 100, 253},
		{windows.Window{100, 200}, 0, 322, 100, 276},
		{windows.Window{100, 200}, 0, 323, 100, 300},
		{windows.Window{100, 200}, 0, 324, 100, 325},
		{windows.Window{100, 200}, 0, 325, 100, 351},

		{windows.Window{101, 202}, 0, 325, 101, 276},
		{windows.Window{102, 204}, 0, 325, 102, 210},
		{windows.Window{103, 206}, 0, 325, 103, 153},
		{windows.Window{104, 208}, 0, 325, 104, 105},
		{windows.Window{105, 210}, 0, 325, 105, 66},

		{windows.Window{50, 100}, 0, 5786, 50, 1326},
		{windows.Window{100, 200}, 0, 5786, 100, 5151},
	}
	for _, tc := range tt {
		got := windows.ExtendCandidate(tc.win, tc.start, tc.stop, tc.minWin)
		if len(got) != tc.expected {
			t.Errorf("Given win:%v, start: %d, stop: %d, min: %d, expected %d, got %d\n",
				tc.win, tc.start, tc.stop, tc.minWin, tc.expected, len(got),
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
