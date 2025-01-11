package indexedhashes3

import "testing"

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
