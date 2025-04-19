package indexedhashes3

// Some sensible bitsPerHashIndex for 15 or 16 years, and for 2 years (for testing)

const digitsPerNumberedFolder = 2

// --------------------------------------------------------------------------
// 15 or 16 years
//
// ------------------------------------------------------------------------- Count at 15 years:
const bitsFor2milBlocks = int64(21) // 2^21 = 2,097,152 blocks				There were 824,204 blocks
const bitsFor2bilTrans = int64(31)  // 2^31 = 2,147,483,648 transactions	There were 947,337,057 transactions
const bitsFor4bilAddrs = int64(32)  // 2^32 = 4,294,967,296 addresses		There must be fewer addresses than txos
// (There were 2,652,374,369 txos, including spent)
const addressesEstimate16Years = int64(3244977536)    // Nased on the size of Hashes.hsh (/32) after 888888 blocks (so an overestimate)
const transactionsEstimate16Years = int64(1169006653) // Based on the size of Hashes.hsh (/32) after 888888 blocks
const blocksEstimate16Years = int64(888888)

func Sensible16YearsAddressHashParams() *HashIndexingParams {
	// Run test Test16YearsAddressParams() to see how these numbers are arrived at
	return NewHashStoreParams(
		bitsFor4bilAddrs,         // bitsPerHashIndex
		addressesEstimate16Years, // hashCountEstimate
		digitsPerNumberedFolder,  // digitsPerNumberedFolder
		16777216,                 // numberOfBins (result of some optimization calculations)
		236,                      // entriesInBinStart (result of some optimization calculations)
		5)                        // bytesPerSortNum (result of some optimization calculations)
}

func Sensible16YearsTransactionHashParams() *HashIndexingParams {
	// Run test Test16YearsTransactionParams() to see how these numbers are arrived at
	return NewHashStoreParams(
		bitsFor2bilTrans,            // bitsPerHashIndex
		transactionsEstimate16Years, // hashCountEstimate
		digitsPerNumberedFolder,     // digitsPerNumberedFolder
		16777216,                    // numberOfBins (result of some optimization calculations)
		101,                         // entriesInBinStart (result of some optimization calculations)
		5)                           // bytesPerSortNum (result of some optimization calculations)
}

func Sensible16YearsBlockHashParams() *HashIndexingParams {
	// Run test Test16YearsBlockParams() to see how these numbers are arrived at
	return NewHashStoreParams(
		bitsFor2bilTrans,            // bitsPerHashIndex
		transactionsEstimate16Years, // hashCountEstimate
		digitsPerNumberedFolder,     // digitsPerNumberedFolder
		65536,                       // numberOfBins (result of some optimization calculations)
		29,                          // entriesInBinStart (result of some optimization calculations)
		6)                           // bytesPerSortNum (result of some optimization calculations)
}

// 2 years (for testing)
// ------------------------------------------------------------------------- Count at 2 years:
const bitsFor101kBlocks = int64(17) // 2^17 = 131,072 blocks (so room for more then the actual 100888 blocks)
const bitsFor220kTrans = int64(18)  // 2^18 = 262,144 transactions (so room for more than the actual 219927)
const bitsFor269kAddrs = int64(19)  // 2^19 = 524,288 addresses (so room from more than the actual 269406)

const blocksEstimate2Years = int64(100888)       // There were actually 100888 blocks in the first two years
const transactionsEstimate2Years = int64(219927) // Based on the size of Hashes.hsh file (/32) for 2 years
const addressesEstimate2Years = int64(269406)    // Based on size of Hashes.hsh file (/32) for 2 years
// This is an OVERESTIMATE due to repeated address use

func Sensible2YearsAddressHashParams() *HashIndexingParams {
	// Run test Test2YearsAddressParams() to see how these numbers are arrived at
	return NewHashStoreParams(
		bitsFor269kAddrs,        // bitsPerHashIndex
		addressesEstimate2Years, // hashCountEstimate
		digitsPerNumberedFolder, // digitsPerNumberedFolder
		65536,                   // numberOfBins (result of some optimization calculations)
		9,                       // entriesInBinStart (result of some optimization calculations)
		6)                       // bytesPerBinEntry (result of some optimization calculations)
}

func Sensible2YearsTransactionHashParams() *HashIndexingParams {
	// Run test Test2YearsTransactionParams() to see how these numbers are arrived at
	return NewHashStoreParams(
		bitsFor220kTrans,           // bitsPerHashIndex
		transactionsEstimate2Years, // hashCountEstimate
		digitsPerNumberedFolder,    // digitsPerNumberedFolder
		65536,                      // numberOfBins (result of some optimization calculations)
		10,                         // entriesInBinStart (result of some optimization calculations)
		6)                          // bytesPerBinEntry (result of some optimization calculations)
}

func Sensible2YearsBlockHashParams() *HashIndexingParams {
	// Run test Test2YearsBlockParams() to see how these numbers are arrived at
	return NewHashStoreParams(
		bitsFor101kBlocks,       // bitsPerHashIndex
		blocksEstimate2Years,    // hashCountEstimate
		digitsPerNumberedFolder, // digitsPerNumberedFolder
		256,                     // numberOfBins (result of some optimization calculations)
		440,                     // entriesInBinStart (result of some optimization calculations)
		7)                       // bytesPerBinEntry (result of some optimization calculations)
}
