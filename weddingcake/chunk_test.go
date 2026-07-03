package weddingcake

import (
	"bytes"
	"fmt"
	"testing"
)

func Test65535HashesLength32(t *testing.T) {
	const count = 65535
	const hashLength = 32
	const presentationOffset = 10
	const masterSeed = 42

	presentationArray := make([]ShallowTreeHash, count)
	for i := range int64(count) {
		hash := helperDeterministicHash(hashLength, masterSeed, i)
		presentationArray[i].Hash = hash
		presentationArray[i].PresentationIndex = i + presentationOffset
	}
	cp := NewChunkPreparation(presentationArray, hashLength, 2, 10,
		ID16[NodeIdType]{}, ID16[HashIndexIdType]{})
	cp.PrepareAndSerialize()

	for i := range count {
		hash := presentationArray[i].Hash
		presentationIndexRecovered := cp.chunk.LookupHash(hash)
		if presentationIndexRecovered == -1 {
			t.Error("Lookup failed, returned -1")
		} else if !bytes.Equal(presentationArray[presentationIndexRecovered-presentationOffset].Hash, hash) {
			t.Error("Lookup failed, returned index of wrong hash")
		} else {
			//fmt.Println("A success")
		}
	}
	randomHash := helperDeterministicHash(hashLength, masterSeed, count)
	presentationIndex := cp.chunk.LookupHash(randomHash)
	if presentationIndex != -1 {
		fmt.Printf("Hash size %d: Random hash returned a match. Surprising? No! false positives need to be filtered by caller\n",
			hashLength)
	}
}

func Test65535HashesVariousLength(t *testing.T) {
	const count = 65535
	const presentationOffset = 10
	const masterSeed = 42

	// We start at 4 to give a low probability (we want zero!) of duplicates
	for hashLength := byte(4); hashLength <= 64; hashLength++ {
		presentationArray := make([]ShallowTreeHash, count)
		for i := range int64(count) {
			hash := helperDeterministicHash(hashLength, masterSeed, i)
			presentationArray[i].Hash = hash
			presentationArray[i].PresentationIndex = i + presentationOffset
		}
		cp := NewChunkPreparation(presentationArray, hashLength, 2, 10,
			ID16[NodeIdType]{}, ID16[HashIndexIdType]{})
		cp.PrepareAndSerialize()

		for i := range count {
			hash := presentationArray[i].Hash
			presentationIndexRecovered := cp.chunk.LookupHash(hash)
			if presentationIndexRecovered == -1 {
				t.Error("Lookup failed, returned -1")
			} else if !bytes.Equal(presentationArray[presentationIndexRecovered-presentationOffset].Hash, hash) {
				t.Error("Lookup failed, returned index of wrong hash")
			} else {
				//fmt.Println("A success")
			}
		}
		randomHash := helperDeterministicHash(hashLength, masterSeed, count)
		presentationIndex := cp.chunk.LookupHash(randomHash)
		if presentationIndex != -1 {
			fmt.Printf("Hash size %d: Random hash returned a match. Surprising? No! false positives need to be filtered by caller\n",
				hashLength)
		}
	}
}
