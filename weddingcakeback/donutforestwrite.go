package weddingcakeback

import (
	"fmt"
	"os"
	"path/filepath"
)

type DonutForestWrite struct {
	Config     *CakeConfig
	SourceTier BakingSourceTier
	Designer   *BakingDesigner
	LevelFiles [65]*os.File
}

func NewDonutForestWrite(sourceTier BakingSourceTier, config *CakeConfig) *DonutForestWrite {
	result := DonutForestWrite{}
	result.Config = config
	result.SourceTier = sourceTier
	result.Designer = NewBakingDesigner()
	for i := range 65 {
		result.LevelFiles[i] = nil // Nil until opened
	}
	return &result
}

func (dfw *DonutForestWrite) Write(cakeFolder string) error {
	destTierIndex := dfw.SourceTier.GetNextTierIndex()

	tierFolder := fmt.Sprintf("Tier%d", destTierIndex)
	tierFolderPath := filepath.Join(cakeFolder, tierFolder)
	err := os.MkdirAll(tierFolderPath, 0755)
	if err != nil {
		return err
	}

	jumpsFilePath := filepath.Join(tierFolderPath, "DonutForestsJumpTables.bin")
	jumpsFile, err := os.OpenFile(jumpsFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer func() { _ = jumpsFile.Close() }()

	dfw.Designer.GatherMetricsFromSourceTier(dfw.SourceTier, dfw.Config)
	bakingDesign := dfw.Designer.DesignTheDesign(dfw.Config, destTierIndex)

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
			tree := GenerateSingleTree(bucket, destPrefixBytesCount, dfw.Config.HashLength,
				dfw.Config.TierBelowConfigs[destTierIndex].ReassuranceBytesCount)
			levelsNodesBytes, rootNodeId := dfw.serializeSingleTreeNodeBytes(tree, bakingDesign)
			err = dfw.appendToLevelsFiles(levelsNodesBytes, tierFolderPath)
			if err != nil {
				return err
			}

			// Append rootNodeId to jumps file
			nodeIdSize := dfw.Config.TierBelowConfigs[destTierIndex].NodeIdConfig.StorageBytes()
			nodeIdBytes := make([]byte, nodeIdSize)
			dfw.Config.TierBelowConfigs[destTierIndex].NodeIdConfig.WriteID(nodeIdBytes, rootNodeId)
			_, err = jumpsFile.Write(nodeIdBytes)
			if err != nil {
				return err
			}
		}
	}
	for _, file := range dfw.LevelFiles {
		if file != nil {
			err := file.Close()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// The first result is a slice of byte slices, one for each level of the tree.
// The second result is a node id for the root node of the tree (to go into the jump table)
func (dfw *DonutForestWrite) serializeSingleTreeNodeBytes(singleTree *SingleTree,
	bakingDesign *BakingDesign) ([][]byte, NodeIdType) {

	levels := len(bakingDesign.LevelSpecs)

	// 1. Group nodes by level just like before
	nodesByLevel := make([][]*SingleTreeNode, levels)
	singleTree.VisitAllNodes(func(node *SingleTreeNode) {
		nodesByLevel[node.Level] = append(nodesByLevel[node.Level], node)
	})

	// 2. We only need ONE map for the "level below us" at any given time
	var nextLevelIdMap map[*SingleTreeNode]NodeIdType

	// Prepare space for results: For each level, a byte slice
	levelsNodesBytes := make([][]byte, levels)
	for i := 0; i < levels; i++ {
		levelsNodesBytes[i] = make([]byte, 0, 10_000)
	}
	// Also as a result, the root node id for the tree
	rootNodeId := NodeIdType(0) // Starts out as zero

	// 3. Process bottom-up
	for levelNum := levels - 1; levelNum >= 0; levelNum-- {
		// fmt.Printf("Processing level %d\n", levelNum)
		currentLevelNodes := nodesByLevel[levelNum]
		levelNodesBytes := &levelsNodesBytes[levelNum]
		// Create a fresh map for the current level allocations
		currentLevelIdMap := make(map[*SingleTreeNode]NodeIdType, len(currentLevelNodes))

		// Pass A: Allocate IDs and populate our current level map
		for _, node := range currentLevelNodes {
			activeSlots := node.activeSlotsCount() // assuming helper attached
			nodeID, _ := bakingDesign.AllocateIdAndSpecForNode(node.Level, activeSlots)
			currentLevelIdMap[node] = nodeID
			// If it's the root of the SingleTree, make a note of the ID
			if node == singleTree.RootSlot.NextNode {
				if rootNodeId != 0 {
					panic("rootNodeId already set")
				}
				rootNodeId = nodeID
			}
		}

		// Pass B: Serialize this level's nodes.
		// When a node looks up a child, it queries nextLevelIdMap in O(1) time!
		for groupIdx, group := range bakingDesign.LevelSpecs[levelNum].Groups {
			spec := &group.Spec

			// Only serialize nodes at this level that belong to the current format group
			for _, node := range currentLevelNodes {
				nodeGroup := bakingDesign.LevelSpecs[levelNum].SlotCountToGroup[node.activeSlotsCount()]
				if int(nodeGroup) != groupIdx {
					continue // Skip until we hit this group's turn
				}

				// Pass the map belonging to levelNum + 1 down to the serializer
				switch spec.Format {
				case NodeFormatLeaf:
					// fmt.Println("Serializing FormatLeaf node")
					dfw.serializeLeafNode(node.LeafNode, levelNodesBytes)
				case NodeFormatFull:
					// fmt.Println("Serializing FormatFull node")
					dfw.serializeFullNode(node.SlotsNode, nextLevelIdMap, levelNodesBytes)
				case NodeFormatMedium:
					// fmt.Println("Serializing FormatMedium node")
					dfw.serializeMediumNode(node.SlotsNode, spec, nextLevelIdMap, levelNodesBytes)
				case NodeFormatTiny:
					// fmt.Println("Serializing FormatTiny node")
					dfw.serializeTinyNode(node.SlotsNode, spec, nextLevelIdMap, levelNodesBytes)
				}
			}
		}

		// Promote the current map to be the "nextLevel" map for the tier above us,
		// allowing the old nextLevelIdMap to be immediately garbage collected!
		nextLevelIdMap = currentLevelIdMap
	}
	return levelsNodesBytes, rootNodeId
}

func (dfw *DonutForestWrite) appendToLevelsFiles(levelsNodesBytes [][]byte, tierFolderPath string) error {
	for levelNum, levelNodesBytes := range levelsNodesBytes {
		if len(levelNodesBytes) != 0 {
			if dfw.LevelFiles[levelNum] == nil {
				// Open for append or create if not yet opened
				filename := fmt.Sprintf("Level%02dNodes.bin", levelNum)
				filePath := filepath.Join(tierFolderPath, filename)
				var err error
				dfw.LevelFiles[levelNum], err = os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					return err
				}
			}
			_, err := dfw.LevelFiles[levelNum].Write(levelNodesBytes)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (dfw *DonutForestWrite) serializeLeafNode(leafNode *SingleTreeLeafNode, bytes *[]byte) {
	// A leaf node is the reassurance bytes followed by the hash index id

	// In ShallowTree, it is clever enough to give fewer reassurance bytes than configured, in cases where
	// there are not enough bytes left to examine in the hash. But our serialized leaf node has a fixed
	// capacity for these, so we need to pad them.
	destTierIndex := dfw.SourceTier.GetNextTierIndex()
	reassuranceBytesCount := dfw.Config.TierBelowConfigs[destTierIndex].ReassuranceBytesCount
	reassurancePadding := reassuranceBytesCount - byte(len(leafNode.ReassuranceHashBytes))
	*bytes = append(*bytes, leafNode.ReassuranceHashBytes...)
	if reassurancePadding > 0 {
		for pad := byte(0); pad < reassurancePadding; pad++ {
			*bytes = append(*bytes, 0)
		}
	}
	pi64 := leafNode.PresentationIndex
	if pi64 == 0 {
		panic("Unexpected presentation index 0")
	}
	if uint64(pi64) > uint64(MaxHashIndexId) {
		panic("Presentation index too big for HashIndexIdType")
	}
	piSmall := HashIndexIdType(pi64)
	hashIndexIdSize := dfw.Config.TierBelowConfigs[destTierIndex].HashIndexIdConfig.StorageBytes()
	const spareRoom = 8
	var hashIndexIdBytes [spareRoom]byte
	dfw.Config.TierBelowConfigs[destTierIndex].HashIndexIdConfig.WriteID(hashIndexIdBytes[:hashIndexIdSize], piSmall)
	*bytes = append(*bytes, hashIndexIdBytes[:hashIndexIdSize]...)
}

func (dfw *DonutForestWrite) serializeFullNode(slotsNode *SingleTreeSlotsNode,
	nextLevelIdMap map[*SingleTreeNode]NodeIdType, bytes *[]byte) {
	// A full node is one byte padding (0), one byte hash byte index, and 256 N-byte nodeId slots.
	// (a nodeId of 0 is used to indicate an empty slot)
	// A full node is therefore fixed size (for a particular nodeIdsize configuration) and can be done in one append
	destTierIndex := dfw.SourceTier.GetNextTierIndex()
	nodeIdConfig := &dfw.Config.TierBelowConfigs[destTierIndex].NodeIdConfig
	nodeIdSize := (*nodeIdConfig).StorageBytes()
	fullNodeSize := 1 + 1 + 256*nodeIdSize
	const spareRoom = 1 + 1 + 256*8
	var nodeBytes [spareRoom]byte
	nodeBytes[0] = 0 // Padding
	nodeBytes[1] = slotsNode.HashByteIndex
	p := 2
	for s := 0; s < 256; s++ {
		if slotsNode.Slots[s].IsEmpty() {
			(*nodeIdConfig).WriteID(nodeBytes[p:p+nodeIdSize], 0)
		} else {
			nodeId, ok := nextLevelIdMap[slotsNode.Slots[s].NextNode]
			if !ok {
				panic("Node pointer not found in map")
			}
			if nodeId == 0 {
				panic("Node id in map is zero")
			}
			(*nodeIdConfig).WriteID(nodeBytes[p:p+nodeIdSize], nodeId)
		}
		p += nodeIdSize
	}
	if p != fullNodeSize {
		panic("Error in byte counting code")
	}
	*bytes = append(*bytes, nodeBytes[:fullNodeSize]...)
}

func (dfw *DonutForestWrite) serializeMediumNode(slotsNode *SingleTreeSlotsNode, spec *NodeFormatSpec,
	nextLevelIdMap map[*SingleTreeNode]NodeIdType, bytes *[]byte) {

	// Total length matching our index bytes estimation:
	// 1 (pad) + 1 (index) + 32 (bitmask flags) + N * SlotsCapacity
	destTierIndex := dfw.SourceTier.GetNextTierIndex()
	nodeIdConfig := &dfw.Config.TierBelowConfigs[destTierIndex].NodeIdConfig
	nodeIdSize := (*nodeIdConfig).StorageBytes()
	totalBytesCount := 1 + 1 + 32 + (nodeIdSize * int(spec.SlotsCapacity))
	nodeBytes := make([]byte, totalBytesCount)

	nodeBytes[0] = 0                       // 1 byte padding
	nodeBytes[1] = slotsNode.HashByteIndex // 1 byte index

	flagsOffset := 2
	payloadOffset := flagsOffset + 32

	// 1. Build out the 256-bit flag bitmask and collect active target nodes sequentially
	activeChildren := make([]*SingleTreeNode, 0, 256)

	for s := 0; s < 256; s++ {
		if !slotsNode.Slots[s].IsEmpty() {
			// Find byte bucket (0-31) and target bit location (0-7)
			byteNum := s >> 3
			bitNum := s & 0x07

			// Set the flag matching our bit layout query
			nodeBytes[flagsOffset+byteNum] |= (1 << bitNum)

			// Collect the target child in strict iteration order
			activeChildren = append(activeChildren, slotsNode.Slots[s].NextNode)
		}
	}

	// 2. Write the 16-bit nodeIDs for active slots into the payload track
	for _, childNode := range activeChildren {
		nodeId, ok := nextLevelIdMap[childNode]
		if !ok {
			panic("Node pointer not found in map")
		}
		if nodeId == 0 {
			panic("Node id in map is zero")
		}

		(*nodeIdConfig).WriteID(nodeBytes[payloadOffset:payloadOffset+nodeIdSize], nodeId)
		payloadOffset += nodeIdSize
	}

	// 3. Right-pad trailing payload space with 0x0000
	// (Unpopulated capacity 'words' remain zero-initialized as bytes automatically from make)

	*bytes = append(*bytes, nodeBytes...)
}

func (dfw *DonutForestWrite) serializeTinyNode(slotsNode *SingleTreeSlotsNode, spec *NodeFormatSpec,
	nextLevelIdMap map[*SingleTreeNode]NodeIdType, bytes *[]byte) {
	// FormatTiny consists of one byte hash byte index (no padding this time) followed
	// by a sequence of {one byte hash byte value, and N-bytes nodeId} with empty slots allowed (nodeId=0).
	// Crucially, the length of the sequence is NOT NECESSARILY equal to the number of non-empty slots.
	destTierIndex := dfw.SourceTier.GetNextTierIndex()
	nodeIdConfig := &dfw.Config.TierBelowConfigs[destTierIndex].NodeIdConfig
	nodeIdSize := (*nodeIdConfig).StorageBytes()
	nodeBytesCount := 1 + (1+nodeIdSize)*int(spec.SlotsCapacity)
	const spareRoom = 1 + (1+8)*5
	if nodeBytesCount > spareRoom {
		panic("Not enough room for tiny node")
	}
	nodeBytes := [spareRoom]byte{}
	nodeBytes[0] = slotsNode.HashByteIndex
	// Find the non-empty slots (which will always fit into the capacity, by prior arrangement)
	p := 1
	for sInt := 0; sInt < 256; sInt++ {
		if slotsNode.Slots[sInt].IsEmpty() {
			// If empty, it simply is not stored as part of the sequence!
		} else {
			nodeBytes[p] = byte(sInt)
			nodeId, ok := nextLevelIdMap[slotsNode.Slots[sInt].NextNode]
			if !ok {
				panic("Node pointer not found in map")
			}
			if nodeId == 0 {
				panic("Node id in map is zero")
			}
			(*nodeIdConfig).WriteID(nodeBytes[p+1:p+1+nodeIdSize], nodeId)
			p += 1 + nodeIdSize
		}
	}
	// If there is remaining capacity, we leave these as zero bytes (the zero bytes for nodeId imply
	// an empty slot)
	*bytes = append(*bytes, nodeBytes[:nodeBytesCount]...)
}
