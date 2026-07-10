package weddingcakeback

type BakingDesign struct {
	// These are indexed by level, with root at index 0
	LevelSpecs []LevelFormat
}

type LevelFormat struct {
	// These are sorted for efficiency, most popular first (but with FormatTiny's at the end which use odd byte numbers)
	Groups []NodeFormatGroup
	// These have the same index as Groups
	NodeIdAllocations []NodeIdAllocation
	// These are indexed by active slot count, and hold indices into the above
	SlotCountToGroup [257]byte
}
type NodeIdAllocation struct {
	NextAvailableNodeId NodeIdType
	AvailableNodeIds    NodeIdType
}

func (bd *BakingDesign) InitializeNodeIdAllocations() {
	levels := len(bd.LevelSpecs)
	for level := 0; level < levels; level++ {
		// Node Id's are now PER LEVEL; you need a level AND a node id to identify a node
		nodeId := NodeIdType(1) // 0 is reserved with special meaning

		groups := len(bd.LevelSpecs[level].Groups)
		bd.LevelSpecs[level].NodeIdAllocations = make([]NodeIdAllocation, groups)
		for group := 0; group < groups; group++ {
			// These will later allocate us NodeIDs from each group in each level
			nodes := bd.LevelSpecs[level].Groups[group].NodesCount
			bd.LevelSpecs[level].NodeIdAllocations[group] = NodeIdAllocation{
				NextAvailableNodeId: nodeId,
				AvailableNodeIds:    NodeIdType(nodes),
			}
			// These will later tell us, for a given active slot count, which
			// group should allocate us NodeIDs
			start := bd.LevelSpecs[level].Groups[group].StartSlotsCount
			end := bd.LevelSpecs[level].Groups[group].EndSlotsCount
			for activeSlotsCount := start; activeSlotsCount < end; activeSlotsCount++ {
				bd.LevelSpecs[level].SlotCountToGroup[activeSlotsCount] = byte(group)
			}

			if uint64(nodeId)+uint64(nodes) > uint64(MaxNodesCount) {
				panic("Too many nodes for node NodeIdType")
			}
			nodeId += NodeIdType(nodes)
		}
	}
}

func (bd *BakingDesign) AllocateIdAndSpecForNode(level byte, activeSlotsCount int) (NodeIdType, NodeFormatSpec) {
	group := bd.LevelSpecs[level].SlotCountToGroup[activeSlotsCount]

	// Read directly from the underlying slice index
	alloc := bd.LevelSpecs[level].NodeIdAllocations[group]
	if alloc.AvailableNodeIds == 0 {
		panic("Too many nodes for datatype")
	}

	nodeID := alloc.NextAvailableNodeId

	// Update the live data tracking fields directly back inside the slice holder
	bd.LevelSpecs[level].NodeIdAllocations[group].NextAvailableNodeId++
	bd.LevelSpecs[level].NodeIdAllocations[group].AvailableNodeIds--

	nodeSpec := bd.LevelSpecs[level].Groups[group].Spec
	return nodeID, nodeSpec
}

func (bd *BakingDesign) countChunkLevelBytes(levelNum byte,
	nodeIdConfig *NByteIdConfig[NodeIdType]) (uint64, uint64) {

	levelData := &bd.LevelSpecs[levelNum]

	indexBytesCount := uint64(0)
	nodesBytesCount := uint64(0)

	// For each level, in indexBytes, the first 2 bytes represent a count of NodeSpecs ("groups") that follow
	indexBytesCount += 2

	nodeCountSize := (*nodeIdConfig).StorageBytes()
	for groupIndex := range levelData.Groups {
		group := &(levelData.Groups[groupIndex])
		// In the indexBytes, for this group (nodespec), N bytes specify the number of nodes,
		// and four bytes describe the formatSpec
		indexBytesCount += uint64(nodeCountSize + 4)
		// In the nodesBytes, for this group (nodespec), the number of bytes has already been determined
		nodesBytesCount += uint64(group.Bytes)
	}
	return indexBytesCount, nodesBytesCount
}
