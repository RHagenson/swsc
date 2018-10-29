package nexus

import (
	"bufio"
	"io"
	"log"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

var (
	// Example: DIMENSIONS NTAX=10 NCHAR=5786;
	rgNtax  = regexp.MustCompile(`NTAX\s*=\s*(?P<ntax>\d+)`)
	rgNchar = regexp.MustCompile(`NCHAR\s*=\s*(?P<nchar>\d+)`)

	// Example: FORMAT DATATYPE=DNA GAP=- MISSING=?;
	rgDatatype = regexp.MustCompile(`DATATYPE\s*=\s*(?P<datatype>.+)`)
	rgGap      = regexp.MustCompile(`GAP\s*=\s*(?P<gap>.)`)
	rgMissing  = regexp.MustCompile(`MISSING\s*=\s*(?P<missing>.)`)
)

// Nexus only understands two blocks: DATA and SETS
// Note: meant for exclusive use in swsc
type Nexus struct {
	mailbox  chan interface{}
	handlers map[string]func(*Nexus, []string)
	data     dataBlock
	sets     setsBlock
}

type dataBlock struct {
	ntax      int       // Number of taxa
	nchar     int       // Number of characters
	dataType  string    // Data type (e.g. DNA, RNA, Nucleotide, Protein)
	gap       byte      // Gap element character
	missing   byte      // Missing element character
	alignment Alignment // All sequences under consideration
}

type Alignment []string

// Column is the letters from each internal sequence at position p
func (aln Alignment) Column(p uint) []byte {
	pos := make([]byte, aln.NSeq())
	for i := uint(0); i < aln.NSeq(); i++ {
		pos[i] = aln.Seq(i)[p]
	}
	return pos
}

func (aln Alignment) NSeq() uint {
	return uint(len(aln))
}

func (aln Alignment) Seq(i uint) string {
	return aln[i]
}

// Subseq creates an array of slices from the original array
// A negative start or end is interpreted as ultimate start or end of alignment
func (aln Alignment) Subseq(s, e int) Alignment {
	subseqs := make(Alignment, 0)
	start := 0
	end := len(aln[0])
	for _, seq := range aln {
		switch {
		case start <= s && e <= end: // Internal slice
			subseqs = append(subseqs, seq[s:e])
		case s < 0 && e < 0: // Whole sequence
			subseqs = append(subseqs, seq[start:end])
		case s < 0 && e <= end: // Ultimate start to defined end
			subseqs = append(subseqs, seq[start:e])
		case start <= s && end < 0: // Defined start to ultimate end
			subseqs = append(subseqs, seq[s:end])
		}
	}
	return subseqs
}

func (aln Alignment) String() string {
	str := ""
	for i := range aln {
		str += aln[i] + "\n"
	}
	return str
}

func (aln Alignment) Len() (length int) {
	for i := range aln {
		length = len(aln[i])
		return
	}
	return
}

type setsBlock struct {
	charSets      map[string][]Pair            // Map from charset-name -> []pair
	charPartition map[string]map[string]string // Map partition-name -> subset-name -> charset-set||charset-name
}

type Pair struct {
	start int
	stop  int
}

func (p *Pair) Start() int {
	return p.start
}

func (p *Pair) Stop() int {
	return p.stop
}

// newPair enforces that start is less or equal to stop, if not it returns the reverse
func newPair(start, stop int) Pair {
	if start > stop {
		return Pair{start: stop, stop: start}
	}
	return Pair{start: start, stop: stop}
}

func (nex *Nexus) mailreader() {
	for {
		block := <-nex.mailbox
		switch block.(type) {
		case *dataBlock:
			nex.data = block.(dataBlock)
		case *setsBlock:
			nex.sets = block.(setsBlock)
		case *struct{}:
			return
		}
	}
}

// New creates a new empty Nexus with registered handlers and deferred block creation
func New() *Nexus {
	nex := &Nexus{
		mailbox: make(chan interface{}),
		handlers: map[string]func(*Nexus, []string){
			"DATA": processDataBlock,
			"SETS": processSetsBlock,
		},
		data: *new(dataBlock),
		sets: *new(setsBlock),
	}
	go nex.mailreader()
	runtime.SetFinalizer(nex, func(nex *Nexus) {
		nex.mailbox <- new(struct{})
		return
	})
	return nex
}

// Read fills in the Nexus with data from a file
func (nex *Nexus) Read(file io.Reader) {
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		handled := false
		for k, f := range nex.handlers {
			if blockByName(scanner.Text(), k) {
				lines := copyLines(scanner)
				f(nex, lines)
				handled = true
			}
		}
		if !handled && scanner.Text() != "" {
			log.Printf("Scanner ignored line:\n%q\n", scanner.Text())
		}
	}
}

