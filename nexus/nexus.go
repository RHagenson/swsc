package nexus

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// Nexus is an Only understands two blocks: DATA and SETS
// meant for exclusive use in swsc
type Nexus struct {
	handlers map[string]func(*Nexus, []string)
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

// New creates a new empty Nexus with registered handlers and deferred block creation
func New() *Nexus {
	return &Nexus{
		handlers: map[string]func(*Nexus, []string){
			"data": processDataBlock,
			"sets": processSetsBlock,
		},
		data: new(dataBlock),
		sets: new(setsBlock),
	}
}

// Read fills in the Nexus with data from a file
func (nex *Nexus) Read(f string) {
	file, err := os.Open(f)
	defer file.Close()
	if err != nil {
		log.Fatalf("Could not process file: %s", err)
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		for k, f := range nex.handlers {
			if blockByName(scanner.Text(), k) {
				block := copyBlock(scanner)
				go f(nex, block)
			} else {
				log.Printf("Could not handle block %q", k)
			}
		}
	}
}

// copyBlock extracts the lines between "BEGIN <block name>;" and "END;", trimming whitespace in the process
// The scanner should be at the BEGIN line and will be on the END line upon return
func copyBlock(s *bufio.Scanner) []string {
	lines := make([]string, 0)
	for s.Scan() {
		if strings.HasPrefix(strings.ToUpper(s.Text()), "END;") {
			return lines
		}
		lines = append(lines, strings.TrimSpace(s.Text()))
	}
	return lines
}

func processSetsBlock(nex *Nexus, lines []string) {
	for _, line := range lines {
		switch {
		case strings.HasPrefix(strings.ToUpper(line), `CHARSET`):
			fields := strings.Fields(strings.TrimRight(line, ";"))
			charsetName := fields[1]
			for _, field := range fields {
				if strings.Contains(field, "-") {
					split := strings.Split(field, "-")
					start, _ := strconv.Atoi(split[0])
					stop, _ := strconv.Atoi(split[1])
					nex.sets.charSets[charsetName] = append(
						nex.sets.charSets[charsetName],
						newPair(start, stop),
					)
				} else if matched, err := regexp.MatchString(`[0-9]+`, field); matched && err == nil {
					val, _ := strconv.Atoi(field)
					nex.sets.charSets[charsetName] = append(
						nex.sets.charSets[charsetName],
						newPair(val, val),
					)
				}
			}
		case strings.HasPrefix(strings.ToUpper(line), "CHARPARTITION"):
			fields := strings.Fields(strings.TrimRight(line, ";"))
			partitionName := fields[1]
			for _, field := range fields {
				if strings.Contains(field, ":") {
					split := strings.Split(field, ":")
					subsetName := split[0]
					charSet := split[len(split)-1]
					nex.sets.charPartition[partitionName][subsetName] = charSet
				}
			}
		default:
			fmt.Printf("Did not understand\n\t%s\nin SETS block", line)
		}
	}
	return
}

func processDataBlock(nex *Nexus, lines []string) {
	var (
		// Example: DIMENSIONS NTAX=10 NCHAR=5786;
		regexDims = regexp.MustCompile(`^DIMENSIONS\s+NTAX\s*=\s*(?P<ntax>\d+)\s+NCHAR\s*=\s*(?P<nchar>\d+);$`)

		// Example: FORMAT DATATYPE=DNA GAP=- MISSING=?;
		regexFormat = regexp.MustCompile(`^FORMAT\s+DATATYPE\s*=\s*(?P<datatype>.+)\s+GAP\s*=\s*(?P<gap>.+)\s+MISSING\s*=\s*(?P<missing>.+);$`)
	)

	for i, line := range lines {
		if matches := regexDims.FindStringSubmatch(strings.ToUpper(line)); matches != nil {
			result := make(map[string]string)
			for i, name := range regexDims.SubexpNames() {
				if name != "" && i != 0 {
					result[name] = matches[i]
				}
			}
			nex.data.ntax, _ = strconv.Atoi(result["ntax"])
			nex.data.nchar, _ = strconv.Atoi(result["nchar"])
		} else if matches := regexFormat.FindStringSubmatch(strings.ToUpper(line)); matches != nil {
			result := make(map[string]string)
			for i, name := range regexFormat.SubexpNames() {
				if name != "" && i != 0 {
					result[name] = matches[i]
				}
			}
			nex.data.dataType = result["datatype"]
			nex.data.gap = result["gap"][0]
			nex.data.missing = result["missing"][0]
		} else if strings.Contains(line, "MATRIX") {
			for _, seqlin := range lines[i+1:] {
				if len(seqlin) < 5 { // Heuristic to ignore non-sequence lines
					continue
				} else {
					fields := strings.Fields(strings.TrimSpace(seqlin))
					nex.data.alignment[fields[0]] = fields[len(fields)-1]
				}
			}
		}
	}
	return
}

func blockByName(s, b string) bool {
	return strings.HasSuffix(strings.ToUpper(s), strings.ToUpper(b+";"))
}

// readLines reads a whole file into memory
// and returns a slice of its lines.
func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// TODO: Unnecessary allocations
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}
