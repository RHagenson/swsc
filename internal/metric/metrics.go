package metric

import (
	"strings"

	"bitbucket.org/rhagenson/swsc/internal/entropy"
	"bitbucket.org/rhagenson/swsc/internal/nexus"
)

// Metric is an enum type denoting possible sitewise metrics to calculate
type Metric int

const (
	// Entropy is Shannon's entropy
	Entropy Metric = iota

	// GC is GC percentage
	GC

	// Multi is the multinomial measurement
	Multi
)

func (m Metric) String() string {
	switch m {
	case Entropy:
		return "Entropy"
	case GC:
		return "GC"
	case Multi:
		return "Multinomial"
	default:
		return ""
	}
}

// func sitewiseMulti(uceAln nexus.Alignment) []float64 {
// 	uceCounts := sitewiseBaseCounts(uceAln)
// 	uceSums := make([]float64, len(uceCounts))
// 	for i := range uceSums {
// 		for _, v := range uceCounts {
// 			for _, z := range v {
// 				uceSums[i] += float64(z)
// 			}
// 		}
// 	}
// 	uceObsFactorial := make([]float64, len(uceSums))
// 	for i, v := range uceSums {
// 		uceObsFactorial[i] = factorial(int(v))
// 	}
// 	uceFactorials := factorialMatrix(uceCounts)

// 	m1 := minFloat64(uceObsFactorial...)
// 	m2 := minFloat64(uceFactorials...)
// 	if m1 < 0 || m2 < 0 {
// 		panic("Sitewise multinomial factorials <0")
// 	}
// 	sitewiseLikelihood := sitewiseMultiCounts(uceCounts, uceFactorials, uceObsFactorial)
// 	logLikelihoods := make([]float64, len(sitewiseLikelihood))
// 	for i, v := range sitewiseLikelihood {
// 		logLikelihoods[i] = math.Log(v)
// 	}
// 	return logLikelihoods
// }

// // Was not in the sample code
// // Implemented similar to sitewiseBaseCount
// func sitewiseMultiCounts(uceAln nexus.Alignment) map[byte][]float64 {
// 	counts := map[byte][]float64{
// 		'A': make([]float64, uceAln.Len()),
// 		'T': make([]float64, uceAln.Len()),
// 		'C': make([]float64, uceAln.Len()),
// 		'G': make([]float64, uceAln.Len()),
// 	}
// 	for i := 0; i < uceAln.Len(); i++ {
// 		site, _ := uceAln.Subseq(i, i+1)
// 		bCounts := sitewiseMulti(site)
// 		for k, v := range bCounts {
// 			counts[k][i] += v
// 		}
// 	}
// 	return counts
// }

func SitewiseEntropy(aln nexus.Alignment, chars []byte) []float64 {
	entropies := make([]float64, aln.Len())
	for i := 0; i < aln.Len(); i++ {
		site := aln.Subseq(i, i+1)
		entropies[i] = entropy.AlignmentEntropy(site, chars)
	}
	return entropies
}

// SitewiseBaseCounts returns a 4xN array of base counts where N is the number of sites
func SitewiseBaseCounts(uceAln nexus.Alignment, letters []byte) map[byte][]int {
	counts := make(map[byte][]int)
	for _, l := range letters {
		counts[l] = make([]int, uceAln.Len())
	}
	for i := 0; i < uceAln.Len(); i++ {
		site := uceAln.Subseq(i, i+1)
		bCounts := site.Count(letters)
		for k, v := range bCounts {
			counts[k][i] += v
		}
	}
	return counts
}

func SitewiseGc(uceAln nexus.Alignment) []float64 {
	gc := make([]float64, uceAln.Len())
	for i := range gc {
		site := strings.ToTitle(string(uceAln.Column(uint(i))))
		for _, nuc := range site {
			if nuc == 'G' || nuc == 'C' {
				gc[i]++
			}
		}
		gc[i] = gc[i] / float64(uceAln.NSeq())
	}
	return gc
}
