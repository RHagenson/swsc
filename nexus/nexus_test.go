package nexus_test

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"bitbucket.org/rhagenson/swsc/nexus"
)

func TestNexus(t *testing.T) {
	datafile := "./testdata/example_input.nex"
	nex := nexus.New()
	in, err := os.Open(datafile)
	if err != nil {
		t.Fatalf("Could not open example input: %s\n", datafile)
	}
	nex.FillFrom(in)

	t.Run("FillFrom matches Read", func(t *testing.T) {
		rin, _ := os.Open(datafile)
		rnex := nexus.Read(rin)

		if nex.NTax() != rnex.NTax() {
			t.Errorf("NTax: FillFrom got %v, Read got %v", nex.NTax(), rnex.NTax())
		}
		if nex.NChar() != rnex.NChar() {
			t.Errorf("NChar: FillFrom got %v, Read got %v", nex.NChar(), rnex.NChar())
		}
		if nex.DataType() != rnex.DataType() {
			t.Errorf("DataType: FillFrom got %v, Read got %v", nex.DataType(), rnex.DataType())
		}
		if nex.Gap() != rnex.Gap() {
			t.Errorf("Gap: FillFrom got %v, Read got %v", nex.Gap(), rnex.Gap())
		}
		if nex.Missing() != rnex.Missing() {
			t.Errorf("Missing: FillFrom got %v, Read got %v", nex.Missing(), rnex.Missing())
		}
		if !reflect.DeepEqual(nex.Charsets(), rnex.Charsets()) {
			t.Errorf("Charsets: FillFrom got %v, Read got %v", nex.Charsets(), rnex.Charsets())
		}
		if !reflect.DeepEqual(nex.Alignment(), rnex.Alignment()) {
			t.Errorf("Alignment: FillFrom got %v, Read got %v", nex.Alignment(), rnex.Alignment())
		}
	})

	t.Run("Data", func(t *testing.T) {
		t.Run("NTax", func(t *testing.T) {
			if nex.NTax() != 10 {
				t.Errorf("NTax() expected %d, Got %d", 10, nex.NTax())
			}
		})
		// TODO: Second half can be replaced by a proptest
		t.Run("NChar", func(t *testing.T) {
			if nex.NChar() != 5786 && nex.NChar() != nex.Alignment().Len() {
				t.Errorf("NChar() expected %d, Got %d", 5786, nex.NChar())
			}
		})
		t.Run("DataType", func(t *testing.T) {
			if nex.DataType() != "DNA" {
				t.Errorf("DataType() expected %s, Got %s", "DNA", nex.DataType())
			}
		})
		t.Run("Gap", func(t *testing.T) {
			if nex.Gap() != '-' {
				t.Errorf("Gap() expected %q, Got %q", '-', nex.Gap())
			}
		})
		t.Run("Missing", func(t *testing.T) {
			if nex.Missing() != '?' {
				t.Errorf("Missing() expected %q, Got %q", '?', nex.Missing())
			}
		})
		t.Run("Alignment", func(t *testing.T) {
			aln := nex.Alignment()
			if aln.Len() != 5786 && aln.Len() != nex.NChar() {
				t.Errorf("Length was %d, got %d\n", aln.Len(), 5786)
			}
			if aln.NSeq() != 10 {
				t.Errorf("Number of sequences was %d, got %d\n", aln.NSeq(), 10)
			}
			// TODO: Can be replace by proptests over message statements
			t.Run("Subseq", func(t *testing.T) {
				if l := aln.Subseq(-1, -1).Len(); l != aln.Len() {
					t.Errorf("Subseq(-1,-1) did not return whole alignment. Got: %d\n", l)
				}
				if l := aln.Subseq(-1, 10).Len(); l != 10 {
					t.Errorf("Subseq(-1, N) should return N length alignment. Got: %d\n", l)
				}
				if l := aln.Subseq(10, 20).Len(); l != 10 {
					t.Errorf("Subseq(N,M) should return M-N length alignment. Got: %d\n", l)
				}
				if l := aln.Subseq(aln.Len()-10, -1).Len(); l != 10 {
					t.Errorf("Subseq(len-N, -1) should return N length alignment. Got: %d\n", l)
				}
			})
			t.Run("Seq", func(t *testing.T) {
				for i := uint(0); i < aln.NSeq(); i++ {
					if aln.Seq(i) != strings.ToUpper(aln.Seq(i)) {
						t.Errorf("Alignment should not be case-sensitive.\n")
					}
				}
			})
			// TODO: Can be replace by a proptest that randomly reassigns `ind`
			t.Run("Column", func(t *testing.T) {
				ind := uint(0)
				col := aln.Column(ind)
				for i := uint(0); i < uint(len(col)); i++ {
					seqi := aln.Seq(i)
					if col[i] != seqi[ind] {
						t.Errorf("Column returned an incorrect slice.\n")
					}
				}
			})
			t.Run("NSeq", func(t *testing.T) {
				if aln.NSeq() != uint(len(aln)) {
					t.Errorf("Length of alignment is not the number of sequences")
				}
			})
			t.Run("Seqs only contain ATGC+missing+gap characters", func(t *testing.T) {
				for i := uint(0); i < aln.NSeq(); i++ {
					seq := aln.Seq(i)
					sum := 0
					for _, r := range nex.Letters() {
						sum += strings.Count(seq, string(r))
					}
					if sum != len(seq) {
						t.Errorf("Sequence contains invalid characters.")
					}
				}
			})
			t.Run("String", func(t *testing.T) {
				str := aln.String()
				seqs := strings.Split(strings.TrimSpace(str), "\n")
				for i := 0; i < len(seqs); i++ {
					if len(seqs[i]) != len(seqs[i/len(seqs)]) {
						t.Errorf("Lengths of sequences %d and %d do not match at %d and %d, respectively",
							i, i/len(seqs), len(seqs[i]), len(seqs[i/len(seqs)]))
					}
				}
				for i := aln.Len(); i < len(str); i += aln.Len() + 1 {
					if str[i] != '\n' {
						t.Errorf("Newline expected, got %q (range %q)", str[i], str[i-5:i+5])
					}
				}
			})
		})
	})
}
