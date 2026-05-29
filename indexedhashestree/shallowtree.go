package indexedhashestree

import (
	"math"
	"sort"
)

// This file is concerned with generating a shallow tree lookup for up to about 60,000 sequential hashes
// by choosing the best order of bytes in the hash to examine (the first byte examined is at the root of
// the generated tree) based on information gain.
// The absolute (unreachable) maximum number of hashes that can be coped with is 65536
// (this "might" happen in the extremely unlikely case of all 65536 hashes being definable based on two selected bytes...
// but even then would be just about unachievable)

// 1 is for the root node
// According to Google Gemini AI, for 50,000 hashes:
// Expecting this root node to have a node at all 256 outpoints (so +256)
// Expecting each of these second nodes to have 154 outpoints
// Expecting each of these outpoints to be a node containing 2 direct entries (ie no further nodes) (so + 256 * 154)
// Plus 10% spare
const initialNodeCapacity = 10_000

// This type (a tree node) represents branching choices based on examining a chosen byte in the hash, after
// some (possibly none) previous chosen bytes have been examined. Bytes 0..31 are examined in an order which
// is optimized for the particular set of hashes presented to the tier.
type shallowTreeNode struct {
	hashByteIndex byte
	lookups       [256]uint16 // Using a uint16 here is what limits us to <65536 hashes (and we are in fact limited further)
}

type shallowTreeContainer struct {
	firstPresentationIndex uint64
	presentationsCount     uint16 // These two summed
	maxSkipNumber          uint16 // cannot exceed 65535 (ie leaving room for a zero too)
	nodesPool              []shallowTreeNode
}

func newShallowTreeContainer() *shallowTreeContainer {
	result := shallowTreeContainer{}
	result.nodesPool = make([]shallowTreeNode, 0, initialNodeCapacity)
	return &result
}

func (stc *shallowTreeContainer) reset() {
	stc.firstPresentationIndex = 0
	stc.presentationsCount = 0
	stc.maxSkipNumber = 0
	stc.nodesPool = stc.nodesPool[:0]
}

// This is used in the parameter to the generate function
type shallowTreeHash struct {
	hash              [32]byte
	presentationIndex uint64
}

// Note that input will be mutated (sorted)
// A true result is an overflow!
func (stc *shallowTreeContainer) generate(input []shallowTreeHash) bool {
	// Start with no nodes
	stc.reset()
	if len(input) == 0 {
		return false // No hashes, empty container (no nodes)
	}
	// The following is used as an offset so we can fit presentationIndices into 16 bits
	stc.firstPresentationIndex = input[0].presentationIndex // ToDo calling generate twice with the same input will break this
	// The following count is used to track whether our 65536 allowed "lookup" values run out
	if len(input) >= 65536 {
		stc.reset()
		return true
	} // Can still overflow even if we get past this check
	stc.presentationsCount = uint16(len(input))
	stc.maxSkipNumber = 0 // So far...
	// Add root node and recursively its children
	unusedBytesFlags := uint32(0xFFFFFFFF)
	_, overflow := stc.recurseGenerateNode(input, unusedBytesFlags, 0)
	// (we throw away the nodeIndex for the root node, it is always zero)
	if overflow {
		stc.reset()
		return true
	}
	return false
}

