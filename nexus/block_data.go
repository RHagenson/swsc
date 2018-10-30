package nexus

import (
	"log"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type dataBlock struct {
	ntax      int       // Number of taxa
	nchar     int       // Number of characters
	dataType  string    // Data type (e.g. DNA, RNA, Nucleotide, Protein)
	gap       byte      // Gap element character
	missing   byte      // Missing element character
	alignment Alignment // All sequences under consideration
}

func processDataBlock(lines []string, nex *Nexus) {
	block := new(dataBlock)
	for i := 0; i < len(lines); i++ {
		fields := strings.Fields(strings.ToUpper(lines[i]))
		if len(fields) != 0 {
			switch fields[0] {
			case "DIMENSIONS":
				var err error
				for _, word := range fields[1:] {
					if strings.HasPrefix(word, "NTAX=") {
						pair := strings.FieldsFunc(word, splitOnChar('='))
						strnum := strings.Trim(pair[1], " ;")
						block.ntax, err = strconv.Atoi(strnum)
						if err != nil {
							err = errors.Wrap(err, "Could not convert to int")
							log.Println(err)
						}
					}
					if strings.HasPrefix(word, "NCHAR") {
						pair := strings.FieldsFunc(word, splitOnChar('='))
						strnum := strings.Trim(pair[1], " ;")
						block.nchar, err = strconv.Atoi(strnum)
						if err != nil {
							err = errors.Wrap(err, "Could not convert to int")
							log.Println(err)
						}
					}
				}
			case "FORMAT":
				for _, word := range fields[1:] {
					if strings.HasPrefix(word, "DATATYPE=") {
						pair := strings.FieldsFunc(word, splitOnChar('='))
						block.dataType = strings.Trim(pair[1], " ;")
					}
					if strings.HasPrefix(word, "GAP=") {
						pair := strings.FieldsFunc(word, splitOnChar('='))
						block.gap = strings.Trim(pair[1], " ;")[0]
					}
					if strings.HasPrefix(word, "MISSING=") {
						pair := strings.FieldsFunc(word, splitOnChar('='))
						block.missing = strings.Trim(pair[1], " ;")[0]
					}
				}
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
	block.makeAlignEqual()
	nex.data = block
	return
}

// makeAlignEqual inflates any shorter alignment sequences with the missing character
func (d *dataBlock) makeAlignEqual() {
	length := d.nchar
	for i, v := range d.alignment {
		if len(v) < length {
			d.alignment[i] = d.alignment[i] +
				strings.Repeat(string(d.missing), len(v)-length)
		}
	}
	return
}
