package indexedhashes

import "testing"

func TestConcreteHashStoreCreator(t *testing.T) {
	useMemFileForLookup := false

	// Test with smaller partial hash bit count (to encourage many more collisions and code coverage)
	// And fewer collisions per chunk setting
	HelperTestConcreteHashStoreCreator(t, 20, 3, 2, 100000, useMemFileForLookup)
	println("Finished short intensive test (1 of 3)")

	// Test with parameters representative of block hashes
	HelperTestConcreteHashStoreCreator(t, 30, 3, 3, 30000, useMemFileForLookup)
	//HelperTestConcreteHashStoreCreator(t, 30, 3, 3, 1000000)	// Passes but takes time and space
	println("Finished block emulation test (2 of 3)")

	// Test with parameters representative of transaction hashes (so more space required)
	HelperTestConcreteHashStoreCreator(t, 30, 4, 3, 20000, useMemFileForLookup)
	//HelperTestConcreteHashStoreCreator(t, 30, 4, 3, 10000000)	// Passes but takes time and space
	println("Finished transaction emulation test (3 of 3)")
}

func HelperTestConcreteHashStoreCreator(t *testing.T, phbc int64, ebc int64, cpc int64, iters uint64, useMemFile bool) {
	var abstractCreator HashStoreCreator

	// Create the creator
	abstractCreator, err := NewConcreteHashStoreCreator("Test", "Temp_Testing", phbc, ebc, cpc, useMemFile)
	if err != nil {
		t.Error("NewConcreteHashStoreCreator() returned error")
	}

	// Create the hash store
	err = abstractCreator.CreateHashStore()
	if err != nil {
		t.Error("CreateHashStore() returned error")
	}

	exists := abstractCreator.HashStoreExists()
	if !exists {
		t.Error("HashStoreExists() returned false after creating")
	}

	var abstractStore HashReadWriter
	abstractStore, err = abstractCreator.OpenHashStore()
	if err != nil {
		t.Error("OpenHashStore() returned error")
	}

	HelperHashStoreBigTest(abstractStore, t, iters)
}