// A true result is an overflow!
func (stc *shallowTreeContainer) recurseGenerateNode(input []shallowTreeHash, unusedByteIndices uint32, level int) (int, bool) {
	// Create a new node appended to the slice in the container
	stc.nodesPool = append(stc.nodesPool, shallowTreeNode{})
	nodeIndex := len(stc.nodesPool) - 1
	// DON'T grab a pointer to the node here. Recursive calls below might re-size nodesPool, invalidating the pointer!

	// Try partitioning by each of the 32 bytes in the hashes. Just the ones we haven't used
	byteIndex := -1
	maxEntropyFound := float64(0)
	maxEntropyIndex := byteIndex
	shiftMask := unusedByteIndices
	for byteIndex = 0; byteIndex < 32; byteIndex++ {
		unused := shiftMask & 1
		if unused == 1 {
			entropy := partitioningEntropy(input, byteIndex)
			if entropy > maxEntropyFound {
				maxEntropyIndex = byteIndex
				maxEntropyFound = entropy
			}
		}
		shiftMask >>= 1
	}
	// Use the best one
	bi := maxEntropyIndex
	unusedByteIndices ^= 1 << bi // Flip the bit
	stc.nodesPool[nodeIndex].hashByteIndex = byte(bi)

	// Now we'll need to sort by that byte, so we can pass subsets of the hash list to each child
	sort.Slice(input, func(i int, j int) bool { return input[i].hash[bi] < input[j].hash[bi] })
	// Find the range for each potential child
	index := 0
	for byteValInt := 0; byteValInt <= 255; byteValInt++ {
		byteVal := byte(byteValInt)
		startIndex := index
		// Look for as many byteVal's in a row that we can find
		for index < len(input) && input[index].hash[bi] == byteVal {
			index++
		}
		if index == startIndex {
			// Didn't find any; Nil entry
			stc.nodesPool[nodeIndex].lookups[byteVal] = 0
		} else if index == startIndex+1 {
			// Found exactly one, a leaf (no node object gets created)
			// The value of the leaf is an indication of the hash's presentationIndex (we subtract the first)
			// This will not overflow
			stc.nodesPool[nodeIndex].lookups[byteVal] = uint16(input[startIndex].presentationIndex - stc.firstPresentationIndex)
		} else if index > startIndex+1 {
			// Found more than one, we need a fully fledged node child
			childIndex, overflow := stc.recurseGenerateNode(input[startIndex:index], unusedByteIndices, level+1)
			if overflow {
				return -1, true
			}

			// Link to it via relative skip
			skipValue := childIndex - nodeIndex
			// Keep a record so we can see how much "room" was left when we finish
			if skipValue > int(stc.maxSkipNumber) {
				stc.maxSkipNumber = uint16(skipValue)
			}

			// "Meeting in the middle" encoding
			encodedValue := uint16(65536 - skipValue)

			// If the encoded downward jump value collides with or drops below
			// our upward-bound hash count, we have run out of 16-bit address space.
			if encodedValue <= stc.presentationsCount {
				stc.reset()
				return -1, true
			}

			stc.nodesPool[nodeIndex].lookups[byteVal] = encodedValue

		} else {
			panic("Error in code logic")
		}
	} // for byteVal
	return nodeIndex, false
}

func (stc *shallowTreeContainer) getNodeSizeStatistics() *[257]int {
	result := [257]int{}
	for n := range len(stc.nodesPool) {
		size := 0
		for slot := range 256 {
			if stc.nodesPool[n].lookups[slot] != 0 {
				size++
			}
		}
		result[size]++
	}
	return &result
}

func (stc *shallowTreeContainer) lookupHash(hash [32]byte) int64 {
	return -1
}

// Partitioning entropy if we were to partition by a particular byte index of the hash
// Shannon Entropy calculation: Maximizes value for an even distribution
func partitioningEntropy(input []shallowTreeHash, hashByteIndex int) float64 {
	//fmt.Printf("Partitioning Entropy() for hashByteIndex %d\n", hashByteIndex)
	var byteValCounts [256]int
	for i := range input {
		byteValCounts[input[i].hash[hashByteIndex]]++
	}
	// Loop over the partitions, calculating probabilities and entropy
	total := float64(len(input))
	entropySum := float64(0)
	for byteVal := 0; byteVal < 256; byteVal++ {
		count := byteValCounts[byteVal]
		if count > 0 { // Avoid log of 0
			prob := float64(count) / total
			entropySum -= prob * math.Log2(prob)
		}
	}
	return entropySum
}
