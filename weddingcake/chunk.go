package weddingcake

import (
	"encoding/binary"
	"fmt"
	"math"
)

type GlobalPiType int64 // Initial presentation indices. Can start at 0 or more. -1 reserved for "no match".
const GlobalPiNoMatch = GlobalPiType(-1)

// Chunk represents a lookup tree for a fixed set of sequential hashes.
// Chunks are created "in one go"; they do not support appending or inserting of hashes. They do not grow.
// Chunks have the various level's of their tree (level 0 is the root) stored in different files.
type Chunk struct {
	Config                              StoreConfig
	FirstGlobalPresentationIndexOfChunk GlobalPiType
	ChunkLevels                         []ChunkLevel
}

func (ch *Chunk) LookupHash(hash []byte) GlobalPiType {
	hashIndexId := ch.lookupHash(hash)
	return ch.GlobalPresentationIndexFromHashIndexId(hashIndexId)
}

func (ch *Chunk) HashIndexIdFromGlobalPresentationIndex(global GlobalPiType) HashIndexIdType {
	// The value reserved for "no match"
	if global == GlobalPiNoMatch {
		return HashIndexIdNoMatch
	}
	// ch.FirstGlobalPresentationIndexOfChunk maps to hashIndexId 1
	hashIndexId := global - ch.FirstGlobalPresentationIndexOfChunk + 1
	if hashIndexId < 1 {
		panic("Global presentation index lower than first presentation index")
	}
	if hashIndexId > GlobalPiType(MaxHashIndexId) {
		panic("Global presentation indices don't fit in HashIndexIdType")
	}
	return HashIndexIdType(hashIndexId)
}

func (ch *Chunk) GlobalPresentationIndexFromHashIndexId(hashIndexId HashIndexIdType) GlobalPiType {
	// The value reserved for "no match"
	if hashIndexId == HashIndexIdNoMatch {
		return GlobalPiNoMatch
	}
	// HashIndexIdType 1 maps to ch.FirstGlobalPresentationIndexOfChunk
	return GlobalPiType(hashIndexId) - 1 + ch.FirstGlobalPresentationIndexOfChunk
}

type ChunkPreparation struct {
	// The input. PresentationIndices are "global" (int64) and don't necessarily start at 0 or 1.
	hashes []ShallowTreeHash
	// Intermediate result.
	// Presentation indices have been converted to "hash index ids" (1 <= n <= ?), but are stored as int64's
	shallowTree *ShallowTree
	// Intermediate result
	treeFormat *TreeFormat
	// Ultimate result
	chunk Chunk
}

func NewChunkPreparation(hashes []ShallowTreeHash,
	hashLength byte, reassuranceBytesCount byte, nodeFormatSpecsPerLevel byte,
	nodeIdConfig NByteIdConfig[NodeIdType], hashIndexIdConfig NByteIdConfig[HashIndexIdType]) *ChunkPreparation {

	result := ChunkPreparation{}
	result.chunk.Config.HashLength = hashLength
	result.chunk.Config.ReassuranceBytesCount = reassuranceBytesCount
	result.chunk.Config.NodeFormatSpecsPerLevel = nodeFormatSpecsPerLevel
	result.chunk.Config.NodeIdConfig = nodeIdConfig
	result.chunk.Config.HashIndexIdConfig = hashIndexIdConfig
	result.hashes = hashes
	return &result
}

func (cp *ChunkPreparation) PrepareAndSerialize() {
	if len(cp.hashes) == 0 {
		cp.chunk.FirstGlobalPresentationIndexOfChunk = 0
	} else {
		cp.chunk.FirstGlobalPresentationIndexOfChunk = GlobalPiType(cp.hashes[0].PresentationIndex)
	}
	cp.shallowTree = cp.generateShallowTree(cp.hashes)
	cp.verifyShallowTree() // Optional debug
	cp.treeFormat = cp.chunk.Config.DesignTreeFormat(cp.shallowTree)
	cp.allocateBytes()
	cp.serializeIndexBytes()
	cp.serializeNodeBytes()
}

