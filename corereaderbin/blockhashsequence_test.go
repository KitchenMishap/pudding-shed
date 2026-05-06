package corereaderbin

import (
	"fmt"
	"testing"

	"github.com/KitchenMishap/pudding-shed/concurrency"
)

func TestStreamBlockHashes(t *testing.T) {
	blockHashesChannel := make(chan BlockBinary)
	fmt.Println("Starting stream block hashes")
	go func() {
		err := StreamBlockHashesFromGenesis(111111, blockHashesChannel)
		if err != nil {
			t.Error(err)
		}
	}()

	// These are used further downstream, but we need seq back here in order to ask if it's bloated
	blockBinariesChannelOrdered := make(chan BlockBinary)
	seq := concurrency.NewSequencerContainer[BlockBinary](0, 100, blockBinariesChannelOrdered)

	blockHashesForBinaryChannel := make(chan BlockBinary)
	fmt.Println("Starting to pass hashes to GetBlocksBinary")
	go func() {
		for hash := range blockHashesChannel {
			seq.WaitForNotBloated()
			blockHashesForBinaryChannel <- hash
		}
		close(blockHashesForBinaryChannel)
	}()

	fmt.Println("Starting get block binaries")
	GetBlocksBinary(blockHashesForBinaryChannel, seq.InChan, 30)

	fmt.Println("Starting output")
	for block := range blockBinariesChannelOrdered {
		if block.Height%10_000 == 0 {
			fmt.Printf("Block %d: %d bytes\n", block.Height, len(block.Binary))
		}
	}
}
