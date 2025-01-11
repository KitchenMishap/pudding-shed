package indexedhashes3

type HashIndexingParams struct {
	bitsPerHashIndex        int64 // A parameter, room for the number of hashes you ever want to append
	hashCountEstimate       int64 // A parameter, a guideline for estimating the amount of memory that preloading will use
	numberOfBins            int64 // A parameter, probably the result of an optimization calculation
	entriesInBinStart       int64 // A parameter, perhaps optimized to result in 4096 binstart bytes per bin
	bytesPerBinEntry        int64 // A parameter, used to calculate bitsPerSortNum, 32 is a likely candidate
	digitsPerNumberedFolder int   // A parameter, used to arrange overflow files into a sensible number of folders
	// Derived values
	bitsPerSortNum       int64
	maskLsbsForHashIndex uint64
	maskLsbsForSortNum   uint64
}

// bitsPerHashIndex is critical, you will only ever be able to add as many hashes whose count fits in these bits.
// hashCountEstimate is not critical, you CAN exceed this number by appending more.
// digitsPerNumberedFolder is a matter of taste, but 2 is a very good balance.
// numberOfBins, entriesInBinStart, and bytesPerBinEntry are the subject of careful and complicated optimization.
func NewHashStoreParams(bitsPerHashIndex int64, hashCountEstimate int64, digitsPerNumberedFolder int,
	numberOfBins int64, entriesInBinStart int64, bytesPerBinEntry int64) *HashIndexingParams {
	params := HashIndexingParams{}
	params.bitsPerHashIndex = bitsPerHashIndex
	params.hashCountEstimate = hashCountEstimate
	params.numberOfBins = numberOfBins
	params.entriesInBinStart = entriesInBinStart
	params.bytesPerBinEntry = bytesPerBinEntry
	params.digitsPerNumberedFolder = digitsPerNumberedFolder

	bitsForHashIndexAndSortNum := params.bytesPerBinEntry*8 - 192 // 192 bits for truncated hash
	params.bitsPerSortNum = bitsForHashIndexAndSortNum - params.bitsPerHashIndex
	params.maskLsbsForHashIndex = uint64(1)<<params.bitsPerHashIndex - 1
	params.maskLsbsForSortNum = uint64(1)<<params.bitsPerSortNum - 1
	return &params
}

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

func (p *HashIndexingParams) HashCountEstimate() int64 {
	return p.hashCountEstimate
}
