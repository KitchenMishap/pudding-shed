package indexedhashestree

import "math"

// Every node format maps 256 byte values to either:
// A) A null, meaning no hash exists with this prefix, or
// B) A leaf, giving a presentation index, or
// C) A node ID, for deeper examination into the tree

// There are three node formats, but two of them have a sizing parameter

// A node format along with any sizing parameter, comprises a node spec

// Crucially, a given node spec has a fixed byte size. There are likely to be about 10 node specs.

// Node IDs increase from 1 through ranges of integers, each range capturing nodes of a certain node spec

// The node container specifies the node ID ranges for each sequential node spec.
// It also specifies the byte offset for the first node in each node spec.
// It is therefore possible, given a container and a node ID, to quickly arrive at the following
// with just a very small number of loookups, additions, and multiplications:
// 1) Byte offset for the node
// 2) Format for the node
// 3) Sizing param for the node
// 4) Bytesize of the node
// In short, everything.

type nodeDetails struct {
	byteOffset int
	nodeFormat int
	sizeParam  int
	byteSize   int
}

type nodeContainer struct {
	rootNodeId                 uint16
	nodeIdStartingEachSpec     []uint16
	byteOffsetStartingEachSpec []int32
	bytes                      []byte
}

type nodeSpec struct {
	nodeFormat int
	sizeParam  int
	byteSize   int
}

func newNodeSpec(nodeFormat int, sizeParam int) *nodeSpec {
	result := nodeSpec{}
	result.nodeFormat = nodeFormat
	result.sizeParam = sizeParam
	if nodeFormat == formatTiny {
		result.byteSize = 1              // For the hash byte index 0..31
		result.byteSize += sizeParam * 3 // 3 bytes for a byteHashVal and a lookup
		if result.byteSize >= 35 {
			panic("sizeParam illegal byteSize matches that for a potential formatMedium(1)")
		}
		if result.byteSize == 513 {
			panic("sizeParam illegal, byteSize matches that for formatFull")
		}
	} else if nodeFormat == formatMedium {
		result.byteSize = 1              // For a hash byte index
		result.byteSize += 32            // 256 bits which say whether each possible byteHashVal is represented below
		result.byteSize += sizeParam * 2 // sizeParam (padded) list of lookups for every bit that's a one above
		if result.byteSize == 513 {
			panic("sizeParam illegal, byteSize matches that for formatFull")
		}
	} else if nodeFormat == formatFull {
		result.byteSize = 1        // For a hash byte index
		result.byteSize += 256 * 2 // Lookup for every conceivable value of byteHashVal
	} else {
		panic("Don't know this format")
	}
	return &result
}

type containerParams struct {
	nodeSpecs []*nodeSpec
}

const formatTiny = 1
const formatMedium = 2
const formatFull = 3

// An example config
// This config is manually optimized for a count of approximately 32,768 hashes
func newContainerParamsConfigA() *containerParams {
	result := containerParams{}
	result.nodeSpecs = make([]*nodeSpec, 0)
	result.nodeSpecs = append(result.nodeSpecs, newNodeSpec(formatTiny, 2)) // Important - 86% of nodes!
	result.nodeSpecs = append(result.nodeSpecs, newNodeSpec(formatTiny, 3)) // 9% of nodes
	result.nodeSpecs = append(result.nodeSpecs, newNodeSpec(formatTiny, 4)) // 0.5% of nodes
	result.nodeSpecs = append(result.nodeSpecs, newNodeSpec(formatTiny, 5)) // A tiny five is vaguely worthwhile
	// For 50_000 hashes (nearly filling our 16 bit budget), there's a big gap
	// with virtually no nodes having between 5 and 127 slots. Let's not waste time there,
	// and instead give a finer grained size choice between 120 and 165 slots which is fairly busy.
	// For 32_768 hashes (a nice number), we find a gap between
	// 5 and 90 slots per node. Then we have a fair number of cases between 90 and 130 slots,
	// with a peak between 105 and 120. We therefore concentrate our range of available sizes more finely there.
	// (There's about 100-86-9-0.5 = 4.5% of nodes spread among the following):
	result.nodeSpecs = append(result.nodeSpecs, newNodeSpec(formatMedium, 90))
	result.nodeSpecs = append(result.nodeSpecs, newNodeSpec(formatMedium, 100))
	result.nodeSpecs = append(result.nodeSpecs, newNodeSpec(formatMedium, 105)) // 103,104,105 typically make up 0.8% of nodes (out of 4.5%)
	result.nodeSpecs = append(result.nodeSpecs, newNodeSpec(formatMedium, 108)) // 106,107,108 typically make up 0.9% of nodes (out of 4.5%)
	result.nodeSpecs = append(result.nodeSpecs, newNodeSpec(formatMedium, 112)) // 109,110,111 typically make up 0.8% of nodes (out of 4.5%)
	result.nodeSpecs = append(result.nodeSpecs, newNodeSpec(formatMedium, 118))
	result.nodeSpecs = append(result.nodeSpecs, newNodeSpec(formatMedium, 126))
	result.nodeSpecs = append(result.nodeSpecs, newNodeSpec(formatMedium, 130))
	result.nodeSpecs = append(result.nodeSpecs, newNodeSpec(formatFull, 256))
	return &result
}

