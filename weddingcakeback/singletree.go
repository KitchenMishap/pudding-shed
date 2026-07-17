package weddingcakeback

import (
	"bytes"
	"math"
	"sort"
)

// SingleTree is a specialized version of previous code called ShallowTree.
// It has now been specialized so that:
// * It assumes that the first n bytes of the hash have been handled elsewhere
//    * More specifically, when generating the tree, it assumes that those bytes of the supplied hashes pertain to this tree
//    * And on lookup, it assumes that those bytes of the hash being looked up have already been found to match
// * (n is a configuration parameter, and in the wider context of DonutForest's, it corresponds to the tier number)
// * It holds entire hashes at its leaves
// * It TYPICALLY supports a count of about 65536 hashes, allowing in-memory use without bloat
// * However there is no limit (due to statistical variation of hashes) and 65536 is often exceeded.

// SingleTree is a mechanism whereby a byte index of the hash (after the first n bytes) is selected
// to achieve the widest possible "spread" across sub-nodes at a node (fork) in the tree.
// SingleTree is a tree graph in which vertices are called Nodes and arcs are called (populated) Slots.
// (A leaf is a Node with no Slots)
// SingleTree does not tolerate duplicate hashes; these are presumed filtered out by prior code.

// Single does not support forward lookup (index to hash.)

const SingleTreeNoMatch = 0

type SingleTreePiType uint64

type SingleTree struct {
	HashLength            byte // eg 32 for SHA-256 hashes, 20 for RIPEMD-160
	PrefixBytesN          byte // Equal to the tier number
	ReassuranceBytesCount byte

	HashCount int
	// The root is a slot, not a node.
	RootSlot SingleTreeSlot // The root slot has "Depth" 0
}

// SingleTreeNode represents a subtree, after a particular set of hash bytes
// (n bytes, in the case of the root node) have been examined, and particular values found in them.
// It specifies which byte of the hash to examine next, and what to do in the case of
// each of the possible 256 values found (256 slots).
// It represents a "fork" in the tree... one slot POINTS to a node, and from a node up to 256 Slots POINT to further Nodes
type SingleTreeNode struct {
	Level     byte                 // The Node at "Level" n is the pointed to by the Slot at "Depth" n
	LeafNode  *SingleTreeLeafNode  // One of these two
	SlotsNode *SingleTreeSlotsNode // pointers will be nil
}

type SingleTreeLeafNode struct {
	ReassuranceHashBytes []byte           // Additional bytes from the hash to give statistical reassurance
	PresentationIndex    SingleTreePiType // The index that was initially presented with the hash corresponding to this leaf
	Hash                 []byte           // The entire hash for this leaf
	//SourceOffset         GlobalPiType     // The offset used when this leaf was captured
}

type SingleTreeSlotsNode struct {
	HashByteIndex byte                // Which byte of the hash to examine to choose between the slots
	Slots         [256]SingleTreeSlot // What to do when each of 256 possible byte values are found at HashByteIndex
}

// SingleTreeSlot represents how the tree progresses, when the next specified byte of the hash has been examined.
// It represents an "arc" (line) of the tree graph, sitting between two Nodes (vertices)
type SingleTreeSlot struct {
	NextNode *SingleTreeNode // If nil, arc does not exist
}

// When created with ShallowTreeSlot{}, they start out as IsEmpty()
func (sts SingleTreeSlot) IsEmpty() bool {
	return sts.NextNode == nil
}

// SingleTreeHash is used in the parameter to GenerateSingleTree()
type SingleTreeHash struct {
	Hash              []byte // len(Hash) must equal ShallowTree.HashLength
	PresentationIndex SingleTreePiType
	//SourceOffset      GlobalPiType
}

