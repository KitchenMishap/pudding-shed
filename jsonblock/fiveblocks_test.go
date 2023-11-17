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
	var ja [expectedCount]*jsonBlockEssential
	for b := int64(0); b < count; b++ {
		sa[b] = HardCodedJsonBlock(b)
		var err error
		ja[b], err = parseJsonBlock([]byte(sa[b]))
		if err != nil {
			t.Error(err)
		}
		if ja[b].Height != b {
			t.Error("block height wrong")
		}
		if len(ja[b].Hash) != 64 {
			t.Error("block hash should be 64 chars")
		}
		if len(ja[b].Tx) < 1 {
			t.Error("block without any transactions")
		}
		if len(ja[b].Tx[0].Txid) != 64 {
			t.Error("transaction id should be 64 chars")
		}
		if len(ja[b].Tx[0].Vout) < 1 {
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