func (nex *Nexus) Charsets() map[string][]Pair {
	copy := make(map[string][]Pair)
	for k, v := range nex.charsets() {
		copy[k] = v
	}
	return copy
}

func (nex *Nexus) charsets() map[string][]Pair {
	return nex.sets.charSets
}

func (nex *Nexus) Alignment() Alignment {
	return nex.data.alignment
}

// copyLines extracts the lines between "BEGIN <block name>;" and "END;", trimming whitespace in the process
// The scanner should be at the BEGIN line and will be on the END line upon return
func copyLines(s *bufio.Scanner) []string {
	lines := make([]string, 0)
	for s.Scan() {
		if strings.HasPrefix(strings.ToUpper(s.Text()), "END;") {
			return lines
		}
		lines = append(lines, strings.TrimSpace(s.Text()))
	}
	return lines
}

func processSetsBlock(n *Nexus, lines []string) {
	block := *new(setsBlock)
	for i := 0; i < len(lines); i++ {
		fields := strings.Fields(strings.ToUpper(lines[i]))
		if len(fields) != 0 {
			switch fields[0] {
			case "CHARSET":
				block.charSets = make(map[string][]Pair, 0)
				charsetName := fields[1]
				for _, field := range fields[3:] {
					setVal := strings.TrimRight(field, ";")
					if strings.Contains(setVal, "-") {
						split := strings.Split(setVal, "-")
						start, _ := strconv.Atoi(split[0])
						stop, _ := strconv.Atoi(split[1])
						block.charSets[charsetName] = append(
							block.charSets[charsetName],
							newPair(start, stop),
						)
					} else if matched, err := regexp.MatchString(`[0-9]+`, setVal); matched && err == nil {
						val, _ := strconv.Atoi(setVal)
						block.charSets[charsetName] = append(
							block.charSets[charsetName],
							newPair(val, val),
						)
					}
				}
			case "CHARPARTITION":
				block.charPartition = make(map[string]map[string]string, 0)
				partitionName := fields[1]
				for _, field := range fields {
					entry := strings.TrimRight(field, ";")
					if strings.Contains(entry, ":") {
						split := strings.Split(entry, ":")
						subsetName := split[0]
						charSet := split[len(split)-1]
						if _, ok := block.charPartition[partitionName][subsetName]; ok {
							block.charPartition[partitionName][subsetName] = charSet
						} else {
							block.charPartition[partitionName] = make(map[string]string, 0)
							block.charPartition[partitionName][subsetName] = charSet
						}

					}
				}
			default:
				log.Printf("SET block processor ignored line:\n%q\n", lines[i])
			}
		}
	}
	n.sets = block
	return
}

func processDataBlock(n *Nexus, lines []string) {
	block := *new(dataBlock)
	for i := 0; i < len(lines); i++ {
		fields := strings.Fields(strings.ToUpper(lines[i]))
		if len(fields) != 0 {
			switch fields[0] {
			case "DIMENSIONS":
				block.ntax, _ = strconv.Atoi(rgNtax.FindString(strings.ToUpper(lines[i])))
				block.nchar, _ = strconv.Atoi(rgNchar.FindString(strings.ToUpper(lines[i])))
			case "FORMAT":
				block.dataType = rgDatatype.FindString(strings.ToUpper(lines[i]))
				block.gap = rgGap.FindString(strings.ToUpper(lines[i]))[0]
				block.missing = rgMissing.FindString(strings.ToUpper(lines[i]))[0]
			case "MATRIX":
				for j := i + 1; j < len(lines); j++ {
					if fields := strings.Fields(lines[j]); len(fields) == 2 {
						block.alignment = append(block.alignment,
							fields[1],
						)
					} else if strings.Contains(lines[j], ";") {
						i = j
						break
					}
				}
			default:
				log.Printf("DATA block processor ignored line:\n%q\n", lines[i])
			}
		}
	}
	n.data = block
	return
}

func blockByName(s, b string) bool {
	return strings.HasSuffix(strings.ToUpper(s), strings.ToUpper(" "+b+";"))
}
