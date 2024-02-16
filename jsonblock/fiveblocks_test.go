package jsonblock

import (
	"bytes"
	"testing"
)

func TestBlocks(t *testing.T) {
	const expectedCount = 5
	count := HardCodedJsonBlockCount()
	if count != expectedCount {
		t.Error("Wrong number of blocks")
	}
	var sa [expectedCount]string
	var ja [expectedCount]*JsonBlockEssential
	for b := int64(0); b < count; b++ {
		sa[b] = HardCodedJsonBlock(b)
		var err error
		ja[b], err = parseJsonBlock([]byte(sa[b]))
		if err != nil {
			t.Error(err)
		}
		if int64(ja[b].J_height) != b {
			t.Error("block height wrong")
		}
		if len(ja[b].J_tx) < 1 {
			t.Error("block without any transactions")
		}
		if len(ja[b].J_tx[0].J_vout) < 1 {
			t.Error("coinbase transaction with no vouts")
		}
		encoded, err := encodeJsonBlock(ja[b])
		if err != nil {
			t.Error(err)
		}
		parsed, err := parseJsonBlock(encoded)
		if err != nil {
			t.Error(err)
		}
		reEncoded, err := encodeJsonBlock(parsed)
		if err != nil {
			t.Error(err)
		}
		if !bytes.Equal(encoded, reEncoded) {
			t.Error("encode/parse/encode doesn't give same result")
		}
	}
}

func TestNonEssentialInts(t *testing.T) {
	indexer := CreateOpenTransactionIndexerFiles("Temp_Testing\\JsonBlock\\Indexing")
	aOneBlockChain := CreateOneBlockChain(&HardCodedBlockFetcher{}, indexer)
	blockHandle := aOneBlockChain.GenesisBlock()
	for !blockHandle.IsInvalid() {
		block, err := aOneBlockChain.BlockInterface(blockHandle)
		if err != nil {
			t.Error("could not get BlockInterface from blockchain")
		}
		nonEssentialInts, err := block.NonEssentialInts()
		if err != nil {
			t.Error("could not get non-essential ints for block")
		}
		size, sizeExists := (*nonEssentialInts)["size"]
		if sizeExists == false {
			t.Error("size should exist")
		}
		if size <= 10 {
			t.Error("size should be more than 10")
		}
		count, err := block.TransactionCount()
		if err != nil {
			t.Error("could not get transaction count from block")
		}
		for tr := int64(0); tr < count; tr++ {
			transHandle, err := block.NthTransaction(tr)
			if err != nil {
				t.Error("could not get transaction handle from block")
			}
			trans, err := aOneBlockChain.TransInterface(transHandle)
			if err != nil {
				t.Error("could not get transaction interface")
			}
			nonEssentialInts, err := trans.NonEssentialInts()
			if err != nil {
				t.Error("could not get non-essential ints for transaction")
			}
			version, versionExists := (*nonEssentialInts)["version"]
			if versionExists == false {
				t.Error("version should exist")
			}
			if version != 1 {
				t.Error("version should be 1")
			}
		}
		blockHandle, err = aOneBlockChain.NextBlock(blockHandle)
		if err != nil {
			t.Error("could not get next block handle")
		}
	}
	indexer.Close()
}

func TestGenesisHandle(t *testing.T) {
	indexer := CreateOpenTransactionIndexerFiles("Temp_Testing\\JsonBlock\\Indexing")
	aOneBlockChain := CreateOneBlockChain(&HardCodedBlockFetcher{}, indexer)
	TestGenesisHandle_helper(aOneBlockChain, t)
	indexer.Close()
}

func TestInvalidHandle(t *testing.T) {
	indexer := CreateOpenTransactionIndexerFiles("Temp_Testing\\JsonBlock\\Indexing")
	aOneBlockChain := CreateOneBlockChain(&HardCodedBlockFetcher{}, indexer)
	TestInvalidHandle_helper(aOneBlockChain, t)
	indexer.Close()
}
