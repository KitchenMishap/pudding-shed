package indexedhashes3

import (
	"fmt"
	"math"
)

type Params struct {
	bytesPerHashIndex int64
	bitsPerBinNum     int64
}

func NewHashStoreParams(storeName string, bytesPerHashIndex int64, hashCountEstimate int64, bytesPerBin int64, maxPercentBinsInOverflows float64) *Params {
	params := OptimizeAndInitializeParams(storeName, bytesPerHashIndex, hashCountEstimate, bytesPerBin, maxPercentBinsInOverflows)
	return params
}

// Note that whole bytes are not enough bits to specify a BinNum
func (p *Params) wholeBytesInBinNum() int64 {
	return p.bitsPerBinNum / 8
}
func wholeBytesInBinNumPrivate(bitsPerBinNum int64) int64 {
	return bitsPerBinNum / 8
}
func bytesInTruncatedHashPrivate(bitsPerBinNum int64) int64 {
	return 32 - wholeBytesInBinNumPrivate(bitsPerBinNum)
}

func (p *Params) bytesInTruncatedHash() int64 {
	return 32 - p.wholeBytesInBinNum()
}

func IterateSmallestSize(hashCountEstimate int64, bytesPerHashIndex int64, percent float64, from int64, to int64, step int64) {
	var minBytes int64
	var minBytesSize int64
	bytesForBinEntryCount := int64(2) // Must be big enough to contain the biggest bin entry count a bin will ever encounter
	for bytesPerBin := from; bytesPerBin <= to; bytesPerBin += step {
		bitsPerBinNum, _, _, _, _ := iterateToOptimizeParams(hashCountEstimate, bytesPerHashIndex, bytesPerBin, bytesForBinEntryCount, percent)
		bytes := bytesPerBin * (1 << bitsPerBinNum)
		if bytesPerBin == from || bytes < minBytes {
			minBytes = bytes
			minBytesSize = bytesPerBin
		}
	}
	giga := float64(minBytes) / 1024.0 / 1024.0 / 1024.0
	fmt.Println("For BytesPerBin=", minBytesSize, ", BinStarts file size:", math.Round(giga*100.0)/100.0, "Gb")
	return

}

func OptimizeAndInitializeParams(storeName string, hashCountEstimate int64, bytesPerHashIndex int64, bytesPerBin int64, maxPercentBinsInOverflows float64) *Params {
	p := Params{}
	p.bytesPerHashIndex = bytesPerHashIndex

	bytesForBinEntryCount := 2 // Must be big enough to contain the biggest bin entry count a bin will ever encounter
	bitsPerBinNum, bytesPerBinEntry, targetEntryCountEachBin, fullEntryCountEachBin, percentAchieved := iterateToOptimizeParams(hashCountEstimate, bytesPerHashIndex, int64(bytesPerBin), int64(bytesForBinEntryCount), maxPercentBinsInOverflows)

	fmt.Println("--------------------------")
	fmt.Println("Hashes Indexing Store for: ", storeName)
	fmt.Println(fullEntryCountEachBin, " x ", bytesPerBinEntry, " bytes")
	fullness := float64(2+targetEntryCountEachBin*bytesPerBinEntry) / float64(bytesPerBin)
	fullnessPercent := fullness * 100.0
	fmt.Println("Target bin fullness: ", math.Round(fullnessPercent*10)/float64(10), "%")
	fmt.Println("Percentage of bins overflowing: ", math.Round((100.0-percentAchieved)*1000)/float64(1000), "%")
	bytes := bytesPerBin * (1 << bitsPerBinNum)
	giga := float64(bytes) / 1024.0 / 1024.0 / 1024.0
	fmt.Println("BinStarts file size:", math.Round(giga*10.0)/10.0, "Gb")

	p.bitsPerBinNum = bitsPerBinNum
	return &p
}

