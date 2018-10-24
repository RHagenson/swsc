package main

import (
	"fmt"
	"sort"

	"github.com/biogo/biogo/seq/multi"
	"gonum.org/v1/gonum/mat"
)

type window [2]int

func (w *window) Start() int {
	return w[0]
}

func (w *window) Stop() int {
	return w[1]
}

func getAllWindows(uceAln *multi.Multi, minWin int) []window {
	// Cannot split into left-flank, core, and right-flank
	if uceAln.Len() < 3*minWin+1 {
		message := fmt.Sprintf(
			"Cannot split UCE into minimum window sized flanks and core.\n"+
				"Minimum window size: %d, UCE length: %d",
			minWin, uceAln.Len(),
		)
		panic(message)
	}

	windows := generateWindows(uceAln.Len(), minWin)
	return windows
}

// generateWindows produces windows of at least a minimum size given a total length
// Windows must be:
//   1) at least minimum window from the start of the UCE (ie, first start at minimum+1)
//   2) at least minimum window from the end of the UCE (ie, last end at length-minimum+1)
//   3) at least minimum window in length (ie, window{start, end)})
func generateWindows(len, min int) []window {
	windows := make([]window, 0)
	// TODO: Should preallocate windows array and fill via start+end-minWin indexing
	for start := min + 1; start+min < len; start++ {
		for end := start + min + 1; end+min <= len; end++ {
			windows = append(windows, window{start, end})
		}
	}
	return windows
}

func getBestWindows(metrics map[string][]float64, windows []window, alnLen int, inVarSites []bool) map[string]window {
	// 1) Make an empty array
	// rows = number of metrics
	// columns = number of windows
	// data = nil, allocate new backing slice
	allSSEs := mat.NewDense(len(metrics), len(windows), nil)

	// 2) Get SSE for each cell in array
	for i, window := range windows {
		// Get SSEs for a given window
		allSSEs.SetCol(i, getSses(metrics, window, inVarSites))
	}

	// 3) get index of minimum value for each metric
	// Indexed the same as metrics argument
	// TODO: Point of non-determinism: Many values can equal the minimum, but the first
	// is chosen using this manner. Improvement:
	length := 0
	for _, v := range metrics {
		length = len(v)
		break
	}
	valMins := make([]float64, length)
	for i := 0; i < length; i++ {
		vals := allSSEs.RawRowView(i)
		copyVals := make([]float64, len(vals))
		copy(copyVals, vals)
		sort.Float64s(copyVals) // Sort done in place
		valMins[i] = copyVals[0]
	}

	// choose windows with the minimum variance in length of l-flank, core, r-flank
	valMinIndexes := make(map[string][]int)
	for m, v := range metrics {
		for i := 0; i < len(v); i++ {
			vals := allSSEs.RawRowView(i)
			for j := range vals {
				if valMins[i] == vals[j] {
					valMinIndexes[m] = append(valMinIndexes[m], j)
				}
			}
		}
	}
	minMetricWindows := make(map[string][]window)
	for k, vs := range valMinIndexes {
		wins := make([]window, 0)
		for v := range vs {
			wins = append(wins, windows[v])
		}
	}
	absMinWindow := make(map[string]window)
	for m, v := range minMetricWindows {
		absMinWindow[m] = getMinVarWindow(minMetricWindows[m], alnLen)
	}
	return absMinWindow
}
