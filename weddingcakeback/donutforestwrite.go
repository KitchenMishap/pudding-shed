package weddingcakeback

type DonutForestWrite struct {
	Config     *CakeConfig
	SourceTier BakingSourceTier
	Designer   *BakingDesigner
}

func NewDonutForestWrite(sourceTier BakingSourceTier, config *CakeConfig) *DonutForestWrite {
	result := DonutForestWrite{}
	result.Config = config
	result.SourceTier = sourceTier
	result.Designer = NewBakingDesigner()
	return &result
}

func (dfw *DonutForestWrite) Write() error {
	destTierIndex := dfw.SourceTier.GetNextTierIndex()
	dfw.Designer.GatherMetricsFromSourceTier(dfw.SourceTier, dfw.Config)
	_ = dfw.Designer.DesignTheDesign(dfw.Config, destTierIndex)

	destPrefixBytesCount := dfw.SourceTier.GetNextTierPrefixBytesCount()
	if destPrefixBytesCount != destTierIndex {
		panic("Expect the prefix bytes count to be equal to the tier index")
	}
	indexRange := dfw.SourceTier.GetIndicesCount()
	for index := range indexRange {
		// index refers to a "tree number" within EACH DonutForrest in the source tier
		// The following call amalgamates the hashes from the multiple "tree at index"'s taken from the source tier's DonutForests
		hashInfos := dfw.SourceTier.GetHashesAtIndex(index, dfw.Config)
		// (There should typically be about 65,536 hashes)

		// Because they were obtained by index (which chooses a tree in each source DonutForest),
		// these hashes will all have the same hash prefix.
		// In the destination tier, we are subdividing these hashes on an ADDITIONAL byte (the prefix gets longer.)
		// We need to put the hashes into buckets based on this "newly examined" byte of the hash.
		buckets := [256][]SingleTreeHash{}
		for i := range 256 {
			buckets[i] = make([]SingleTreeHash, 0, 300)
		}

		// EXCEPTION: If the destination tier index is 0 (source tier was TierTop), the "new longer prefix" is
		// still 0 bytes, so we just use bucket[0] for the all
		if destTierIndex == 0 {
			buckets[0] = hashInfos
		} else {
			byteIndex := destPrefixBytesCount - 1
			for _, hashInfo := range hashInfos {
				examinedByte := hashInfo.Hash[byteIndex]
				buckets[examinedByte] = append(buckets[examinedByte], hashInfo)
			}
		}
		// Now we either have one (in the case of TierTop) or 256 (in the case of TierBottom) buckets of hashes.
		// These buckets correspond to either one or 256 of the 256^n trees in the DonutForest we are writing to.
		// One by one, turn them into SingleTree's and write them.
		// We throw each tree away before starting on the next one, to conserve memory.
		var treeCount int
		if destTierIndex == 0 {
			treeCount = 1
		} else {
			treeCount = 256
		}
		for t := range treeCount {
			bucket := buckets[t]
			_ = GenerateSingleTree(bucket, destPrefixBytesCount, dfw.Config.HashLength,
				dfw.Config.TierBelowConfigs[destTierIndex].ReassuranceBytesCount)
			// ToDo Serialize
		}
	}
	return nil
}
