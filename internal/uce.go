package internal

import "bitbucket.org/rhagenson/swsc/internal/nexus"

// ProcessUce computes the corresponding metrics within the minimum window size,
// returning the best window and list of values for each metric
func ProcessUce(uceAln nexus.Alignment, metrics []Metric, minWin int, chars []byte) (map[Metric]Window, map[Metric][]float64) {
	var (
		metricBestWindow = make(map[Metric]Window, len(metrics))
		metricBestVals   = make(map[Metric][]float64, len(metrics))
	)

	windows := getAllWindows(uceAln, minWin)
	inVarSites := invariantSites(uceAln, chars)

	for _, m := range metrics {
		switch m {
		case Entropy:
			metricBestVals[Entropy] = sitewiseEntropy(uceAln, chars)
		case GC:
			metricBestVals[GC] = sitewiseGc(uceAln)
			// case "multi":
			// 	metricBestVals["multi"] = sitewiseMulti(uceAln)
		}
	}
	if len(windows) > 1 {
		metricBestWindow = getBestWindows(metricBestVals, windows, uceAln.Len(), inVarSites)
	} else {
		for _, k := range metrics {
			metricBestWindow[k] = windows[0]
		}
	}
	return metricBestWindow, metricBestVals
}
