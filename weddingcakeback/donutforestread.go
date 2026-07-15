package weddingcakeback

import (
	"encoding/binary"
	"math/bits"

	"github.com/edsrzf/mmap-go"
)

type DonutForestInfo struct {
	FirstGlobalPresentationIndex GlobalPiType
	Levels                       []DonutForestLevelSlices
}
type DonutForestLevelSlices struct {
	// These are slices into the mmap'ed files
	IndexBytes mmap.MMap
	NodesBytes mmap.MMap
}

// IndexBytes (for a level) is formatted as follows
// WHERE: N = StoreConfig.NodeIdConfig.StorageBytes()
// ----------------------------------
// FormatSpecCount		- 2 bytes
// FormatSpecNodeCount	- N bytes	\__ This pair of fields are repeated FormatSpecCount times
// FormatSpecBytes		- 4 bytes	/

// The first node of those described by the first {FormatSpecNodeCount,FormatSpecBytes} pair
// has node id = 1.

// chunkNode: For processing (eg lookup) a node is temporarily internally represented by a chunkNode
type donutForestNode struct {
	formatSpecBytes uint32
	nodeBytes       []byte
}

// ExtractNode extracts to an existing chunkNode to avoid a busy heap
// Returns true for success
func (dfls *DonutForestLevelSlices) ExtractNode(nodeId NodeIdType, target *donutForestNode,
	nodeIdConfig *NByteIdConfig[NodeIdType]) bool {

	currentSpecNodeId := NodeIdType(1)
	currentSpecNodeByteOffset := BytesCountType(0)
	byteCount := BytesCountType(0)
	formatSpecCount := binary.LittleEndian.Uint16(dfls.IndexBytes[byteCount : byteCount+2])
	byteCount += 2

	// The results of the following loop
	found := false
	var formatSpecBytes uint32
	var nodeByteOffset BytesCountType
	var nodeByteSize uint16

	nodeCountSize := (*nodeIdConfig).StorageBytes()
	for fs := uint16(0); fs < formatSpecCount; fs++ {
		formatSpecNodeCount := NodeCountType((*nodeIdConfig).ReadID(dfls.IndexBytes[byteCount : byteCount+BytesCountType(nodeCountSize)]))
		byteCount += BytesCountType(nodeCountSize)
		formatSpecBytes = binary.LittleEndian.Uint32(dfls.IndexBytes[byteCount : byteCount+4])
		byteCount += 4
		nodeByteSize = uint16(formatSpecBytes & 0xFFFF) // Held in the bottom 16 bits
		if currentSpecNodeId+NodeIdType(formatSpecNodeCount) > nodeId {
			// Found it in the current spec
			found = true
			nodeByteOffset = currentSpecNodeByteOffset + BytesCountType(nodeId-currentSpecNodeId)*BytesCountType(nodeByteSize)
			break
		}
		currentSpecNodeId += NodeIdType(formatSpecNodeCount)
		currentSpecNodeByteOffset += BytesCountType(formatSpecNodeCount) * BytesCountType(nodeByteSize)
	}
	if !found {
		return false
	}
	target.formatSpecBytes = formatSpecBytes
	target.nodeBytes = dfls.NodesBytes[nodeByteOffset : nodeByteOffset+BytesCountType(nodeByteSize)]
	return true
}

// Here are various functions to ask various questions of a node.
// donutForestNode.nodeBytes will be interpreted in different ways depending on donutForestNode.formatSpecBytes.

// detailsIfLeaf returns (true, reassuranceBytes, presentationIndex) if cn is a leaf
// returns (false, nil, 0) otherwise
// reassuranceBytesCount is a configuration pertaining to the overall store
func (dfn *donutForestNode) detailsIfLeaf(reassuranceBytesCount byte,
	hashIndexIdConfig *NByteIdConfig[HashIndexIdType]) (bool, []byte, HashIndexIdType) {

	// A leaf node is identified as two MSB zero bytes followed by an appropriate two LSB bytes bytesCount
	hashIndexIdSize := (*hashIndexIdConfig).StorageBytes()
	if dfn.formatSpecBytes != uint32(reassuranceBytesCount)+uint32(hashIndexIdSize) {
		return false, nil, 0
	}
	// cn.nodeBytes interpreted as FormatLeaf
	reassuranceBytes := dfn.nodeBytes[0:reassuranceBytesCount]
	hashIndexId := (*hashIndexIdConfig).ReadID(dfn.nodeBytes[reassuranceBytesCount : int(reassuranceBytesCount)+hashIndexIdSize])
	return true, reassuranceBytes, hashIndexId
}

