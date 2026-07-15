package weddingcakeback

import (
	"io"
	"os"
	"path/filepath"
)

// This file endows tierbelow with a BakingSourceTier interface, so that one TierBelow[n] can
// be baked into a DonutForest in the next tier TierBelow[n+1]

// Check that implements
var _ BakingSourceTier = (*TierBelow)(nil)

func (tb *TierBelow) GetNextTierPrefixBytesCount() byte {
	return tb.NextTierConfig.PrefixBytesCount
}
func (tb *TierBelow) GetNextTierIndex() byte {
	return tb.TierIndex + 1
}
func (tb *TierBelow) GetIndicesCount() uint64 {
	// The indices count is 256 ^ prefixBytesCount
	return 1 << (8 * uint(tb.ThisTierConfig.PrefixBytesCount))
}
func (tb *TierBelow) GetHashesAtIndex(index uint64, offsetToUse GlobalPiType) []SingleTreeHash {
	if index >= tb.GetIndicesCount() {
		panic("TierBelow.GetHashesForIndex() should only be called with index<GetIndicesCount()")
	}

	// index is an entry number in the jump-table.
	// But a TierBelow comprises multiple DonutForest's, each of which has its own jump-table.
	// "GetHashesAtIndex" really means "Concatenate the hashes from the SingleTrees indexed by 'index' across all
	// DonutForests".
	jumpTableIndex := index

	result := make([]SingleTreeHash, 0, 1000)

	// Work out the size of each jump table
	nodeIdConfig := &tb.ThisTierConfig.NodeIdConfig
	hashIndexIdConfig := &tb.ThisTierConfig.HashIndexIdConfig
	reassuranceBytesCount := tb.ThisTierConfig.ReassuranceBytesCount
	nodeIdSize := uint64((*nodeIdConfig).StorageBytes())
	prefixBytesCount := tb.ThisTierConfig.PrefixBytesCount
	jumpTableEntries := uint64(1) << (prefixBytesCount * 8) // = 256 ^ prefixBytesCount
	jumpTableSize := jumpTableEntries * nodeIdSize

	// Now iterate over all DonutForests
	donutForestsCount := uint64(len(tb.DonutForestsInfo))
	for donutForestIndex := range donutForestsCount {
		donutForestInfo := &tb.DonutForestsInfo[donutForestIndex]

		// Read root of SingleTree from jump table
		jumpTableByteOffset := donutForestIndex*jumpTableSize + jumpTableIndex*nodeIdSize
		singleTreeNodeId := (*nodeIdConfig).ReadID(tb.JumpTableMemoryMap[jumpTableByteOffset : jumpTableByteOffset+nodeIdSize])
		if singleTreeNodeId != 0 {
			level := prefixBytesCount
			tb.recurseVisitEveryLeaf(singleTreeNodeId, level, donutForestInfo,
				nodeIdConfig, hashIndexIdConfig, reassuranceBytesCount,
				func(hashIndexId HashIndexIdType) {

					// hashIndexId is local to the DonutForest, and it was snapshotted against
					// that forest's own FirstGlobalPresentationIndex when the forest was baked.
					// The hashesFile is global to the tier, so we still need the tier base for fileIndex.
					globalPi := GlobalPiFromHashIndexId(hashIndexId, donutForestInfo.FirstGlobalPresentationIndex)
					// Subtract the offset relevant to the tier
					relativeToTier := globalPi - tb.GetFirstPresentationIndex()
					fileIndex := int64(relativeToTier)

					// For each hashIndexId found at a leaf...
					// Look up the hash (Todo do this quicker with an mmap)
					hash, err := tb.hashesFile.ReadHashAt(fileIndex)
					if err != nil {
						panic(err)
					}

					// singleTreeOffset is used to convert globalPi to singleTreePi.
					// It needs to be the same in all cases for this bake, as these pairs will be sent to construct a SingleTree.
					// It is therefore the firstPresentationIndex from the source tier (ie, this TierBelow)
					//singleTreePiOffset := donutForestInfo.FirstGlobalPresentationIndex
					singleTreePi := SingleTreePiFromGlobalPi(globalPi)
					pair := SingleTreeHash{PresentationIndex: singleTreePi, Hash: hash[:]}
					result = append(result, pair)
				})
		}
	}
	return result
}

