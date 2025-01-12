package indexedhashes3

import "encoding/binary"

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
	MS64Bits := uint64(hi) // Put hashIndex in the LSBs to start with
	if hi >= 1<<p.BitsPerHashIndex() {
		panic("hashIndex doesn't fit ths number of bits")
	}
	MS64Bits <<= p.BitsPerSortNum() // Shift left to make room for SortNum
	if sn >= 1<<p.BitsPerSortNum() {
		panic("sortNum doesn't fit ths number of bits")
	}
	MS64Bits |= uint64(sn)
	// hashIndex followed by sortNum are currently packed into the LSBs
	// Shift them to the MSBs
	MS64Bits <<= (64 - (p.BitsPerHashIndex() + p.BitsPerSortNum()))
	// Write MS64Bits into the most significant bytes (LittleEndian) of the byte slice
	binary.LittleEndian.PutUint64(result[p.BytesPerBinEntry()-8:p.BytesPerBinEntry()], MS64Bits)
	// Write in (including overwrite some zero bytes) the truncated hash
	copy(result[0:24], (*t)[0:24])
	return (binEntryBytes)(result)
}

func (beb *binEntryBytes) getTruncatedHash() *truncatedHash {
	result := truncatedHash{}
	copy(result[0:24], (*beb)[0:24])
	return &result
}

func (beb *binEntryBytes) getMS64Bits(p *HashIndexingParams) uint64 {
	return binary.LittleEndian.Uint64((*beb)[p.BytesPerBinEntry()-8 : p.BytesPerBinEntry()])
}

func (beb *binEntryBytes) getHashIndexSortNum(p *HashIndexingParams) (hashIndex, sortNum) {
	MS64Bits := beb.getMS64Bits(p)
	MS64Bits >>= (64 - p.BitsPerHashIndex() - p.BitsPerSortNum())
	sn := MS64Bits & p.MaskForSortNum()
	MS64Bits >>= p.BitsPerSortNum()
	return hashIndex(MS64Bits & p.MaskForHashIndex()), sortNum(sn)
}
