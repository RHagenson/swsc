package nexus_test

import (
	"os"
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
	nex.Read(in)

	t.Run("Alignment", func(t *testing.T) {
		aln := nex.Alignment()
		if aln.Len() != 5786 {
			t.Errorf("Length was %d, got %d\n", aln.Len(), 5786)
		}
		if aln.NSeq() != 10 {
			t.Errorf("Number of sequences was %d, got %d\n", aln.NSeq(), 10)
		}
		t.Run("Subseq", func(t *testing.T) {
			if aln.Subseq(-1, -1).Len() != aln.Len() {
				t.Errorf("Subseq(-1,-1) did not return whole alignment.\n")
			}
			if aln.Subseq(-1, 10).Len() != 10 {
				t.Errorf("Subseq(-1, N) should return N length alignment.\n")
			}
			if aln.Subseq(10, 20).Len() != 10 {
				t.Errorf("Subseq(N,M) should return M-N length alignment.\n")
			}
			if aln.Subseq(aln.Len()-10, -1).Len() != 10 {
				t.Errorf("Subseq(len-N, -1) should return N length alignment.\n")
			}
		})
		t.Run("Seq", func(t *testing.T) {
			for i := uint(0); i < aln.NSeq(); i++ {
				if aln.Seq(i) != strings.ToUpper(aln.Seq(i)) {
					t.Errorf("Alignment should not be case-sensitive.\n")
				}
			}
		})
	})

}
