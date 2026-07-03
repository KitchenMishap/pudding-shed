package weddingcake

import (
	"bytes"
	"fmt"
	"testing"
)

func Test65535HashesVariousLengths(t *testing.T) {
	const count = 65535
	const presentationOffset = 10
	firstHashLength := byte(8) // So that we're "very unlikely" to get a duplicate
	for hashLength := firstHashLength; hashLength <= 64; hashLength++ {
		presentationArray := make([]ShallowTreeHash, count)
		for i := range count {
			hash := helperRandomHash(hashLength)
			presentationArray[i].Hash = hash
			presentationArray[i].PresentationIndex = int64(i + presentationOffset)
		}
		cp := NewChunkPreparation(presentationArray, hashLength, 2, 10,
			ID16[NodeIdType]{}, ID16[HashIndexIdType]{})
		cp.PrepareAndSerialize()

		for i := range count {
			hash := presentationArray[i].Hash
			presentationIndexRecovered := cp.chunk.LookupHash(hash)
			if presentationIndexRecovered == 0 {
				t.Error("Lookup failed, returned \"unmatched\"")
			}
			if !bytes.Equal(presentationArray[presentationIndexRecovered-presentationOffset].Hash, hash) {
				t.Error("Lookup failed, returned index of wrong hash")
			}
		}
		randomHash := helperRandomHash(hashLength)
		presentationIndexRecovered := cp.chunk.LookupHash(randomHash)
		if presentationIndexRecovered != -1 {
			fmt.Printf("Hash size %d: Random hash returned a match. Surprising? No! false positives need to be filtered by caller\n",
				hashLength)
		}
	}
}
