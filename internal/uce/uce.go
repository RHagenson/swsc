package uce

import (
	"bitbucket.org/rhagenson/swsc/internal/invariants"
	"bitbucket.org/rhagenson/swsc/internal/metrics"
	"bitbucket.org/rhagenson/swsc/internal/nexus"
	"bitbucket.org/rhagenson/swsc/internal/windows"
)

// ProcessUce computes the corresponding metrics within the minimum window size,
// returning the best window and list of values for each metric
func ProcessUce(uceAln nexus.Alignment, mets []metrics.Metric, minWin int, chars []byte) (map[metrics.Metric]windows.Window, map[metrics.Metric][]float64) {
	var (
		metricBestWindow = make(map[metrics.Metric]windows.Window, len(mets))
		metricBestVals   = make(map[metrics.Metric][]float64, len(mets))
	)

	wins := windows.GetAll(uceAln, minWin)
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
	if len(wins) > 1 {
		metricBestWindow = windows.GetBest(metricBestVals, wins, uceAln.Len(), inVarSites)
	} else {
		for _, k := range mets {
			metricBestWindow[k] = wins[0]
		}
	}
	return metricBestWindow, metricBestVals
}
