package indexedhashes3

import "fmt"

// For graphing and optimizing suitable parameters

func graphGigabytes(bitsPerHashIndex int64, hashCountEstimate int64) {
	const minLambda = float64(20)
	for bitsForSortNum := int64(8); bitsForSortNum <= 64-bitsPerHashIndex; bitsForSortNum++ {
		bits := bitsPerHashIndex + bitsForSortNum
		bytesForHashIndexSortNum := ((bits - 1) / 8) + 1

		// DON'T Assume sortNum bits fully used
		for mult := 1.0; mult <= 1.0; mult += 1.0 {
			sortNumsPerBin := int64(mult * float64(int64(1)<<bitsForSortNum))
			divider := sortNumsPerBin
			if divider%2 == 0 {
				numberOfBins := int64((uint64(1) << uint64(63)) / (uint64(divider) >> 1))
				if numberOfBins%2 == 0 && numberOfBins > 0 {
					lambda := float64(hashCountEstimate) / float64(numberOfBins)
					// For lambda <= 20, Poission distribution wouldn't be adequately modelled by Normal distribution
					if lambda > minLambda {
						for percentOverflows := 10.0; percentOverflows >= 0.000001; percentOverflows /= 10.0 {

							entriesPerBinStart := xLimitBigEnoughForForPoissonCumulativeExceedsPercentageAtX(lambda, 100.0-percentOverflows)
							bytes, overflows := estimateBytes(hashCountEstimate, numberOfBins, 24+bytesForHashIndexSortNum,
								entriesPerBinStart)

							bytesPerBinStart := entriesPerBinStart * (bytesForHashIndexSortNum + 24)

							if entriesPerBinStart > 0 {
								fmt.Println("numberOfBins:", numberOfBins, "\tentriesInBinStart:", entriesPerBinStart, "\tbytesPerBinEntry:", bytesPerBinStart/entriesPerBinStart, "\t%Overflows:", percentOverflows, "\tOverflowFiles:", overflows, "\tBytesPerBinStart:", bytesPerBinStart, "\t", float64(bytes/1000000)/1000.0, "GB")
							}
						}
					} else {
						// Min lambda not met. Many items unknown, but number of bins still useful
						entriesPerBinStart := int64(10) // Arbitrary small number
						bytesPerBinStart := entriesPerBinStart * (bytesForHashIndexSortNum + 24)
						fmt.Println("numberOfBins:", numberOfBins, "\tentriesInBinStart ARBITRARY:", entriesPerBinStart, "\tbytesPerBinEntry:", bytesPerBinStart/entriesPerBinStart, "\t%Overflows: ???", "\tOverflowFiles: ???", "\tBytesPerBinStart: ", bytesPerBinStart, "\t???GB")
					}
				}
			}
		}
	}
}

func estimateBytes(hashCountEstimate int64,
	numberOfBins int64, bytesPerBinEntry int64, entriesPerBinStart int64) (bytes int64, overflows int64) {
	if numberOfBins%2 == 1 {
		panic("numberOfBins must be even")
	}
	//divider := (uint64(1) << uint64(63)) / (uint64(numberOfBins) >> uint64(1))

	bytes = int64(0)

	// First the binstarts file
	binBytes := int64(0)
	binBytes += entriesPerBinStart * bytesPerBinEntry // The bin entries
	bytes += binBytes * numberOfBins                  // Multiply by number of bins

	// Then the overflow files
	averageBinEntriesPerBin := float64(hashCountEstimate) / float64(numberOfBins)
	lambda := averageBinEntriesPerBin
	overflows = 0
	for overflow := int64(1); overflow <= 20; overflow++ {
		// Chance of any particular bin overflowing by 'overflow' entries
		chance := poissonApproximation(lambda, float64(entriesPerBinStart+overflow))
		// Likely number of bins to overflow by 'overflow' entries
		likely := int64(chance * float64(numberOfBins))
		overflows += likely
		// file size of such an overflow file
		size := overflow * bytesPerBinEntry
		allocationSize := int64(4096)
		sizeOnDisk := (int64(float64(size)/float64(allocationSize)) + 1) * allocationSize
		bytes += sizeOnDisk * likely
	}

	// Then the (new) hash index to bin index file
	num := numberOfBins
	var bytesNeeded int64
	for bytesNeeded = 1; true; bytesNeeded++ {
		num >>= 8
		if num == 0 {
			break
		}
	}
	bytes += hashCountEstimate * bytesNeeded

	return bytes, overflows
}
