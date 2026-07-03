package weddingcake

import "encoding/binary"

// Bounded semantic types
// The following two types are often the same (since the number of nodes is "similiar" to the number of hashes in a chunk)
// NodeId's are used to identify nodes within a level within a chunk
type NodeIdType uint32 // Could be uint64 some day

// HashIndexId's are used to identify Hash Presentation Indices within a chunk
type HashIndexIdType uint32 // Could be uint64 some day
const MaxHashIndexId = ^HashIndexIdType(0)
const HashIndexIdNoMatch = HashIndexIdType(0)

type BytesCountType uint32 // Often needs to be bigger than NodeIdType, Might be uint64 some day
const MaxBytesCount = ^BytesCountType(0)

// IdTypeConstraint defines the underlying unsigned integers we support parameterizing on
type IdTypeConstraint interface {
	~uint16 | ~uint32 | ~uint64
}

// NByteIdConfig handles the dynamic bytes sizing for any underlying IdTypeConstraint
type NByteIdConfig[T IdTypeConstraint] interface {
	StorageBytes() int
	WriteID(b []byte, id T)
	ReadID(b []byte) T
}

// ID16 implements 16-bit IDs (2 bytes)
type ID16[T IdTypeConstraint] struct{}

func (ID16[T]) StorageBytes() int      { return 2 }
func (ID16[T]) WriteID(b []byte, id T) { binary.LittleEndian.PutUint16(b[:2], uint16(id)) }
func (ID16[T]) ReadID(b []byte) T      { return T(binary.LittleEndian.Uint16(b[:2])) }

// ID24 implements 24-bit IDs (3 bytes)
type ID24[T IdTypeConstraint] struct{}

func (ID24[T]) StorageBytes() int { return 3 }
func (ID24[T]) WriteID(b []byte, id T) {
	val := uint32(id)
	b[0] = byte(val)
	b[1] = byte(val >> 8)
	b[2] = byte(val >> 16)
}
func (ID24[T]) ReadID(b []byte) T {
	return T(uint32(b[0]) | (uint32(b[1]) << 8) | (uint32(b[2]) << 16))
}

// ID32 implements 32-bit IDs (4 bytes)
type ID32[T IdTypeConstraint] struct{}

func (ID32[T]) StorageBytes() int      { return 4 }
func (ID32[T]) WriteID(b []byte, id T) { binary.LittleEndian.PutUint32(b[:4], uint32(id)) }
func (ID32[T]) ReadID(b []byte) T      { return T(binary.LittleEndian.Uint32(b[:4])) }

/* Uncomment this some day? (Will need to define 'type NodeIdType uint64' above)
// ID40 implements 40-bit node IDs (5 bytes)
type ID40 struct{}
func (ID40) StorageBytes() int { return 5 }
func (ID40) WriteID(b []byte, id NodeIdType) {
	b[0] = byte(id)
	b[1] = byte(id >> 8)
	b[2] = byte(id >> 16)
	b[3] = byte(id >> 24)
	b[4] = byte(id >> 32)
}
func (ID40) ReadID(b []byte) NodeIdType {
	return NodeIdType(b[0]) | (NodeIdType(b[1]) << 8) | (NodeIdType(b[2]) << 16) |
		NodeIdType(b[3]) << 24 | NodeIdType(b[4]) << 32
}*/
