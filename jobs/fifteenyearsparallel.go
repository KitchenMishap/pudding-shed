package jobs

import (
	"fmt"
	"github.com/KitchenMishap/pudding-shed/chainstorage"
	"github.com/KitchenMishap/pudding-shed/concurrency"
	"github.com/KitchenMishap/pudding-shed/corereader"
	"github.com/KitchenMishap/pudding-shed/jsonblock"
	"github.com/KitchenMishap/pudding-shed/transactionindexing"
	"os"
	"runtime"
	"sync"
	"time"
)

type preprocessTask struct {
	blockBytes []byte
	jsonBlock  *jsonblock.JsonBlockEssential
	err        error
	outChan    *chan *preprocessTask
}

func (ppt *preprocessTask) Process() error {
	var err error
	ppt.jsonBlock, err = jsonblock.ParseJsonBlock(ppt.blockBytes)
	jsonblock.PostJsonRemoveCoinbaseTxis(ppt.jsonBlock)
	jsonblock.PostJsonCalculateSatoshis(ppt.jsonBlock)
	jsonblock.PostJsonEncodeAddressHashes(ppt.jsonBlock)
	jsonblock.PostJsonEncodeSha256s(ppt.jsonBlock)
	return err
}
func (ppt *preprocessTask) SetError(err error) {
	ppt.err = err
}
func (ppt *preprocessTask) Done() {
	*ppt.outChan <- ppt
}
func (ppt *preprocessTask) GetError() error {
	return ppt.err
}

