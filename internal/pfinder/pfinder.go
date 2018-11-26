package pfinder

import (
	"fmt"
)

// StartBlock writes PartitionFinder2 configuration header/start block
func StartBlock(datasetName string) string {
	branchLengths := "linked"
	models := "GTR+G"
	modelSelection := "aicc"

	block := "## ALIGNMENT FILE ##\n" +
		fmt.Sprintf("alignment = %s.nex;\n\n", datasetName) +
		"## BRANCHLENGTHS: linked | unlinked ##\n" +
		fmt.Sprintf("branchlengths = %s;\n\n", branchLengths) +
		"MODELS OF EVOLUTION: all | allx | mybayes | beast | gamma | gammai <list> ##\n" +
		fmt.Sprintf("models = %s;\n\n", models) +
		"# MODEL SELECTION: AIC | AICc | BIC #\n" +
		fmt.Sprintf("model_selection = %s;\n\n", modelSelection) +
		"## DATA BLOCKS: see manual for how to define ##\n" +
		"[data_blocks]\n"
	return block
}

// ConfigBlock appends the proper window size for the UCE
// If their are either undetermined or blocks w/o all sites the fullRange should be used
func ConfigBlock(name string, bestWindow [2]int, start, stop int, fullRange bool) string {
	block := ""
	if fullRange || bestWindow[1]-bestWindow[0] == stop-start {
		block = fmt.Sprintf("%s_all = %d-%d;\n", name, start, stop)
	} else {
		// left UCE
		leftStart := start
		leftEnd := start + bestWindow[0]
		// core UCE
		coreStart := leftEnd + 1
		coreEnd := start + bestWindow[1]
		// right UCE
		rightStart := coreEnd + 1
		rightEnd := stop
		block = fmt.Sprintf("%s_left = %d-%d;\n", name, leftStart, leftEnd) +
			fmt.Sprintf("%s_core = %d-%d;\n", name, coreStart, coreEnd) +
			fmt.Sprintf("%s_right = %d-%d;\n", name, rightStart, rightEnd)
	}

	return block
}

// EndBlock appends the end block to the specified .cfg file
func EndBlock() string {
	search := "rclusterf"
	block := "\n" +
		"## SCHEMES, search: all | user | greedy | rcluster | hcluster | kmeans ##\n" +
		"[schemes]\n" +
		fmt.Sprintf("search = %s;\n\n", search)
	return block
}
