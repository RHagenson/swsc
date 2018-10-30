package nexus

// Pair is a pair of integer values (typically represents a range)
type Pair [2]int

// First is the first value of the Pair
func (p *Pair) First() int {
	return p[0]
}

// Second is the second value of the Pair
func (p *Pair) Second() int {
	return p[1]
}

// newPair enforces that start is less than or equal to stop
func newPair(start, stop int) Pair {
	if start <= stop {
		return Pair{start, stop}
	}
	panic("Pair with start > stop attempted")
}
