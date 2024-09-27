package justhashes

import (
	"fmt"
	"github.com/KitchenMishap/pudding-shed/corereader"
	"github.com/KitchenMishap/pudding-shed/jsonblock"
	"github.com/KitchenMishap/pudding-shed/memfile"
	"github.com/KitchenMishap/pudding-shed/wordfile"
	"os"
	"time"
)

func JustHashes(folder string, blocks int64) error {
	fmt.Println("Removing old files")
	err := os.RemoveAll(folder)
	if err != nil {
		return err
	}
	err = os.MkdirAll(folder, 0755)
	if err != nil {
		return err
	}
	fmt.Println("Starting")
	blkUnderlying, err := os.Create(folder + "/Block.hsh")
	if err != nil {
		return err
	}
	transUnderlying, err := os.Create(folder + "/Transaction.hsh")
	if err != nil {
		return err
	}

	blkAppendable, err := memfile.NewAppendOptimizedFile(blkUnderlying)
	if err != nil {
		return err
	}
	transAppendable, err := memfile.NewAppendOptimizedFile(transUnderlying)
	if err != nil {
		return err
	}

	blkHashFile := wordfile.NewHashFile(blkAppendable, 0)
	transHashFile := wordfile.NewHashFile(transAppendable, 0)

	reader := corereader.CoreReader{}

	blks := int64(0)
	trns := int64(0)

	var myErr error = nil
	blockNumbersChan := make(chan int64, 20)
	jsonBytesChan := make(chan []byte, 20)
	jsonBlockChan := make(chan *jsonblock.JsonBlockEssential, 20)
	processedJsonChan := make(chan *jsonblock.JsonBlockEssential, 20)
	finishedChan := make(chan bool)

	// We'll be squirting block numbers in once we've set up the goroutines

	// First go routine, get the bytes
	go func() {
		for height := range blockNumbersChan {
			jsonBytes, err := reader.FetchBlockJsonBytes(height)
			if err != nil {
				myErr = err
				fmt.Println(err.Error())
				close(jsonBytesChan)
				return
			}

			jsonBytesChan <- jsonBytes
		}
		close(jsonBytesChan)
	}()

	// Second goroutine, parse the bytes
	go func() {
		for bytes := range jsonBytesChan {
			jsonBlock, err := jsonblock.ParseJsonBlock(bytes)
			if err != nil {
				myErr = err
				fmt.Println(err.Error())
				close(jsonBlockChan)
				return
			}

			jsonBlockChan <- jsonBlock
		}
		close(jsonBlockChan)
	}()

	// Third goroutine, process the json
	go func() {
		for jsonBlock := range jsonBlockChan {
			jsonblock.PostJsonRemoveCoinbaseTxis(jsonBlock)
			err = jsonblock.PostJsonEncodeSha256s(jsonBlock)
			if err != nil {
				myErr = err
				fmt.Println(err.Error())
				close(processedJsonChan)
				return
			}

			processedJsonChan <- jsonBlock
		}
		close(processedJsonChan)
	}()

	// Fourth goroutine, output hashes to files
	go func() {
		for processedJson := range processedJsonChan {
			blkHash := processedJson.BlockHash()
			err = blkHashFile.WriteHashAt(blkHash, blks)
			if err != nil {
				myErr = err
				fmt.Println(err.Error())
				close(finishedChan)
				return
			}
			blks++

			for _, trans := range processedJson.J_tx {
				trnHash := trans.TransHash()
				err = transHashFile.WriteHashAt(trnHash, trns)
				if err != nil {
					myErr = err
					fmt.Println(err.Error())
					close(finishedChan)
					return
				}
				trns++

				if trns%100000 == 0 {
					t := time.Now()
					fmt.Println("\rBlocks:", blks, " Transactions:", trns, t.Format("Mon Jan 2 15:04:05"))
				}
			}
		}
		close(finishedChan)
	}()

	// From the main goroutine, squirt in the block numbers
	for height := int64(0); height < blocks; height++ {
		blockNumbersChan <- height
	}
	close(blockNumbersChan)
	for _ = range finishedChan {
	}
	if myErr != nil {
		fmt.Println(err.Error())
		return myErr
	}
	fmt.Println("Blocks:", blks, " Transactions:", trns)
	err = blkHashFile.Sync()
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	err = transHashFile.Sync()
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	err = blkHashFile.Close()
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	err = transHashFile.Close()
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	fmt.Println("Finished")
	return nil
}
