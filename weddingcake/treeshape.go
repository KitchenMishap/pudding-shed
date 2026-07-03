package weddingcake

import (
	"fmt"
	"math"
)

// A ShallowTree has zero or more levels, with level 0 being pointed to by the root slot.
// A ShallowTree containing no hashes has no nodes, and so has no levels.
// The activeSlotsCount of a node is the number of its slots for which IsEmpty() is false.
// The "shape" of a level in a tree holds the following:
// 		* the histogram of activeSlotCounts of all the tree's nodes at that level
// The level shapes of a tree can be used to optimize the storage format of a tree.

// An empty ShallowTree has a single empty slot.
// A ShallowTree containing only one hash, has a single root slot pointing to a single leaf node.

type LevelShape struct {
	// Leaf nodes appear at index 0 in this histogram (as they have no subnodes
	ActiveSlotCountHistogram [257]NodeCountType // Nodes at each level can have between 0 and 256 subnodes
}

// As up to 64 bytes per hash are supported, and because leaf nodes are nodes,
// a tree can have a maximum of 65 levels (0 to 64).
// Level 0 is the root, where a choice is made based on the first selected byte of the hash
// Level 63 is where a choice is made based on the last available byte of a 64 byte hash
// Level 64 may then contain leaf nodes, hence the maximum 65 levels for a 64 byte hash

type TreeShape struct {
	NonEmptyLevels        byte
	LevelShapes           [65]LevelShape
	GreatestNodesPerLevel NodeCountType
}

func (st *ShallowTree) CountLevelShapes() *TreeShape {
	result := TreeShape{} // Already a valid (empty) TreeShape
	result.NonEmptyLevels = 0
	result.GreatestNodesPerLevel = 0

	st.VisitAllNodes(func(node *ShallowTreeNode) {
		if node == nil {
			panic("Didn't expect nil node")
		}
		if node.Level+1 > result.NonEmptyLevels {
			result.NonEmptyLevels = node.Level + 1
		}
		// If this node is a slots node, count this node's activeSlotsCount
		activeSlotsCount := 0
		if node.LeafNode == nil {
			for hashByteVal := 0; hashByteVal < 256; hashByteVal++ {
				if !node.SlotsNode.Slots[hashByteVal].IsEmpty() {
					activeSlotsCount++
				}
			}
		}
		// We want to count even when activeSlotsCount is zero (ie a leaf node)
		result.LevelShapes[node.Level].ActiveSlotCountHistogram[activeSlotsCount]++
	})
	for level := byte(0); level <= result.NonEmptyLevels; level++ {
		nodesPerLevel := NodeCountType(0)
		// nodeCount at this level is the sum of the histogram
		for slotsCount := 0; slotsCount <= 256; slotsCount++ {
			nodesPerLevel += result.LevelShapes[level].ActiveSlotCountHistogram[slotsCount]
		}
		// We don't panic here, as ShallowTree can support more nodes per level than ChunkLevel's restriction
		if nodesPerLevel > result.GreatestNodesPerLevel {
			result.GreatestNodesPerLevel = nodesPerLevel
		}
	}
	return &result
}

func (ts *TreeShape) Print() {
	fmt.Printf("Out of the non-zero counts:\n")
	for level := 0; level < 64; level++ {
		totalNodesAtThisLevel := NodeCountType(0)
		totalActiveSlotsSeen := NodeCountType(0)
		maxActiveSlotsSeen := 0
		minActiveSlotsSeen := math.MaxInt

		// Scan through the histogram buckets (0 to 256 active slots possible)
		for activeSlotsCount := 0; activeSlotsCount <= 256; activeSlotsCount++ {
			nodeCount := ts.LevelShapes[level].ActiveSlotCountHistogram[activeSlotsCount]

			if nodeCount > 0 {
				totalNodesAtThisLevel += nodeCount
				totalActiveSlotsSeen += NodeCountType(activeSlotsCount) * nodeCount

				if activeSlotsCount > maxActiveSlotsSeen {
					maxActiveSlotsSeen = activeSlotsCount
				}
				if activeSlotsCount < minActiveSlotsSeen {
					minActiveSlotsSeen = activeSlotsCount
				}
			}
		}

		if totalNodesAtThisLevel > 0 {
			averageActiveSlots := float64(totalActiveSlotsSeen) / float64(totalNodesAtThisLevel)
			fmt.Printf("Level %d: %d total nodes. Active slots per node metrics: (min: %d, av: %.1f, max: %d)\n",
				level, totalNodesAtThisLevel, minActiveSlotsSeen, averageActiveSlots, maxActiveSlotsSeen)
		}
	}
}
