package indexedhashestree

import (
	"math"
	"sort"
)

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

type nodeContainer struct {
	firstPresentationIndex     int64
	presentationsCount         uint16
	rootNodeId                 uint16
	nodeIdStartingEachSpec     []uint16 // These start at 1
	byteOffsetStartingEachSpec []int32
	bytes                      []byte
	nodeLevelBytes             [][]byte // Each outer array represents a level of the tree. Slices into the above slice
}

type nodeSpec struct {
	nodeLevel  int
	nodeFormat int
	sizeParam  int
	byteSize   int
}

func newNodeSpec(nodeLevel int, nodeFormat int, sizeParam int) *nodeSpec {
	result := nodeSpec{}
	result.nodeLevel = nodeLevel
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
	} else if nodeFormat == formatDuplicate {
		result.byteSize = 2 // The smallest of the presentionIndex's that were presented with this hash
	} else {
		panic("Don't know this format")
	}
	return &result
}

type containerParams struct {
	nodeSpecs []*nodeSpec
}

// Less reports whether specI should be ordered before specJ.
// This establishes the canonical sorting rule for the entire package.
// We don't depend on a particular sort order here, but it does always need to be the same!
func (cp *containerParams) Less(specI, specJ *nodeSpec) bool {
	if specI.nodeLevel != specJ.nodeLevel {
		return specI.nodeLevel < specJ.nodeLevel
	}
	if specI.byteSize != specJ.byteSize {
		return specI.byteSize < specJ.byteSize
	}
	if specI.nodeFormat != specJ.nodeFormat {
		return specI.nodeFormat < specJ.nodeFormat
	}
	return specI.sizeParam < specJ.sizeParam
}

const formatTiny = 1
const formatMedium = 2
const formatFull = 3
const formatDuplicate = 4

// An example config
// This config is manually optimized for a count of approximately 32,768 hashes
func newContainerParamsConfigA() *containerParams {
	result := containerParams{}
	result.nodeSpecs = make([]*nodeSpec, 0)
	for nodeLevel := 0; nodeLevel < 32; nodeLevel++ {
		result.nodeSpecs = append(result.nodeSpecs, newNodeSpec(nodeLevel, formatTiny, 2)) // Important - 86% of nodes!
		result.nodeSpecs = append(result.nodeSpecs, newNodeSpec(nodeLevel, formatTiny, 3)) // 9% of nodes
		result.nodeSpecs = append(result.nodeSpecs, newNodeSpec(nodeLevel, formatTiny, 4)) // 0.5% of nodes
		result.nodeSpecs = append(result.nodeSpecs, newNodeSpec(nodeLevel, formatTiny, 5)) // A tiny five is vaguely worthwhile
		// For 50_000 hashes (nearly filling our 16 bit budget), there's a big gap
		// with virtually no nodes having between 5 and 127 slots. Let's not waste time there,
		// and instead give a finer grained size choice between 120 and 165 slots which is fairly busy.
		// For 32_768 hashes (a nice number), we find a gap between
		// 5 and 90 slots per node. Then we have a fair number of cases between 90 and 130 slots,
		// with a peak between 105 and 120. We therefore concentrate our range of available sizes more finely there.
		// (There's about 100-86-9-0.5 = 4.5% of nodes spread among the following):
		result.nodeSpecs = append(result.nodeSpecs, newNodeSpec(nodeLevel, formatMedium, 90))
		result.nodeSpecs = append(result.nodeSpecs, newNodeSpec(nodeLevel, formatMedium, 100))
		result.nodeSpecs = append(result.nodeSpecs, newNodeSpec(nodeLevel, formatMedium, 105)) // 103,104,105 typically make up 0.8% of nodes (out of 4.5%)
		result.nodeSpecs = append(result.nodeSpecs, newNodeSpec(nodeLevel, formatMedium, 108)) // 106,107,108 typically make up 0.9% of nodes (out of 4.5%)
		result.nodeSpecs = append(result.nodeSpecs, newNodeSpec(nodeLevel, formatMedium, 112)) // 109,110,111 typically make up 0.8% of nodes (out of 4.5%)
		result.nodeSpecs = append(result.nodeSpecs, newNodeSpec(nodeLevel, formatMedium, 118))
		result.nodeSpecs = append(result.nodeSpecs, newNodeSpec(nodeLevel, formatMedium, 126))
		result.nodeSpecs = append(result.nodeSpecs, newNodeSpec(nodeLevel, formatMedium, 130))
		result.nodeSpecs = append(result.nodeSpecs, newNodeSpec(nodeLevel, formatFull, 256))
		result.nodeSpecs = append(result.nodeSpecs, newNodeSpec(nodeLevel, formatDuplicate, 1))
	}
	// Automatically sort the specs using our canonical rule
	sort.Slice(result.nodeSpecs, func(i, j int) bool { return result.Less(result.nodeSpecs[i], result.nodeSpecs[j]) })

	return &result
}