func (cp *ChunkPreparation) generateShallowTree(input []ShallowTreeHash) *ShallowTree {
	// We must convert all global presentation indices to local indices
	updated := make([]ShallowTreeHash, len(input))
	for ind := range input {
		updated[ind].Hash = input[ind].Hash // Copy by slice (basically a pointer). We won't be modifying them.
		updated[ind].PresentationIndex = int64(cp.chunk.HashIndexIdFromGlobalPresentationIndex(GlobalPiType(input[ind].PresentationIndex)))
	}
	return GenerateShallowTree(updated, cp.chunk.Config.HashLength, cp.chunk.Config.ReassuranceBytesCount)
}

func (cp *ChunkPreparation) verifyShallowTree() {
	cp.shallowTree.VisitAllNodes(func(node *ShallowTreeNode) {
		leafNode := node.LeafNode
		if leafNode != nil {
			if len(leafNode.ReassuranceHashBytes) > int(cp.chunk.Config.ReassuranceBytesCount) {
				panic("Found wrong number of reassurance bytes")
			}
			if leafNode.PresentationIndex == 0 {
				panic("Found zero presentation index")
			}
			if leafNode.PresentationIndex > int64(MaxHashIndexId) {
				panic("Found presentation index > MaxHashIndexId")
			}
		}
	})
}

func (cp *ChunkPreparation) allocateBytes() {
	levels := len(cp.treeFormat.LevelSpecs)
	cp.chunk.ChunkLevels = make([]ChunkLevel, levels)
	for levelNum := 0; levelNum < levels; levelNum++ {
		indexBytes, nodeBytes := countChunkLevelBytes(&cp.treeFormat.LevelSpecs[levelNum], cp.chunk.Config.NodeIdConfig)
		cp.chunk.ChunkLevels[levelNum].IndexBytes = make([]byte, 0, indexBytes)
		cp.chunk.ChunkLevels[levelNum].NodeBytes = make([]byte, 0, nodeBytes)
		// fmt.Printf("allocateBytes(): Level %d: nodeBytes = %d\n", levelNum, nodeBytes)
	}
}

