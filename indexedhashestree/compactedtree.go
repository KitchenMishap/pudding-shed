package indexedhashestree

// A compacted tree is built out of a shallowTree.
// Once a shallowTree is known, we can count the distribution of node sizes (how many 256-nodes, down to how many
// 2-nodes).
// We aim to optimize the parameters of the compact tree to minimize total byte size.
// There are at least 3 size-types of node available to a compacted tree.
// Trivial: "No slots" is represented by a zero field (no node required).
//			"One slot" is represented by the presentation index (no node required).
// Tiny: Two to five (an optimized parameter "t") slots in node. Can carry between 2 and t slots. Size = 1 + 3 x t bytes.
// <where n is number of occupied slots>:
// Medium (variable size): 2 to 256 slots in node. Can carry between 2 and 256 slots. Size = 1 + 32 + n*2 bytes.
// Full: 256 slots in node (no parameter). Can carry between 2 and 256 slots. Size = 1 + 256*2 bytes

func tryParamsForSize(tParam int, nodeSlots int, nodeCount int) int {
	// First we try "Full"
	bytesFull := nodeCount * 513
	// Then we try "Medium"
	bytesMedium := nodeCount * (33 + nodeSlots*2)
	// Best so far
	bestBytes := bytesFull
	if bytesMedium < bestBytes {
		bestBytes = bytesMedium
	}
	// Then we try "tiny"
	if tParam >= nodeSlots {
		bytesTiny := nodeCount * (1 + 3*tParam)
		if bytesTiny < bestBytes {
			bestBytes = bytesTiny
		}
	}
	return bestBytes
}