func SeveralYearsParallel(years int, transactionIndexingMethod string) error {
	const path = "F:\\Data\\CurrentJob"
	lastBlock := int64(10000) // Default
	if years == 1 {
		lastBlock = 33000
	} else if years == 2 {
		lastBlock = 66000
	} else if years == 3 {
		lastBlock = 99000
	} else if years == 4 {
		lastBlock = 132000
	} else if years == 5 {
		lastBlock = 165000
	} else if years == 6 {
		lastBlock = 315360
	} else if years == 7 {
		lastBlock = 391000 // 7 years = mnemonic for 100 million transactions
	} else if years == 15 {
		lastBlock = 840000 // Halving
	} else if years == 16 {
		lastBlock = 860530 // 09 Sep 2024
	}

	fmt.Println("Removing previous files...")
	err := os.RemoveAll(path)
	if err != nil {
		return err
	}

	delegated := transactionIndexingMethod == "delegated"

	fmt.Println("Creating appendable chain...")
	acc, err := chainstorage.NewConcreteAppendableChainCreator(path,
		[]string{"time", "mediantime", "difficulty", "strippedsize", "size", "weight"},
		[]string{"size", "vsize", "weight"},
		delegated, false, true)
	if err != nil {
		return err
	}
	err = acc.Create()
	if err != nil {
		return err
	}
	ac, ind, err := acc.OpenWithIndexer()
	if err != nil {
		return err
	}

	fmt.Println("Creating transaction indexer...")
	var transactionIndexer transactionindexing.ITransactionIndexer = nil
	if transactionIndexingMethod == "delegated" {
		transactionIndexer = ind
	} else if transactionIndexingMethod == "separate files" {
		transactionIndexer = jsonblock.CreateOpenTransactionIndexerFiles(path)
	} else {
		panic("incorrect parameter " + transactionIndexingMethod)
	}

	readerPool := corereader.NewPool(1)
	workerPool := concurrency.NewWorkerPool(20)

	statusMap := sync.Map{}

	// Going parallel now

	// These two are much further downstream. But we create them here as here is the only good place
	// for the cutoff valve when the sequencer becomes bloated!
	sequencedChan := make(chan *jsonblock.JsonBlockEssential)
	sequencer := concurrency.NewSequencerContainer(0, 30, &sequencedChan)

	// A channel to receive []byte slices a block at a time
	haveReadChannel := make(chan *corereader.Task)

	// A go routine to squirt all the block requests through readerPool
	go func() {
		for height := int64(0); height <= lastBlock+1; height++ {
			task := corereader.NewTask(height, &haveReadChannel)
			// Squirt it into the pool
			statusMap.Store("startCoreReader", "WaitingBloated")
			sequencer.WaitForNotBloated()
			statusMap.Store("startCoreReader", "WaitedNotBloated")
			statusMap.Store("startCoreReader", "Squirting")
			readerPool.InChan <- task
			statusMap.Store("startCoreReader", "Squirted")
		}
		statusMap.Store("startCoreReader", "FINISHING...")
		close(readerPool.InChan)
		readerPool.Flush()
		close(haveReadChannel)
		statusMap.Store("startCoreReader", "FINISHED")
	}()

	// Once each block's []byte slice is available, we need to parse it
	// We use the worker pool for this

	// We'll need a new Task type that exposes the correct interface
	// (see above)

	preprocessedChan := make(chan *preprocessTask)

	// A go func to take things out of haveReadChannel and squirt them
	// through the worker pool
	go func() {
		for obj := range haveReadChannel {
			statusMap.Store("haveReadToWorker", "transferring")
			byts := obj.ResultBytes
			tsk := preprocessTask{}
			tsk.blockBytes = byts
			tsk.outChan = &preprocessedChan

			// This is NOT a good place to sequencer.WaitForNotBloated()
			// A good place for doing so MUST be somewhere where the items
			// are still in sequence!

			statusMap.Store("haveReadToWorker", "squirting")
			workerPool.InChan <- &tsk
			statusMap.Store("haveReadToWorker", "squirted")
		}
		statusMap.Store("haveReadToWorker", "FINISHING...")
		close(workerPool.InChan)
		statusMap.Store("haveReadToWorker", "FINISHED")
	}()

	// Preprocessed blocks will now come through preprocessedChan
	// We now need to squirt them through a SequencerContainer to
	// get them back in the right order

	// Squirt them into the sequencer
	go func() {
		statusMap.Store("preprocessedToSequencer", "Starting")
		for blk := range preprocessedChan {
			statusMap.Store("preprocessedToSequencer", "Squirting")
			sequencer.InChan <- blk.jsonBlock
			statusMap.Store("preprocessedToSequencer", "Squirted")
		}
		statusMap.Store("preprocessedToSequencer", "FINISHING...")
		close(sequencer.InChan)
		statusMap.Store("preprocessedToSequencer", "FINISHED")
	}()

	// Final destination for the sequenced blocks
	var aOneBlockHolder = jsonblock.CreateOneBlockHolder(transactionIndexer)

	// Squirt the sequenced blocks into the holder
	go func() {
		statusMap.Store("sequencedIntoHolder", "Starting")
		for b := range sequencedChan {
			statusMap.Store("sequencedIntoHolder", "Squirting")
			aOneBlockHolder.InChan <- b
			statusMap.Store("sequencedIntoHolder", "Squirted")
		}
		statusMap.Store("sequencedIntoHolder", "FINISHING...")
		close(aOneBlockHolder.InChan)
		statusMap.Store("sequencedIntoHolder", "FINISHED")
	}()

	hBlock := aOneBlockHolder.GenesisBlock()
	block, err := aOneBlockHolder.BlockInterface(hBlock)
	if err != nil {
		return err
	}
	height := block.Height()
	transactions := int64(0)
	for height <= lastBlock {
		if height%1000 == 0 {
			t := time.Now()
			fmt.Println(t.Format("Mon Jan 2 15:04:05"))
			fmt.Println("Block ", height, " Transaction ", transactions)
			// The sync is just so we can see
			// file sizes in explorer during processing
			if height%10000 == 0 {
				err := ac.Sync()
				if err != nil {
					return err
				}
				runtime.GC()
			}
		}

		/*		i := 0
				statusMap.Range(func(key, value interface{}) bool {
					fmt.Printf("\t%v: %v\n", key, value)
					i++
					return true
				})*/

		err = ac.AppendBlock(aOneBlockHolder, block)
		if err != nil {
			ac.Close()
			return err
		}

		count, _ := block.TransactionCount()
		transactions += count

		/*		i = 0
				statusMap.Range(func(key, value interface{}) bool {
					fmt.Printf("\t%v: %v\n", key, value)
					i++
					return true
				})*/

		hBlock, err = aOneBlockHolder.NextBlock(hBlock)
		if err != nil {
			ac.Close()
			return err
		}
		block, err = aOneBlockHolder.BlockInterface(hBlock)
		if err != nil {
			ac.Close()
			return err
		}
		height = block.Height()
	}
	ac.Close()
	if transactionIndexingMethod != "delegated" {
		transactionIndexer.Close()
	}
	fmt.Println("Done Several Years")
	return nil
}
