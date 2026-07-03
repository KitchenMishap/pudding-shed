package weddingcake

import (
	"bytes"
	"fmt"
	"math/rand"
	"testing"
)

const reassuranceBytesCount = 2

func TestEmptyTree(t *testing.T) {
	for hashLength := byte(reassuranceBytesCount); hashLength <= 64; hashLength++ {
		st := GenerateShallowTree(make([]ShallowTreeHash, 0), hashLength, reassuranceBytesCount)
		if st.LookupHash(helperRandomHash(hashLength)) != ShallowTreeNoMatch {
			t.Error("LookupHash should return ShallowTreeNoMatch, when looking up in an empty tree")
		}
	}
}

func TestSingleHashPresent(t *testing.T) {
	for hashLength := byte(reassuranceBytesCount); hashLength <= 64; hashLength++ {
		presentationArray := make([]ShallowTreeHash, 1)
		hash := helperRandomHash(hashLength)
		presentationArray[0].Hash = hash
		presentationArray[0].PresentationIndex = 0
		st := GenerateShallowTree(presentationArray, hashLength, reassuranceBytesCount)
		presentationIndex := st.LookupHash(hash)
		if presentationIndex != 0 {
			t.Error("Expected presentationIndex 0")
		}
	}
}

func TestSingleHashAbsent(t *testing.T) {
	for hashLength := byte(reassuranceBytesCount); hashLength <= 64; hashLength++ {
		presentationArray := make([]ShallowTreeHash, 1)
		hash := helperRandomHash(hashLength)
		presentationArray[0].Hash = hash
		presentationArray[0].PresentationIndex = 0
		st := GenerateShallowTree(presentationArray, hashLength, reassuranceBytesCount)
		presentationIndex := st.LookupHash(hash)
		if presentationIndex != 0 {
			t.Error("Expected presentationIndex 0, even when hash doesn't fully match")
		}
	}
}

func TestThousandHashes(t *testing.T) {
	count := 1000
	firstHashLength := byte(8) // So that we're "very unlikely" to get a duplicate
	for hashLength := firstHashLength; hashLength <= 64; hashLength++ {
		presentationArray := make([]ShallowTreeHash, count)
		for i := range count {
			hash := helperRandomHash(hashLength)
			presentationArray[i].Hash = hash
			presentationArray[i].PresentationIndex = int64(i)
		}
		st := GenerateShallowTree(presentationArray, hashLength, reassuranceBytesCount)
		for i := range count {
			hash := presentationArray[i].Hash
			presentationIndex := st.LookupHash(hash)
			if presentationIndex == ShallowTreeNoMatch {
				t.Error("Lookup failed, returned ShallowTreeNoMatch")
			}
			if !bytes.Equal(presentationArray[presentationIndex].Hash, hash) {
				t.Error("Lookup failed, returned index of wrong hash")
			}
		}
		randomHash := helperRandomHash(hashLength)
		presentationIndex := st.LookupHash(randomHash)
		if presentationIndex != ShallowTreeNoMatch {
			fmt.Printf("Hash size %d: Random hash returned a match. Surprising? No! false positives need to be filtered by caller\n",
				hashLength)
		}
	}
}

/* Duplicates are no longer tolerated
func TestThousandWithTripleDuplicate(t *testing.T) {
	count := 1000
	location := 500
	for hashLength := byte(1); hashLength <= 64; hashLength++ {
		presentationArray := make([]ShallowTreeHash, count)
		for i := 0; i < count; i++ {
			hash := helperRandomHash(hashLength)
			presentationArray[i].Hash = hash
			presentationArray[i].PresentationIndex = int64(i)
		}
		// Overwrite to cause a triple duplicate at locations 500,501,502
		presentationArray[501].Hash = presentationArray[500].Hash
		presentationArray[502].Hash = presentationArray[500].Hash

		st := GenerateShallowTree(presentationArray, hashLength)
		for i := range count {
			hash := presentationArray[i].Hash
			presentationIndices := st.LookupHash(hash)
			if len(presentationIndices) == 0 {
				t.Error("Lookup failed, returning 0 matches")
			}
			if i >= location && i <= location+2 {
				// It must find at least your 3 intentional duplicates (could be more due to natural collisions)
				if len(presentationIndices) < 3 {
					t.Errorf("HashLength %d: Expected at least 3 duplicates, got %d", hashLength, len(presentationIndices))
				} else {
					// Track if our three explicit presentation indices are present
					found500, found501, found502 := false, false, false
					for _, pi := range presentationIndices {
						switch pi {
						case 500:
							found500 = true
						case 501:
							found501 = true
						case 502:
							found502 = true
						}
					}
					if !found500 || !found501 || !found502 {
						t.Errorf("HashLength %d: Missing one of our explicit duplicates in returned slice: %v", hashLength, presentationIndices)
					}
				}
			}
			for p := range presentationIndices {
				pi := presentationIndices[p]
				if !bytes.Equal(presentationArray[pi].Hash, hash) {
					t.Error("Lookup failed, returned index of wrong hash")
				}
			}
		}
	}
}*/

func helperRandomHash(hashLength byte) []byte {
	result := make([]byte, hashLength)
	for i := range hashLength {
		result[i] = byte(rand.Intn(256))
	}
	return result
}

// helperDeterministicHash generates a reproducible pseudo-random byte slice based on a seed and an index.
func helperDeterministicHash(hashLength byte, seed int64, counter int64) []byte {
	// Create a unique source for this specific hash to ensure variety across the loop,
	// while keeping the sequence perfectly locked to the master seed.
	rng := rand.New(rand.NewSource(seed + counter))

	result := make([]byte, hashLength)
	for i := range hashLength {
		result[i] = byte(rng.Intn(256))
	}
	return result
}
