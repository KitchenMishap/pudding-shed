package weddingcake

import (
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
)

func TestTierZero(t *testing.T) {
	// 1. Completely wipe and recreate the testing directory to clear out stale LevelXX.bin files
	testDir := filepath.Join("Temp_Testing")
	_ = os.RemoveAll(testDir) // Ignore error if it doesn't exist yet
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatal(err) // Stop execution immediately if environment setup fails
	}

	creator := NewTierZeroCreator(testDir)
	if creator.HashStoreExists() {
		t.Fatal("HashStoreExists should return false")
	}

	err := creator.CreateHashStore()
	if err != nil {
		t.Fatal(err)
	}

	hashReadWriter, err := creator.OpenHashStore()
	if err != nil {
		t.Fatal("OpenHashStore() failed:", err)
	}

	const hashCount = 65535
	const hashLength = 32
	const masterSeed = 42
	const presentationOffset = 0
	presentationArray := make([]ShallowTreeHash, hashCount)
	for i := range hashCount {
		hash := helperDeterministicHashArray32(masterSeed, int64(i))
		presentationArray[i].Hash = hash[:hashLength]
		presentationArray[i].PresentationIndex = int64(i) + presentationOffset
		sha := Sha256(hash)
		_, err = hashReadWriter.AppendHash(&sha)
		if err != nil {
			t.Fatal(err)
		}
	}

	for i := range hashCount {
		hash := presentationArray[i].Hash
		hashArray := [32]byte{}
		copy(hashArray[:], hash)
		hash256 := Sha256(hashArray)
		presentationIndexRecovered, err := hashReadWriter.IndexOfHash(&hash256)
		if err != nil {
			t.Fatal(err)
		}
		if presentationIndexRecovered == int64(GlobalPiNoMatch) {
			t.Fatal("Lookup failed, returned GlobalPiNoMatch")
		} else if !bytes.Equal(presentationArray[presentationIndexRecovered-presentationOffset].Hash, hash) {
			t.Fatal("Lookup failed, returned index of wrong hash")
		} else {
			//fmt.Println("A success")
		}
	}
	randomHash := helperRandomHashArray32()
	hash256 := Sha256(randomHash)
	presentationIndexRecovered, err := hashReadWriter.IndexOfHash(&hash256)
	if err != nil {
		t.Fatal(err)
	}
	if presentationIndexRecovered != int64(GlobalPiNoMatch) {
		fmt.Printf("Random hash returned a match. Surprising? Maybe? But false positives must be filtered by caller\n")
	}

	err = hashReadWriter.Close()
	if err != nil {
		t.Fatal(err)
	}
}

// helperDeterministicHash generates a reproducible pseudo-random byte slice based on a seed and an index.
func helperDeterministicHashArray32(seed int64, counter int64) [32]byte {
	// Create a unique source for this specific hash to ensure variety across the loop,
	// while keeping the sequence perfectly locked to the master seed.
	rng := rand.New(rand.NewSource(seed + counter))

	result := [32]byte{}
	for i := 0; i < 32; i++ {
		result[i] = byte(rng.Intn(256))
	}
	return result
}

func helperRandomHashArray32() [32]byte {
	result := [32]byte{}
	for i := range 32 {
		result[i] = byte(rand.Intn(256))
	}
	return result
}