// GetHashesAtIndex() requires a VisitEveryLeaf tree walker...
func (tb *TierBelow) recurseVisitEveryLeaf(nodeIdWithinLevel NodeIdType, levelNum byte, donutForestInfo *DonutForestInfo,
	nodeIdConfig *NByteIdConfig[NodeIdType], hashIndexIdConfig *NByteIdConfig[HashIndexIdType],
	reassuranceBytesCount byte, leafVisitor func(hashIndexId HashIndexIdType)) {

	// Look at the node we were directed to
	var node donutForestNode
	donutForestInfo.Levels[levelNum].ExtractNode(nodeIdWithinLevel, &node, nodeIdConfig)

	// Have we reached a leaf node?
	isLeaf, _, hashIndexId := node.detailsIfLeaf(reassuranceBytesCount, hashIndexIdConfig)
	if isLeaf {
		// Here we are not interested in reassurance bytes, as we are not trying to match against a particular hash
		leafVisitor(hashIndexId)
	} else {
		// Not a leaf.
		// This node is instructing us to dig deeper, by recursing all slots
		// Even though we are not trying to match against a particular hash, (and so we are not interested in
		// knowing which byte to examine), we still need to call the following function, to get the mediumSlots
		// and tinySlots values
		_, mediumSlots, tinySlots := node.hashByteIndexToExamine(nodeIdConfig)
		// Get a list of the slots
		nodeIdList := node.getAllNextLevelNodeIds(mediumSlots, tinySlots, nodeIdConfig)
		// Go deeper, with these various node id's at the next level...
		for _, nodeIdNextLevel := range nodeIdList {
			tb.recurseVisitEveryLeaf(nodeIdNextLevel, levelNum+1, donutForestInfo,
				nodeIdConfig, hashIndexIdConfig, reassuranceBytesCount, leafVisitor)
		}
	}
}
func (tb *TierBelow) AppendHashesFile(hashesFile *os.File) error {
	srcFilename := filepath.Join(tb.TierFolder, "Hashes.hsh")
	srcFile, err := os.Open(srcFilename)
	if err != nil {
		return err
	}
	defer func() { _ = srcFile.Close() }()

	defer func() { _ = hashesFile.Close() }()

	// 3. Efficiently stream/copy the data from source to destination
	// io.Copy uses a small internal buffer, preventing high memory usage for large files
	_, err = io.Copy(hashesFile, srcFile)
	if err != nil {
		return err
	}

	return nil
}
func (tb *TierBelow) GetFirstPresentationIndex() GlobalPiType {
	// This is only called when TierBelow is the source tier in the context of baking.
	// That only happens when the source tier fills.
	// We can therefore assume that the tier is not empty.
	return tb.DonutForestsInfo[0].FirstGlobalPresentationIndex
}
func (tb *TierBelow) MakeEmptyAfterBaking() error {
	// It took a bit of thought.
	// But I'm pretty sure that an "Empty after baking" BelowTier[n] is as simple as a non-existent "Tier<N>" folder!
	// Tiers above (and they WILL exist because this isn't a TierTop) are perfectly capable of recreating this tier
	// when they fill.
	err := tb.CloseThis()
	if err != nil {
		return err
	}

	// Close the subsequent tier before deleting this tier's files.
	// On Windows the parent directory removal is sensitive to any still-open handles in the tier chain.
	if tb.nextTier != nil {
		err = tb.nextTier.CloseThis()
		if err != nil {
			return err
		}
		tb.nextTier = nil
	}

	// Delete the whole tier folder
	err = os.RemoveAll(tb.TierFolder)
	if err != nil {
		return err
	}

	// Indicate empty tier by emptying DonutForestsInfo
	// In fact, since we have a function for it, clear everything
	tb.OpenAsEmptyTier()

	return nil
}
func (tb *TierBelow) SetNextTier(tierReadable TierReadable) {
	tb.nextTier = tierReadable
}
