package weddingcakeback

import (
	"bytes"
	"fmt"
	"sort"
)

// BakingDesigner is introduced in the effort to bake a (smaller/newer) tier n-1 of the "cake" into a
// new DonutForest in the (larger/older) tier n underneath it. Tier n-1 will become empty ready for injection of
// new hashes (if it is TierTop), or ready for new rings/chunks to be consolidated/baked into it from the (even
// smaller/newer) tier n-2.

// The "Design" task involves, for each level (hash byte index) for the new DonutForest, determining
// a choice of a set of NodeFormatSpecs (the "design"). These designs determine how different nodes are expressed as a
// compact serialized sequence of bytes on disk. There is one such design per level for the DonutForest we aim to bake.

// Prior to the actual design phase, the designer must first gather metrics to base the design on.
// Crucially, these metrics must be accumulated in many stages from the source tier, as we don't want
// to hold large arrays of SingleTree's in memory at any time. This implementation holds just one SingleTree
// in memory at any time.

type BakingDesigner struct {
	Metrics TierMetrics
}

func NewBakingDesigner() *BakingDesigner {
	result := BakingDesigner{}
	return &result
}

func (bd *BakingDesigner) GatherMetricsFromSourceTier(sourceTierInfo BakingSourceTier,
	config *CakeConfig) {

	offsetToUse := sourceTierInfo.GetFirstPresentationIndex()

	destTierIndex := sourceTierInfo.GetNextTierIndex()
	destPrefixBytesCount := sourceTierInfo.GetNextTierPrefixBytesCount()
	indexRange := sourceTierInfo.GetIndicesCount()
	for index := uint64(0); index < indexRange; index++ {
		hashInfos := sourceTierInfo.GetHashesAtIndex(index, offsetToUse)
		// (there should typically be about 65,536 hashes)

		// Because they were obtained by index (which chooses a tree in each source DonutForest),
		// these hashes will all have the same hash prefix.
		// In the destination tier, we are subdividing these hashes on an ADDITIONAL byte (the prefix gets longer.)
		// We need to put the hashes into buckets based on this "newly examined" byte of the hash.
		buckets := [256][]SingleTreeHash{}
		for i := range 256 {
			buckets[i] = make([]SingleTreeHash, 0, 300)
		}

		// EXCEPTION: If the destination tier index is 0 (source tier was TierTop), the "new longer prefix" is
		// still 0 bytes, so we just use bucket[0] for all
		if destTierIndex == 0 {
			buckets[0] = hashInfos
		} else {
			byteIndex := destPrefixBytesCount - 1
			for _, hashInfo := range hashInfos {
				examinedByte := hashInfo.Hash[byteIndex]
				buckets[examinedByte] = append(buckets[examinedByte], hashInfo)
			}
		}
		// Now we either have one or 256 buckets of hashes.
		// These buckets correspond to either one or 256 of the 256^n trees in the DonutForest we are writing to.
		// One by one, turn them into SingleTree's and measure them.
		// We throw each tree away before starting on the next one, to conserve memory.
		var treeCount int
		if destTierIndex == 0 {
			treeCount = 1
		} else {
			treeCount = 256
		}
		for t := range treeCount {
			bucket := buckets[t]
			bd.RecordSubtreeDistribution(bucket, config, destTierIndex, destPrefixBytesCount)
		}
	}
}

func PanicIfDuplicates(hashes []SingleTreeHash) {
	count := len(hashes)
	if count < 2 {
		return
	}
	for A := 1; A < count; A++ {
		for B := 0; B < A; B++ {
			if bytes.Equal(hashes[A].Hash, hashes[B].Hash) {
				fmt.Printf("Duplicate hash found: indices A,B = %d, %d\n", A, B)
				fmt.Printf("Duplicate hash found: Presentation indices = %d, %d\n", hashes[A].PresentationIndex, hashes[B].PresentationIndex)
				fmt.Printf("Duplicate hash found: Hash value = %x\n", hashes[A].Hash)
				panic("Duplicate hash found")
			}
		}
	}
}

// RecordSubtreeDistribution analyzes a raw list of hashes representing a future
// downstream subtree, measures it, and discards it to protect memory.
// This will be called several times over for a given tier that we are rebaking into a DonutForest.
func (bd *BakingDesigner) RecordSubtreeDistribution(hashes []SingleTreeHash,
	config *CakeConfig, tierIndex byte, prefixBytesCount byte) {

	if tierIndex == 1 {
		PanicIfDuplicates(hashes)
	}

	if len(hashes) == 0 {
		return
	}

	// 1. Generate the temporary SingleTree for this specific prefix bucket
	// We "WILL" be needing this later, BUT to conserve memory and disk space,
	// we will DISCARD IT and RECREATE it. For large hash counts this is very important.
	reassuranceBytesCount := config.TierBelowConfigs[tierIndex].ReassuranceBytesCount
	tempTree := GenerateSingleTree(hashes, prefixBytesCount, config.HashLength, reassuranceBytesCount)

	// 2. Measure its nodes into our persistent, lightweight metrics matrix
	if tempTree.RootSlot.IsEmpty() {
		panic("Empty root slot should have been caught by len(hashes)==0")
	}
	bd.Metrics.RecurseAccumulateSubtreeMetrics(tempTree.RootSlot.NextNode)

	// 3. By returning, tempTree is dropped and cleared for Go's Garbage Collector
}

