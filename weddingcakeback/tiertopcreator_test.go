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

	config := NewCakeConfig(32)

	creator := NewTierTopCreator(testDir, config)
	if creator.Exists() {
		t.Fatal("Creator should not have a tier zero yet")
	}

	err = creator.Create(0)
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

	const count = 65534
	const masterSeed = 42

	presentationArray := make([][]byte, count)
	for i := range int64(count) {
		hash := helperDeterministicHash(32, masterSeed, i)
		presentationArray[i] = hash
		index, err := tierTop.AppendHash(hash)
		if err != nil {
			t.Fatal(err)
		}
		if index != i {
			t.Fatal("Hash index mismatch")
		}
	}

	for i := range int64(count) {
		hash := presentationArray[i]
		presentationIndexRecovered, found, err := tierTop.TryIndexOfHash(hash)
		if err != nil {
			t.Fatal(err)
		}
		if !found {
			t.Error("Lookup failed, returned found=false")
		} else if presentationIndexRecovered == -1 {
			t.Error("Lookup failed, returned -1")
		} else if !bytes.Equal(presentationArray[presentationIndexRecovered][:], hash[:]) {
			t.Error("Lookup failed, returned index of wrong hash")
		} else {
			//fmt.Println("A success")
		}
	}
	randomHash := helperRandomHash(32)
	_, found, err := tierTop.TryIndexOfHash(randomHash)
	if err != nil {
		t.Fatal(err)
	} else if found {
		fmt.Printf("Random hash returned a match. Surprising? Maybe? But false positives need to be filtered by caller\n")
	}
}

// helperDeterministicHash generates a reproducible pseudo-random byte slice based on a seed and an index.
func helperDeterministicHash(hashLength int, seed int64, counter int64) []byte {
	// Create a unique source for this specific hash to ensure variety across the loop,
	// while keeping the sequence perfectly locked to the master seed.
	rng := rand.New(rand.NewSource(seed + counter))

	result := [64]byte{}
	for i := 0; i < hashLength; i++ {
		result[i] = byte(rng.Intn(256))
	}
	return result[:hashLength]
}

func helperRandomHash(hashLength int) []byte {
	result := [64]byte{}
	for i := range hashLength {
		result[i] = byte(rand.Intn(256))
	}
	return result[:hashLength]
}