func (cp *ChunkPreparation) serializeIndexBytes() {
	levels := len(cp.treeFormat.LevelSpecs)
	nodeIdSize := cp.chunk.Config.NodeIdConfig.StorageBytes()
	nodesCountSize := nodeIdSize
	hashIndexIdSize := cp.chunk.Config.HashIndexIdConfig.StorageBytes()
	for levelNum := 0; levelNum < levels; levelNum++ {
		formatSpecGroups := &cp.treeFormat.LevelSpecs[levelNum].Groups
		levelIndexBytes := &cp.chunk.ChunkLevels[levelNum].IndexBytes

		// In each level, we start with two bytes representing the count of NodeSpec's ("group"s) that follow
		var serializedGroupsCountBytes [2]byte
		binary.LittleEndian.PutUint16(serializedGroupsCountBytes[:], uint16(len(*formatSpecGroups)))
		*levelIndexBytes = append(*levelIndexBytes, serializedGroupsCountBytes[:]...)

		for groupIndex := range *formatSpecGroups {
			group := (*formatSpecGroups)[groupIndex]
			// Whilst we call this a "group", this has only come about by merging of individual
			// formatSpecs in StoreConfig.DesignTreeFormat(). The "group" is in fact governed
			// by a single FormatSpec, which we serialize here.
			const spareRoom = 8 // The most space we will ever need
			if nodesCountSize > spareRoom {
				panic("Not enough bytes")
			}
			serializedNodesCountBytes := [spareRoom]byte{} // The count of nodes expressed as "some" bytes
			cp.chunk.Config.NodeIdConfig.WriteID(serializedNodesCountBytes[:nodesCountSize], NodeIdType(group.NodesCount))
			*levelIndexBytes = append(*levelIndexBytes, serializedNodesCountBytes[:nodesCountSize]...)
			serializedNodeSpecBytes := [4]byte{} // The details of the FormatSpecs for these nodes
			switch group.Spec.Format {
			case NodeFormatFull:
				// Most significant bytes pair = zero, LS byte pair = number of bytes per node
				// Number of bytes per node is (1) pad + (1) hashByteIndex + (256 * N) node ids
				bytesPerNodeFull := 1 + 1 + 256*nodeIdSize
				binary.LittleEndian.PutUint32(serializedNodeSpecBytes[:], uint32(bytesPerNodeFull))
			case NodeFormatLeaf:
				// Most significant bytes pair = zero, LS byte pair = number of bytes per node
				// Number of bytes per node is (Reassurance bytes count) + (size of a hash index id)
				bytesPerNodeLeaf := uint32(int(cp.chunk.Config.ReassuranceBytesCount) + hashIndexIdSize)
				binary.LittleEndian.PutUint32(serializedNodeSpecBytes[:], bytesPerNodeLeaf)
			case NodeFormatMedium:
				// MS byte = zero, then slots byte, LS byte pair = number of bytes per node
				slotsFields := uint32(group.Spec.SlotsCapacity) << 16
				// Bytes per node = 1 (pad) + 1 (hash byte index) + 32 (slot flags) + N (node id) * slotsCapacity
				bytesPerNodeField := uint32(1 + 1 + 32 + nodeIdSize*int(group.Spec.SlotsCapacity))
				binary.LittleEndian.PutUint32(serializedNodeSpecBytes[:], slotsFields|bytesPerNodeField)
			case NodeFormatTiny:
				// MS byte slots capacity byte, then zero, LS byte pair = number of bytes per node
				slotsFields := uint32(group.Spec.SlotsCapacity) << 24
				// Bytes per node = 1 (hash byte index) + slots capacity * (1 (hash byte value) + N (node id))
				bytesPerNodeField := uint32(1 + int(group.Spec.SlotsCapacity)*(1+nodeIdSize))
				binary.LittleEndian.PutUint32(serializedNodeSpecBytes[:], slotsFields|bytesPerNodeField)
			}
			*levelIndexBytes = append(*levelIndexBytes, serializedNodeSpecBytes[:]...)
		}
		// Check (because we can) that we have exactly reached capacity
		if len(*levelIndexBytes) != cap(*levelIndexBytes) {
			panic("Error in byte counting code")
		}
	}
}

