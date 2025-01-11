package indexedhashes3

type HashIndexingParams struct {
	// A parameter, room for the number of hashes you ever want to append
	BitsPerHashIndex_ int64 `json:"bitsPerHashIndex"`
	// A parameter, a guideline for estimating the amount of memory that preloading will use
	HashCountEstimate_ int64 `json:"hashCountEstimate"`
	// A parameter, probably the result of an optimization calculation
	NumberOfBins_ int64 `json:"numberOfBins"`
	// A parameter, perhaps optimized to result in 4096 binstart bytes per bin
	EntriesInBinStart_ int64 `json:"entriesInBinStart"`
	// A parameter, used to calculate bitsPerSortNum, 32 is a likely candidate
	BytesPerBinEntry_ int64 `json:"bytesPerBinEntry"`
	// A parameter, used to arrange overflow files into a sensible number of folders
	DigitsPerNumberedFolder_ int `json:"digitsPerNumberedFolder"`

	// Derived values, not stored to JSON
	bitsPerSortNum       int64  `json:"-"`
	bytesRoomForBinNum   int64  `json:"-"`
	maskLsbsForHashIndex uint64 `json:"-"`
	maskLsbsForSortNum   uint64 `json:"-"`
	divider              uint64 `json:"-"`
}

// bitsPerHashIndex is critical, you will only ever be able to add as many hashes whose count fits in these bits.
// hashCountEstimate is not critical, you CAN exceed this number by appending more.
// digitsPerNumberedFolder is a matter of taste, but 2 is a very good balance.
// numberOfBins, entriesInBinStart, and bytesPerBinEntry are the subject of careful and complicated optimization.
func NewHashStoreParams(bitsPerHashIndex int64, hashCountEstimate int64, digitsPerNumberedFolder int,
	numberOfBins int64, entriesInBinStart int64, bytesPerBinEntry int64) *HashIndexingParams {
	params := HashIndexingParams{}
	params.BitsPerHashIndex_ = bitsPerHashIndex
	params.HashCountEstimate_ = hashCountEstimate
	params.NumberOfBins_ = numberOfBins
	params.EntriesInBinStart_ = entriesInBinStart
	params.BytesPerBinEntry_ = bytesPerBinEntry
	params.DigitsPerNumberedFolder_ = digitsPerNumberedFolder

	params.calculateDerivedValues()

	return &params
}

func (p *HashIndexingParams) calculateDerivedValues() {
	for bytes := int64(0); bytes < 8; bytes++ {
		numbersStorable := int64(1) << (8 * bytes)
		if numbersStorable >= p.NumberOfBins_ {
			p.bytesRoomForBinNum = bytes
			break
		}
	}
	bitsForHashIndexAndSortNum := p.BytesPerBinEntry_*8 - 192 // 192 bits for truncated hash
	p.bitsPerSortNum = bitsForHashIndexAndSortNum - p.BitsPerHashIndex_
	p.maskLsbsForHashIndex = uint64(1)<<p.BitsPerHashIndex_ - 1
	p.maskLsbsForSortNum = uint64(1)<<p.bitsPerSortNum - 1
	p.divider = (uint64(1) << uint64(63)) / (uint64(p.NumberOfBins_) >> uint64(1))
}

func (p *HashIndexingParams) NumberOfBins() int64 {
	return p.NumberOfBins_
}

func (p *HashIndexingParams) BytesPerBinEntry() int64 { return p.BytesPerBinEntry_ }

func (p *HashIndexingParams) BitsPerHashIndex() int64 {
	return p.BitsPerHashIndex_
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
	return p.EntriesInBinStart_
}

func (p *HashIndexingParams) HashCountEstimate() int64 {
	return p.HashCountEstimate_
}

func (p *HashIndexingParams) BytesRoomForBinNum() int64 { return p.bytesRoomForBinNum }

func (p *HashIndexingParams) DigitsPerNumberedFolder() int { return p.DigitsPerNumberedFolder_ }

func (p *HashIndexingParams) Divider() uint64 { return p.divider }
