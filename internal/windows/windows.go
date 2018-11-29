package windows

import (
	"math"
	"sort"

	"bitbucket.org/rhagenson/swsc/internal/metrics"
	"bitbucket.org/rhagenson/swsc/internal/nexus"
	"bitbucket.org/rhagenson/swsc/internal/utils"
	"gonum.org/v1/gonum/floats"
	"gonum.org/v1/gonum/stat"
)

// Window is an inclusive window into a UCE
type Window [2]int

func New(start, stop int) Window {
	if stop < start {
		return Window{stop, start}
	}
	return Window{start, stop}
}

// Start is the starting position of a window
func (w *Window) Start() int {
	return w[0]
}

// Stop is the inclusive stopping position of a window
func (w *Window) Stop() int {
	return w[1]
}

func GetAll(uceAln nexus.Alignment, minWin int) []Window {
	windows := GenerateWindows(uceAln.Len(), minWin)
	return windows
}

func GetBest(mets map[metrics.Metric][]float64, windows []Window, alnLen int, inVarSites []bool, largeCore bool) map[metrics.Metric]Window {
	// 1) Make an empty array
	// rows = number of metrics
	// columns = number of windows
	// data = nil, allocate new backing slice
	// Each "cell" of the matrix created by {metric}x{window} is the position-wise SSE for that combination
	sses := make(map[metrics.Metric]map[Window]float64)

	// 2) Get SSE for each cell in array
	for _, win := range windows {
		// Get SSEs for a given Window
		for m, v := range getSses(mets, win, inVarSites) {
			if _, ok := sses[m]; !ok {
				sses[m] = make(map[Window]float64, 1)
				sses[m][win] = v
			} else {
				sses[m][win] = v
			}
		}
	}

	// Find minimum values and record the window(s) they occur in
	minMetricWindows := make(map[metrics.Metric][]Window)
	for m, windows := range sses {
		bestVal := math.MaxFloat64
		for w, val := range windows {
			if val < bestVal {
				bestVal = val
				minMetricWindows[m] = []Window{w}
			} else if floats.EqualWithinAbs(val, bestVal, 1e-10) {
				minMetricWindows[m] = append(minMetricWindows[m], w)
			}
		}
	}

	absMinWindow := make(map[metrics.Metric]Window)
	for m := range minMetricWindows {
		/*
			Sort windows before calculating window variances
			Must be done or the random order means equal sized windows are equivalent
			TODO: Devise a better method to resolve ties. This way an earlier window is preferred.
		*/
		wins := minMetricWindows[m]
		sort.SliceStable(wins[:], func(i, j int) bool {
			// Lowest Start first
			return wins[i].Start() < wins[j].Start()
		})
		if largeCore {
			sort.SliceStable(wins[:], func(i, j int) bool {
				// Smallest window first
				wini := wins[i].Stop() - wins[i].Start()
				winj := wins[j].Stop() - wins[j].Start()
				return wini > winj
			})
		} else {
			sort.SliceStable(wins[:], func(i, j int) bool {
				// Smallest window first
				wini := wins[i].Stop() - wins[i].Start()
				winj := wins[j].Stop() - wins[j].Start()
				return wini < winj
			})
		}
		absMinWindow[m] = getMinVarWindow(wins, alnLen)
	}

	return absMinWindow
}

// GenerateWindows produces windows of at least a minimum size given a total length
// Windows must be:
//   1) at least minimum window from the start of the UCE (ie, first start at minimum+1)
//   2) at least minimum window from the end of the UCE (ie, last end at length-minimum+1)
//   3) at least minimum window in length (ie, window{start, end)})
// Input is treated inclusively, but returned with exclusive stop indexes
func GenerateWindows(length, min int) []Window {
	n := (length - min*3) + 1            // Make range inclusive
	windows := make([]Window, n*(n+1)/2) // Length is sum of all numbers n and below
	i := 0
	for start := min; start+min+min <= length; start++ {
		for end := start + min; end+min <= length; end++ {
			windows[i] = Window{start, end + 1}
			i++
		}
	}
	return windows
}

