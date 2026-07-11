package weddingcakeback

import (
	"bytes"
	"fmt"
	"math/rand"
	"testing"
)

const reassuranceBytesCount = 2

func TestEmptyTree(t *testing.T) {
	for prefixBytesN := byte(0); prefixBytesN <= 4; prefixBytesN++ {
		for hashLength := byte(8); hashLength <= 64; hashLength += 4 {
			st := GenerateSingleTree(make([]SingleTreeHash, 0), prefixBytesN, hashLength, reassuranceBytesCount)
			if st.LookupHash(helperRandomHash(hashLength)) != SingleTreeNoMatch {
				t.Error("LookupHash should return SingleTreeNoMatch, when looking up in an empty tree")
			}
		}
	}
}

func TestSingleHashPresent(t *testing.T) {
	for prefixBytesN := byte(0); prefixBytesN <= 4; prefixBytesN++ {
		for hashLength := byte(8); hashLength <= 64; hashLength += 4 {
			presentationArray := make([]SingleTreeHash, 1)
			hash := helperRandomHash(hashLength)
			presentationArray[0].Hash = hash
			presentationArray[0].PresentationIndex = 1
			st := GenerateSingleTree(presentationArray, prefixBytesN, hashLength, reassuranceBytesCount)
			presentationIndex := st.LookupHash(hash)
			if presentationIndex != 1 {
				t.Error("Expected presentationIndex 1")
			}
		}
	}
}

func TestSingleHashAbsent(t *testing.T) {
	for prefixBytesN := byte(0); prefixBytesN <= 4; prefixBytesN++ {
		for hashLength := byte(8); hashLength <= 64; hashLength += 4 {
			presentationArray := make([]SingleTreeHash, 1)
			hash := helperRandomHash(hashLength)
			presentationArray[0].Hash = hash
			presentationArray[0].PresentationIndex = 1
			st := GenerateSingleTree(presentationArray, prefixBytesN, hashLength, reassuranceBytesCount)
			hash = helperRandomHash(hashLength)
			presentationIndex := st.LookupHash(hash)
			if presentationIndex != SingleTreeNoMatch {
				t.Error("Expected no match")
			}
		}
	}
}

func Test65535Hashes(t *testing.T) {
	const count = 65535
	const prefixHashBytesCount = byte(2)

	firstHashLength := byte(8) // So that we're "very unlikely" to get a duplicate
	for hashLength := firstHashLength; hashLength <= 64; hashLength += 4 {
		fmt.Printf("Hash size %d...\n", hashLength)
		presentationArray := make([]SingleTreeHash, count)
		for i := range count {
			hash := helperRandomHash(hashLength)
			presentationArray[i].Hash = hash
			presentationArray[i].PresentationIndex = SingleTreePiType(i + 1) // 0 is now reserved as SingleTreeNoMatch
		}
		st := GenerateSingleTree(presentationArray, prefixHashBytesCount, hashLength, reassuranceBytesCount)
		for i := range count {
			hash := presentationArray[i].Hash
			presentationIndex := st.LookupHash(hash)
			if presentationIndex == SingleTreeNoMatch {
				t.Error("Lookup failed, returned SingleTreeNoMatch")
			}
			if !bytes.Equal(presentationArray[presentationIndex-1].Hash, hash) {
				t.Error("Lookup failed, returned index of wrong hash")
			}
		}
		randomHash := helperRandomHash(hashLength)
		presentationIndex := st.LookupHash(randomHash)
		if presentationIndex != SingleTreeNoMatch {
			fmt.Printf("Hash size %d: Random hash returned a match. Surprising? Yes as we now check the full hash!",
				hashLength)
		}
	}
}

func helperRandomHash(hashLength byte) []byte {
	result := make([]byte, hashLength)
	for i := range hashLength {
		result[i] = byte(rand.Intn(256))
	}
	return result
}
