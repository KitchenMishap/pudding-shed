package indexedhashes3

import "encoding/binary"

// The following comment is WRONG ToDo

// binEntryBytes is a byte slice whose size is fixed for a given hash indexing store, somewhere between 24 and 32 bytes.
// The 24 LSBytes are the truncatedHash (least significant 24 bytes of the full hash).
// The 8 MSBytes are called the MS64Bits and are obtained as a uint64.
// Note that for len(binEntryBytes) < 32, the 24 LSBytes and MS64Bits overlap! So some masking and shifting is involved.
// MS64Bits holds in its various bits, the HashIndex (with a known number of bits) and the SortNum (known num of bits).
// The number of bits of each is fixed for a given hash indexing store, but varies between stores (it is specified
// in a Params object you supply). And beware that MS64Bits is also polluted by some bits from the truncatedHash,
// due to the overlap mentioned above.

// HashIndex is in the most significant m bits of MS64Bits, where m = p.BitsPerHashIndex
// SortNum is in the subsequent n most significant bits of MS64Bits, where n = p.BitsPerSortNum

type binEntryBytes []byte // The number of bytes in the slice is fixed for a given hash indexing store

func newBinEntryBytes(t *truncatedHash, hi hashIndex, sn sortNum, p *HashIndexingParams) binEntryBytes {
	result := make([]byte, p.BytesPerBinEntry())

	// Write in the truncated hash
	copy(result[0:24], (*t)[0:24])

	// Write the hash index
	byteCount1 := p.BytesPerHashIndex()
	bytesForHashIndex := [8]byte{}
	binary.LittleEndian.PutUint64(bytesForHashIndex[0:8], uint64(hi))
	for i := byteCount1; i < 8; i++ {
		if bytesForHashIndex[i] != 0 {
			panic("hash index didn't fit into bytes")
		}
	}
	copy(result[24:24+byteCount1], bytesForHashIndex[0:byteCount1])

	//Write the sort num
	byteCount2 := p.BytesPerSortNum()
	bytesForSortNum := [8]byte{}
	binary.LittleEndian.PutUint64(bytesForSortNum[0:8], uint64(sn))
	for i := byteCount2; i < 8; i++ {
		if bytesForHashIndex[i] != 0 {
			panic("sort num didn't fit into bytes")
		}
	}
	copy(result[24+byteCount1:24+byteCount1+byteCount2], bytesForSortNum[0:byteCount2])

	return result
}

func (beb *binEntryBytes) getTruncatedHash() *truncatedHash {
	result := truncatedHash{}
	copy(result[0:24], (*beb)[0:24])
	return &result
}

func (beb *binEntryBytes) getHashIndexSortNum(p *HashIndexingParams) (hashIndex, sortNum) {
	hashIndexBytes := [8]byte{}
	sortNumBytes := [8]byte{}

	hiByteCount := p.BytesPerHashIndex()
	snByteCount := p.BytesPerSortNum()
	copy((*beb)[24:24+hiByteCount], hashIndexBytes[0:hiByteCount])
	copy((*beb)[24+hiByteCount:24+hiByteCount+snByteCount], sortNumBytes[0:snByteCount])
	hi := binary.LittleEndian.Uint64(hashIndexBytes[0:8])
	sn := binary.LittleEndian.Uint64(sortNumBytes[0:8])
	return hashIndex(hi), sortNum(sn)
}
