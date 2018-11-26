package utils_test

import (
	"fmt"
	"testing"

	"bitbucket.org/rhagenson/swsc/internal/utils"
	"github.com/pkg/errors"
)

func TestValidateMinWin(t *testing.T) {
	tt := []struct {
		length int
		minWin int
		valid  bool
	}{
		{50, 50, false},
		{100, 50, false},
		{150, 50, true},
		{200, 50, true},
		{300, 10, true},
	}

	for _, tc := range tt {
		got := utils.ValidateMinWin(tc.length, tc.minWin)
		msg := fmt.Sprintf(
			"minWin is too large, maximum allowed value is length/3 or %d\n",
			tc.length/3,
		)
		if tc.valid {
			if got == errors.New(msg) {
				t.Errorf("ValidateMinWin(%d, %d) => Got: %v, Expected: %v", tc.length, tc.minWin, got, nil)
			}
		} else {
			if got == nil {
				t.Errorf("ValidateMinWin(%d, %d) => Got: %v, Expected: %v", tc.length, tc.minWin, got, "!nil")
			}
		}
	}
}

func TestMinInCountsMap(t *testing.T) {
	tt := []struct {
		counts map[byte]int
		exp    int
	}{
		{
			map[byte]int{
				'A': 1,
				'T': 2,
				'G': 3,
				'C': 4,
			}, 1,
		},
		{
			map[byte]int{
				'A': 1,
				'T': 1,
				'G': 1,
				'C': 1,
			}, 1,
		},
		{
			map[byte]int{
				'A': 4,
				'T': 3,
				'G': 2,
				'C': 1,
			}, 1,
		},
		{
			map[byte]int{
				'A': 1,
				'T': 2,
				'G': 3,
				'C': 4,
				'Z': 0,
			}, 0,
		},
		{
			map[byte]int{
				'A': 1,
				'T': 2,
				'G': 3,
				'C': 4,
				'Q': -1,
			}, -1,
		},
	}

	for _, tc := range tt {
		got := utils.MinInCountsMap(tc.counts)
		if got != tc.exp {
			t.Errorf("Got %v, Expected %v", got, tc.exp)
		}
	}
}

func TestMaxInFreqsMap(t *testing.T) {
	tt := []struct {
		counts map[byte]float64
		exp    float64
	}{
		{
			map[byte]float64{
				'A': 1.0,
				'T': 2.0,
				'G': 3.0,
				'C': 4.0,
			}, 4.0,
		},
		{
			map[byte]float64{
				'A': 1.0,
				'T': 1.0,
				'G': 1.0,
				'C': 1.0,
			}, 1,
		},
		{
			map[byte]float64{
				'A': 4.0,
				'T': 3.0,
				'G': 2.0,
				'C': 1.0,
			}, 4.0,
		},
		{
			map[byte]float64{
				'A': 1.0,
				'T': 2.0,
				'G': 3.0,
				'C': 4.0,
				'Z': 0.0,
			}, 4.0,
		},
		{
			map[byte]float64{
				'A': 1.0,
				'T': 2.0,
				'G': 3.0,
				'C': 4.0,
				'Q': -1.0,
			}, 4.0,
		},
	}

	for _, tc := range tt {
		got := utils.MaxInFreqMap(tc.counts)
		if got != tc.exp {
			t.Errorf("Got %v, Expected %v", got, tc.exp)
		}
	}
}
