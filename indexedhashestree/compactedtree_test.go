package indexedhashestree

import (
	"fmt"
	"os"
	"testing"
)

// This test shows that the "compactedtree" method could perhaps come close to 4 bytes per hash for 53,000 hashes.
// However it does not propose a way to index into an array of variable-sized nodes.
// One idea was to have an array of "offset bytes", though this was not investigated as it would add a fair amount of data.
// A better way was arrived at sooner, "Multiple Fixed Node Sizes", see multiplefixednodesize.go.
func TestTParam(t *testing.T) {
	fmt.Println("Reading hashes")
	file, err := os.Open("Hashes.hsh")
	if err != nil {
		t.Error(err)
	}
	defer func() { _ = file.Close() }()

	numHashes := 53000
	input := make([]shallowTreeHash, numHashes)
	for i := range numHashes {
		var n int
		n, err = file.Read(input[i].hash[:])
		if err != nil {
			t.Error(err)
		}
		if n != 32 {
			t.Error("Couldn't read 32 byte hash")
		}
		input[i].presentationIndex = uint64(i)
	}

	fmt.Println("Indexing hashes")
	container := newShallowTreeContainer()
	overflow := container.generate(input)
	if overflow {
		t.Error("Overflowed!")
	} else {
		fmt.Println("Succeeded")
		fmt.Printf("Hashes: %d\n", numHashes)
		fmt.Printf("Nodes used: %d\n", len(container.nodesPool))
		fmt.Printf("MaxSkip used: %d\n", container.maxSkipNumber)
		fmt.Printf("Spare lookup values: %d\n", 65536-1-numHashes-int(container.maxSkipNumber))

		stats := container.getNodeSizeStatistics()

		fmt.Printf("Old shallowTree: Bytes per hash %d\n", len(container.nodesPool)*513/numHashes)
		for tParam := range 10 {
			bytes := 0
			for nodeSize := range 257 {
				bytes += tryParamsForSize(tParam, nodeSize, stats[nodeSize])
			}
			fmt.Printf("New compactTree: tParam = %d, \tBytes per hash %.02f, \tTotal bytes %d\n", tParam, float64(bytes)/float64(numHashes), bytes)
		}
	}
}
