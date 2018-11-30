package uce

import (
	"bitbucket.org/rhagenson/swsc/internal/invariants"
	"bitbucket.org/rhagenson/swsc/internal/metrics"
	"bitbucket.org/rhagenson/swsc/internal/nexus"
	"bitbucket.org/rhagenson/swsc/internal/windows"
)

// ProcessUce computes the corresponding metrics within the minimum window size,
// returning the best window and list of values for each metric
func ProcessUce(uceAln nexus.Alignment, inVars []bool, mets map[metrics.Metric][]float64, minWin int, chars []byte, largeCore bool) map[metrics.Metric]windows.Window {
	var (
		metricBestWindow = make(map[metrics.Metric]windows.Window, len(mets))
	)

	// Heuristic: Get nonoverlapping candidate windows
	canWins := windows.GenerateCandidates(uceAln.Len(), minWin)

	// Determine the best candidate window
	bestCanWins := windows.GetBest(metricBestVals, canWins, uceAln.Len(), inVarSites, largeCore)

	// Extend the best candidate and retest
	for _, w := range bestCanWins {
		canWins = windows.ExtendCandidate(w, uceAln.Len(), minWin)
	}

	metricBestWindow = windows.GetBest(mets, canWins, uceAln.Len(), inVars, largeCore)

	return metricBestWindow
}
