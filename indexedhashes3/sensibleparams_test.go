package indexedhashes3

import "testing"

func Test16YearsAddressParams(t *testing.T) {
	graphGigabytes(bitsFor4bilAddrs, addressesEstimate16Years, 4096, 4096)
}
func Test16YearsTransactionParams(t *testing.T) {
	graphGigabytes(bitsFor2bilTrans, transactionsEstimate16Years, 4092, 4096)
}
