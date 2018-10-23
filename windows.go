package main

import (
	"sort"

	"gonum.org/v1/gonum/mat"
)

type window [2]int

func (w *window) Start() int {
	return w[0]
}

func (w *window) Stop() int {
	return w[1]
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
