package main

import (
	"math"

	"github.com/biogo/biogo/seq/multi"
)

func sitewiseMulti(uceAln *multi.Multi) []float64 {
	uceCounts := sitewiseBaseCounts(uceAln)
	uceSums := make([]float64, len(uceCounts))
	for i := range uceSums {
		for _, v := range uceCounts {
			for _, z := range v {
				uceSums[i] += float64(z)
			}
		}
	}
	uceObsFactorial := make([]float64, len(uceSums))
	for i, v := range uceSums {
		uceObsFactorial[i] = factorial(int(v))
	}
	uceFactorials := factorialMatrix(uceCounts)

	m1 := minFloat64(uceObsFactorial...)
	m2 := minFloat64(uceFactorials...)
	if m1 < 0 || m2 < 0 {
		panic("Sitewise multinomial factorials <0")
	}
	sitewiseLikelihood := sitewiseMultiCounts(uceCounts, uceFactorials, uceObsFactorial)
	logLikelihoods := make([]float64, len(sitewiseLikelihood))
	for i, v := range sitewiseLikelihood {
		logLikelihoods[i] = math.Log(v)
	}
	return logLikelihoods
}

// Was not in the sample code
// Implemented similar to sitewiseBaseCount
func sitewiseMultiCounts(uceAln *multi.Multi) map[byte][]float64 {
	counts := map[byte][]float64{
		'A': make([]float64, uceAln.Len()),
		'T': make([]float64, uceAln.Len()),
		'C': make([]float64, uceAln.Len()),
		'G': make([]float64, uceAln.Len()),
	}
	for i := 0; i < uceAln.Len(); i++ {
		site, _ := uceAln.Subseq(i, i+1)
		bCounts := sitewiseMulti(site)
		for k, v := range bCounts {
			counts[k][i] += v
		}
	}
	return counts
}

func sitewiseEntropy(aln *multi.Multi) []float64 {
	entropies := make([]float64, aln.Len())
	for i := 0; i < aln.Len(); i++ {
		site, _ := aln.Subseq(i, i+1)
		entropies[i] = alignmentEntropy(site)
	}
	return entropies
}

// sitewiseBaseCounts returns a 4xN aeeay of base counts where N is the number of sites
func sitewiseBaseCounts(uceAln *multi.Multi) map[byte][]int {
	counts := map[byte][]int{
		'A': make([]int, uceAln.Len()),
		'T': make([]int, uceAln.Len()),
		'C': make([]int, uceAln.Len()),
		'G': make([]int, uceAln.Len()),
	}
	for i := 0; i < uceAln.Len(); i++ {
		site, _ := uceAln.Subseq(i, i+1)
		bCounts := countBases(site)
		for k, v := range bCounts {
			counts[k][i] += v
		}
	}
	return counts
}
