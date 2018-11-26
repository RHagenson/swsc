package utils

import (
	"fmt"
	"math"

	"github.com/pkg/errors"
)

// ValidateMinWin checks if minWin has been set too large to create proper flanks and core
func ValidateMinWin(length, minWin int) error {
	if length/3 < minWin {
		msg := fmt.Sprintf(
			"minWin is too large, maximum allowed value is length/3 or %d\n",
			length/3,
		)
		return errors.New(msg)
	}
	return nil
}

// MinInCountsMap returns the minimum value in the map
func MinInCountsMap(counts map[byte]int) int {
	min := math.MaxInt16
	for _, val := range counts {
		if val < min {
			min = val
		}
	}
	return min
}

// MaxInFreqMap returns the maximum value in the map
func MaxInFreqMap(freqs map[byte]float64) float64 {
	max := math.SmallestNonzeroFloat64
	for _, val := range freqs {
		if max < val {
			max = val
		}
	}
	return max
}