// GenerateShallowTree generates a tree from the supplied hashes and presentationIndices
func GenerateSingleTree(input []SingleTreeHash, PrefixBytesN byte, hashLength byte, reassuranceBytes byte) *SingleTree {
	if PrefixBytesN > hashLength {
		panic("PrefixBytesN cannot exceed hashLength")
	}
	if hashLength < 1 || hashLength > 64 {
		panic("Only 1 to 64 byte hashes are supported")
	}
	if reassuranceBytes > hashLength {
		panic("Reassurance bytes must be less than or equal to hash length")
	}
	for i := range input {
		if input[i].Hash == nil {
			panic("Malformed input: ShallowTreeHash contains a nil Hash slice")
		}
		if len(input[i].Hash) != int(hashLength) {
			panic("Malformed input: ShallowTreeHash slice length does not match specified hashLength")
		}
	}
	result := SingleTree{}
	result.HashLength = hashLength
	result.PrefixBytesN = PrefixBytesN
	result.ReassuranceBytesCount = reassuranceBytes
	result.HashCount = len(input)
	if len(input) == 0 {
		// No hashes, empty tree (no nodes)
		// The root slot is already IsEmpty()
		return &result
	}
	// Important special case for a lone hash, because recurseGenerateNode() assumes at least two hashes.
	if len(input) == 1 {
		// On creation of SingleTree{] above, result.RootSlot is currently IsEmpty().
		// We need to point it to single leaf node
		leafNode := SingleTreeLeafNode{}
		leafNode.PresentationIndex = input[0].PresentationIndex
		//leafNode.SourceOffset = input[0].SourceOffset
		leafNode.ReassuranceHashBytes = make([]byte, reassuranceBytes)
		copy(leafNode.ReassuranceHashBytes, input[0].Hash[:reassuranceBytes])
		leafNode.Hash = make([]byte, hashLength)
		copy(leafNode.Hash, input[0].Hash)
		node := SingleTreeNode{}
		node.Level = PrefixBytesN
		node.LeafNode = &leafNode
		node.SlotsNode = nil
		result.RootSlot.NextNode = &node
		return &result
	}
	// Because we will be mutating it (sorting it), we take a copy of the input so as not to surprise the caller
	inputCopy := make([]SingleTreeHash, len(input))
	copy(inputCopy, input)

	// Create root node and recursively its children
	var unusedBytesFlags uint64
	if hashLength == 64 {
		unusedBytesFlags = math.MaxUint64 // Special case to avoid overflow below
	} else {
		unusedBytesFlags = (uint64(1) << hashLength) - 1 // eg 0xFFFFFFFF for hashLength = 32
	}
	// Shift left to mark the first PrefixBytesN bytes as "examined"
	unusedBytesFlags = (unusedBytesFlags << PrefixBytesN) & unusedBytesFlags
	// We start recursing at hash byte index PrefixBytesN
	rootNode := result.recurseGenerateNode(inputCopy, unusedBytesFlags, PrefixBytesN)
	result.RootSlot.NextNode = rootNode
	return &result
}

// LookupHash uses SingleTree to lookup one presentationIndex if it exists.
// If the tree contains no matches for the hash, SingleTreeNoMatch is returned.
func (st *SingleTree) LookupHash(hash []byte) SingleTreePiType {
	if len(hash) != int(st.HashLength) {
		panic("Wrong hash length")
	}
	// A tree that has no nodes (it contains no hashes), will always fail without even looking at the hash
	if st.RootSlot.NextNode == nil {
		return SingleTreeNoMatch
	}
	node := st.RootSlot.NextNode

	// Keep track (by way of a mask) of which bytes of the mask have been examined
	var unusedBytesFlags uint64
	if st.HashLength == 64 {
		unusedBytesFlags = math.MaxUint64 // Special case to avoid overflow below
	} else {
		unusedBytesFlags = (uint64(1) << st.HashLength) - 1 // eg 0xFFFFFFFF for hashLength = 32
	}
	// Shift left to mark the first PrefixBytesN bytes as "examined"
	unusedBytesFlags = (unusedBytesFlags << st.PrefixBytesN) & unusedBytesFlags

	for {
		leafNode := node.LeafNode
		if leafNode != nil {
			// We've reached a leaf node, a potential match

			// Check whole hash (not just the reassurance bytes)
			if !bytes.Equal(leafNode.Hash, hash) {
				return SingleTreeNoMatch
			}

			/*
				// Check reassurance bytes
				// The reassurance bytes are (sequentially) the first few that haven't yet been examined
				mask := uint64(1)
				ind := 0
				for i := 0; i < len(leafNode.ReassuranceHashBytes); i++ {
					// Find the next hash byte that has not yet been examined
					if unusedBytesFlags == 0 {
						panic("No bytes left to examine for reassurance")
					}
					for unusedBytesFlags&mask == 0 {
						mask <<= 1
						ind++
					}
					// Mark it as examined
					unusedBytesFlags ^= mask
					// Compare the reassurance byte
					byteValue := hash[ind]
					reassuranceValue := leafNode.ReassuranceHashBytes[i]
					if byteValue != reassuranceValue {
						return ShallowTreeNoMatch
					} // No match
				}*/
			return leafNode.PresentationIndex // Reassured potential match
		}
		// Not a leaf node. It's a slots node. Examine the slots...
		byteIndexToExamine := node.SlotsNode.HashByteIndex
		// Let's check we're not being to examine one of the prefix bytes
		if byteIndexToExamine < st.PrefixBytesN {
			panic("Byte index is part of the prefix")
		}
		// Mark byte index as examined
		mask := uint64(1) << byteIndexToExamine
		if unusedBytesFlags&mask == 0 {
			panic("Byte index already examined")
		}
		unusedBytesFlags ^= mask // Flip bit (clear flag)
		examinedByteValue := hash[byteIndexToExamine]
		if node.SlotsNode.Slots[examinedByteValue].IsEmpty() {
			return SingleTreeNoMatch
		}
		node = node.SlotsNode.Slots[examinedByteValue].NextNode
	}
}

