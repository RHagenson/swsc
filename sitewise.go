package main

import "github.com/biogo/biogo/seq/multi"

func sitewiseGc(uceAln *multi.Multi) []float64 {
	gc := make([]float64, uceAln.Len())
	for i := range gc {
		site := uceAln.Column(i, false)
		// TODO: Will not compute properly if using lowercase letters
		for _, nuc := range site {
			if byte(nuc) == 'G' || byte(nuc) == 'C' {
				gc[i]++
			}
		}
		gc[i] = gc[i] / float64(uceAln.Rows())
	}
	return gc
}
