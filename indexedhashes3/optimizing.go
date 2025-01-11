package indexedhashes3

import "fmt"

// For graphing and optimizing suitable parameters

func graphGigabytes(bitsPerHashIndex int64, hashCountEstimate int64, minBytesPerBinStart int64, maxBytesPerBinStart int64) {
	//wholeBytesForHashIndex := bitsPerHashIndex / 8
	for bitsForSortNum := int64(16); bitsForSortNum <= 32; bitsForSortNum++ {
		// Assume bytes fully used
		//bytesForHashIndexSortNum := wholeBytesForHashIndex + extraBytes
		//bitsForSortNum := bytesForHashIndexSortNum*8 - bitsPerHashIndex

		bits := bitsPerHashIndex + bitsForSortNum
		bytesForHashIndexSortNum := ((bits - 1) / 8) + 1

		// DON'T Assume sortNum bits fully used
		for mult := 0.505; mult <= 1.0; mult += 0.005 {
			numberOfBins := int64(mult * float64((int64(1) << bitsForSortNum)))

			lambda := float64(hashCountEstimate) / float64(numberOfBins)
			// For lambda <= 20, Poission distribution wouldn't be adequately modelled by Normal distribution
			if lambda > 20 {
				for percentOverflows := 10.0; percentOverflows >= 0.01; percentOverflows /= 10.0 {

					entriesPerBinStart := xLimitBigEnoughForForPoissonCumulativeExceedsPercentageAtX(lambda, 100.0-percentOverflows)
					bytes, overflows := estimateBytes(hashCountEstimate, numberOfBins, 24+bytesForHashIndexSortNum,
						entriesPerBinStart, 2)

					//bytesPerBinStart := 2 + entriesPerBinStart*(bytesForHashIndexSortNum+24)
					bytesPerBinStart := 0 + entriesPerBinStart*(bytesForHashIndexSortNum+24)

					if bytesPerBinStart >= minBytesPerBinStart && bytesPerBinStart <= maxBytesPerBinStart {
						fmt.Println("Bins:", numberOfBins, "\t%Overflows:", percentOverflows, "\tBinStartEntries:", entriesPerBinStart, "\tBytesPerBinStart:", bytesPerBinStart, "\t", float64(bytes/100000000)/10.0, "GB", "\tOverflowFiles:", overflows)
					}
				}
			}
		}
	}
}

func estimateBytes(hashCountEstimate int64,
	numberOfBins int64, bytesPerBinEntry int64, entriesPerBinStart int64,
	bytesToCountBinEntries int64) (bytes int64, overflows int64) {
	bytes = int64(0)

	// First the binstarts file
	binBytes := int64(0)
	binBytes += bytesToCountBinEntries                // The count at the start of the bin
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