// DesignTheDesign optimizes and builds the BakingDesign using the gathered TierMetrics.
func (bd *BakingDesigner) DesignTheDesign(config *CakeConfig, tierIndex byte) *BakingDesign {
	if config.TierBelowConfigs[tierIndex].NodeFormatSpecsPerLevel < 5 {
		panic("nodeSpecFormatSpecsPerLevel must be at least 5")
	}

	result := &BakingDesign{}
	// To start with, the levels array length matches the 65 max support depth specified in TierMetrics
	result.LevelSpecs = make([]LevelFormat, len(bd.Metrics.Levels))

	// Determine how many levels contain actual measured nodes.
	// We scan backwards to find the physical boundary of our tree shape.
	maxNonEmptyLevel := 0
	for l := len(bd.Metrics.Levels) - 1; l >= 0; l-- {
		hasNodes := false
		for slots := 0; slots <= 256; slots++ {
			if bd.Metrics.Levels[l].ActiveSlotHistogram[slots] > 0 {
				hasNodes = true
				break // ToDo are you sure?
			}
		}
		if hasNodes {
			maxNonEmptyLevel = l + 1
			break // ToDo are you sure?
		}
	}

	// Truncate the configured format specs to match the actual populated level bounds
	result.LevelSpecs = result.LevelSpecs[:maxNonEmptyLevel]

	for level := 0; level < maxNonEmptyLevel; level++ {
		// Safety check: Ensure a level doesn't overflow our data type capacity limits
		totalNodesOnLevel := NodeCountType(0)
		for slots := 0; slots <= 256; slots++ {
			totalNodesOnLevel += bd.Metrics.Levels[level].ActiveSlotHistogram[slots]
		}
		if totalNodesOnLevel > MaxNodesCount {
			panic("Too many nodes accumulated at a single level")
		}

		// Initially propose a separate NodeFormatSpec for each active slots count found on this level
		nfgs := make([]NodeFormatGroup, 0, 257)
		for slotsCount := 0; slotsCount <= 256; slotsCount++ {
			count := bd.Metrics.Levels[level].ActiveSlotHistogram[slotsCount]
			if count > 0 {
				group := NodeFormatGroup{
					StartSlotsCount: slotsCount,
					EndSlotsCount:   slotsCount + 1,
					NodesCount:      count,
					Spec:            bd.ProposeNodeFormatForSlotsCount(slotsCount),
				}
				bytesCost := group.groupByteSize(config, tierIndex)
				if bytesCost > uint64(MaxBytesCount) {
					panic("Too many bytes for BytesCountType")
				}
				group.Bytes = BytesCountType(bytesCost)
				nfgs = append(nfgs, group)
			}
		}

		// If this level was untouched (e.g. fixed jump table levels), leave it empty
		if len(nfgs) == 0 {
			continue
		}

		// Run your exact reduction loop from nodeformat.go to compress specs down to NodeFormatSpecsPerLevel
		for {
			if len(nfgs) <= int(config.TierBelowConfigs[tierIndex].NodeFormatSpecsPerLevel) {
				break
			}
			proposed, reduced := bd.RefineNodeFormatGroups(nfgs, config, tierIndex)
			if !reduced {
				panic("Could not reduce node format specs! nodeFormatSpecsPerLevel too low?")
			}
			nfgs = proposed
		}

		// Sort primarily to push FormatTiny to the end, and secondarily by NodesCount descending
		sort.SliceStable(nfgs, func(i, j int) bool {
			iIsTiny := nfgs[i].Spec.Format == NodeFormatTiny
			jIsTiny := nfgs[j].Spec.Format == NodeFormatTiny

			if iIsTiny != jIsTiny {
				return jIsTiny
			}
			return nfgs[i].NodesCount > nfgs[j].NodesCount
		})

		result.LevelSpecs[level].Groups = nfgs
	}

	// Initialize ID allocation bounds across the entire planned structure
	result.InitializeNodeIdAllocations()
	return result
}

// SlotCount represents how many active paths fan out from a node (0 to 256).
type SlotCount uint16