func (cp *ChunkPreparation) serializeNodeBytes() {
	levels := len(cp.treeFormat.LevelSpecs)

	// 1. Group nodes by level just like before
	nodesByLevel := make([][]*ShallowTreeNode, levels)
	cp.shallowTree.VisitAllNodes(func(node *ShallowTreeNode) {
		nodesByLevel[node.Level] = append(nodesByLevel[node.Level], node)
	})

	// 2. We only need ONE map for the "level below us" at any given time
	var nextLevelIdMap map[*ShallowTreeNode]NodeIdType

	// 3. Process bottom-up
	for levelNum := levels - 1; levelNum >= 0; levelNum-- {
		// fmt.Printf("Processing level %d\n", levelNum)
		currentLevelNodes := nodesByLevel[levelNum]
		levelNodesBytes := &cp.chunk.ChunkLevels[levelNum].NodeBytes
		// Create a fresh map for the current level allocations
		currentLevelIdMap := make(map[*ShallowTreeNode]NodeIdType, len(currentLevelNodes))

		// Pass A: Allocate IDs and populate our current level map
		for _, node := range currentLevelNodes {
			activeSlots := node.activeSlotsCount() // assuming helper attached
			nodeID, _ := cp.treeFormat.AllocateIdAndSpecForNode(node.Level, activeSlots)
			currentLevelIdMap[node] = nodeID
		}

		// Pass B: Serialize this level's nodes.
		// When a node looks up a child, it queries nextLevelIdMap in O(1) time!
		for groupIdx, group := range cp.treeFormat.LevelSpecs[levelNum].Groups {
			spec := &group.Spec

			// Only serialize nodes at this level that belong to the current format group
			for _, node := range currentLevelNodes {
				nodeGroup := cp.treeFormat.LevelSpecs[levelNum].SlotCountToGroup[node.activeSlotsCount()]
				if int(nodeGroup) != groupIdx {
					continue // Skip until we hit this group's turn
				}

				// Pass the map belonging to levelNum + 1 down to the serializer
				switch spec.Format {
				case NodeFormatLeaf:
					// fmt.Println("Serializing FormatLeaf node")
					cp.serializeLeafNode(node.LeafNode, levelNodesBytes)
				case NodeFormatFull:
					// fmt.Println("Serializing FormatFull node")
					cp.serializeFullNode(node.SlotsNode, nextLevelIdMap, levelNodesBytes)
				case NodeFormatMedium:
					// fmt.Println("Serializing FormatMedium node")
					cp.serializeMediumNode(node.SlotsNode, spec, nextLevelIdMap, levelNodesBytes)
				case NodeFormatTiny:
					// fmt.Println("Serializing FormatTiny node")
					cp.serializeTinyNode(node.SlotsNode, spec, nextLevelIdMap, levelNodesBytes)
				}
			}
		}
		// Promote the current map to be the "nextLevel" map for the tier above us,
		// allowing the old nextLevelIdMap to be immediately garbage collected!
		nextLevelIdMap = currentLevelIdMap
		// Just because we can, check that level nodes bytes are full to capacity
		if len(*levelNodesBytes) != cap(*levelNodesBytes) {
			fmt.Printf("Level %d: len(*levelNodeBytes) = %d, cap(*levelNodeBytes) = %d\n", levelNum, len(*levelNodesBytes), cap(*levelNodesBytes))
			panic("Expected nodes bytes to be full to capacity")
		}
	}
}

func (cp *ChunkPreparation) serializeLeafNode(leafNode *ShallowTreeLeafNode, bytes *[]byte) {
	// A leaf node is the reassurance bytes followed by the hash index id

	// In ShallowTree, it is clever enough to give fewer reassurance bytes than configured, in cases where
	// there are not enough bytes left to examine in the hash. But our serialized leaf node has a fixed
	// capacity for these, so we need to pad them.
	reassurancePadding := cp.chunk.Config.ReassuranceBytesCount - byte(len(leafNode.ReassuranceHashBytes))
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
	if pi64 > int64(MaxHashIndexId) {
		panic("Presentation index too big for HashIndexIdType")
	}
	piSmall := HashIndexIdType(pi64)
	hashIndexIdSize := cp.chunk.Config.HashIndexIdConfig.StorageBytes()
	const spareRoom = 8
	var hashIndexIdBytes [spareRoom]byte
	cp.chunk.Config.HashIndexIdConfig.WriteID(hashIndexIdBytes[:hashIndexIdSize], piSmall)
	*bytes = append(*bytes, hashIndexIdBytes[:hashIndexIdSize]...)
}

