package indexedhashestree

import (
	"fmt"
	"os"
	"testing"
)

func TestShallowTree(t *testing.T) {
	fmt.Println("Reading hashes")
	file, err := os.Open("Hashes.hsh")
	if err != nil {
		t.Error(err)
	}
	defer func() { _ = file.Close() }()

	numHashes := 53000 // The largest multiple of 1000 that I find works here, for these specific hashes
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
	}
}

// This test shows that a single fixed node size takes up too many bytes per hash!
// Switch to multifixednodesize.go...
func TestBytesPerHash(t *testing.T) {
	for numHashes := 10_000; numHashes <= 50_000; numHashes += 1000 {
		file, err := os.Open("Hashes.hsh")
		if err != nil {
			t.Error(err)
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
		err = file.Close()
		if err != nil {
			t.Error(err)
		}

		container := newShallowTreeContainer()
		overflow := container.generate(input)
		if overflow {
			t.Error("Overflowed!")
		} else {
			bytes := len(container.nodesPool) * (1 + 256*2) // + 32*numHashes
			bytesPerHash := bytes / numHashes
			fmt.Printf("%d\t%d\t%d\n", numHashes, bytes, bytesPerHash)
		}
	}
}

func TestAllHashesShallowTree(t *testing.T) {
	fmt.Println("Reading hashes")
	file, err := os.Open("Hashes.hsh")
	if err != nil {
		t.Error(err)
	}
	defer func() { _ = file.Close() }()

	numHashes := 53000 // The largest multiple of 1000 that I find works here, for these specific hashes
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

		fmt.Println("Testing all hashes...")
		for i := int64(0); i < int64(numHashes); i++ {
			hash := input[i].hash
			presentationIndex, _ := container.lookupHash(hash)
			if presentationIndex != i {
				if presentationIndex == -1 {
					t.Error("Couldn't find hash at index", i)
				} else {
					t.Error(fmt.Sprintf("Found hash at %d instead of %d", presentationIndex, i))
				}
			}
		}
	}
}

func TestAllHashesWithDuplicateShallowTree(t *testing.T) {
	fmt.Println("Reading hashes")
	file, err := os.Open("Hashes.hsh")
	if err != nil {
		t.Error(err)
	}
	defer func() { _ = file.Close() }()

	numHashes := 53000 // The largest multiple of 1000 that I find works here, for these specific hashes
	input := make([]shallowTreeHash, numHashes)
	for i := 0; i < numHashes; i++ {
		var n int
		n, err = file.Read(input[i].hash[:])
		if err != nil {
			t.Error(err)
		}
		if n != 32 {
			t.Error("Couldn't read 32 byte hash")
		}
		input[i].presentationIndex = int64(i)

		// Throw things off balance by introducing a fake duplicate...
		if i == 1000 {
			i++
			input[i].presentationIndex = int64(i)
			input[i].hash = input[1000].hash
		}
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

		fmt.Println("Testing all hashes...")
		for i := int64(0); i < int64(numHashes); i++ {
			hash := input[i].hash
			presentationIndex, duplicate := container.lookupHash(hash)
			if duplicate {
				fmt.Printf("Duplicate hash found for index %d, sent back %d\n", i, presentationIndex)
			} else if presentationIndex != i {
				if presentationIndex == -1 {
					t.Error("Couldn't find hash at index", i)
				} else {
					t.Error(fmt.Sprintf("Found hash at %d instead of %d", presentationIndex, i))
				}
			}
		}
	}
}
