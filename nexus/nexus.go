package nexus

import (
	"bufio"
	"io"
	"log"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
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

// Nexus is an Only understands two blocks: DATA and SETS
// meant for exclusive use in swsc
type Nexus struct {
	handChan chan interface{}
	handlers map[string]func(chan interface{}, []string)
	data     *dataBlock
	sets     *setsBlock
}

type dataBlock struct {
	ntax      int               // Number of taxa
	nchar     int               // Number of characters
	dataType  string            // Data type (e.g. DNA, RNA, Nucleotide, Protein)
	gap       byte              // Gap element character
	missing   byte              // Missing element character
	alignment map[string]string // Mapping for ID -> sequence
}

type setsBlock struct {
	charSets      map[string][]pair            // Map from charset-name -> []pair
	charPartition map[string]map[string]string // Map partition-name -> subset-name -> charset-set||charset-name
}

type pair struct {
	start int
	stop  int
}

// newPair enforces that start is less or equal to stop, if not it returns the reverse
func newPair(start, stop int) pair {
	if start > stop {
		return pair{start: stop, stop: start}
	}
	return pair{start: start, stop: stop}
}

func (nex *Nexus) mutator() {
	for {
		block := <-nex.handChan
		switch block.(type) {
		case *dataBlock:
			nex.data = block.(*dataBlock)
		case *setsBlock:
			nex.sets = block.(*setsBlock)
		case *struct{}:
			return
		}
	}
}

// New creates a new empty Nexus with registered handlers and deferred block creation
func New() *Nexus {
	nex := &Nexus{
		handChan: make(chan interface{}),
		handlers: map[string]func(chan interface{}, []string){
			"DATA": processDataBlock,
			"SETS": processSetsBlock,
		},
		data: new(dataBlock),
		sets: new(setsBlock),
	}
	go nex.mutator()
	runtime.SetFinalizer(nex, func(nex *Nexus) {
		nex.handChan <- new(struct{})
		return
	})
	return nex
}

// Read fills in the Nexus with data from a file
func (nex *Nexus) Read(file io.Reader) {
	inProgressBlocks := new(sync.WaitGroup)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		for k, f := range nex.handlers {
			if blockByName(scanner.Text(), k) {
				lines := copyLines(scanner)
				c := make(chan interface{})
				go func(c1, c2 chan interface{}) { // Spawn a goroutine to track blocks still in process
					inProgressBlocks.Add(1)
					c1 <- <-c2
					inProgressBlocks.Done()
				}(nex.handChan, c)
				go f(c, lines)
			} else {
				log.Printf("Ignored line:\n\t%q\n", k)
			}
		}
	}
	inProgressBlocks.Wait()
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

func processSetsBlock(c chan interface{}, lines []string) {
	block := new(setsBlock)
	for i := 0; i < len(lines); i++ {
		fields := strings.Fields(strings.ToUpper(lines[i]))
		if len(fields) != 0 {
			switch fields[0] {
			case "CHARSET":
				charsetName := fields[1]
				for _, field := range fields[2:] {
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
				partitionName := fields[1]
				for _, field := range fields {
					entry := strings.TrimRight(field, ";")
					if strings.Contains(entry, ":") {
						split := strings.Split(entry, ":")
						subsetName := split[0]
						charSet := split[len(split)-1]
						block.charPartition[partitionName][subsetName] = charSet
					}
				}
			default:
				log.Printf("Did not understand\n\t%s\nin SETS block\n", lines[i])
			}
		}
	}
	c <- block
	return
}

func processDataBlock(c chan interface{}, lines []string) {
	block := new(dataBlock)
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
				for j, seqlin := range lines[i+1:] {
					if fields := strings.Fields(seqlin); len(fields) == 2 {
						block.alignment[fields[0]] = fields[1]
					} else if strings.Contains(seqlin, ";") {
						i = j
						break
					}
				}
			default:
				log.Printf("Did not understand\n\t%s\nin DATA block\n", lines[i])
			}
		}
	}
	c <- block
	return
}

func blockByName(s, b string) bool {
	return strings.HasSuffix(strings.ToUpper(s), strings.ToUpper(" "+b+";"))
}
