package indexedhashes3

import "testing"

func TestRoundTrip(t *testing.T) {
	tr1 := [24]byte{0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xAA, 0xBB, 0xCC, 0xDD, 0xEE, 0xFF,
		0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77}
	trBytes := tr1
	tr := truncatedHash(trBytes)
	hi := hashIndex(12345)
	sn := sortNum(23456)
	params := Sensible2YearsBlockHashParams()
	entry := newBinEntryBytes(&tr, hi, sn, params)

	if !entry.getTruncatedHash().equals(&tr) {
		t.Error("Truncated hash doesn't match")
	}
	hi2, sn2 := entry.getHashIndexSortNum(params)
	if hi2 != hashIndex(12345) {
		t.Error("Hash Index doesn't match")
	}
	if sn2 != sortNum(23456) {
		t.Error("Sort Number doesn't match")
	}
}
