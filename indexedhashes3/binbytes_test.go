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

func TestInsert12(t *testing.T) {
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

	if b.insertBinEntry(sn, hi, &tr, params) != 0 {
		t.Error("first insertion didn't return 0")
	}
	if len(b) != 1 {
		t.Error("inserted bin, size not 1")
	}
	hi2, sn2 := b[0].getHashIndexSortNum(params)
	if hi2 != hi {
		t.Error("inserted bin entry, hashIndex doesn't match")
	}
	if sn2 != sn {
		t.Error("inserted bin entry, sortNum doesn't match")
	}

	hi = hashIndex(12346)
	sn = sortNum(23457)

	if b.insertBinEntry(sn, hi, &tr, params) != 1 {
		t.Error("second insertion didn't return 1")
	}
	if len(b) != 2 {
		t.Error("inserted 2nd bin, size not 2")
	}

	ind := b.findIndexBasedOnSortNum(sn, params)
	if ind != 1 {
		t.Error("inserted two entries, findIndexBasedOnSortNum not 1")
	}

	n := b.lookupByHash(&tr, sn, params)
	if n != hi {
		t.Error("inserted bin, lookup not same")
	}
}
