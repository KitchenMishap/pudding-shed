package indexedhashes3

type HashIndexingParams struct {
	bitsPerHashIndex        int64 // A parameter
	numberOfBins            int64
	bytesPerBinEntry        int64
	entriesInBinStart       int64
	digitsPerNumberedFolder int
	// Derived values
	bitsPerSortNum       int64
	maskLsbsForHashIndex uint64
	maskLsbsForSortNum   uint64
}

/*
func NewHashStoreParams(bitsPerHashIndex int64, hashCountEstimate int64) *Params {
	params := Params{}
	params.bitsPerHashIndex = bitsPerHashIndex

	params.numberOfBins, params.bytesPerBinEntry = optimize(bitsPerHashIndex, hashCountEstimate)

	bitsForHashIndexAndSortNum := params.bytesPerBinEntry*8 - 192 // 192 bits for truncated hash
	params.bitsPerSortNum = bitsForHashIndexAndSortNum - params.bitsPerHashIndex
	params.maskLsbsForHashIndex = uint64(1)<<params.bitsPerHashIndex - 1
	params.maskLsbsForSortNum = uint64(1)<<params.bitsPerSortNum - 1
	return &params
}*/

func (p *HashIndexingParams) NumberOfBins() int64 {
	return p.numberOfBins
}

func (p *HashIndexingParams) BytesPerBinEntry() int64 {
	return p.bytesPerBinEntry
}

func (p *HashIndexingParams) BitsPerHashIndex() int64 {
	return p.bitsPerHashIndex
}

func (p *HashIndexingParams) BitsPerSortNum() int64 {
	return p.bitsPerSortNum
}

func (p *HashIndexingParams) MaskForHashIndex() uint64 {
	return p.maskLsbsForHashIndex
}

func (p *HashIndexingParams) MaskForSortNum() uint64 {
	return p.maskLsbsForSortNum
}

func (p *HashIndexingParams) EntriesInBinStart() int64 {
	return p.entriesInBinStart
}
