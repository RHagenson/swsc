package invariants

import (
	"bitbucket.org/rhagenson/swsc/internal/metric"
	"bitbucket.org/rhagenson/swsc/internal/nexus"
)

// InvariantSites streams across an alignment and calls sites invariant by their entropy
func InvariantSites(aln nexus.Alignment, chars []byte) []bool {
	entropies := metric.SitewiseEntropy(aln, chars)
	calls := make([]bool, len(entropies))
	for i, v := range entropies {
		if v > 0 {
			calls[i] = true
		}
	}
	return calls
}

func AllInvariantSites(vs []bool) bool {
	for _, v := range vs {
		if v {
			return false
		}
	}
	return true
}
