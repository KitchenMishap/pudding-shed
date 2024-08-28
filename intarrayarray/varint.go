package intarrayarray

import "encoding/binary"

type VarInt struct {
	val int64
}

func (vi *VarInt) FromInt64(v int64) {
	vi.val = v
}

func (vi *VarInt) ToInt64() int64 {
	return vi.val
}

// FromBytes returns number of bytes read
func (vi *VarInt) FromBytes(byts []byte) int {
	if byts[0] < 0xFD {
		vi.val = int64(byts[0])
		return 1
	} else if byts[0] == 0xFD {
		vi.val = int64(binary.LittleEndian.Uint16(byts[1:3]))
		return 3
	} else if byts[0] == 0xFE {
		vi.val = int64(binary.LittleEndian.Uint32(byts[1:5]))
		return 5
	} else if byts[0] == 0xFF {
		vi.val = int64(binary.LittleEndian.Uint64(byts[1:9]))
		return 9
	}
	return 0 // Can't happen
}

func (vi *VarInt) ToBytes() []byte {
	resultSpace := [9]byte{}
	binary.LittleEndian.PutUint64(resultSpace[1:9], uint64(vi.val))
	if vi.val < 0xFD {
		return resultSpace[1:2]
	} else if vi.val <= 0xFFFF {
		resultSpace[0] = 0xFD
		return resultSpace[0:3]
	} else if vi.val <= 0xFFFFFFFF {
		resultSpace[0] = 0xFE
		return resultSpace[0:5]
	} else {
		resultSpace[0] = 0xFF
		return resultSpace[0:9]
	}
}
