package main

import (
	"math"

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

func getSse(metric []float64, win window, siteVar []bool) float64 {
	leftAln := allInvariantSites(siteVar[:win.Start()])
	coreAln := allInvariantSites(siteVar[win.Start():win.Stop()])
	rightAln := allInvariantSites(siteVar[win.Stop():])

	if leftAln || coreAln || rightAln {
		return math.MaxFloat64
	}

	left := sse(metric[:win.Start()])
	core := sse(metric[win.Start():win.Stop()])
	right := sse(metric[win.Stop():])
	return left + core + right
}

// getSses generalized getSse over each site window.
// metrics is [metric name][value index] arranged
func getSses(metrics map[string][]float64, win window, siteVar []bool) map[string][]float64 {
	// TODO: Preallocate array
	sses := make(map[string][]float64, len(metrics))
	for m := range metrics {
		sses[m] = append(sses[m], getSse(metrics[m], win, siteVar))
	}
	return sses
}