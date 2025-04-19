package indexedhashes3

import "testing"

// Functions to optimize and print params for 16 Years of bitcoin
func Test16YearsAddressParams(t *testing.T) {
	// At time of writing, based on the printed outputs of this call, we choose:
	// numberOfBins = 16777216, entriesInBinStart = 236, bytesPerSortNum = 5.
	// (see Sensible16YearsAddressHashParams() where these values are set)
	// This gives IN THEORY a prediction of 16310 overflow files (0.1%) with 143.707 GB.
	// Watch this space for actual result!
	graphGigabytes(bitsFor4bilAddrs, addressesEstimate16Years)
}
func Test16YearsTransactionParams(t *testing.T) {
	// At time of writing, based on the printed outputs of this call, we choose:
	// numberOfBins = 16777216, entriesInBinStart = 101, bytesPerSortNum = 5.
	// (see Sensible16YearsTransactionHashParams() where these values are set)
	// This gives IN THEORY a prediction of 1325 overflow files (0.01%) with 60.599 GB.
	// Watch this space for actual result!
	graphGigabytes(bitsFor2bilTrans, transactionsEstimate16Years)
}
func Test16YearsBlockParams(t *testing.T) {
	// At time of writing, based on the printed outputs of this call, we choose:
	// numberOfBins = 65536, entriesInBinStart = 29, bytesPerSortNum = 6.
	// (see Sensible16YearsBlockHashParams() where these values are set)
	// This gives IN THEORY a prediction of 5 overflow files (0.01%) with 0.065 GB.
	graphGigabytes(bitsFor2milBlocks, blocksEstimate16Years)
}

// Functions to optimize and print params for 2 Years of bitcoin (for testing)
func Test2YearsAddressParams(t *testing.T) {
	// At time of writing, based on the printed outputs of this call, we choose:
	// numberOfBins = 65536, entriesInBinStart = 9, bytesPerSortNum = 6.
	// (see Sensible2YearsAddressHashParams() where these values are set)
	// This gives IN THEORY a prediction of 636 overflow files (1%) with 0.022 GB.
	// IN PRACTISE we end up with only 36 overflow files, with 0.020 GB.
	// We GUESS/HOPE that the difference is due to the re-use of some addresses in the actual blockchain.
	graphGigabytes(bitsFor269kAddrs, addressesEstimate2Years)
}
func Test2YearsTransactionParams(t *testing.T) {
	// At time of writing, based on the printed outputs of this call, we choose:
	// numberOfBins = 65536, entriesInBinStart = 10, bytesPerSortNum = 6.
	// (see Sensible2YearsTransactionHashParams() where these values are set)
	// This gives IN THEORY a prediction of 48 overflow files (0.1%) with 0.022 GB.
	// IN PRACTISE we find 38 overflows, with 0.022 GB!
	// We are VERY PLEASED with this accuracy of prediction!
	graphGigabytes(bitsFor220kTrans, transactionsEstimate2Years)
}
func Test2YearsBlockParams(t *testing.T) {
	// At time of writing, based on the printed outputs of this call, we choose:
	// Number of bins = 256, entriesInBinStart = 455 (quite high), bytesPerSortNum = 7.
	// (see Sensible2YearsTransactionHashParams() where these values are set)
	// This gives IN THEORY a prediction of 2 overflow files (1%) with 0.004 GB.
	// IN PRACTISE we find 5 overflows, with 0.004 GB!
	// AGAIN We are VERY PLEASED with this accuracy of prediction!
	graphGigabytes(bitsFor101kBlocks, blocksEstimate2Years)
}