// A nodeSpec is suitable if it gives the smallest byte size of those that can support this number of slots
// A formatDuplicate nodeSpec is specifically treated in a separate function below
func (cp *containerParams) nodeSpecSuitableFor(nodeLevel int, slots int) *nodeSpec {
	bestByteSize := math.MaxInt
	bestNodeSpec := (*nodeSpec)(nil)
	for _, spec := range cp.nodeSpecs {
		if spec.sizeParam >= slots && spec.nodeFormat != formatDuplicate && spec.nodeLevel == nodeLevel {
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
func (cp *containerParams) nodeSpecSuitableForDuplicate(nodeLevel int) *nodeSpec {
	bestNodeSpec := (*nodeSpec)(nil)
	for _, spec := range cp.nodeSpecs {
		if spec.nodeFormat == formatDuplicate && spec.nodeLevel == nodeLevel {
			bestNodeSpec = spec
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
	fixedNodeSpec    *nodeSpec // We will be sorting on this
}

// The result will need to be sorted by nodeSpec afterwards
func (cp *containerParams) sliceOfNodesFromShallowTree(stc *shallowTreeContainer) []*nodeReconfigShallow {
	result := make([]*nodeReconfigShallow, len(stc.nodesPool))
	for i := range stc.nodesPool {
		result[i] = &nodeReconfigShallow{}
		result[i].shallowTreeIndex = uint16(i)
		result[i].aShallowTreeNode = &stc.nodesPool[i]
		lookupsCount := 0
		isDuplicate := stc.nodesPool[i].hashByteIndex == 32
		if isDuplicate {
			result[i].fixedNodeSpec = cp.nodeSpecSuitableForDuplicate(stc.nodesPool[i].nodeLevel)
		} else {
			for j := 0; j <= 255; j++ {
				if stc.nodesPool[i].lookups[j] != 0 {
					lookupsCount++
				}
			}
			result[i].fixedNodeSpec = cp.nodeSpecSuitableFor(stc.nodesPool[i].nodeLevel, lookupsCount)
		}
	}
	return result[:]
}

// sortedNodeReconfig parameter must be already sorted by nodeSpec
func (cp *containerParams) serializeMultiFixedSizedNodeTree(sortedNodeReconfig []*nodeReconfigShallow,
	shallowContainer *shallowTreeContainer) *nodeContainer {
	// First, a default object
	result := nodeContainer{}
	result.presentationsCount = shallowContainer.presentationsCount
	result.firstPresentationIndex = shallowContainer.firstPresentationIndex
	result.nodeIdStartingEachSpec = make([]uint16, len(cp.nodeSpecs))
	result.byteOffsetStartingEachSpec = make([]int32, len(cp.nodeSpecs))
	result.bytes = make([]byte, 0, 120*1024) // 120 Kb is about right we find
	// Then the offsets to each batch of nodes having identical nodeSpecs
	byteOffset := int32(0)
	nextReconfig := uint16(0)
	levelByteOffsets := make([]int32, 33)
	currentLevel := 0
	for i := range cp.nodeSpecs {
		result.nodeIdStartingEachSpec[i] = nextReconfig + 1
		result.byteOffsetStartingEachSpec[i] = byteOffset

		// Smoothly fill intermediate skipped levels if any exist
		if nextReconfig < uint16(len(sortedNodeReconfig)) {
			nodeLevel := sortedNodeReconfig[nextReconfig].fixedNodeSpec.nodeLevel
			for currentLevel < nodeLevel {
				currentLevel++
				levelByteOffsets[currentLevel] = byteOffset
			}
		}

		for nextReconfig < uint16(len(sortedNodeReconfig)) && *sortedNodeReconfig[nextReconfig].fixedNodeSpec == *cp.nodeSpecs[i] {
			byteOffset += int32(cp.nodeSpecs[i].byteSize)
			nextReconfig++
		}
	}

	// Cap off all remaining trailing unused levels
	for lev := currentLevel + 1; lev <= 32; lev++ {
		levelByteOffsets[lev] = byteOffset
	}
	result.bytes = make([]byte, byteOffset) // Allocate bytes for all the levels
	result.nodeLevelBytes = make([][]byte, 32)
	for level := 0; level <= 31; level++ {
		result.nodeLevelBytes[level] = result.bytes[levelByteOffsets[level]:levelByteOffsets[level+1]]
	}
	result.bytes = result.bytes[:0] // Empty it (keep capacity) so we can conveniently append to it

	// Gather a mapping from old shallow node indices to new node indices (which are sorted by nodeSpec.)
	// Old node index representations are zero based, with 0 indicating the root node which
	// is never referred to by other nodes (and other nodes reserve zero to mean something else.)
	// New node index representations are one based, with the root rarely being the first item.
	nodeCount := len(sortedNodeReconfig)
	mapNewIdFromOld := make([]uint16, nodeCount)
	for newNodeIndex := 0; newNodeIndex < nodeCount; newNodeIndex++ {
		oldNodeIndex := sortedNodeReconfig[newNodeIndex].shallowTreeIndex
		mapNewIdFromOld[oldNodeIndex] = uint16(newNodeIndex + 1)
	}
	// Store the root nodeId
	result.rootNodeId = mapNewIdFromOld[0]

	// Then, finally, the actual bytes
	nextReconfig = uint16(0)
	for i := range cp.nodeSpecs {
		format := cp.nodeSpecs[i].nodeFormat
		size := cp.nodeSpecs[i].sizeParam
		if int32(len(result.bytes)) != result.byteOffsetStartingEachSpec[i] {
			panic("Mismatch between bytes written and predicted")
		}
		if nextReconfig+1 != result.nodeIdStartingEachSpec[i] {
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
					newId := mapNewIdFromOld[oldIndex] // These start at one
					// Once again, this time for newIndex, it is represented in memory and file as 65536 - newIndex
					updatedLookup := uint16(65536 - int(newId))
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
				// A formatFull is {1 byte: hashByteIndex}...
				bytes[0] = shallowNode.hashByteIndex

				// ... followed by 256 x {2 bytes: lookup}
				for v := 0; v < 256; v++ {
					updatedLookup := updatedLookupsWithZero[v].updatedLookup
					bytes[1+2*v] = byte(updatedLookup & 0xFF)
					bytes[2+2*v] = byte((updatedLookup & 0xFF00) >> 8)
				}
				result.bytes = append(result.bytes, bytes[:]...)
			} else if format == formatDuplicate {
				// A formatDuplicate is simply a presentation index in 2 bytes
				// It corresponds to a shallowTree node for which only index 0 is populated
				updatedLookup := updatedLookupsWithZero[0].updatedLookup
				bytes := [2]byte{}
				bytes[0] = byte(updatedLookup & 0xFF) // LittleEndian
				bytes[1] = byte((updatedLookup & 0xFF00) >> 8)
				result.bytes = append(result.bytes, bytes[:]...)
			} else {
				panic("format code not recognized")
			}

			nextReconfig++
		} // for nextReconfig is current nodeSpec
	} // for i = nodeSpecs
	return &result
}

// A returned -1 means "not currently loaded". Presumably the batch has been truncated to a certain number of levels.
func (nc *nodeContainer) nodeIdToByteIndex(nodeId uint16, params *containerParams) (int32, *nodeSpec) {
	nodeSpecDetailsPrev := params.nodeSpecs[0]
	firstNodeIdPrev := nc.nodeIdStartingEachSpec[0]
	firstBytePrev := nc.byteOffsetStartingEachSpec[0]
	// The above give us for nodeSpec[0],
	// We start below at nodeSpec[1]
	for i := 1; i < len(params.nodeSpecs); i++ {
		firstNodeId := nc.nodeIdStartingEachSpec[i]
		if nodeId < firstNodeId {
			// Calculations are based on the previous nodeSpec (the one containing nodeId)
			sinceStartOfSpec := int32(nodeId - firstNodeIdPrev)
			bytesOffset := firstBytePrev + sinceStartOfSpec*int32(nodeSpecDetailsPrev.byteSize)
			if bytesOffset >= int32(len(nc.bytes)) {
				return -1, nodeSpecDetailsPrev
			}
			return bytesOffset, nodeSpecDetailsPrev
		}
		nodeSpecDetailsPrev = params.nodeSpecs[i]
		firstNodeIdPrev = firstNodeId
		firstBytePrev = nc.byteOffsetStartingEachSpec[i]
	}
	// nodeId must be in the section for the last nodeSpec
	// Calculations are based on the previous nodeSpec (the final one)
	sinceStartOfSpec := int32(nodeId - firstNodeIdPrev)
	bytesOffset := firstBytePrev + sinceStartOfSpec*int32(nodeSpecDetailsPrev.byteSize)
	if bytesOffset >= int32(len(nc.bytes)) {
		return -1, nodeSpecDetailsPrev
	}
	return bytesOffset, nodeSpecDetailsPrev
}

// "index>=0, false, byteBits" is returned for a hash which is unique within this batch. byteBits has 1 for each byte examined.
// "index>=0, true, byteBits" is returned for a hash which is duplicated within this batch. byteBits has 1 for each byte examined.
// "-1, false, byteBits" is returned for a hash that is not present in this batch. byteBits has 1 for each byte examined.
// "-2, false, byteBits" is returned for "hash is present but insufficient levels are loaded". byteBits has 1 for each byte examined.
func (nc *nodeContainer) lookupHash(hash [32]byte, params *containerParams) (int64, bool, uint32) {
	byteBits := uint32(0) // A "1" bit for every byte that is examined
	numHashes := nc.presentationsCount
	if numHashes == 0 {
		return -1, false, byteBits
	}

	nextNodeId := nc.rootNodeId
	for {
		nodeByteOffset, nodeSpecParams := nc.nodeIdToByteIndex(nextNodeId, params)
		if nodeByteOffset == -1 {
			return -2, false, byteBits
		}
		// The hashByteIndex is present for all of formatTiny, formatMedium, formatFull
		hashByteIndex := byte(0)
		if nodeSpecParams.nodeFormat == formatTiny || nodeSpecParams.nodeFormat == formatMedium || nodeSpecParams.nodeFormat == formatFull {
			hashByteIndex = nc.bytes[nodeByteOffset]
		}

		// Get the 256 possible lookups, based on the nodeSpecParams for this node
		lookups := [256]uint16{}
		if nodeSpecParams.nodeFormat == formatTiny {
			// The following is repeated sizeParam times
			for item := range nodeSpecParams.sizeParam {
				// There are three bytes, the hashByteVal, and a two byte lookup
				hashByteVal := nc.bytes[nodeByteOffset+1+int32(item)*3+0]
				lookup0 := nc.bytes[nodeByteOffset+1+int32(item)*3+1]
				lookup1 := nc.bytes[nodeByteOffset+1+int32(item)*3+2]
				lookup := uint16(lookup0) + uint16(lookup1)<<8
				lookups[hashByteVal] = lookup
			}
		} else if nodeSpecParams.nodeFormat == formatMedium {
			// There are 256 bits of flags, each '1' indicating there is a lookup to be read
			lookupOffset := nodeByteOffset + 1 + 32
			lookupIndex := 0
			for flagByteIndex := range 32 {
				flagByte := nc.bytes[nodeByteOffset+1+int32(flagByteIndex)]
				if flagByte == 0 {
					lookupIndex += 8
				} else {
					for mask := 1; mask <= 128; mask <<= 1 {
						if (flagByte & byte(mask)) != 0 {
							// Read a lookup
							lookup0 := nc.bytes[lookupOffset]
							lookup1 := nc.bytes[lookupOffset+1]
							lookupOffset += 2
							lookup := uint16(lookup0) + uint16(lookup1)<<8
							lookups[lookupIndex] = lookup
						}
						lookupIndex++
					}
				}
			}
		} else if nodeSpecParams.nodeFormat == formatFull {
			// There are just 256 16 bit lookups to be read
			for hashByteVal := range 256 {
				lookup0 := nc.bytes[nodeByteOffset+1+int32(hashByteVal)*2]
				lookup1 := nc.bytes[nodeByteOffset+2+int32(hashByteVal)*2]
				lookup := uint16(lookup0) + uint16(lookup1)<<8
				lookups[hashByteVal] = lookup
			}
		} else if nodeSpecParams.nodeFormat == formatDuplicate {
			// There is only one item
			lookup0 := nc.bytes[nodeByteOffset+0]
			lookup1 := nc.bytes[nodeByteOffset+1]
			encodedLookup := uint16(lookup0) + uint16(lookup1)<<8
			// Subtract 1 because internally, 1 means the first presentation (0 is reserved meaning something else)
			return nc.firstPresentationIndex + int64(encodedLookup-1), true, byteBits
		} else {
			panic("Unrecognized node format")
		}

		var encodedLookup uint16
		hashByte := hash[hashByteIndex]
		byteBits |= 1 << hashByteIndex
		encodedLookup = lookups[hashByte]
		if encodedLookup == 0 {
			// Hash not found
			return -1, false, byteBits
		} else if encodedLookup <= nc.presentationsCount {
			// Presentation index found
			// Subtract 1 because internally, 1 means the first presentation (0 is reserved meaning something else)
			return nc.firstPresentationIndex + int64(encodedLookup-1), false, byteBits
		}
		// Must be a link to a node
		linkIndex := uint16(65536 - uint32(encodedLookup))
		nextNodeId = linkIndex
		if byteBits == 0xFFFFFFFF {
			return -1, true, byteBits // ToDo: What does this mean? Should we panic?
		} // A link too far
	}
}