func (cp *ChunkPreparation) serializeFullNode(slotsNode *ShallowTreeSlotsNode,
	nextLevelIdMap map[*ShallowTreeNode]NodeIdType, bytes *[]byte) {
	// A full node is one byte padding (0), one byte hash byte index, and 256 N-byte nodeId slots.
	// (a nodeId of 0 is used to indicate an empty slot)
	// A full node is therefore fixed size (for a particular nodeIdsize configuration) and can be done in one append
	nodeIdSize := cp.chunk.Config.NodeIdConfig.StorageBytes()
	fullNodeSize := 1 + 1 + 256*nodeIdSize
	const spareRoom = 1 + 1 + 256*8
	var nodeBytes [spareRoom]byte
	nodeBytes[0] = 0 // Padding
	nodeBytes[1] = slotsNode.HashByteIndex
	p := 2
	for s := 0; s < 256; s++ {
		if slotsNode.Slots[s].IsEmpty() {
			cp.chunk.Config.NodeIdConfig.WriteID(nodeBytes[p:p+nodeIdSize], 0)
		} else {
			nodeId, ok := nextLevelIdMap[slotsNode.Slots[s].NextNode]
			if !ok {
				panic("Node pointer not found in map")
			}
			if nodeId == 0 {
				panic("Node id in map is zero")
			}
			cp.chunk.Config.NodeIdConfig.WriteID(nodeBytes[p:p+nodeIdSize], nodeId)
		}
		p += nodeIdSize
	}
	if p != fullNodeSize {
		panic("Error in byte counting code")
	}
	*bytes = append(*bytes, nodeBytes[:fullNodeSize]...)
}

func (cp *ChunkPreparation) serializeMediumNode(slotsNode *ShallowTreeSlotsNode, spec *NodeFormatSpec,
	nextLevelIdMap map[*ShallowTreeNode]NodeIdType, bytes *[]byte) {

	// Total length matching our index bytes estimation:
	// 1 (pad) + 1 (index) + 32 (bitmask flags) + N * SlotsCapacity
	nodeIdSize := cp.chunk.Config.NodeIdConfig.StorageBytes()
	totalBytesCount := 1 + 1 + 32 + (nodeIdSize * int(spec.SlotsCapacity))
	nodeBytes := make([]byte, totalBytesCount)

	nodeBytes[0] = 0                       // 1 byte padding
	nodeBytes[1] = slotsNode.HashByteIndex // 1 byte index

	flagsOffset := 2
	payloadOffset := flagsOffset + 32

	// 1. Build out the 256-bit flag bitmask and collect active target nodes sequentially
	activeChildren := make([]*ShallowTreeNode, 0, 256)

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

		cp.chunk.Config.NodeIdConfig.WriteID(nodeBytes[payloadOffset:payloadOffset+nodeIdSize], nodeId)
		payloadOffset += nodeIdSize
	}

	// 3. Right-pad trailing payload space with 0x0000
	// (Unpopulated capacity 'words' remain zero-initialized as bytes automatically from make)

	*bytes = append(*bytes, nodeBytes...)
}

func (cp *ChunkPreparation) serializeTinyNode(slotsNode *ShallowTreeSlotsNode, spec *NodeFormatSpec,
	nextLevelIdMap map[*ShallowTreeNode]NodeIdType, bytes *[]byte) {
	// FormatTiny consists of one byte hash byte index (no padding this time) followed
	// by a sequence of {one byte hash byte value, and N-bytes nodeId} with empty slots allowed (nodeId=0).
	// Crucially, the length of the sequence is NOT NECESSARILY equal to the number of non-empty slots.
	nodeIdSize := cp.chunk.Config.NodeIdConfig.StorageBytes()
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
			cp.chunk.Config.NodeIdConfig.WriteID(nodeBytes[p+1:p+1+nodeIdSize], nodeId)
			p += 1 + nodeIdSize
		}
	}
	// If there is remaining capacity, we leave these as zero bytes (the zero bytes for nodeId imply
	// an empty slot)
	*bytes = append(*bytes, nodeBytes[:nodeBytesCount]...)
}

