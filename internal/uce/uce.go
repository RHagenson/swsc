package uce

import (
	"math"

	"github.com/rhagenson/swsc/internal/metrics"
	"github.com/rhagenson/swsc/internal/windows"
)

// ProcessUce computes the corresponding metrics within the minimum window size,
// returning the best window and list of values for each metric
func ProcessUce(start, stop int, mets map[metrics.Metric][]float64, minWin uint, chars []byte, largeCore bool, n uint) map[metrics.Metric]windows.Window {
	var (
		metricBestWindow = make(map[metrics.Metric]windows.Window, len(mets))
	)

	// Heuristic: Get nonoverlapping candidate windows
	canWins := windows.GenerateCandidates(start, stop, int(minWin))

	// Determine the best candidate window
	bestCanWins := windows.GetBestN(mets, canWins, stop, largeCore, n)

	// Extend the best candidates and retest
	// Also find the encompassing Window for testing (could be whole sequence)
	var (
		extWins  []windows.Window
		winStart = math.MaxInt64
		winStop  = math.MinInt64
	)
	for _, wins := range bestCanWins {
		for _, w := range wins {
			extWins = append(extWins, windows.ExtendCandidate(w, start, stop, int(minWin))...)
			if w.Start() < winStart {
				winStart = w.Start()
			}
			if winStop < w.Stop() {
				winStop = w.Stop()
			}
		}
	}
	extWins = append(extWins, windows.New(winStart, winStop))

	metricBestWindow = windows.GetBest(mets, extWins, stop, largeCore)

	return metricBestWindow
}
