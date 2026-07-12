package weddingcakeback

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestBakingDesigner(t *testing.T) {
	// 1. Completely wipe and recreate the testing directory to clear out stale files
	testDir := filepath.Join("Temp_Testing")
	_ = os.RemoveAll(testDir) // Ignore error if it doesn't exist yet
	err := os.MkdirAll(testDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	creator := NewTierTopCreator(testDir)
	err = creator.Create(0)
	if err != nil {
		t.Fatal(err)
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

	// We have created a tier zero with 65535 hashes
	// Now bake the first DonutForest in tier 1, from the entirity of tier zero

	config := NewCakeConfig(32)
	writer := NewDonutForestWrite(tierTop, config)
	err = writer.Write(testDir)
	if err != nil {
		t.Fatal(err)
	}

	tb := NewTierBelow(testDir, 0, config)
	err = tb.Open()
	if err != nil {
		t.Fatal(err)
	}

	for i := range int64(count) {
		hash := presentationArray[i]
		globalPi := tb.LookupHash(hash[:])
		if globalPi == GlobalPiNoMatch {
			t.Fatal("Hash not found in tierBelow[0]")
		}
		if !bytes.Equal(hash[:], presentationArray[globalPi][:]) {
			t.Fatal("Hash mismatch (wrong presentation index)")
		}
	}

	err = tb.Close()
	if err != nil {
		t.Fatal(err)
	}
}
