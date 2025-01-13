package indexedhashes3

import (
	"fmt"
	"math"
)

// For graphing and optimizing suitable parameters

func graphGigabytes(bitsPerHashIndex int64, hashCountEstimate int64) {
	const minLambda = float64(20)
	for bitsForSortNum := int64(16); bitsForSortNum <= 63; bitsForSortNum += 8 {
		bits := bitsPerHashIndex
		bytesForHashIndex := ((bits - 1) / 8) + 1
		bits = bitsForSortNum
		bytesForSortNum := ((bits - 1) / 8) + 1
		bytesPerBinEntry := 24 + bytesForSortNum + bytesForHashIndex

		sortNumsPerBin := int64(1) << bitsForSortNum
		divider := sortNumsPerBin
		if divider%2 == 0 {
			numberOfBins := int64((uint64(1) << uint64(63)) / (uint64(divider) >> 1))
			if numberOfBins%2 == 0 && numberOfBins > 0 {
				lambda := float64(hashCountEstimate) / float64(numberOfBins)
				// For lambda <= 20, Poission distribution wouldn't be adequately modelled by Normal distribution
				if lambda > minLambda {
					for percentOverflows := 1.0; percentOverflows >= 0.01; percentOverflows /= 10.0 {
						entriesPerBinStart := xLimitBigEnoughForForPoissonCumulativeExceedsPercentageAtX(lambda, 100.0-percentOverflows)
						bytes, overflows := estimateBytes(hashCountEstimate, numberOfBins, bytesPerBinEntry, entriesPerBinStart)

						bytesPerBinStart := entriesPerBinStart * bytesPerBinEntry

						if entriesPerBinStart > 0 {
							fmt.Println("numberOfBins:", numberOfBins, "\tentriesInBinStart:", entriesPerBinStart, "bytesPerSortNum:", bitsForSortNum/8, "\tbytesPerBinEntry:", bytesPerBinStart/entriesPerBinStart, "\t%Overflows:", percentOverflows, "\tOverflowFiles:", overflows, "\tBytesPerBinStart:", bytesPerBinStart, "\t", float64(bytes/1000000)/1000.0, "GB")
						}
					}
				} else {
					// Min lambda not met. Many items unknown, but number of bins still useful
					fmt.Println("Number of bins:", numberOfBins, "Lambda too small:", math.Round(lambda*10.0)/10.0)
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
