package windows

import (
	"math"

	"github.com/rhagenson/swsc/internal/metrics"
	"gonum.org/v1/gonum/stat"
)

// Compute Sum of Square Errors
func sse(vs []float64) float64 {
	mean := stat.Mean(vs, nil)
	total := 0.0
	for _, v := range vs {
		total += math.Pow((v / mean), 2)
	}
	return total
}

func getSse(metric []float64, win Window) float64 {
	left := sse(metric[:win.Start()])
	core := sse(metric[win.Start():win.Stop()])
	right := sse(metric[win.Stop():])
	return left + core + right
}

// getSses generalized getSse over each site window.
func getSses(mets map[metrics.Metric][]float64, win Window) map[metrics.Metric]float64 {
	sses := make(map[metrics.Metric]float64, len(mets))
	for m := range mets {
		sses[m] = getSse(mets[m], win)
	}
	return sses
}
