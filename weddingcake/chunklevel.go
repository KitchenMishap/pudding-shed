package weddingcake

import (
	"encoding/binary"
	"math/bits"
)

// A ChunkLevel is an object (typically a slice of bytes, possibly part of an mmap'd file) which
// holds a particular level of the tree of a particular chunk.

// A ChunkLevel breaks down into "indexbytes" followed by "nodebytes".
// "nodebytes" is a sequence of nodes, which can each be represented by different formats with different bytesizes.
// Each node is identified by its level, and a node id within that level.
// "indexbytes" provides a mapping from each node id within a level to:
// (1) a byte offset into "nodebytes"
// (2) a format spec which describes the way a particular node is stored as bytes

type ChunkLevel struct {
	IndexBytes []byte
	NodeBytes  []byte
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
type chunkNode struct {
	formatSpecBytes uint32
	nodeBytes       []byte
}

// ExtractNode extracts to an existing chunkNode to avoid a busy heap
// Returns true for success
func (cl *ChunkLevel) ExtractNode(nodeId NodeIdType, target *chunkNode,
	nodeIdConfig NByteIdConfig[NodeIdType]) bool {

	currentSpecNodeId := NodeIdType(1)
	currentSpecNodeByteOffset := BytesCountType(0)
	byteCount := BytesCountType(0)
	formatSpecCount := binary.LittleEndian.Uint16(cl.IndexBytes[byteCount : byteCount+2])
	byteCount += 2

	// The results of the following loop
	found := false
	var formatSpecBytes uint32
	var nodeByteOffset BytesCountType
	var nodeByteSize uint16

	nodeCountSize := nodeIdConfig.StorageBytes()
	for fs := uint16(0); fs < formatSpecCount; fs++ {
		formatSpecNodeCount := NodeCountType(nodeIdConfig.ReadID(cl.IndexBytes[byteCount : byteCount+BytesCountType(nodeCountSize)]))
		byteCount += BytesCountType(nodeCountSize)
		formatSpecBytes = binary.LittleEndian.Uint32(cl.IndexBytes[byteCount : byteCount+4])
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
	target.nodeBytes = cl.NodeBytes[nodeByteOffset : nodeByteOffset+BytesCountType(nodeByteSize)]
	return true
}

// Here are various functions to ask various questions of a node.
// chunkNode.nodeBytes will be interpreted in different ways depending on chunkNode.formatSpecBytes.

// detailsIfLeaf returns (true, reassuranceBytes, presentationIndex) if cn is a leaf
// returns (false, nil, 0) otherwise
// reassuranceBytesCount is a configuration pertaining to the overall store
func (cn *chunkNode) detailsIfLeaf(reassuranceBytesCount byte,
	hashIndexIdConfig NByteIdConfig[HashIndexIdType]) (bool, []byte, HashIndexIdType) {

	// A leaf node is identified as two MSB zero bytes by an appropriate two LSB bytes bytesCount
	hashIndexIdSize := hashIndexIdConfig.StorageBytes()
	if cn.formatSpecBytes != uint32(reassuranceBytesCount)+uint32(hashIndexIdSize) {
		return false, nil, 0
	}
	// cn.nodeBytes interpreted as FormatLeaf
	reassuranceBytes := cn.nodeBytes[0:reassuranceBytesCount]
	hashIndexId := hashIndexIdConfig.ReadID(cn.nodeBytes[reassuranceBytesCount : int(reassuranceBytesCount)+hashIndexIdSize])
	return true, reassuranceBytes, hashIndexId
}

// hashByteIndexToExamine() should only be called if detailsIfLeaf() has already returned false
// It returns (index, 0, 0) for a FormatFull,
// or (index, slots, 0) for a FormatMedium,
// or (index, 0, slots) for a FormatTiny
func (cn *chunkNode) hashByteIndexToExamine(nodeIdConfig NByteIdConfig[NodeIdType]) (byte, byte, byte) {
	// Is it a FormatFull?
	nodeIdSize := nodeIdConfig.StorageBytes()
	if cn.formatSpecBytes == uint32(1+1+256*nodeIdSize) {
		if cn.nodeBytes[0] != 0 {
			panic("Expected a zero byte for padding")
		}
		return cn.nodeBytes[1], 0, 0
	}
	// Is it a FormatMedium?
	mediumSlots := byte((cn.formatSpecBytes & 0x00FF0000) >> 16)
	if mediumSlots > 0 {
		if cn.nodeBytes[0] != 0 {
			panic("Expected a zero byte for padding")
		}
		return cn.nodeBytes[1], mediumSlots, 0
	}
	// Is it a FormatTiny?
	tinySlots := byte((cn.formatSpecBytes & 0xFF000000) >> 24)
	if tinySlots > 0 {
		// No padding in this case, because FormatTiny is often a small odd number of bytes, and we want
		// them tightly packed
		return cn.nodeBytes[0], 0, tinySlots
	}
	panic("Unrecognised format spec")
}

// nextLevelNodeId should only be called if detailsIfLeaf() returned false and you have extracted the
// hash byte specified by hashByteIndexToExamine().
// It returns 0 for a "dead end" "empty slot" "hash not found" scenario
func (cn *chunkNode) nextLevelNodeId(examinedByte byte, mediumSlots byte, tinySlots byte,
	nodeIdConfig NByteIdConfig[NodeIdType]) NodeIdType {

	// Is it a FormatFull?
	nodeIdSize := nodeIdConfig.StorageBytes()
	if cn.formatSpecBytes == uint32(1+1+256*nodeIdSize) {
		//fmt.Println("Visiting a FormatFull node")
		byteIndex := 1 + 1 + int(examinedByte)*nodeIdSize
		return nodeIdConfig.ReadID(cn.nodeBytes[byteIndex : byteIndex+nodeIdSize])
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
		flagsByte := cn.nodeBytes[flagsOffset+int(byteNumber)]
		if flagsByte&(1<<bitNumberWithinByte) == 0 {
			return 0 // Dead end: Bit is '0', slot is unpopulated
		}

		// 2. Identify which of the 4 uint64 buckets our target bit belongs to
		targetBlock := examinedByte >> 6 // 0 to 3 (Which uint64)
		// Calculate the exact bit shift position inside the 64-bit integer (0 to 63)
		bitNumberWithinBlock := examinedByte & 0x3F

		// Read the four 64-bit blocks out of the node bytes stream
		u0 := binary.LittleEndian.Uint64(cn.nodeBytes[flagsOffset+0 : flagsOffset+8])
		u1 := binary.LittleEndian.Uint64(cn.nodeBytes[flagsOffset+8 : flagsOffset+16])
		u2 := binary.LittleEndian.Uint64(cn.nodeBytes[flagsOffset+16 : flagsOffset+24])
		u3 := binary.LittleEndian.Uint64(cn.nodeBytes[flagsOffset+24 : flagsOffset+32])

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

		return nodeIdConfig.ReadID(cn.nodeBytes[nodeIdByteOffset : nodeIdByteOffset+nodeIdSize])
	}
	// Is it a FormatTiny?
	if tinySlots > 0 {
		//fmt.Println("Visiting a FormatTiny node")
		// FormatTiny is a hashByteIndex (byte) followed by
		// a series of {byteValue (byte), (NodeIdType)} pairs
		offset := 1
		for slot := 0; slot < int(tinySlots); slot++ {
			byteValue := cn.nodeBytes[offset]
			offset++
			if examinedByte == byteValue {
				return nodeIdConfig.ReadID(cn.nodeBytes[offset : offset+nodeIdSize])
			}
			offset += nodeIdSize
		}
		return 0
	}
	panic("Unrecognized format")
}

// countChunkLevelBytes() returns the number of bytes for indexBytes and for nodeBytes
func countChunkLevelBytes(levelData *LevelFormat,
	nodeIdConfig NByteIdConfig[NodeIdType]) (uint64, uint64) {
	indexBytesCount := uint64(0)
	nodesBytesCount := uint64(0)

	// For each level, in indexBytes, the first 2 bytes represent a count of NodeSpecs ("groups") that follow
	indexBytesCount += 2

	nodeCountSize := nodeIdConfig.StorageBytes()
	for groupIndex := range levelData.Groups {
		group := &(levelData.Groups[groupIndex])
		// In the indexBytes, for this group (nodespec), N bytes specify the number of nodes,
		// and four bytes describe the formatSpec
		indexBytesCount += uint64(nodeCountSize + 4)
		// In the nodesBytes, for this group (nodespec), the number of bytes has already been determined
		nodesBytesCount += uint64(group.Bytes)
		// fmt.Printf("countChunkLevelBytes(): group index %d bytes %d\n", groupIndex, group.Bytes)
	}
	return indexBytesCount, nodesBytesCount
}
