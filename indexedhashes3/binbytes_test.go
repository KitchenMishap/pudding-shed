package indexedhashes3

import "testing"

func TestInsert(t *testing.T) {
	b := newEmptyBin()
	if len(b) != 0 {
		t.Error("new bin not empty")
	}

	tr1 := [24]byte{0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xAA, 0xBB, 0xCC, 0xDD, 0xEE, 0xFF,
		0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77}
	trBytes := tr1
	tr := truncatedHash(trBytes)
	hi := hashIndex(12345)
	sn := sortNum(23456)
	params := Sensible2YearsBlockHashParams()

	b.insertBinEntry(sn, hi, &tr, params)
	if len(b) != 1 {
		t.Error("inserted bin, size not 1")
	}

	ind := b.findIndexBasedOnSortNum(sn, params)
	if ind != 0 {
		t.Error("inserted bin, findIndexBasedOnSortNum not 0")
	}

	n := b.lookupByHash(&tr, sn, params)
	if n != hi {
		t.Error("inserted bin, lookup not same")
	}
}
