package weddingcake

import (
	"sort"
)

// nodeformat.go is concerned with choosing, optimizing, and specifying a variety of variously
// sized representations of nodes.
// Different node formats are better for representing nodes with different numbers of non-empty slots.

type NodeCountType NodeIdType // Enough room to count the number of nodes (could be bigger than int on 32 bit platforms)
const MaxNodeCount = ^NodeCountType(0)

// StoreConfig holds some parameters of the store being built
type StoreConfig struct {
	HashLength              byte
	ReassuranceBytesCount   byte
	NodeFormatSpecsPerLevel byte
	NodeIdConfig            NByteIdConfig[NodeIdType]
	HashIndexIdConfig       NByteIdConfig[HashIndexIdType]
}

// NodeFormatChoice is a type for holding the choice of node format
type NodeFormatChoice byte

// NodeFormatSlotCapacity is a type for configuring the maximum number of slots a format spec can hold
type NodeFormatSlotCapacity uint16 // Needs to represent up to 256

const (
	NodeFormatTiny NodeFormatChoice = iota
	NodeFormatMedium
	NodeFormatFull
	NodeFormatLeaf
)

// NodeFormatSpec is a way of describing a specific way of representing a node
type NodeFormatSpec struct {
	Format        NodeFormatChoice
	SlotsCapacity NodeFormatSlotCapacity
}

func (sc *StoreConfig) ByteSize(nfs *NodeFormatSpec) int {
	// Cache the dynamic width cost for this store configuration
	idSize := sc.NodeIdConfig.StorageBytes()
	hashIdSize := sc.HashIndexIdConfig.StorageBytes()

	switch nfs.Format {
	case NodeFormatTiny:
		// FormatTiny: 1 byte (hash byte index) + Slots * (1 byte key + Node ID size)
		return 1 + int(nfs.SlotsCapacity)*(1+idSize)
	case NodeFormatMedium:
		// FormatMedium: 1 byte pad + 1 byte index + 32 bytes bitmask + (Slots * Node ID size)
		return 1 + 1 + 32 + (int(nfs.SlotsCapacity) * idSize)
	case NodeFormatFull:
		// FormatFull: 1 byte pad + 1 byte index + (256 * Node ID size)
		return 1 + 1 + (256 * idSize)
	case NodeFormatLeaf:
		// FormatLeaf: Reassurance bytes payload + Hash Id size
		return int(sc.ReassuranceBytesCount) + hashIdSize
	default:
		panic("unknown node format")
	}
}

type NodeFormatGroup struct {
	StartSlotsCount int            // The first slots count value that this group applies to
	EndSlotsCount   int            // One past the last slots count value that this group applies to
	NodesCount      NodeCountType  // The number of nodes that fall within this group, for the tree considered
	Spec            NodeFormatSpec // The node format spec that is currently proposed for this group
	Bytes           BytesCountType // Number of bytes used for all the nodes in this group
}

func (sc *StoreConfig) groupByteSize(nfg *NodeFormatGroup) uint64 {
	return uint64(nfg.NodesCount) * uint64(sc.ByteSize(&nfg.Spec))
}

