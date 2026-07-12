package weddingcakefront

import (
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
)

func TestCakeCreatorTierTop(t *testing.T) {
	CakeCreatorHelper(t, 65534) // Because 65535 would trigger a rebaking
}

func TestCakeCreatorTierTopEmpty(t *testing.T) {
	CakeCreatorHelper(t, 65535) // 65535 in TierBelow[0], none in TierTop
}
func TestCakeCreatorTierBelow0(t *testing.T) {
	CakeCreatorHelper(t, 65536) // 65535 in TierBelow[0], one in TierTop
}
func TestCakeCreator131070(t *testing.T) {
	CakeCreatorHelper(t, 65535*2) // 65535 in TierBelow[0], 65535 in a second DonutForest
}

func CakeCreatorHelper(t *testing.T, count int) {
	// 1. Completely wipe and recreate the testing directory to clear out stale files
	testDir := filepath.Join("Temp_Testing")
	// Make the directory so that doesn't trigger an error
	err := os.MkdirAll(testDir, 0755)
	if err != nil {
		t.Fatal(err)
	}
	err = os.RemoveAll(testDir) // Ignore error if it doesn't exist yet
	if err != nil {
		t.Fatal(err)
	}
	// Remake the dir
	err = os.MkdirAll(testDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	creator := NewCakeCreator(testDir)
	if creator.HashStoreExists() {
		t.Fatal("Cake creator should not have a hash store yet")
	}

	err = creator.CreateHashStore()
	if err != nil {
		t.Fatal(err)
	}
	if !creator.HashStoreExists() {
		t.Fatal("Hash store should now exist")
	}

	store, err := creator.OpenHashStore()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = store.Close() }()

	const masterSeed = 42

	presentationArray := make([]Sha256, count)
	for i := range int64(count) {
		hash := helperDeterministicHashSha256(masterSeed, i)
		presentationArray[i] = hash
		index, err := store.AppendHash(&hash)
		if err != nil {
			t.Fatal(err)
		}
		if index != i {
			t.Fatal("Hash index mismatch")
		}
	}

	for i := range int64(count) {
		hash := presentationArray[i]
		presentationIndexRecovered, err := store.IndexOfHash(&hash)
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
	presentationIndex, err := store.IndexOfHash(&randomHash)
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
