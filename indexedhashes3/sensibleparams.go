package indexedhashes3

// Some sensible bitsPerHashIndex for 15 or 16 years, and for 2 years (for testing)

// --------------------------------------------------------------------------
// 15 or 16 years
//
//	Count at 15 years:
const bitsFor2milBlocks = int64(21) // 2^21 = 2,097,152 blocks				There were 824,204 blocks
const bitsFor2bilTrans = int64(31)  // 2^31 = 2,147,483,648 transactions	There were 947,337,057 transactions
const bitsFor4bilAddrs = int64(32)  // 2^32 = 4,294,967,296 addresses		There must be fewer addresses than txos
// (There were 2,652,374,369 txos, including spent)
const addressesEstimate16Years = int64(3000000000)
const transactionsEstimate16Years = int64(1000000000)
const blocksEstimate16Years = int64(1000000)

func Sensible16YearAddressHashParams() *HashIndexingParams {
	// Run test Test16YearsAddressParams() to see how these numbers are arrived at
	// The following values give:
	// BytesPerBinStart = 4096 (good as file accesses will be aligned to hard disk allocation unit boundaries)
	// OverflowFilesEstimate = 28137 (good, a nice balance as a million or more files take a very long time to copy)
	// 0.1% of bins resorting to an overflow file
	// Estimated hash store size on disk = 137.9GB (good, a nice balance. The minumum found for 4096 bytes was
	//   not much lower at 129.6GB, but had 2 million overflow files)
	return NewHashStoreParams(
		bitsFor4bilAddrs,         // bitsPerHashIndex
		addressesEstimate16Years, // hashCountEstimate
		2,                        // digitsPerNumberedFolder
		30702305,                 // numberOfBins (result of some optimization calculations)
		128,                      // entriesInBinStart (result of some optimization calculations)
		32)                       // bytesPerBinEntry (result of some optimization calculations)
}

func Sensible16YearsTransactionHashParams() *HashIndexingParams {
	// Run test Test16YearsTransactionParams() to see how these numbers are arrived at
	// The following values give:
	// BytesPerBinStart = 4092 (might be a good idea to pad to file allocation unit 4096?!)
	// OverflowFilesEstimate = 8560 (good, a nice balance as a million or so files take a very long time to copy)
	// 0.1% of bins resorting to an overflow file
	// Estimated hash store size on disk = 43.5GB (good, a nice balance. The minumum found for 4092 bytes was
	//   not much lower at 40.8GB, but had nearly a million overflow files)
	return NewHashStoreParams(
		bitsFor2bilTrans,            // bitsPerHashIndex
		transactionsEstimate16Years, // hashCountEstimate
		2,                           // digitsPerNumberedFolder
		9898557,                     // numberOfBins (result of some optimization calculations)
		132,                         // entriesInBinStart (result of some optimization calculations)
		31)                          // bytesPerBinEntry (result of some optimization calculations)
}

func Sensible16YearsBlockHashParams() *HashIndexingParams {
	// Run test Test16YearsBlockParams() to see how these numbers are arrived at
	// The following values give:
	// BytesPerBinStart = 1015 (didn't get anywhere near 4096)
	// OverflowFilesEstimate = 238 (good, a nice balance as a million or so files take a very long time to copy)
	// 1% of bins resorting to an overflow file
	// Estimated hash store size on disk = 0GB (good, insignificant)
	return NewHashStoreParams(
		bitsFor2bilTrans,            // bitsPerHashIndex
		transactionsEstimate16Years, // hashCountEstimate
		2,                           // digitsPerNumberedFolder
		42926,                       // numberOfBins (result of some optimization calculations)
		35,                          // entriesInBinStart (result of some optimization calculations)
		29)                          // bytesPerBinEntry (result of some optimization calculations)
}

// 2 years (for testing)
// ToDo