// We implement a visitor pattern to enable you to visit every node in the tree

// SingleTreeNodeVisitor is the signature for your custom processing functions
type SingleTreeNodeVisitor func(node *SingleTreeNode)

type traversalFrame struct {
	node *SingleTreeNode
}

// VisitAllNodes performs a complete iterative stack traversal of the tree,
// invoking the supplied visitor function on every leaf node or slots node.
func (st *SingleTree) VisitAllNodes(visitor SingleTreeNodeVisitor) {
	// 1. Return if there are no nodes
	if st.RootSlot.IsEmpty() {
		return
	}

	// 2. Initialize our iterative LIFO stack with the first branch node frame
	stack := []traversalFrame{{
		node: st.RootSlot.NextNode,
	}}

	// 3. Process the explicit stack loop
	for len(stack) > 0 {
		// Pop the top frame off the heap slice
		currentFrame := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		// Whether its a leaf node or a slots node, visit it (user supplied callback)
		visitor(currentFrame.node)

		// If it's a slot node, go through its slots
		if currentFrame.node.LeafNode == nil {
			// Iterate across the full fixed 256-slot routing block
			for byteValInt := 0; byteValInt <= 255; byteValInt++ {
				byteVal := byte(byteValInt)
				slot := &currentFrame.node.SlotsNode.Slots[byteVal]

				if !slot.IsEmpty() {
					// push the node that the slot points to (its frame) onto the stack to explore later
					nextFrame := traversalFrame{
						node: slot.NextNode,
					}
					stack = append(stack, nextFrame)
				}
			}
		}
	}
}

func (st *SingleTree) CountNodes() int {
	count := 0
	st.VisitAllNodes(func(node *SingleTreeNode) {
		count++
	})
	return count
}

// SingleTreeLeavesVisitor a similar pattern for visiting all leaves
type SingleTreeLeavesVisitor func(node *SingleTreeLeafNode)

// VisitAllLeaves performs a complete iterative stack traversal of the tree,
// invoking the supplied visitor function on every leaf node (NOT slots nodes)
func (st *SingleTree) VisitAllLeaves(visitor SingleTreeLeavesVisitor) {
	st.VisitAllNodes(func(node *SingleTreeNode) {
		if node.SlotsNode == nil {
			visitor(node.LeafNode)
		}
	})
}

func (st *SingleTree) CountLeaves() int {
	count := 0
	st.VisitAllLeaves(func(leaf *SingleTreeLeafNode) {
		count++
	})
	return count
}

// Helper to quickly find the active slots count for a node
func (stn *SingleTreeNode) activeSlotsCount() int {
	if stn.SlotsNode == nil {
		return 0
	}
	count := 0
	for i := 0; i < 256; i++ {
		if !stn.SlotsNode.Slots[i].IsEmpty() {
			count++
		}
	}
	return count
}

