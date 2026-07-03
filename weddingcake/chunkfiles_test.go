package weddingcake

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// The various pairs of "ways to represent" node ids and hash ids as multiple bytes
type nodeAndHashIdConfigPair struct {
	forNodeId NByteIdConfig[NodeIdType]
	forHashId NByteIdConfig[HashIndexIdType]
}

func TestChunkFilesMultiConfig(t *testing.T) {
	const bits24 = true  // Choose between testing 16 bit or 24 bit scenarios
	const quicker = true // For a quicker 10m test (slow test is 1.5h!)

	hashLengthsSlow := []byte{8, 20, 32, 64}
	hashLengthsQuicker := []byte{32}

	// Use this one if NodeIdType, HashIndexIdType are uint16
	hashCounts16Bit := []int64{0, 1, 4, 10, 100, 1000, 10000, 65535}
	// Use this one if NodeIdType, HashIndexIdType are > uint16
	hashCounts24Bit := []int64{0, 1, 4, 10, 100, 1000, 10_000, 65535, 65536, 100_000, 1000_000, 256*256*256 - 1}
	hashCounts24BitQuicker := []int64{256*256*256 - 1}

	// The various pairs of "ways to represent" node ids and hash ids as multiple bytes
	nodeAndHashIdConfigPairs16Bit := []nodeAndHashIdConfigPair{
		{forNodeId: ID16[NodeIdType]{}, forHashId: ID16[HashIndexIdType]{}},
		{forNodeId: ID24[NodeIdType]{}, forHashId: ID24[HashIndexIdType]{}},
		{forNodeId: ID32[NodeIdType]{}, forHashId: ID32[HashIndexIdType]{}},
	}
	nodeAndHashIdConfigPairs24Bit := []nodeAndHashIdConfigPair{
		{forNodeId: ID24[NodeIdType]{}, forHashId: ID24[HashIndexIdType]{}},
		{forNodeId: ID32[NodeIdType]{}, forHashId: ID32[HashIndexIdType]{}},
	}

	var nodeAndHashIdConfigPairs *[]nodeAndHashIdConfigPair
	var hashCounts *[]int64
	var hashLengths *[]byte
	if bits24 {
		nodeAndHashIdConfigPairs = &nodeAndHashIdConfigPairs24Bit
		if quicker {
			hashCounts = &hashCounts24BitQuicker
			hashLengths = &hashLengthsQuicker
		} else {
			hashCounts = &hashCounts24Bit
			hashLengths = &hashLengthsSlow
		}
	} else {
		nodeAndHashIdConfigPairs = &nodeAndHashIdConfigPairs16Bit
		hashCounts = &hashCounts16Bit
		if quicker {
			hashLengths = &hashLengthsQuicker
		} else {
			hashLengths = &hashLengthsSlow
		}
	}

	for _, hashLength := range *hashLengths {
		for _, hashCount := range *hashCounts {
			for _, pair := range *nodeAndHashIdConfigPairs {
				startTime := time.Now()
				fmt.Printf("Hash Length: %d, Hash Count: %d, node id bits:%d, hash id bits:%d\n",
					hashLength, hashCount, pair.forNodeId.StorageBytes()*8, pair.forHashId.StorageBytes()*8)
				chunkFilesHelper(hashCount, hashLength, pair, t)
				duration := time.Since(startTime)
				durationMins := duration.Minutes()
				fmt.Printf("Time taken: %.2f minutes\n", durationMins)
			}
		}
	}
}

func chunkFilesHelper(hashCount int64, hashLength byte, configPair nodeAndHashIdConfigPair, t *testing.T) {
	const presentationOffset = 10
	const masterSeed = 42

	// 1. Completely wipe and recreate the testing directory to clear out stale LevelXX.bin files
	testDir := filepath.Join("Temp_Testing")
	_ = os.RemoveAll(testDir) // Ignore error if it doesn't exist yet
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatal(err) // Stop execution immediately if environment setup fails
	}

	// 2. Safely create your initial empty metadata tracking file
	chunksInfoPath := filepath.Join(testDir, "ChunksInfo.bin")
	file, err := os.Create(chunksInfoPath)
	if err != nil {
		t.Fatal(err)
	}
	_ = file.Close()

	presentationArray := make([]ShallowTreeHash, hashCount)
	for i := range hashCount {
		hash := helperDeterministicHash(hashLength, masterSeed, i)
		presentationArray[i].Hash = hash
		presentationArray[i].PresentationIndex = i + presentationOffset
	}
	cp := NewChunkPreparation(presentationArray, hashLength, 2, 10,
		configPair.forNodeId, configPair.forHashId)
	cp.PrepareAndSerialize()

	filesWrite := NewChunkFilesWrite("Temp_Testing")
	err = filesWrite.AppendChunk(&cp.chunk)
	if err != nil {
		t.Fatal(err)
	}

	filesRead := NewChunkFilesRead("Temp_Testing")
	err = filesRead.ReadAndMmap(cp.chunk.Config)
	if err != nil {
		t.Fatal(err)
	}

	for i := range hashCount {
		hash := presentationArray[i].Hash
		presentationIndexRecovered := filesRead.chunks[0].LookupHash(hash)
		if presentationIndexRecovered == GlobalPiNoMatch {
			t.Fatal("Lookup failed, returned GlobalPiNoMatch")
		} else if !bytes.Equal(presentationArray[presentationIndexRecovered-presentationOffset].Hash, hash) {
			t.Fatal("Lookup failed, returned index of wrong hash")
		} else {
			//fmt.Println("A success")
		}
	}
	randomHash := helperRandomHash(hashLength)
	presentationIndex := filesRead.chunks[0].LookupHash(randomHash)
	if presentationIndex != GlobalPiNoMatch {
		fmt.Printf("Hash size %d: Random hash returned a match. Surprising? Maybe? But false positives must be filtered by caller\n",
			hashLength)
	}

	err = filesRead.UnMmapLevelFiles()
	if err != nil {
		t.Fatal(err)
	}
}
