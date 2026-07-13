package weddingcakefront

import (
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
)

// The following tests pass with no reassurance bytes in the leaves, but only because there
// are so few hashes.
func TestCakeCreatorTierTop(t *testing.T) {
	CakeCreatorHelper(t, 65534, 0) // Because 65535 would trigger a rebaking
}
func TestCakeCreatorTierTopEmpty(t *testing.T) {
	CakeCreatorHelper(t, 65535, 0) // 65535 in TierBelow[0], none in TierTop
}
func TestCakeCreatorTierBelow0(t *testing.T) {
	CakeCreatorHelper(t, 65536, 0) // 65535 in TierBelow[0], one in TierTop
}

// With more hashes, two reassurance bytes are needed. It still only "probably" passes because
// we don't yet check the whole hash when a leaf is encountered.
func TestCakeCreator131070(t *testing.T) {
	CakeCreatorHelper(t, 65535*2, 2) // 65535 in TierBelow[0], 65535 in a second DonutForest
}

// Flamegraph for the following test (mainly for comparison with the test below that)
// 38.93% helperDeterministicHashSha256()
// 35.54% AppendHash()
// 4.36% TryIndexOfHash()
// Other stuff: 11.95% gcBgMarkWorker(), 6.13% systemstack, 1.5% mcall, 1.13% bgsweep
// Executive summary: For every append hash, this test does 3 read checks.
// (guess 1.45% of overall time spent on one of the three lookups)
// versus 35.54% appending the hash. So it is 24.5 times quicker to do a lookup versus append. Nice.
func TestCakeCreatorTierBelow0ThreeDonutForests(t *testing.T) {
	fmt.Printf("This will take about ten seconds...\n")
	CakeCreatorHelper(t, 65535*3, 3) // 3 DonutForests in TierBelow[0]
}

// Flamegraph for the following test:
// 8.93% helperDeterministicHashSha256()
// 8.6% AppendHash (8.41% DonutForestWrite.Write(), 3.46% + 3.47% GenerateSingleTree)
// 80.67% TryIndexOfHash()
// One particular recurseLookupHash() (second recursive call): 67.69%, and inside that:
// 27.05% next recurseLookupHash(), 16.41% nextLevelNodeId(), 10.76% hashByteIndexToExamine(), 10% ExtractNode(),
// 1.21% detailsIfLeaf()
// Executive summary: For every append hash, this test does 3 read checks.
// According to the above, 80.67%  of overall time is spent on the three reads (guess 26.89% each)
// versus 8.6% of overall time on each append. So it takes 3.13 times as long for a lookup as an append!
// Perhaps a major issue can be that we loop through up to 255 DonutForest's when looking up within a tier...
func TestCakeCreatorTierBelow0Full(t *testing.T) {
	fmt.Printf("This will take more than an hour...\n")
	CakeCreatorHelper(t, 65535*254, 4) // 254 DonutForests in TierBelow[0]
}

func CakeCreatorHelper(t *testing.T, count int, reassuranceBytes byte) {
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

	creator := NewCakeCreator(testDir, reassuranceBytes)
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

	const masterSeed = 42

	fmt.Printf("Writing the hashes\n")
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
	fmt.Printf("Wrote hashes\n")

	// Each passConfig is {reopen, readonly}
	var storeRo LegacyHashReader
	passConfig := [3][2]bool{{false, false}, {true, false}, {true, true}}
	for pass := 0; pass < 3; pass++ {
		reopen := passConfig[pass][0]
		readonly := passConfig[pass][1]
		if reopen {
			if readonly {
				fmt.Printf("Pass %d: Reading back after closing and reopening readonly\n", pass)
				err = store.Close()
				if err != nil {
					t.Fatal(err)
				}
				store = nil
				storeRo, err = creator.OpenHashStoreReadOnly()
				if err != nil {
					t.Fatal(err)
				}
			} else {
				fmt.Printf("Pass %d: Reading back after closing and reopening read/write\n", pass)
				err = store.Close()
				if err != nil {
					t.Fatal(err)
				}
				store, err = creator.OpenHashStore()
				if err != nil {
					t.Fatal(err)
				}
			}
		} else {
			fmt.Printf("Pass %d: Reading back without closing \n", pass)
		}
		for i := range int64(count) {
			hash := presentationArray[i]
			var presentationIndexRecovered int64
			if readonly {
				presentationIndexRecovered, err = storeRo.IndexOfHash(&hash)
			} else {
				presentationIndexRecovered, err = store.IndexOfHash(&hash)
			}
			if err != nil {
				t.Fatal(err)
			}
			if presentationIndexRecovered == -1 {
				t.Fatal("Lookup failed, returned -1")
			} else if !bytes.Equal(presentationArray[presentationIndexRecovered][:], hash[:]) {
				t.Fatal("Lookup failed, returned index of wrong hash")
			} else {
				//fmt.Println("A success")
			}
		}
		randomHash := helperRandomHashSha256()
		var presentationIndexRecovered int64
		if readonly {
			presentationIndexRecovered, err = storeRo.IndexOfHash(&randomHash)
		} else {
			presentationIndexRecovered, err = store.IndexOfHash(&randomHash)
		}
		if err != nil {
			t.Fatal(err)
		}
		if presentationIndexRecovered != -1 {
			fmt.Printf("Random hash returned a match. Surprising? Maybe? But false positives need to be filtered by caller\n")
		}
	}
	if store != nil {
		err = store.Close()
	} else if storeRo != nil {
		err = storeRo.Close()
	}
	if err != nil {
		t.Fatal(err)
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
