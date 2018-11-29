package uce

import (
	"bitbucket.org/rhagenson/swsc/internal/invariants"
	"bitbucket.org/rhagenson/swsc/internal/metrics"
	"bitbucket.org/rhagenson/swsc/internal/nexus"
	"bitbucket.org/rhagenson/swsc/internal/windows"
)

// ProcessUce computes the corresponding metrics within the minimum window size,
// returning the best window and list of values for each metric
func ProcessUce(uceAln nexus.Alignment, mets []metrics.Metric, minWin int, chars []byte, largeCore bool) (map[metrics.Metric]windows.Window, map[metrics.Metric][]float64) {
	var (
		metricBestWindow = make(map[metrics.Metric]windows.Window, len(mets))
		metricBestVals   = make(map[metrics.Metric][]float64, len(mets))
	)
	inVarSites := invariants.InvariantSites(uceAln, chars)
	for _, m := range mets {
		switch m {
		case metrics.Entropy:
			metricBestVals[metrics.Entropy] = metrics.SitewiseEntropy(uceAln, chars)
		case metrics.GC:
			metricBestVals[metrics.GC] = metrics.SitewiseGc(uceAln)
			// case "multi":
			// 	metricBestVals["multi"] = sitewiseMulti(uceAln)
		}
	}

	// Heuristic: Get nonoverlapping candidate windows
	canWins := windows.GenerateCandidates(uceAln.Len(), minWin)

	// Determine the best candidate window
	bestCanWins := windows.GetBest(metricBestVals, canWins, uceAln.Len(), inVarSites, largeCore)

	// Extend the best candidate and retest
	for _, w := range bestCanWins {
		canWins = windows.ExtendCandidate(w, uceAln.Len(), minWin)
	}

	metricBestWindow = windows.GetBest(metricBestVals, canWins, uceAln.Len(), inVarSites, largeCore)

	return metricBestWindow, metricBestVals
}
