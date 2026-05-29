package indexedhashestree

import (
	"fmt"
	"os"
	"sort"
	"testing"
)

func TestMultipleFixedNodeSize(t *testing.T) {
	file, err := os.Open("Hashes.hsh")
	if err != nil {
		t.Error(err)
	}
	defer func() { _ = file.Close() }()

	numHashes := 32768

	type percentString struct {
		percent     float64
		description string
	}

	// Read an arbitrary number of hashes first (checking for stability)
	for _ = range 100_000 {
		dummy := [32]byte{}
		var n int
		n, err = file.Read(dummy[:])
		if err != nil {
			t.Error(err)
		}
		if n != 32 {
			t.Error("Couldn't read 32 byte hash")
		}
	}

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
		input[i].presentationIndex = int64(i)
	}

	container := newShallowTreeContainer()
	overflow := container.generate(input)
	if overflow {
		t.Error("Overflowed!")
	} else {
		fmt.Printf("Hashes: %d\n", numHashes)
		fmt.Printf("Nodes used: %d\n", len(container.nodesPool))
		fmt.Printf("Spare lookup values: %d\n", 65536-1-numHashes-len(container.nodesPool))

		stats := container.getNodeSizeStatistics()

		fmt.Printf("Old shallowTree: Bytes per hash %d\n", len(container.nodesPool)*513/numHashes)

		config := newContainerParamsConfigA()
		totalBytes := 0
		totalNodes := 0
		for _, count := range stats {
			totalNodes += count
		}
		results := make([]percentString, 0, 256)
		for i, count := range stats {
			percentOfNodes := float64(100) * float64(count) / float64(totalNodes)
			bytes := config.nodeBytesizeSuitableFor(i)
			if count > 0 {
				entry := percentString{}
				entry.percent = percentOfNodes
				entry.description = fmt.Sprintf("%d slots: %.03f%% of nodes, %d bytes (%.2f bytes per slot)", i, percentOfNodes, bytes, float64(bytes)/float64(i))
				results = append(results, entry)
			}
			totalBytes += bytes * count
		}
		sort.Slice(results, func(i int, j int) bool { return results[i].percent > results[j].percent })
		for x := range results {
			fmt.Println(results[x].description)
		}
		fmt.Printf("New multipleFixedNodeSizes: Bytes per hash %.2f\n", float64(totalBytes)/float64(numHashes))
		fmt.Printf("%d hashes indexed in %.1f KB\n", numHashes, float64(totalBytes)/(1024))
	}
}

// This test shows a minimum for "indexing space used" of about 3.02 bytes per hash, close to 32,768 hashes
// It rises to 4 or 6 above and below that count, but is fairly flat at the minimum (<3.1 between 23,000 and 35,000)
func TestMultipleFixedNodeSweetSpot(t *testing.T) {
	fmt.Println("Reading hashes")

	for numHashes := 10_000; numHashes <= 53000; numHashes += 1_000 {
		input := make([]shallowTreeHash, numHashes)
		file, err := os.Open("Hashes.hsh")
		if err != nil {
			t.Error(err)
		}

		// Read an arbitrary number of hashes first (checking for stability)
		for _ = range 150_000 {
			dummy := [32]byte{}
			var n int
			n, err = file.Read(dummy[:])
			if err != nil {
				t.Error(err)
			}
			if n != 32 {
				t.Error("Couldn't read 32 byte hash")
			}
		}

		for i := range numHashes {
			var n int
			n, err = file.Read(input[i].hash[:])
			if err != nil {
				t.Error(err)
			}
			if n != 32 {
				t.Error("Couldn't read 32 byte hash")
			}
			input[i].presentationIndex = int64(i)
		}

		err = file.Close()
		if err != nil {
			t.Error(err)
		}

		container := newShallowTreeContainer()
		overflow := container.generate(input)
		if overflow {
			t.Error("Overflowed!")
		} else {
			//			fmt.Printf("Hashes: %d\n", numHashes)
			//			fmt.Printf("Nodes used: %d\n", len(container.nodesPool))
			//			fmt.Printf("Spare lookup values: %d\n", 65536-1-numHashes-len(container.nodesPool))

			stats := container.getNodeSizeStatistics()

			config := newContainerParamsConfigA()
			totalBytes := 0
			totalNodes := 0
			for _, count := range stats {
				totalNodes += count
			}
			for i, count := range stats {
				//percentOfNodes := float64(100) * float64(count) / float64(totalNodes)
				bytes := config.nodeBytesizeSuitableFor(i)
				//if count > 0 {
				//	fmt.Printf("%d slots: %f%% of nodes, %d bytes (%.1f bytes per slot)\n", i, percentOfNodes, bytes, float64(bytes)/float64(i))
				//}
				totalBytes += bytes * count
			}
			fmt.Printf("%d Hashes: Bytes per hash %.2f\n", numHashes, float64(totalBytes)/float64(numHashes))
		}
	}
}
