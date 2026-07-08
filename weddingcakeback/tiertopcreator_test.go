package weddingcakeback

import (
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
)

func TestTierTopCreator(t *testing.T) {

	// 1. Completely wipe and recreate the testing directory to clear out stale files
	testDir := filepath.Join("Temp_Testing")
	_ = os.RemoveAll(testDir) // Ignore error if it doesn't exist yet
	err := os.MkdirAll(testDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	creator := NewTierTopCreator(testDir)
	if creator.Exists() {
		t.Fatal("Creator should not have a tier zero yet")
	}

	err = creator.Create()
	if err != nil {
		t.Fatal(err)
	}
	if !creator.Exists() {
		t.Fatal("Tier zero should now exist")
	}

	tierTop, err := creator.Open()
	if err != nil {
		t.Fatal(err)
	}

	const count = 65535
	const masterSeed = 42

	presentationArray := make([]Sha256, count)
	for i := range int64(count) {
		hash := helperDeterministicHashSha256(masterSeed, i)
		presentationArray[i] = hash
		index, err := tierTop.AppendHash(&hash)
		if err != nil {
			t.Fatal(err)
		}
		if index != i {
			t.Fatal("Hash index mismatch")
		}
	}

	for i := range int64(count) {
		hash := presentationArray[i]
		presentationIndexRecovered, err := tierTop.IndexOfHash(&hash)
		if err != nil {
			t.Fatal(err)
		}
		if presentationIndexRecovered == -1 {
			t.Error("Lookup failed, returned -1")
		} else if !bytes.Equal(presentationArray[presentationIndexRecovered][:], hash[:]) {
			t.Error("Lookup failed, returned index of wrong hash")
		} else {
			//fmt.Println("A success")
		}
	}
	randomHash := helperRandomHashSha256()
	presentationIndex, err := tierTop.IndexOfHash(&randomHash)
	if err != nil {
		t.Fatal(err)
	}
	if presentationIndex != -1 {
		fmt.Printf("Random hash returned a match. Surprising? Maybe? But false positives need to be filtered by caller\n")
	}

}

// helperDeterministicHash generates a reproducible pseudo-random byte slice based on a seed and an index.
func helperDeterministicHashSha256(seed int64, counter int64) Sha256 {
	// Create a unique source for this specific hash to ensure variety across the loop,
	// while keeping the sequence perfectly locked to the master seed.
	rng := rand.New(rand.NewSource(seed + counter))

	result := [32]byte{}
	for i := 0; i < 32; i++ {
		result[i] = byte(rng.Intn(256))
	}
	return result
}

func helperRandomHashSha256() Sha256 {
	result := [32]byte{}
	for i := range 32 {
		result[i] = byte(rand.Intn(256))
	}
	return result
}