func (sc *StoreConfig) ProposeNodeFormatForSlotsCount(activeSlots int) NodeFormatSpec {
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

func (sc *StoreConfig) ProposeNodeFormatGroupsForLevelShape(shape LevelShape) []NodeFormatGroup {
	result := make([]NodeFormatGroup, 0, 257)
	for slotsCount := 0; slotsCount <= 256; slotsCount++ {
		if shape.ActiveSlotCountHistogram[slotsCount] > 0 {
			// At least one node needs exactly slotsCount slots
			group := NodeFormatGroup{}
			group.StartSlotsCount = slotsCount
			group.EndSlotsCount = slotsCount + 1
			group.NodesCount = shape.ActiveSlotCountHistogram[slotsCount]
			group.Spec = sc.ProposeNodeFormatForSlotsCount(slotsCount)
			bytes := sc.groupByteSize(&group) // We store this to avoid repeatedly recalculating
			if bytes > uint64(MaxBytesCount) {
				panic("Too many bytes for BytesCountType")
			}
			group.Bytes = BytesCountType(bytes)
			result = append(result, group)
		}
	}
	return result
}

func (sc *StoreConfig) AllowMergeGroups(left *NodeFormatGroup, right *NodeFormatGroup) bool {
	return right.Spec.Format == left.Spec.Format
}

func (sc *StoreConfig) ProposeMergeGroups(left *NodeFormatGroup, right *NodeFormatGroup) NodeFormatGroup {
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
	bytes := sc.groupByteSize(&result)
	if bytes > uint64(MaxBytesCount) {
		panic("Too many bytes for BytesCountType")
	}
	result.Bytes = BytesCountType(bytes)
	return result
}

func (sc *StoreConfig) RefineNodeFormatGroups(groups []NodeFormatGroup) ([]NodeFormatGroup, bool) {
	// Try each neighbouring pair of groups
	// We are looking for the lowest cost merge (counted in bytes)
	lowestCost := MaxBytesCount
	bestProposalLeft := -1
	bestProposedMerge := NodeFormatGroup{}
	for left := 0; left < len(groups)-1; left++ {
		right := left + 1
		if sc.AllowMergeGroups(&groups[left], &groups[right]) {
			leftBytes := groups[left].Bytes
			rightBytes := groups[right].Bytes
			proposedGroup := sc.ProposeMergeGroups(&groups[left], &groups[right])
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

type LevelFormat struct {
	// These are sorted for efficiency, most popular first (but with FormatTiny's at the end which use odd byte numbers)
	Groups []NodeFormatGroup
	// These have the same index as Groups
	NodeIdAllocations []NodeIdAllocation
	// These are indexed by active slot count, and hold indices into the above
	SlotCountToGroup [257]byte
}
type TreeFormat struct {
	// These are indexed by level, with root at index 0
	LevelSpecs []LevelFormat
}
type NodeIdAllocation struct {
	NextAvailableNodeId NodeIdType
	AvailableNodeIds    NodeIdType
}

func (tns *TreeFormat) InitializeNodeIdAllocations() {
	levels := len(tns.LevelSpecs)
	for level := 0; level < levels; level++ {
		// Node Id's are now PER LEVEL; you need a level AND a node id to identify a node
		nodeId := NodeIdType(1) // 0 is reserved with special meaning

		// Note that duplicates are no longer tolerated

		groups := len(tns.LevelSpecs[level].Groups)
		tns.LevelSpecs[level].NodeIdAllocations = make([]NodeIdAllocation, groups)
		for group := 0; group < groups; group++ {
			// These will later allocate us NodeIDs from each group in each level
			nodes := tns.LevelSpecs[level].Groups[group].NodesCount
			tns.LevelSpecs[level].NodeIdAllocations[group] = NodeIdAllocation{
				NextAvailableNodeId: nodeId,
				AvailableNodeIds:    NodeIdType(nodes),
			}
			// These will later tell us, for a given active slot count, which
			// group should allocate us NodeIDs
			start := tns.LevelSpecs[level].Groups[group].StartSlotsCount
			end := tns.LevelSpecs[level].Groups[group].EndSlotsCount
			for activeSlotsCount := start; activeSlotsCount < end; activeSlotsCount++ {
				tns.LevelSpecs[level].SlotCountToGroup[activeSlotsCount] = byte(group)
			}

			if uint64(nodeId)+uint64(nodes) > uint64(MaxNodeCount) {
				panic("Too many nodes for node NodeIdType")
			}
			nodeId += NodeIdType(nodes)
		}
	}
}

func (tns *TreeFormat) AllocateIdAndSpecForNode(level byte, activeSlotsCount int) (NodeIdType, NodeFormatSpec) {
	group := tns.LevelSpecs[level].SlotCountToGroup[activeSlotsCount]

	// Read directly from the underlying slice index
	alloc := tns.LevelSpecs[level].NodeIdAllocations[group]
	if alloc.AvailableNodeIds == 0 {
		panic("Too many nodes for uint16")
	}

	nodeID := alloc.NextAvailableNodeId

	// Update the live data tracking fields directly back inside the slice holder
	tns.LevelSpecs[level].NodeIdAllocations[group].NextAvailableNodeId++
	tns.LevelSpecs[level].NodeIdAllocations[group].AvailableNodeIds--

	nodeSpec := tns.LevelSpecs[level].Groups[group].Spec
	return nodeID, nodeSpec
}

func (sc *StoreConfig) ChooseNodeFormatSpecsForTreeShape(treeShape *TreeShape) *TreeFormat {
	// 4 are required to allow for a single formatTiny, a formatMedium, a formatFull. and a formatLeaf
	// BUT 4 is not sensible as 4 could be achieved without this expensive procedure!
	// A sensible number is perhaps 8 or more
	if sc.NodeFormatSpecsPerLevel < 5 {
		panic("nodeSpecFormatSpecsPerLevel must be at least 5")
	}
	result := TreeFormat{}
	levels := treeShape.NonEmptyLevels
	result.LevelSpecs = make([]LevelFormat, levels)
	for level := byte(0); level < levels; level++ {
		// Initially we propose a separate NodeFormatSpec for each count of slots represented in the level
		nfgs := sc.ProposeNodeFormatGroupsForLevelShape(treeShape.LevelShapes[level])
		// originalCount := len(nfgs)
		originalCost := BytesCountType(0)
		for _, nfg := range nfgs {
			originalCost += nfg.Bytes
		}
		// Then we repeatedly reduce by merging pairs until we reach nodeFormatSpecsPerLevel
		for {
			if len(nfgs) <= int(sc.NodeFormatSpecsPerLevel) {
				break
			}
			proposed, reduced := sc.RefineNodeFormatGroups(nfgs)
			if !reduced {
				panic("Could not reduce node format specs! nodeFormatSpecsPerLevel too low?")
			}
			nfgs = proposed
		}
		cost := BytesCountType(0)
		for _, nfg := range nfgs {
			cost += nfg.Bytes
		}
		// fmt.Printf("Level %d optimization: Reduced nfgs %d -> %d, bytes %d -> %d\n",
		//	level, originalCount, len(nfgs), originalCost, cost)

		// Sort descending by NodesCount (so popular nodeFormatSpecs appear first)
		// Now with FormatTiny forced to the end of the sort (they have odd numbers of bytes, so we don't want them
		// to result in the others being on odd byte boundaries)
		// Sort primarily to push FormatTiny to the end, and secondarily by NodesCount descending
		sort.SliceStable(nfgs, func(i, j int) bool {
			iIsTiny := nfgs[i].Spec.Format == NodeFormatTiny
			jIsTiny := nfgs[j].Spec.Format == NodeFormatTiny

			// Rule 1 (Primary): If one is Tiny and the other isn't, non-Tiny comes first
			if iIsTiny != jIsTiny {
				return jIsTiny // Returns true if j is Tiny (meaning i is non-Tiny, so i comes first)
			}

			// Rule 2 (Secondary): If they are of the same category, bigger NodesCount comes first
			return nfgs[i].NodesCount > nfgs[j].NodesCount
		})
		result.LevelSpecs[level].Groups = nfgs
	}
	return &result
}

func (sc *StoreConfig) DesignTreeFormat(tree *ShallowTree) *TreeFormat {
	treeShape := tree.CountLevelShapes()
	if treeShape.GreatestNodesPerLevel > MaxNodeCount {
		panic("Too many nodes at a single level")
	}
	result := sc.ChooseNodeFormatSpecsForTreeShape(treeShape)
	result.InitializeNodeIdAllocations()
	return result
}
