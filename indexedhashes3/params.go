package indexedhashes3

type Params struct {
	bitsPerHashIndex int64 // A parameter
	numberOfBins     int64
	bytesPerBinEntry int64
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

func (p *Params) NumberOfBins() int64 {
	return p.numberOfBins
}

func (p *Params) BytesPerBinEntry() int64 {
	return p.bytesPerBinEntry
}

func (p *Params) BitsPerHashIndex() int64 {
	return p.bitsPerHashIndex
}

func (p *Params) BitsPerSortNum() int64 {
	return p.bitsPerSortNum
}

func (p *Params) MaskForHashIndex() uint64 {
	return p.maskLsbsForHashIndex
}

func (p *Params) MaskForSortNum() uint64 {
	return p.maskLsbsForSortNum
}

func lambdaSmallEnoughForForPoissionCumulativeExceedsPercentageAtXLimit(percentage float64, xLimit int64) (lambdaResult int64, percentAchieved float64) {
	// Start with lambda = xLimit to give about 50% percentage
	// (The peak of the Poisson distribution is positioned horizontally at the top limit xLimit,
	// with half to the left of xLimit, and half to the right,
	// tailing down towards zero at x=infinity and x=-infinity)
	lambda := xLimit
	fraction := float64(0.5) // 50%
	for fraction < percentage/100.0 {
		// Decrease lambda to "squish" distribution leftwards,
		// bringing more of the area under the distribution to the left of xLimit
		// (increasing fraction)
		lambda--
		// Fraction is the cumulative sum of the Poisson Distribution up to xLimit
		fraction = 0.0
		for x := int64(0); x < xLimit; x++ {
			fraction += poissonApproximation(float64(lambda), float64(x))
		}
	}
	lambdaResult = lambda
	percentAchieved = fraction * 100.0
	return lambdaResult, percentAchieved
}

func xLimitBigEnoughForForPoissionCumulativeExceedsPercentageAtX(lambda float64, percentage float64) (xLimitResult int64) {
	fraction := 0.0
	for x := int64(0); true; x++ {
		fraction += poissonApproximation(lambda, float64(x))
		if fraction >= percentage/100.0 {
			return x
		}
	}
	return 1
}