// A nodeSpec is suitable if it gives the smallest byte size of those that can support this number of slots
func (cp *containerParams) nodeSpecSuitableFor(slots int) *nodeSpec {
	bestByteSize := math.MaxInt
	bestNodeSpec := (*nodeSpec)(nil)
	for _, spec := range cp.nodeSpecs {
		if spec.sizeParam >= slots {
			byteSize := spec.byteSize
			if byteSize < bestByteSize {
				bestByteSize = byteSize
				bestNodeSpec = spec
			}
		}
	}
	if bestNodeSpec == nil {
		panic("Not found")
	}
	return bestNodeSpec
}

// A pair of structures to aid the mapping from a "shallowTree" set of indexed nodes, to a "multiFixedNodeSize" tree
// of nodes having consequently different indices
type nodeReconfigShallow struct {
	shallowTreeIndex uint16
	aShallowTreeNode *shallowTreeNode
	fixedNodeSpec    *nodeSpec // We will be sorting on the bytesize field of this
}

// The result will need to be sorted by nodeSpec afterwards
func (cp *containerParams) sliceOfNodesFromShallowTree(stc *shallowTreeContainer) []*nodeReconfigShallow {
	result := make([]*nodeReconfigShallow, len(stc.nodesPool))
	for i := range stc.nodesPool {
		result[i] = &nodeReconfigShallow{}
		result[i].shallowTreeIndex = uint16(i)
		result[i].aShallowTreeNode = &stc.nodesPool[i]
		lookupsCount := 0
		for j := 0; j <= 255; j++ {
			if stc.nodesPool[i].lookups[j] != 0 {
				lookupsCount++
			}
		}
		result[i].fixedNodeSpec = cp.nodeSpecSuitableFor(lookupsCount)
	}
	return result[:]
}

