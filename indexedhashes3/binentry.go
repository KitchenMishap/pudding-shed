package indexedhashes3

// Bin entries are found in a binMem, and also in the BinStarts.bin file and the overflow files xx.ovf
// The truncHash field can be expanded to a Hash using the binNum number
// The hashIndex field is stored to a file using the number of bytes specified in Params.bytesPerHashIndex

type binEntry struct {
	truncHash truncatedHash
	hashIndex hashIndex
}

func (be *binEntry) toBytes(p *Params) []byte {
	hashBytesCount := p.bytesInTruncatedHash()
	buf := make([]byte, hashBytesCount+p.bytesPerHashIndex)
	truncHashSlice := be.truncHash.toBytes(p)
	copy(buf[0:hashBytesCount], truncHashSlice)
	hashIndexSlice := be.hashIndex.toBytes(p)
	copy(buf[hashBytesCount:hashBytesCount+p.bytesPerHashIndex], hashIndexSlice)
	return truncHashSlice
}

func (be *binEntry) fromBytes(buf []byte, p *Params) {
	hashBytesCount := p.bytesInTruncatedHash()
	copy(be.truncHash[0:hashBytesCount], buf[0:hashBytesCount])
	be.hashIndex.fromBytes(buf[hashBytesCount:hashBytesCount+p.bytesPerHashIndex], p)
}
