package uce

import (
	"bitbucket.org/rhagenson/swsc/internal/metrics"
	"bitbucket.org/rhagenson/swsc/internal/windows"
)

// ProcessUce computes the corresponding metrics within the minimum window size,
// returning the best window and list of values for each metric
func ProcessUce(start, stop int, mets map[metrics.Metric][]float64, minWin int, chars []byte, largeCore bool) map[metrics.Metric]windows.Window {
	var (
		metricBestWindow = make(map[metrics.Metric]windows.Window, len(mets))
	)

	// Heuristic: Get nonoverlapping candidate windows
	canWins := windows.GenerateCandidates(start, stop, minWin)

	// Determine the best candidate window
	bestCanWins := windows.GetBest(mets, canWins, stop, largeCore)

	// Extend the best candidate and retest
	var extWins []windows.Window
	for _, w := range bestCanWins {
		extWins = append(extWins, windows.ExtendCandidate(w, start, stop, minWin)...)
	}

	metricBestWindow = windows.GetBest(mets, extWins, stop, largeCore)

	return metricBestWindow
}