// hashByteIndexToExamine() should only be called if detailsIfLeaf() has already returned false
// It returns (index, 0, 0) for a FormatFull,
// or (index, slots, 0) for a FormatMedium,
// or (index, 0, slots) for a FormatTiny
func (dfn *donutForestNode) hashByteIndexToExamine(nodeIdConfig *NByteIdConfig[NodeIdType]) (byte, byte, byte) {
	// Is it a FormatFull?
	nodeIdSize := (*nodeIdConfig).StorageBytes()
	if dfn.formatSpecBytes == uint32(1+1+256*nodeIdSize) {
		if dfn.nodeBytes[0] != 0xAA {
			panic("Expected an 0xAA byte for FormatFull padding")
		}
		return dfn.nodeBytes[1], 0, 0
	}
	// Is it a FormatMedium?
	mediumSlots := byte((dfn.formatSpecBytes & 0x00FF0000) >> 16)
	if mediumSlots > 0 {
		if dfn.nodeBytes[0] != 0x55 {
			panic("Expected an 0x55 byte for FormatMedium padding")
		}
		return dfn.nodeBytes[1], mediumSlots, 0
	}
	// Is it a FormatTiny?
	tinySlots := byte((dfn.formatSpecBytes & 0xFF000000) >> 24)
	if tinySlots > 0 {
		// No padding in this case, because FormatTiny is often a small odd number of bytes, and we want
		// them tightly packed
		return dfn.nodeBytes[0], 0, tinySlots
	}
	panic("Unrecognised format spec")
}

// nextLevelNodeId should only be called if detailsIfLeaf() returned false and you have extracted the
// hash byte specified by hashByteIndexToExamine().
// It returns 0 for a "dead end" "empty slot" "hash not found" scenario
func (dfn *donutForestNode) nextLevelNodeId(examinedByte byte, mediumSlots byte, tinySlots byte,
	nodeIdConfig *NByteIdConfig[NodeIdType]) NodeIdType {

	// Is it a FormatFull?
	nodeIdSize := (*nodeIdConfig).StorageBytes()
	if dfn.formatSpecBytes == uint32(1+1+256*nodeIdSize) {
		//fmt.Println("Visiting a FormatFull node")
		byteIndex := 1 + 1 + int(examinedByte)*nodeIdSize
		return (*nodeIdConfig).ReadID(dfn.nodeBytes[byteIndex : byteIndex+nodeIdSize])
	}
	// Is it a FormatMedium?
	if mediumSlots > 0 {
		//fmt.Println("Visiting a FormatMedium node")
		// There are 256 bits within cn.nodeBytes which tell us (if each is a '1') which slots are represented
		// in the NodeIdType's which follow.
		// The first question is, for the value of examinedByte, is the corresponding bit a zero?
		// 1. Isolate the target bit's coordinates inside the 256-bit space
		byteNumber := examinedByte >> 3
		bitNumberWithinByte := examinedByte & 0x07 // Keep this for the byte presence check

		// The 32-byte bitmask flags slice starts at offset 2 of cn.nodeBytes
		flagsOffset := 2
		flagsByte := dfn.nodeBytes[flagsOffset+int(byteNumber)]
		if flagsByte&(1<<bitNumberWithinByte) == 0 {
			return 0 // Dead end: Bit is '0', slot is unpopulated
		}

		// 2. Identify which of the 4 uint64 buckets our target bit belongs to
		targetBlock := examinedByte >> 6 // 0 to 3 (Which uint64)
		// Calculate the exact bit shift position inside the 64-bit integer (0 to 63)
		bitNumberWithinBlock := examinedByte & 0x3F

		// Read the four 64-bit blocks out of the node bytes stream
		u0 := binary.LittleEndian.Uint64(dfn.nodeBytes[flagsOffset+0 : flagsOffset+8])
		u1 := binary.LittleEndian.Uint64(dfn.nodeBytes[flagsOffset+8 : flagsOffset+16])
		u2 := binary.LittleEndian.Uint64(dfn.nodeBytes[flagsOffset+16 : flagsOffset+24])
		u3 := binary.LittleEndian.Uint64(dfn.nodeBytes[flagsOffset+24 : flagsOffset+32])

		var onesBefore int

		switch targetBlock {
		case 0:
			mask := (uint64(1) << bitNumberWithinBlock) - 1
			onesBefore = bits.OnesCount64(u0 & mask)
		case 1:
			mask := (uint64(1) << bitNumberWithinBlock) - 1
			onesBefore = bits.OnesCount64(u0) + bits.OnesCount64(u1&mask)
		case 2:
			mask := (uint64(1) << bitNumberWithinBlock) - 1
			onesBefore = bits.OnesCount64(u0) + bits.OnesCount64(u1) + bits.OnesCount64(u2&mask)
		case 3:
			mask := (uint64(1) << bitNumberWithinBlock) - 1
			onesBefore = bits.OnesCount64(u0) + bits.OnesCount64(u1) + bits.OnesCount64(u2) + bits.OnesCount64(u3&mask)
		}
		// 3. Compute physical NodeIdType payload layout offset
		// The NodeIdType data payloads start directly after our 32-byte bitmask (offset 34).
		uint16PayloadStart := flagsOffset + 32
		nodeIdByteOffset := uint16PayloadStart + (onesBefore * nodeIdSize)

		return (*nodeIdConfig).ReadID(dfn.nodeBytes[nodeIdByteOffset : nodeIdByteOffset+nodeIdSize])
	}
	// Is it a FormatTiny?
	if tinySlots > 0 {
		//fmt.Println("Visiting a FormatTiny node")
		// FormatTiny is a hashByteIndex (byte) followed by
		// a series of {byteValue (byte), (NodeIdType)} pairs
		offset := 1
		for slot := 0; slot < int(tinySlots); slot++ {
			byteValue := dfn.nodeBytes[offset]
			offset++
			if examinedByte == byteValue {
				return (*nodeIdConfig).ReadID(dfn.nodeBytes[offset : offset+nodeIdSize])
			}
			offset += nodeIdSize
		}
		return 0
	}
	panic("Unrecognized format")
}

