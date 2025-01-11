package indexedhashes3

import "testing"

// Functions to optimize and print params for 16 Years of bitcoin
func Test16YearsAddressParams(t *testing.T) {
	graphGigabytes(bitsFor4bilAddrs, addressesEstimate16Years, 4096, 4096)
}
func Test16YearsTransactionParams(t *testing.T) {
	// We don't get any 4096 byte results, so we use 4092 (132 * 31 bytes)
	graphGigabytes(bitsFor2bilTrans, transactionsEstimate16Years, 4092, 4096)
}
func Test16YearsBlockParams(t *testing.T) {
	// We don't get anywhere near 4096 byte results, so we use 1015 (35 * 29 bytes)
	graphGigabytes(bitsFor2milBlocks, blocksEstimate16Years, 1000, 1024)
}

// Functions to optimize and print params for 2 Years of bitcoin (for testing)
func Test2YearsAddressParams(t *testing.T) {
	// We don't get any 4096 byte results, so we use 4092 (132 * 31 bytes)
	graphGigabytes(bitsFor200milAddrs, addressesEstimate2Years, 4092, 4096)
}
func Test2YearsTransactionParams(t *testing.T) {
	// We don't get any 4096 byte results, so we use 4092 (132 * 31 bytes)
	graphGigabytes(bitsFor100milTrans, transactionsEstimate2Years, 4050, 4096)
}
func Test2YearsBlockParams(t *testing.T) {
	graphGigabytes(bitsFor100ThouBlocks, blocksEstimate2Years, 1550, 1600)
}