// sortedNodeReconfig parameter must be already sorted by nodeSpec
func (cp *containerParams) serializeMultiFixedSizedNodeTree(sortedNodeReconfig []*nodeReconfigShallow,
	shallowContainer *shallowTreeContainer) *nodeContainer {
	// First, a default object
	result := nodeContainer{}
	result.nodeIdStartingEachSpec = make([]uint16, len(cp.nodeSpecs))
	result.byteOffsetStartingEachSpec = make([]int32, len(cp.nodeSpecs))
	result.bytes = make([]byte, 0, 120*1024) // 120 Kb is about right we find
	// Then the offsets to each batch of nodes having identical nodeSpecs
	byteOffset := int32(0)
	nextReconfig := uint16(0)
	for i := range cp.nodeSpecs {
		result.nodeIdStartingEachSpec[i] = nextReconfig
		result.byteOffsetStartingEachSpec[i] = byteOffset
		for nextReconfig < uint16(len(sortedNodeReconfig)) && *sortedNodeReconfig[nextReconfig].fixedNodeSpec == *cp.nodeSpecs[i] {
			// while we're on this nodeSpec
			byteOffset += int32(cp.nodeSpecs[i].byteSize) // Move forward by the size of the node
			nextReconfig++
		}
	}

	// Gather a mapping from old shallow node indices to new node indices (which are sorted by nodeSpec.)
	// Old node index representations are zero based, with 0 indicating the root node which
	// is never referred to by other nodes (and other nodes reserve zero to mean something else.)
	// New node index representations are one based, with the root rarely being the first item.
	nodeCount := len(sortedNodeReconfig)
	mapNewIndexFromOld := make([]uint16, nodeCount)
	for newNodeIndex := 0; newNodeIndex < nodeCount; newNodeIndex++ {
		oldNodeIndex := sortedNodeReconfig[newNodeIndex].shallowTreeIndex
		mapNewIndexFromOld[oldNodeIndex] = uint16(newNodeIndex + 1)
	}
	// Store the root nodeId
	result.rootNodeId = mapNewIndexFromOld[0]

	// Then, finally, the actual bytes
	nextReconfig = uint16(0)
	for i := range cp.nodeSpecs {
		format := cp.nodeSpecs[i].nodeFormat
		size := cp.nodeSpecs[i].sizeParam
		if int32(len(result.bytes)) != result.byteOffsetStartingEachSpec[i] {
			panic("Mismatch between bytes written and predicted")
		}
		if nextReconfig != result.nodeIdStartingEachSpec[i] {
			panic("Mismatch between nodes written and predicted")
		}
		// While we're on this nodeSpec
		// Do the bytes of each node with this nodespec
		for nextReconfig < uint16(len(sortedNodeReconfig)) && *sortedNodeReconfig[nextReconfig].fixedNodeSpec == *cp.nodeSpecs[i] {
			// This is the shallowNode object we're "copying" from
			shallowNode := sortedNodeReconfig[nextReconfig].aShallowTreeNode
			// Here we step through the 256 lookups and condense them (throw away the nils) and
			// update those that refer to node indices with their new node indices
			type hashByteValAndLookup struct {
				byteVal       byte
				updatedLookup uint16
			}
			updatedLookupsNonZero := make([]hashByteValAndLookup, 0, 256)
			updatedLookupsWithZero := [256]hashByteValAndLookup{}
			for v := 0; v < 256; v++ {
				lookup := shallowNode.lookups[v]
				if lookup == 0 {
					updatedLookupsWithZero[v] = hashByteValAndLookup{byteVal: byte(v), updatedLookup: 0}
				} else if lookup <= shallowContainer.presentationsCount {
					// It's a presentation index, keep it as it is
					updatedLookupsWithZero[v] = hashByteValAndLookup{byteVal: byte(v), updatedLookup: lookup}
					updatedLookupsNonZero = append(updatedLookupsNonZero, updatedLookupsWithZero[v])
				} else {
					// lookup refers to a shallowNode index, replace it with the corresponding new node index
					// lookup is in the shallowTree format, in which the oldIndex is 65536 - lookup
					oldIndex := 65536 - int(lookup)
					newIndex := mapNewIndexFromOld[oldIndex] // These start at one
					// Once again, this time for newIndex, it is represented in memory and file as 65536 - newIndex
					updatedLookup := uint16(65536 - int(newIndex))
					updatedLookupsWithZero[v] = hashByteValAndLookup{byteVal: byte(v), updatedLookup: updatedLookup}
					updatedLookupsNonZero = append(updatedLookupsNonZero, updatedLookupsWithZero[v])
				}
			}

			// Here we go through different serialization for each possible format
			if format == formatTiny {
				// A formatTiny is {1 byte: hashByteIndex}...
				result.bytes = append(result.bytes, shallowNode.hashByteIndex)

				// plus size (padded) lots of...
				for item := 0; item < size; item++ {
					if item < len(updatedLookupsNonZero) {
						// ...{1 byte: hashByteVal, 2 bytes: lookup}
						result.bytes = append(result.bytes, updatedLookupsNonZero[item].byteVal)
						updatedLookup := updatedLookupsNonZero[item].updatedLookup
						result.bytes = append(result.bytes, byte(updatedLookup&0xFF)) // LittleEndian
						result.bytes = append(result.bytes, byte((updatedLookup&0xFF00)>>8))
					} else {
						padding := [3]byte{}
						result.bytes = append(result.bytes, padding[:]...)
					}
				} // for item in node

			} else if format == formatMedium {
				// A formatMedium is {1 byte: hashByteIndex}...
				result.bytes = append(result.bytes, shallowNode.hashByteIndex)

				// Followed by 256 bits flagging whether each hash byte value is represented below...
				bitBytes := [32]byte{}
				for item := 0; item < len(updatedLookupsNonZero); item++ {
					byteVal := updatedLookupsNonZero[item].byteVal
					bitIndex := byteVal & 0x07         // 3 bit, between 0 and 7
					byteIndex := (byteVal & 0xF8) >> 3 // 5 bit, between 0 and 31
					bitBytes[byteIndex] |= 1 << bitIndex
				}
				result.bytes = append(result.bytes, bitBytes[:]...)

				// Followed by size (padded) lots of {2 bytes: lookup}
				for item := 0; item < size; item++ {
					if item < len(updatedLookupsNonZero) {
						updatedLookup := updatedLookupsNonZero[item].updatedLookup
						result.bytes = append(result.bytes, byte(updatedLookup&0xFF)) // LittleEndian
						result.bytes = append(result.bytes, byte((updatedLookup&0xFF00)>>8))
					} else {
						padding := [2]byte{}
						result.bytes = append(result.bytes, padding[:]...) // Two bytes padding instead
					}
				} // for item in node

			} else if format == formatFull {
				// A formatFull is always 513 bytes
				bytes := [513]byte{}
				// A formatMedium is {1 byte: hashByteIndex}...
				bytes[0] = shallowNode.hashByteIndex

				for v := 0; v < 256; v++ {
					updatedLookup := updatedLookupsWithZero[v].updatedLookup
					bytes[1+2*v] = byte(updatedLookup & 0xFF)
					bytes[2+2*v] = byte((updatedLookup & 0xFF00) >> 8)
				}
				result.bytes = append(result.bytes, bytes[:]...)
			}

			byteOffset += int32(cp.nodeSpecs[i].byteSize) // Move forward by the size of the node
			nextReconfig++
		} // for nextReconfig is current nodeSpec
	} // for i = nodeSpecs
	return &result
}