func (ch *Chunk) lookupHash(hash []byte) HashIndexIdType {
	if ch.Config.HashLength == 0 {
		panic("Cannot deal with zero length hashes")
	}
	// Firstly, is there a level 0?
	if len(ch.ChunkLevels) == 0 {
		return 0 // There is no root node, and the chunk contains no hashes (and so certainly not the one specified)
	}
	// We start with node 1 of level 0, which should be the root node
	levelNum := byte(0)
	nodeIdWithinLevel := NodeIdType(1)
	// We also keep track of which hash byte indices have NOT yet been examined so far
	var flagsHashByteIndicesUnexamined uint64
	if ch.Config.HashLength == 64 {
		flagsHashByteIndicesUnexamined = math.MaxUint64
	} else {
		flagsHashByteIndicesUnexamined = (uint64(1) << ch.Config.HashLength) - 1
	}
	return ch.recurseLookupHash(hash, levelNum, nodeIdWithinLevel, flagsHashByteIndicesUnexamined)
}

func (ch *Chunk) recurseLookupHash(hash []byte, levelNum byte,
	nodeIdWithinLevel NodeIdType, flagsHashByteIndicesUnexamined uint64) HashIndexIdType {

	// Sanity check we haven't run out of levels
	if levelNum >= byte(len(ch.ChunkLevels)) {
		panic("Level does not exist")
	}

	// Look at the node we were directed to
	var node chunkNode
	ch.ChunkLevels[levelNum].ExtractNode(nodeIdWithinLevel, &node, ch.Config.NodeIdConfig)

	// Have we reached a leaf node?
	isLeaf, reassuranceBytes, hashIndexId := node.detailsIfLeaf(ch.Config.ReassuranceBytesCount, ch.Config.HashIndexIdConfig)
	if isLeaf {
		// Now we need to check some bytes of our hash... the "next" ones that haven't been examined yet...
		// against the reassurance bytes specified by the leaf
		mask := flagsHashByteIndicesUnexamined
		byteToMaybeExamine := 0
		reassuranceByteCounter := 0
		for mask != 0 && reassuranceByteCounter < len(reassuranceBytes) {
			if mask&uint64(1) == 1 {
				// Yes we need to examine it
				match := hash[byteToMaybeExamine] == reassuranceBytes[reassuranceByteCounter]
				if !match {
					// This leaf was our only shot at a match, but the reassurance bytes were not reassuring!
					//fmt.Printf("Leaf failed reassurance bytes match at level %d\n", levelNum)
					return 0 // Not a match
				}
				// A byte matched... now lets see if there are any more reassurance bytes available...
				reassuranceByteCounter++
			}
			// Either we didn't need to examine byteToExamine (it had already been examined),
			// or it matched. In either case, we may have more reassurance bytes to check...
			// so keep looking...
			mask >>= 1
			byteToMaybeExamine++
		}
		// All the available reassurance bytes matched... It's a potential match!
		// It is for the caller to double check the ENTIRE hash, now that we've identified a unique strong match
		// with the data available
		return hashIndexId
	}
	// Not a leaf.
	// This node is instructing us to dig deeper, by examining another byte in the hash.
	byteIndexToExamine, mediumSlots, tinySlots := node.hashByteIndexToExamine(ch.Config.NodeIdConfig)
	// Examine the specified byte
	byteThatWasFound := hash[byteIndexToExamine]
	// See if that byte get's us to another node at the next level...
	nextNodeId := node.nextLevelNodeId(byteThatWasFound, mediumSlots, tinySlots, ch.Config.NodeIdConfig)
	if nextNodeId == 0 {
		return 0 // A dead end. No match. The hash definitely isn't in this chunk.
	}
	// Make a note that this byte index of the hash has now been examined
	bitMask := uint64(1) << byteIndexToExamine
	if flagsHashByteIndicesUnexamined&bitMask == 0 {
		panic("We examined the same byte of the hash twice")
	}
	flagsHashByteIndicesUnexamined ^= bitMask // Clear the bit to say it is no longer unexamined

	// Go deeper, with our new node id at the next level...
	return ch.recurseLookupHash(hash, levelNum+1, nextNodeId, flagsHashByteIndicesUnexamined)
}
