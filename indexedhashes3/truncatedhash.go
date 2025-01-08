package indexedhashes3

import "encoding/binary"

// A truncated hash is taken from a hash by discarding the bytes which are whole bytes of the binNum
// It is stored in file and memory using binary.LittleEndian (ie LSB first)

type truncatedHash []byte

func (th *truncatedHash) toHash(bn binNum, p *Params) *Hash {
	wholeBytesInBinNum := p.wholeBytesInBinNum()
	hash := Hash{}

	// The LSB bytes of hash are from the binNum
	binNumBytes := [8]byte{}
	binary.LittleEndian.PutUint64(binNumBytes[:], uint64(bn))
	for i := int64(0); i < wholeBytesInBinNum; i++ {
		hash[i] = binNumBytes[i]
	}

	// The MSB bytes of hash are from the truncatedHash
	otherBytes := 32 - wholeBytesInBinNum
	for i := int64(0); i < otherBytes; i++ {
		hash[wholeBytesInBinNum+i] = (*th)[i]
	}
	return &hash
}

// The sortNum of a truncatedHash comes from bytes 1,2,3,4 (not byte 0).
// Byte zero may contain as its MSBits, the top MSBits of the binNum, which will always be the same within one bin.
// Byte zero is therefore not included in the sort operation.
func (th *truncatedHash) toSortNum(p *Params) sortNum {
	if len(*th) < 5 {
		panic("truncatedHash too short to determine a sortNum")
	}
	i32 := binary.LittleEndian.Uint32((*th)[1:5])
	return sortNum(float32(i32))
}

func (th *truncatedHash) toBytes(p *Params) []byte {
	result := make([]byte, p.bytesInTruncatedHash())
	copy(result, *th)
	return result
}

func (th *truncatedHash) fromBytes(bytes []byte, p *Params) {
	copy(*th, bytes)
}