// getAllNextLevelNodeIds should only be called if detailsIfLeaf() returned false and you are in the process
// of recursing to all leaves.
func (dfn *donutForestNode) getAllNextLevelNodeIds(mediumSlots byte, tinySlots byte,
	nodeIdConfig *NByteIdConfig[NodeIdType]) []NodeIdType {

	result := make([]NodeIdType, 0, 256)

	// Is it a FormatFull?
	nodeIdSize := (*nodeIdConfig).StorageBytes()
	if dfn.formatSpecBytes == uint32(1+1+256*nodeIdSize) {
		//fmt.Println("Visiting a FormatFull node")
		// We "imagine" what would happen if the "next examined byte" was each of its 256 possible values,
		// and return all the non-zero resulting node ids
		for imaginedExaminedByteInt := range 256 {
			byteIndex := 1 + 1 + imaginedExaminedByteInt*nodeIdSize
			nextLevelNodeId := (*nodeIdConfig).ReadID(dfn.nodeBytes[byteIndex : byteIndex+nodeIdSize])
			if nextLevelNodeId != 0 {
				result = append(result, nextLevelNodeId)
			}
		}
	} else if mediumSlots > 0 {
		// Is it a FormatMedium?
		//fmt.Println("Visiting a FormatMedium node")
		// There are 256 bits within cn.nodeBytes which tell us (if each is a '1') which slots are represented
		// in the NodeIdType's which follow.
		// HOWEVER, since we are visiting all nodes (on our way to all leaves), the "mediumSlots" value
		// already directly tells us a maximum number of slots to examine.

		// The 32-byte bitmask flags slice starts at offset 2 of cn.nodeBytes
		flagsOffset := 2

		// Compute physical NodeIdType payload layout offset
		// The NodeIdType data payloads start directly after our 32-byte bitmask (offset 34).
		uint16PayloadStart := flagsOffset + 32
		for onesBefore := range int(mediumSlots) {
			nodeIdByteOffset := uint16PayloadStart + (onesBefore * nodeIdSize)
			singleNodeId := (*nodeIdConfig).ReadID(dfn.nodeBytes[nodeIdByteOffset : nodeIdByteOffset+nodeIdSize])
			if singleNodeId != 0 {
				result = append(result, singleNodeId)
			}
		}
	} else if tinySlots > 0 {
		// Is it a FormatTiny?
		//fmt.Println("Visiting a FormatTiny node")
		// FormatTiny is a hashByteIndex (byte) followed by
		// a series of {byteValue (byte), (NodeIdType)} pairs
		offset := 1
		for slot := 0; slot < int(tinySlots); slot++ {
			offset++ // Skip the byte value field
			singleNodeId := (*nodeIdConfig).ReadID(dfn.nodeBytes[offset : offset+nodeIdSize])
			offset += nodeIdSize
			if singleNodeId != 0 {
				result = append(result, singleNodeId)
			}
		}
	} else {
		panic("Unrecognized format")
	}
	return result
}
