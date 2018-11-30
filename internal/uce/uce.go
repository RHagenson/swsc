package uce

import (
	"bitbucket.org/rhagenson/swsc/internal/metrics"
	"bitbucket.org/rhagenson/swsc/internal/windows"
)

// ProcessUce computes the corresponding metrics within the minimum window size,
// returning the best window and list of values for each metric
func ProcessUce(start, stop int, inVars []bool, mets map[metrics.Metric][]float64, minWin int, chars []byte, largeCore bool) map[metrics.Metric]windows.Window {
	var (
		metricBestWindow = make(map[metrics.Metric]windows.Window, len(mets))
	)

	// Heuristic: Get nonoverlapping candidate windows
	canWins := windows.GenerateCandidates(start, stop, minWin)

	// Determine the best candidate window
	bestCanWins := windows.GetBest(mets, canWins, stop-start, inVars, largeCore)

	// Extend the best candidate and retest
	for _, w := range bestCanWins {
		canWins = windows.ExtendCandidate(w, start, stop, minWin)
	}

	metricBestWindow = windows.GetBest(mets, canWins, stop, inVars, largeCore)

	return metricBestWindow
}
