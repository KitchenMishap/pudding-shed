package indexedhashes3

// An abbreviated hash is a uint64, being the 64 LSBits of the hash

type abbreviatedHash uint64

func newAbbreviatedHashFromBinNumSortNum(bn binNum, sn sortNum, p *HashIndexingParams) abbreviatedHash {
	return abbreviatedHash(uint64(int64(bn)*p.NumberOfBins()) + uint64(sn))
}

func (ah *abbreviatedHash) toBinNum(p *HashIndexingParams) binNum {
	// Integer division
	return binNum(uint64(*ah) / uint64(p.NumberOfBins()))
}

func (ah *abbreviatedHash) toSortNum(p *HashIndexingParams) sortNum {
	// Remainder
	return sortNum(float64(uint64(*ah) % uint64(p.NumberOfBins())))
}
