package jobs

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/KitchenMishap/pudding-shed/chainstorage"
	"github.com/KitchenMishap/pudding-shed/concurrency"
	"github.com/KitchenMishap/pudding-shed/corereaderbin"
)

func TestStreamBlockHashes(t *testing.T) {

	path := "E:\\Data\\OneYearBinary"
	dateFormat := "Mon Jan 2 15:04:05"
	fmt.Println(time.Now().Format(dateFormat) + "\t\tRemoving previous files...")
	err := os.RemoveAll(path)
	if err != nil {
		t.Error(err)
	}
	hcc := chainstorage.NewConcreteHashesChainCreator(path)
	err = hcc.Create()
	if err != nil {
		t.Error(err)
	}
	hc, err := hcc.Open()
	if err != nil {
		t.Error(err)
	}

	blockHashesChannel := make(chan corereaderbin.BlockBinary)
	fmt.Println("Starting stream block hashes")
	go func() {
		err := corereaderbin.StreamBlockHashesFromGenesis(32879, blockHashesChannel)
		if err != nil {
			t.Error(err)
		}
	}()

	// These are used further downstream, but we need seq back here in order to ask if it's bloated
	blockBinariesChannelOrdered := make(chan corereaderbin.BlockBinary)
	seq := concurrency.NewSequencerContainer[corereaderbin.BlockBinary](0, 100, blockBinariesChannelOrdered)

	blockHashesForBinaryChannel := make(chan corereaderbin.BlockBinary)
	fmt.Println("Starting to pass hashes to GetBlocksBinary")
	go func() {
		for hash := range blockHashesChannel {
			seq.WaitForNotBloated()
			blockHashesForBinaryChannel <- hash
		}
		close(blockHashesForBinaryChannel)
	}()

	fmt.Println("Starting get block binaries")
	corereaderbin.GetBlocksBinary(blockHashesForBinaryChannel, seq.InChan, 30)

	fmt.Println("Starting output")
	for block := range blockBinariesChannelOrdered {
		if block.Height%10_000 == 0 {
			fmt.Printf("Block %d: %d bytes\n", block.Height, len(block.Binary))
		}
		err = hc.AppendHashesBinary(&block)
	}
	hc.Close()
}