// GenerateCandidates produces candidate windows of minimum size
// spanning the total length with a minimum/2 overlap
// Windows must be:
//   1) at least minimum window from the start of the UCE (ie, first start at minimum+1)
//   2) at least minimum window from the end of the UCE (ie, last end at length-minimum+1)
//   3) at least minimum window in length (ie, window{start, end)})
// Input is treated inclusively, but returned with exclusive stop indexes
// TODO: When length % min != 0, must produce candidates from the end as well
func GenerateCandidates(length, min int) []Window {
	nNonOverlapping := length / min
	var wins []Window

	mod := length % min

	if mod == 0 { // Only need to produce forward series
		wins = make([]Window, 0, nNonOverlapping+1)
		for start := min; start+min+min <= length; start += min / 2 {
			wins = append(wins, Window{start, start + min})
		}
	} else { // Need to produce forward and reverse series (revese series is forward+mod)
		wins = make([]Window, 0, (nNonOverlapping+1)*2)
		for start := min; start+min+min <= length; start += min / 2 {
			wins = append(wins, Window{start, start + min}, Window{start + mod, start + min + mod})
		}
	}

	return wins
}

func getMinVarWindow(windows []Window, alnLength int) Window {
	best := math.MaxFloat64
	var bestWindow Window

	for _, w := range windows {
		left := float64(w.Start())
		core := float64(w.Stop() - w.Start())
		right := float64(alnLength - w.Stop())
		variance := stat.Variance([]float64{left, core, right}, nil)
		if variance < best {
			best = variance
			bestWindow = w
		}
	}
	return bestWindow
}

// anyUndeterminedBlocks checks if any blocks are only undetermined/ambiguous characters
// Not the same as anyBlocksWoAllSites()
func anyUndeterminedBlocks(bestWindow Window, uceAln nexus.Alignment, chars []byte) bool {
	leftAln := uceAln.Subseq(-1, bestWindow.Start())
	coreAln := uceAln.Subseq(bestWindow.Start(), bestWindow.Stop())
	rightAln := uceAln.Subseq(bestWindow.Stop(), -1)

	leftFreq := leftAln.Frequency(chars)
	coreFreq := coreAln.Frequency(chars)
	rightFreq := rightAln.Frequency(chars)

	// If any frequency is NaN
	// TODO: Likely better with bpFreqCalc returning an error value
	if utils.MaxInFreqMap(leftFreq) == 0 || utils.MaxInFreqMap(coreFreq) == 0 || utils.MaxInFreqMap(rightFreq) == 0 {
		return true
	}
	return false
}

// anyBlocksWoAllSites checks for blocks with only undetermined/ambiguous characters
// Not the same as anyUndeterminedBlocks()
func anyBlocksWoAllSites(bestWindow Window, uceAln nexus.Alignment, chars []byte) bool {
	leftAln := uceAln.Subseq(-1, bestWindow.Start())
	coreAln := uceAln.Subseq(bestWindow.Start(), bestWindow.Stop())
	rightAln := uceAln.Subseq(bestWindow.Stop(), -1)

	leftCounts := leftAln.Count(chars)
	coreCounts := coreAln.Count(chars)
	rightCounts := rightAln.Count(chars)

	if utils.MinInCountsMap(leftCounts) == 0 || utils.MinInCountsMap(coreCounts) == 0 || utils.MinInCountsMap(rightCounts) == 0 {
		return true
	}
	return false
}

// UseFullRange checks invariant conditions and returns if any are true
func UseFullRange(bestWindow Window, uceAln nexus.Alignment, chars []byte) bool {
	return anyBlocksWoAllSites(bestWindow, uceAln, chars) || anyUndeterminedBlocks(bestWindow, uceAln, chars)
}
