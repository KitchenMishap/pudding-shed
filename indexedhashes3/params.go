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
	// A parameter, used to arrange overflow files into a sensible number of folders
	DigitsPerNumberedFolder_ int `json:"digitsPerNumberedFolder"`
	// A parameter
	BytesPerSortNum_ int64 `json:"bytesPerSortNum"`

	// Derived values, not stored to JSON
	bytesRoomForBinNum int64  `json:"-"`
	divider            uint64 `json:"-"`
	bytesPerBinEntry   int64  `json:"-"`
}

// bitsPerHashIndex is critical, you will only ever be able to add as many hashes whose count fits in these bits.
// hashCountEstimate is not critical, you CAN exceed this number by appending more.
// digitsPerNumberedFolder is a matter of taste, but 2 is a very good balance.
// numberOfBins, entriesInBinStart, bytesPerSortNum are the subject of careful and complicated optimization.
func NewHashStoreParams(bitsPerHashIndex int64, hashCountEstimate int64, digitsPerNumberedFolder int,
	numberOfBins int64, entriesInBinStart int64, bytesPerSortNum int64) *HashIndexingParams {
	if numberOfBins%2 == 1 {
		panic("number of bins must be even")
	}
	params := HashIndexingParams{}
	params.BitsPerHashIndex_ = bitsPerHashIndex
	params.HashCountEstimate_ = hashCountEstimate
	params.NumberOfBins_ = numberOfBins
	params.EntriesInBinStart_ = entriesInBinStart
	params.BytesPerSortNum_ = bytesPerSortNum
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
	p.divider = (uint64(1) << uint64(63)) / (uint64(p.NumberOfBins_) >> uint64(1))
	// 24 is the number of bytes in a truncated hash
	p.bytesPerBinEntry = 24 + p.BytesPerSortNum_ + p.BytesPerHashIndex()
}

func (p *HashIndexingParams) BytesPerHashIndex() int64 { return (p.BitsPerHashIndex_-1)/8 + 1 }

func (p *HashIndexingParams) BytesPerSortNum() int64 { return p.BytesPerSortNum_ }

func (p *HashIndexingParams) NumberOfBins() int64 {
	return p.NumberOfBins_
}

func (p *HashIndexingParams) BytesPerBinEntry() int64 { return p.bytesPerBinEntry }

func (p *HashIndexingParams) BitsPerHashIndex() int64 {
	return p.BitsPerHashIndex_
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