// LevelMetrics captures the statistical distribution of nodes on a single level.
// It tracks exactly how many nodes have a specific number of active slots.
type LevelMetrics struct {
	// Index represents the active slot count (0 to 256).
	// Value represents the total number of nodes matching that active slot density.
	ActiveSlotHistogram [257]NodeCountType
}

// TierMetrics Accumulates structural profiles across the entire new DonutForest, at each level.
type TierMetrics struct {
	// Max supported levels
	// (64 for each byte of the max supported hash size, plus 1 level for potential final leaf nodes)
	Levels [65]LevelMetrics
}

// RecurseAccumulateSubtreeMetrics analyzes a single hydrated subtree and aggregates its
// layout fingerprint before the subtree is dropped from memory.
func (tm *TierMetrics) RecurseAccumulateSubtreeMetrics(node *SingleTreeNode) {
	if node == nil {
		return
	}

	// Track the structural density of this specific node
	activeSlots := node.activeSlotsCount()
	targetLevel := node.Level
	tm.Levels[targetLevel].ActiveSlotHistogram[activeSlots]++

	// Recurse down through this local subtree's children
	if node.SlotsNode != nil {
		for s := 0; s < 256; s++ {
			if !node.SlotsNode.Slots[s].IsEmpty() {
				tm.RecurseAccumulateSubtreeMetrics(node.SlotsNode.Slots[s].NextNode)
			}
		}
	}
}

func (bd *BakingDesigner) ProposeNodeFormatForSlotsCount(activeSlots int) NodeFormatSpec {
	result := NodeFormatSpec{}
	result.SlotsCapacity = NodeFormatSlotCapacity(activeSlots)
	if activeSlots == 0 {
		// With no active slots, we have a leaf node
		result.Format = NodeFormatLeaf
		return result
	}
	if activeSlots == 1 {
		panic("There should be no nodes with one active slot") // As it should already be a leaf
	}
	if activeSlots <= 5 {
		result.Format = NodeFormatTiny
		return result
	}
	if activeSlots >= 245 {
		result.Format = NodeFormatFull
		result.SlotsCapacity = 256
		return result
	}
	result.Format = NodeFormatMedium
	return result
}

func (bd *BakingDesigner) RefineNodeFormatGroups(groups []NodeFormatGroup, config *CakeConfig,
	tierIndex byte) ([]NodeFormatGroup, bool) {
	// Try each neighbouring pair of groups
	// We are looking for the lowest cost merge (counted in bytes)
	lowestCost := MaxBytesCount
	bestProposalLeft := -1
	bestProposedMerge := NodeFormatGroup{}
	for left := 0; left < len(groups)-1; left++ {
		right := left + 1
		if bd.AllowMergeGroups(&groups[left], &groups[right]) {
			leftBytes := groups[left].Bytes
			rightBytes := groups[right].Bytes
			proposedGroup := bd.ProposeMergeGroups(&groups[left], &groups[right], config, tierIndex)
			mergedBytes := proposedGroup.Bytes
			proposedCost := mergedBytes - (leftBytes + rightBytes)
			if proposedCost < lowestCost {
				lowestCost = proposedCost
				bestProposalLeft = left
				bestProposedMerge = proposedGroup
			}
		}
	}
	if bestProposalLeft == -1 {
		return groups, false
	}
	// Replace (left, right) with (merged)
	result := make([]NodeFormatGroup, 0, len(groups)-1)
	result = append(result, groups[:bestProposalLeft]...)
	result = append(result, bestProposedMerge)
	result = append(result, groups[bestProposalLeft+2:]...)

	return result, true
}

func (bd *BakingDesigner) AllowMergeGroups(left *NodeFormatGroup, right *NodeFormatGroup) bool {
	return right.Spec.Format == left.Spec.Format
}

func (bd *BakingDesigner) ProposeMergeGroups(left *NodeFormatGroup, right *NodeFormatGroup,
	config *CakeConfig, tierIndex byte) NodeFormatGroup {

	if right.StartSlotsCount < left.EndSlotsCount {
		panic("Illegal group merge: overlapping")
	}
	if right.Spec.Format != left.Spec.Format {
		panic("Illegal group merge: different formats")
	}
	result := NodeFormatGroup{}
	result.StartSlotsCount = left.StartSlotsCount
	result.EndSlotsCount = right.EndSlotsCount
	result.NodesCount = left.NodesCount + right.NodesCount
	result.Spec.Format = left.Spec.Format
	result.Spec.SlotsCapacity = right.Spec.SlotsCapacity
	byts := result.groupByteSize(config, tierIndex)
	if byts > uint64(MaxBytesCount) {
		panic("Too many bytes for BytesCountType")
	}
	result.Bytes = BytesCountType(byts)
	return result
}
