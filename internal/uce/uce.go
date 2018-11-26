package uce

import (
	"bitbucket.org/rhagenson/swsc/internal/invariants"
	"bitbucket.org/rhagenson/swsc/internal/metric"
	"bitbucket.org/rhagenson/swsc/internal/nexus"
	"bitbucket.org/rhagenson/swsc/internal/windows"
)

// ProcessUce computes the corresponding metrics within the minimum window size,
// returning the best window and list of values for each metric
func ProcessUce(uceAln nexus.Alignment, metrics []metric.Metric, minWin int, chars []byte) (map[metric.Metric]windows.Window, map[metric.Metric][]float64) {
	var (
		metricBestWindow = make(map[metric.Metric]windows.Window, len(metrics))
		metricBestVals   = make(map[metric.Metric][]float64, len(metrics))
	)

	wins := windows.GetAll(uceAln, minWin)
	inVarSites := invariants.InvariantSites(uceAln, chars)

	for _, m := range metrics {
		switch m {
		case metric.Entropy:
			metricBestVals[metric.Entropy] = metric.SitewiseEntropy(uceAln, chars)
		case metric.GC:
			metricBestVals[metric.GC] = metric.SitewiseGc(uceAln)
			// case "multi":
			// 	metricBestVals["multi"] = sitewiseMulti(uceAln)
		}
	}
	if len(wins) > 1 {
		metricBestWindow = windows.GetBest(metricBestVals, wins, uceAln.Len(), inVarSites)
	} else {
		for _, k := range metrics {
			metricBestWindow[k] = wins[0]
		}
	}
	return metricBestWindow, metricBestVals
}
