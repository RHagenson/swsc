package nexus

import (
	"log"
	"regexp"
	"strconv"
	"strings"
)

// setsBlock stores sets of objects (characters, states, taxa, etc.)
// Currently only concerned with sets of characters (as needed by swsc)
type setsBlock struct {
	charSets      map[string][]Pair            // Map from charset-name -> []pair
	charPartition map[string]map[string]string // Map partition-name -> subset-name -> charset-set||charset-name
}

// processSetsBlock parses the SETS block' lines and writes to the passed Nexus
func processSetsBlock(lines []string, nex *Nexus) {
	block := new(setsBlock)
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
	nex.sets = block
	return
}
