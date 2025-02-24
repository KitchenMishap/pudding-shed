package indexedhashes3

// An abbreviated hash is a uint64, being the 64 LSBits of the hash

type abbreviatedHash uint64

func newAbbreviatedHashFromBinNumSortNum(bn binNum, sn sortNum, p *HashIndexingParams) abbreviatedHash {
	return abbreviatedHash(uint64(bn)*p.Divider() + uint64(sn))
}

func (ah *abbreviatedHash) toBinNum(p *HashIndexingParams) binNum {
	// Integer division
	return binNum(uint64(*ah) / p.Divider())
}

func (ah *abbreviatedHash) toSortNum(p *HashIndexingParams) sortNum {
	// Remainder
	return sortNum(uint64(*ah) % p.Divider())
}
