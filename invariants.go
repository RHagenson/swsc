package main

import "bitbucket.org/rhagenson/swsc/nexus"

// invariantSites streams across an alignment and calls sites invariant by their entropy
func invariantSites(aln nexus.Alignment) []bool {
	entropies := sitewiseEntropy(aln)
	calls := make([]bool, len(entropies))
	for i, v := range entropies {
		if v > 0 {
			calls[i] = true
		}
	}
	return calls
}

func allInvariantSites(vs []bool) bool {
	for _, v := range vs {
		if v {
			return false
		}
	}
	return true
}