func iterateToOptimizeParams(hashCountEstimate int64, bytesPerHashIndex int64, bytesPerBin int64, bytesForBinEntriesCount int64, maxPercentBinsInOverflows float64) (bitsPerBinNum int64, bytesPerBinEntry int64, targetEntryCountEachBin int64, fullEntryCountEachBin int64, percentAchieved float64) {
	bytesAvailableForEntriesInBin := bytesPerBin - bytesForBinEntriesCount

	// Start by assuming truncated hashes are full hashes, and that we aim to have 50% (!) of entries in overflows
	bytesPerTruncatedHash := int64(32)
	// ### FIRST ESTIMATE FOR FULL BINS ###
	bytesPerBinEntry = bytesPerTruncatedHash + bytesPerHashIndex
	numEntriesInFullBin := bytesAvailableForEntriesInBin / bytesPerBinEntry // Integer divide
	// How many bins do we end up with?
	binCount := hashCountEstimate / numEntriesInFullBin
	// How many bits do we need to index this many bins?
	bitsPerBinNum = powerOfTwoExponentToAccommodateMaxVal(binCount)
	// We can now reduce from 32 bit hashes to some sort of truncated hash
	bytesPerTruncatedHash = bytesInTruncatedHashPrivate(bitsPerBinNum)

	// Now do another iteration of the above...
	// ### SECOND ESTIMATE FOR FULL BINS ###
	bytesPerBinEntry = bytesPerTruncatedHash + bytesPerHashIndex
	numEntriesInFullBin = bytesAvailableForEntriesInBin / bytesPerBinEntry // Integer divide

	// So far (as a first estimate) our target was to fill the bins.
	// So the target percentage of bin entries ending up in overflow bins was effectively 50%.
	overflowPercentageTarget := float64(50)
	// Now we will use our actual target percentage
	overflowPercentageTarget = maxPercentBinsInOverflows
	percentageTarget := float64(100) - overflowPercentageTarget
	// Find a lambda (the target for the average count of entries in a bin) that achieves this percentage
	// ### FIRST "Partial" BIN ESTIMATE ###
	lambda, percentAchieved := lambdaSmallEnoughForForPoissionCumulativeExceedsPercentageAtXLimit(percentageTarget, numEntriesInFullBin)
	numEntriesInAverageBin := lambda
	// This results in a different number of bins...
	binCount = hashCountEstimate / numEntriesInAverageBin
	// How many bits do we need to index this many bins?
	bitsPerBinNum = powerOfTwoExponentToAccommodateMaxVal(binCount)
	// And a new size for truncated hashes
	bytesPerTruncatedHash = bytesInTruncatedHashPrivate(bitsPerBinNum)

	// ### ANOTHER ESTIMATE FOR FULL BINS ###
	bytesPerBinEntry = bytesPerTruncatedHash + bytesPerHashIndex
	numEntriesInFullBin = bytesAvailableForEntriesInBin / bytesPerBinEntry // Integer divide

	// ### AND OUR FINAL ESTIMATE FOR PARTIAL BINS ###
	lambda, percentAchieved = lambdaSmallEnoughForForPoissionCumulativeExceedsPercentageAtXLimit(percentageTarget, numEntriesInFullBin)
	numEntriesInAverageBin = lambda
	// This results in a different number of bins...
	binCount = hashCountEstimate / numEntriesInAverageBin
	// How many bits do we need to index this many bins?
	bitsPerBinNum = powerOfTwoExponentToAccommodateMaxVal(binCount)
	// And a new size for truncated hashes
	bytesPerTruncatedHash = bytesInTruncatedHashPrivate(bitsPerBinNum)

	// THE RESULTS
	targetEntryCountEachBin = numEntriesInAverageBin
	fullEntryCountEachBin = numEntriesInFullBin
	return bitsPerBinNum, bytesPerBinEntry, targetEntryCountEachBin, fullEntryCountEachBin, percentAchieved
}

func powerOfTwoExponentToAccommodateMaxVal(maxVal int64) int64 {
	exponent := int64(0)
	powerOfTwo := int64(1) << exponent
	for maxVal > powerOfTwo {
		exponent++
		powerOfTwo = int64(1) << exponent
	}
	return exponent
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

func poissonApproximation(lambda float64, x float64) float64 {
	// Normal distribution with the following parameters mu and sigma,
	// is a usable approximation to Poisson distribution for lambda > 20
	mu := lambda
	sigma := math.Sqrt(lambda)
	return normalDistribution(mu, sigma, x)
}

func normalDistribution(mu float64, sigma float64, x float64) float64 {
	sigmaSquared := sigma * sigma
	return math.Exp(-(x-mu)*(x-mu)/(2*sigmaSquared)) / math.Sqrt(2*math.Pi*sigmaSquared)
}
