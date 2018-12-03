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

type winWVals struct {
	win      Window
	sqerr    float64
	variance float64
}

// GetBestN gets the best N windows for each metric.
// Quality is determined by sum of square error of metric, variance, and user-preference for size of core.
func GetBestN(mets map[metrics.Metric][]float64, wins []Window, stop int, largeCore bool, n uint) map[metrics.Metric][]Window {
	// 1) Init necessary space
	sses := make(map[metrics.Metric][]winWVals, len(mets))
	for m := range mets {
		sses[m] = make([]winWVals, len(wins))
	}

	// 2) Get SSE and variance values for each cell in array
	for i, win := range wins {
		for m, v := range getSses(mets, win) {
			sses[m][i].win = win
			sses[m][i].sqerr = v
			sses[m][i].variance = winVariance(win, stop)
		}
	}

	// 3) Sort by lowest sum of square errors, lowest variance, then user-preference for window size.
	for m := range sses {
		sort.SliceStable(sses[m], func(i, j int) bool {
			return sses[m][i].sqerr < sses[m][j].sqerr
		})
		sort.SliceStable(sses[m], func(i, j int) bool {
			return sses[m][i].variance < sses[m][j].variance
		})
		if largeCore {
			sort.SliceStable(sses[m], func(i, j int) bool {
				// Largest window first
				wini := sses[m][i].win.Stop() - sses[m][i].win.Start()
				winj := sses[m][j].win.Stop() - sses[m][j].win.Start()
				return wini > winj
			})
		} else {
			sort.SliceStable(sses[m], func(i, j int) bool {
				// Smallest window first
				wini := sses[m][i].win.Stop() - sses[m][i].win.Start()
				winj := sses[m][j].win.Stop() - sses[m][j].win.Start()
				return wini < winj
			})
		}
	}

	// 4) Pull out the best windows
	out := make(map[metrics.Metric][]Window, len(mets))
	for m := range sses {
		out[m] = make([]Window, n)
		for i := range out[m] {
			out[m][i] = sses[m][i].win
		}
	}

	return out
}

func GetBest(mets map[metrics.Metric][]float64, wins []Window, stop int, largeCore bool) map[metrics.Metric]Window {
	// 1) Make an empty array
	// rows = number of metrics
	// columns = number of windows
	// data = nil, allocate new backing slice
	// Each "cell" of the matrix created by {metric}x{window} is the position-wise SSE for that combination
	sses := make(map[metrics.Metric]map[Window]float64)

	// 2) Get SSE for each cell in array
	for _, win := range wins {
		// Get SSEs for a given Window
		for m, v := range getSses(mets, win) {
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
	for m, wins := range minMetricWindows {
		/*
			Sort windows before calculating window variances
			Must be done or the random order means equal sized windows are equivalent
		*/
		sort.SliceStable(wins, func(i, j int) bool {
			// Lowest Start first
			return wins[i].Start() < wins[j].Start()
		})
		if largeCore {
			sort.SliceStable(wins, func(i, j int) bool {
				// Largest window first
				wini := wins[i].Stop() - wins[i].Start()
				winj := wins[j].Stop() - wins[j].Start()
				return wini > winj
			})
		} else {
			sort.SliceStable(wins, func(i, j int) bool {
				// Smallest window first
				wini := wins[i].Stop() - wins[i].Start()
				winj := wins[j].Stop() - wins[j].Start()
				return wini < winj
			})
		}
		absMinWindow[m] = getMinVarWindow(wins, stop)
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
func GenerateCandidates(start, stop, min int) []Window {
	fwdWins := (stop - start - min - min) / min
	var wins []Window

	mod := (stop - start) % min

	if mod == 0 { // Only need to produce forward series
		wins = make([]Window, fwdWins)
		for i := range wins {
			offset := min * (i + 1)
			wins[i] = Window{start + offset, start + offset + min - 1}
		}
	} else { // Need to produce forward and reverse series (revese series is forward+mod)
		wins = make([]Window, (fwdWins)*2)
		for i := 0; i < len(wins)/2; i++ {
			offset := min * (i + 1)
			wins[i] = Window{start + offset, start + offset + min - 1}
			wins[len(wins)/2+i] = Window{start + offset + mod, start + offset + min + mod - 1}
		}
	}

	return wins
}

func ExtendCandidate(w Window, start, stop, minWin int) []Window {
	var wins []Window
	firstStart := int(math.Max(float64(w.Start()-minWin), float64(start+minWin)))
	lastEnd := int(math.Min(float64(w.Stop()+minWin), float64(stop-minWin)))

	for start := firstStart; start <= lastEnd-minWin; start++ {
		for end := start + minWin; end <= lastEnd; end++ {
			wins = append(wins, Window{start, end})
		}
	}

	return wins
}

func getMinVarWindowN(n int, windows []Window, stop int) []Window {
	var (
		bestWindow = make([]Window, n)
		vars       = make([]struct {
			w Window
			v float64
		}, len(windows))
	)
	for i, w := range windows {
		vars[i].w = w
		vars[i].v = winVariance(w, stop)
	}
	sort.Slice(vars, func(i, j int) bool {
		return vars[i].v < vars[i].v
	})

	for i := 0; i < n; i++ {
		bestWindow[i] = vars[i].w
	}

	return bestWindow
}

func winVariance(w Window, stop int) float64 {
	left := float64(w.Start())
	core := float64(w.Stop() - w.Start())
	right := float64(stop - w.Stop())
	return stat.Variance([]float64{left, core, right}, nil)
}

func getMinVarWindow(windows []Window, stop int) Window {
	return getMinVarWindowN(1, windows, stop)[0]
}

// anyUndeterminedBlocks checks if any blocks are only undetermined/ambiguous characters
// Not the same as anyBlocksWoAllSites()
func anyUndeterminedBlocks(bestWindow Window, aln *nexus.Alignment, chars []byte) bool {
	leftAln := aln.Subseq(-1, bestWindow.Start())
	coreAln := aln.Subseq(bestWindow.Start(), bestWindow.Stop())
	rightAln := aln.Subseq(bestWindow.Stop(), -1)

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
func anyBlocksWoAllSites(bestWindow Window, aln *nexus.Alignment, chars []byte) bool {
	leftAln := aln.Subseq(-1, bestWindow.Start())
	coreAln := aln.Subseq(bestWindow.Start(), bestWindow.Stop())
	rightAln := aln.Subseq(bestWindow.Stop(), -1)

	leftCounts := leftAln.Count(chars)
	coreCounts := coreAln.Count(chars)
	rightCounts := rightAln.Count(chars)

	if utils.MinInCountsMap(leftCounts) == 0 || utils.MinInCountsMap(coreCounts) == 0 || utils.MinInCountsMap(rightCounts) == 0 {
		return true
	}
	return false
}

// UseFullRange checks invariant conditions and returns if any are true
func UseFullRange(bestWindow Window, aln *nexus.Alignment, chars []byte) bool {
	return anyBlocksWoAllSites(bestWindow, aln, chars) || anyUndeterminedBlocks(bestWindow, aln, chars)
}
