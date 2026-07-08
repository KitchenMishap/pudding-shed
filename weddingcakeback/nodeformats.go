package weddingcakeback

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

func (nfs *NodeFormatSpec) ByteSize(config *CakeConfig, tierIndex byte) int {
	// Cache the dynamic width cost for this store configuration
	idSize := config.TierBelowConfigs[tierIndex].NodeIdConfig.StorageBytes()
	hashIdSize := config.TierBelowConfigs[tierIndex].HashIndexIdConfig.StorageBytes()

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
		return int(config.TierBelowConfigs[tierIndex].ReassuranceBytesCount) + hashIdSize
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

func (nfg *NodeFormatGroup) groupByteSize(config *CakeConfig, tierIndex byte) uint64 {
	return uint64(nfg.NodesCount) * uint64(nfg.Spec.ByteSize(config, tierIndex))
}