// recurseGenerateNode() is a recursive call to populate a SingleTree based on a slice of SingleTreeHash.
// The SingleTreeHash will be modified (sorted), so send in a copy if this is not tolerated.
// It returns a pointer to a new node. Duplicate hashes are not tolerated.
func (st *SingleTree) recurseGenerateNode(inputCopy []SingleTreeHash,
	unusedByteIndices uint64, level byte) *SingleTreeNode {
	if len(inputCopy) < 2 {
		panic("recurseGenerateNode() should only be called with multiple hashes")
	}
	node := SingleTreeNode{}
	node.Level = level
	slotsNode := SingleTreeSlotsNode{}
	node.SlotsNode = &slotsNode
	node.LeafNode = nil // It won't be a leaf node because we know we have multiple hashes

	// Try partitioning by each of the (up to 64) bytes in the hashes. Just the ones we haven't used
	byteIndex := -1
	maxEntropyFound := float64(0)
	maxEntropyIndex := byteIndex
	shiftMask := unusedByteIndices
	for byteIndex = 0; byteIndex < int(st.HashLength); byteIndex++ {
		unused := (shiftMask & 1) != 0
		if unused {
			entropy := partitioningEntropy(inputCopy, byteIndex)
			if entropy > maxEntropyFound {
				maxEntropyIndex = byteIndex
				maxEntropyFound = entropy
			}
		}
		shiftMask >>= 1
	}

	if maxEntropyFound == 0.0 {
		// We know there were multiple hashes input to this function.
		// An entropy of 0 indicates these hashes are all duplicates.
		// Duplicates are not tolerated! They should have first been removed by a higher authority
		panic("Duplicate hashes are not tolerated")
	}

	// Use the best one
	bi := maxEntropyIndex
	if bi < int(st.PrefixBytesN) {
		panic("Byte index is part of the prefix")
	}
	mask := uint64(1) << bi
	if unusedByteIndices&mask == 0 {
		panic("We chose to use a byte index that is already used")
	}
	unusedByteIndices ^= 1 << bi // Flip (clear) the bit
	node.SlotsNode.HashByteIndex = byte(bi)

	// Now we'll need to sort by that byte, so we can pass subsets of the hash list to each child.
	// Use a stable sort to prevent Go from randomly scrambling the presentation order of duplicate hashes.
	sort.SliceStable(inputCopy, func(i int, j int) bool {
		return inputCopy[i].Hash[bi] < inputCopy[j].Hash[bi]
	})

	// We have decided to split this node into 256 (fork the tree) based on the value found in the hashes
	// at byte index bi. Consider each value we might find at bi, and what to do.
	index := 0
	for byteValInt := 0; byteValInt <= 255; byteValInt++ {
		byteVal := byte(byteValInt)
		startIndex := index // index into the list of hashes
		// Look for as many "byteVal's at bi" in a row that we can find in the list of hashes
		for index < len(inputCopy) && inputCopy[index].Hash[bi] == byteVal {
			index++
		}
		if index == startIndex {
			// Didn't find any; empty slot (the bytes examined up to this point in the tree lead to no hash entries)
			// (and the slot was already created empty; nothing to do)
		} else if index == startIndex+1 {
			// Found exactly one hash at startIndex, so we need a leaf node, and don't recurse
			leafNode := SingleTreeLeafNode{}
			leafNode.PresentationIndex = inputCopy[startIndex].PresentationIndex
			//leafNode.SourceOffset = inputCopy[startIndex].SourceOffset
			leafNode.Hash = make([]byte, st.HashLength)
			copy(leafNode.Hash, inputCopy[startIndex].Hash)

			// The reassurance hash bytes are (sequentially) a maximum of st.ReassuranceBytesCount
			// bytes, out of the hash bytes that haven't been examined yet
			leafNode.ReassuranceHashBytes = make([]byte, 0, st.ReassuranceBytesCount)
			localUnusedFlags := unusedByteIndices
			mask := uint64(1)
			ind := 0
			for b := byte(0); b < st.ReassuranceBytesCount; b++ {
				// Abort if all bytes have been examined
				if localUnusedFlags == 0 {
					break
				}
				// Find the next hash byte that has not yet been examined
				for localUnusedFlags&mask == 0 {
					mask <<= 1
					ind++
				}
				// Mark it as examined in our local copy
				localUnusedFlags ^= mask
				// Record the byte value
				byteValue := inputCopy[startIndex].Hash[ind]
				leafNode.ReassuranceHashBytes = append(leafNode.ReassuranceHashBytes, byteValue)
			}

			newNode := SingleTreeNode{}
			newNode.Level = level + 1
			newNode.LeafNode = &leafNode
			newNode.SlotsNode = nil
			node.SlotsNode.Slots[byteVal].NextNode = &newNode
		} else if index > startIndex+1 {
			// Found more than one, we need a fully fledged slots node child
			childNode := st.recurseGenerateNode(inputCopy[startIndex:index], unusedByteIndices, level+1)
			node.SlotsNode.Slots[byteVal].NextNode = childNode
		} else {
			panic("Error in code logic")
		}
	} // for byteVal
	return &node
}

// Partitioning entropy if we were to partition by a particular byte index of the hash
// Shannon Entropy calculation: Maximizes value for an even distribution
func partitioningEntropy(input []SingleTreeHash, hashByteIndex int) float64 {
	var byteValCounts [256]int
	for i := range input {
		byteValCounts[input[i].Hash[hashByteIndex]]++
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
