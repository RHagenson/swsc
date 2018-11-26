package utils

import (
	"fmt"
	"math"
	"math/big"

	"bitbucket.org/rhagenson/swsc/internal/ui"
	"github.com/pkg/errors"
)

// ValidateMinWin checks if minWin has been set too large to create proper flanks and core
func ValidateMinWin(length, minWin int) error {
	if length/3 <= minWin {
		msg := fmt.Sprintf(
			"minWin is too large, maximum allowed value is length/3 or %d\n",
			length/3,
		)
		return errors.New(msg)
	}
	return nil
}

func minFloat64(vs ...float64) float64 {
	min := math.MaxFloat64
	for _, v := range vs {
		if v < min {
			min = v
		}
	}
	return min
}

func factorial(v int) (float64, error) {
	fact := big.NewFloat(1)
	for i := 1; i <= v; i++ {
		fact.Mul(fact, big.NewFloat(float64(i)))
	}
	val, acc := fact.Float64()
	if acc == big.Exact {
		return val, nil
	}
	return val, errors.Errorf("factorial of %d was %s the true value", v, acc)
}

func factorialMatrix(vs map[byte][]int) []float64 {
	length := 0
	for _, v := range vs {
		length = len(v)
	}
	product := make([]float64, length) // vs['A'][i] * vs['T'][i] * vs['G'][i] * vs['C'][i]
	for i := range product {
		product[i] = 1.0
	}
	for i := range product {
		for nuc := range vs {
			val, err := factorial(vs[nuc][i])
			product[i] *= val
			if err != nil {
				ui.Errorf("%v", err)
			}
		}
	}
	return product
}

func MinInCountsMap(counts map[byte]int) int {
	min := math.MaxInt16
	for _, val := range counts {
		if val < min {
			min = val
		}
	}
	return min
}

func MaxInFreqMap(freqs map[byte]float64) float64 {
	max := math.SmallestNonzeroFloat64
	for _, val := range freqs {
		if max < val {
			max = val
		}
	}
	return max
}
