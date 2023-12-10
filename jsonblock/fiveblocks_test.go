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

func TestGenesisHandle(t *testing.T) {
	aOneBlockChain := CreateOneBlockChain(&HardCodedBlockFetcher{}, "Temp_Testing\\JsonBlock")
	TestGenesisHandle_helper(aOneBlockChain, t)
}

func TestInvalidHandle(t *testing.T) {
	aOneBlockChain := CreateOneBlockChain(&HardCodedBlockFetcher{}, "Temp_Testing\\JsonBlock")
	TestInvalidHandle_helper(aOneBlockChain, t)
}
