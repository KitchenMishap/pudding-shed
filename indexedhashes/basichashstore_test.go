package indexedhashes

import (
	"crypto/sha256"
	"encoding/binary"
	"log"
	"os"
	"testing"
)

func TestNewBasicHashStore(t *testing.T) {
	file, _ := os.Create("Temp_Testing\\Test.hsh")
	defer file.Close()
	bhs := NewBasicHashStore(file)
	HelperHashStoreSmallTest(bhs, t)
}

func HelperHashStoreSmallTest(hs HashReadWriter, t *testing.T) {
	count, err := hs.CountHashes()
	if err != nil {
		t.Error("CountHashes() on new file should not give error")
	}
	if count != 0 {
		t.Error("Count of hashes in new file should be zero")
	}
	hash0 := Sha256{}
	hash0[0] = byte(100)
	count, err = hs.AppendHash(&hash0)
	if err != nil {
		t.Error("AppendHash() on new file should not give error")
	}
	if count != 0 {
		t.Error("AppendHash() on new file should return index of 0")
	}
	hash1 := Sha256{}
	hash1[0] = byte(101)
	count, err = hs.AppendHash(&hash1)
	if err != nil {
		t.Error("Second AppendHash() on new file should not give error")
	}
	if count != 1 {
		t.Error("Second AppendHash() on new file should return index of 1")
	}
	count, err = hs.CountHashes()
	if err != nil {
		t.Error("CountHashes() after two appends should not give error")
	}
	if count != 2 {
		t.Error("CountHashes() after two appends should return 2")
	}
	hashCheck := Sha256{}
	err = hs.GetHashAtIndex(0, &hashCheck)
	if err != nil {
		t.Error("GetHashAtIndex(0) should not give error")
	}
	if hashCheck != hash0 {
		t.Error("GetHashAtIndex(0) should give the first hash")
	}
	err = hs.GetHashAtIndex(1, &hashCheck)
	if err != nil {
		t.Error("GetHashAtIndex(1) should not give error")
	}
	if hashCheck != hash1 {
		t.Error("GetHashAtIndex(0) should give the second hash")
	}
	log.Println("Note: An EOF here is a PASS")
	err = hs.GetHashAtIndex(2, &hashCheck)
	if err == nil {
		t.Error("GetHashAtIndex(2) should give error")
	}
	index, err := hs.IndexOfHash(&hash1)
	if err != nil {
		t.Error("IndexOfHash(...) should not give error for an existing hash")
	}
	if index != 1 {
		t.Error("IndexOfHash(...) should give an index of 1")
	}
	index, err = hs.IndexOfHash(&hash0)
	if err != nil {
		t.Error("IndexOfHash(...) should not give error for an existing hash")
	}
	if index != 0 {
		t.Error("IndexOfHash(...) should give an index of 0")
	}
}

func HelperHashStoreBigTest(hs HashReadWriter, t *testing.T, testSize uint64) {
	count, err := hs.CountHashes()
	if err != nil {
		t.Error("CountHashes() on new file should not give error")
	}
	if count != 0 {
		t.Error("Count of hashes in new file should be zero")
	}

	// Store hashes that are hash of iterator
	iter := Sha256{}
	for i := uint64(0); i < testSize; i++ {
		binary.LittleEndian.PutUint64(iter[:], i)
		hash := Sha256{}
		HelperHashOfHash(&iter, &hash)
		index, err := hs.AppendHash(&hash)
		if err != nil {
			t.Error("AppendHash() should not give error")
		}
		if uint64(index) != i {
			t.Error("AppendHash() should return correct index")
		}
	}
	// Read hashes that are equal to iterator
	readHash := Sha256{}
	for i := uint64(0); i < testSize; i++ {
		iter := Sha256{}
		binary.LittleEndian.PutUint64(iter[:], i)
		hash := Sha256{}
		HelperHashOfHash(&iter, &hash)
		index, err := hs.IndexOfHash(&hash)
		if err != nil {
			t.Error("IndexOfHash() should not give error")
		}
		if uint64(index) != i {
			t.Error("IndexOfHash() should return correct index")
		}
		err = hs.GetHashAtIndex(int64(i), &readHash)
		if err != nil {
			t.Error("GetHashAtIndex() should not give error")
		}
		if readHash != hash {
			t.Error("GetHashAtIndex() should match hash of index in this test")
		}
	}
}

func HelperHashOfHash(in *Sha256, out *Sha256) {
	// Calculate the sha256 of a sha256
	h := sha256.New()

	h.Write((*in)[0:32])

	o := h.Sum(nil)
	for i := 0; i < len(o); i++ {
		(*out)[i] = o[i]
	}
}
