package jobs

import (
	"fmt"
	"github.com/KitchenMishap/pudding-shed/chainstorage"
	"github.com/KitchenMishap/pudding-shed/concurrency"
	"github.com/KitchenMishap/pudding-shed/corereader"
	"github.com/KitchenMishap/pudding-shed/jsonblock"
	"github.com/KitchenMishap/pudding-shed/transactionindexing"
	"os"
	"time"
)

type preprocessTask struct {
	blockBytes []byte
	jsonBlock  *jsonblock.JsonBlockEssential
	err        error
	outChan    chan *preprocessTask
}

func (ppt *preprocessTask) Process() error {
	var err error
	ppt.jsonBlock, err = jsonblock.ParseJsonBlock(ppt.blockBytes)
	return err
}
func (ppt *preprocessTask) SetError(err error) {
	ppt.err = err
}
func (ppt *preprocessTask) Done() {
	ppt.outChan <- ppt
}
func (ppt *preprocessTask) GetError() error {
	return ppt.err
}

func SeveralYearsParallel(years int, transactionIndexingMethod string) error {
	const path = "D:\\Data\\CurrentJob"
	lastBlock := int64(10000) // Default
	if years == 1 {
		lastBlock = 33000
	} else if years == 2 {
		lastBlock = 66000
	} else if years == 3 {
		lastBlock = 99000
	} else if years == 4 {
		lastBlock = 132000
	} else if years == 6 {
		lastBlock = 315360
	} else if years == 15 {
		lastBlock = 824196
	} else if years == 16 {
		lastBlock = 860530 // 09 Sep 2024
	}

	println("Removing previous files...")
	err := os.RemoveAll(path)
	if err != nil {
		return err
	}

	delegated := transactionIndexingMethod == "delegated"

	println("Creating appendable chain...")
	acc, err := chainstorage.NewConcreteAppendableChainCreator(path,
		[]string{"time", "mediantime", "difficulty", "strippedsize", "size", "weight"},
		[]string{"size", "vsize", "weight"},
		delegated)
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

	println("Creating transaction indexer...")
	var transactionIndexer transactionindexing.ITransactionIndexer = nil
	if transactionIndexingMethod == "delegated" {
		transactionIndexer = ind
	} else if transactionIndexingMethod == "separate files" {
		transactionIndexer = jsonblock.CreateOpenTransactionIndexerFiles(path)
	} else {
		panic("incorrect parameter " + transactionIndexingMethod)
	}

	readerPool := corereader.NewPool(10)
	workerPool := concurrency.NewWorkerPool(30)

	// Going parallel now

	// A channel to receive []byte slices a block at a time
	haveReadChannel := make(chan *corereader.Task)

	// A go routine to squirt all the block requests through readerPool
	go func() {
		for height := int64(0); height <= lastBlock; {
			task := corereader.NewTask(height, &haveReadChannel)
			// Squirt it into the pool
			readerPool.InChan <- task
		}
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
			byts := obj.ResultBytes
			tsk := preprocessTask{}
			tsk.blockBytes = byts
			tsk.outChan = preprocessedChan
			workerPool.InChan <- &tsk
		}
	}()

	// Preprocessed blocks will now come through preprocessedChan
	// We now need to squirt them through a SequencerContainer to
	// get them back in the right order

	sequencedChan := make(chan *jsonblock.JsonBlockEssential)

	sequencer := concurrency.NewSequencerContainer(0, 100, sequencedChan)

	// Squirt them into the sequencer
	go func() {
		for blk := range preprocessedChan {
			sequencer.InChan <- blk.jsonBlock
		}
	}()

	// Final destination for the sequenced blocks
	var aOneBlockHolder = jsonblock.CreateOneBlockHolder(transactionIndexer)

	// Squirt the sequenced blocks into the holder
	go func() {
		for b := range sequencedChan {
			aOneBlockHolder.InChan <- b
		}
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
			}
		}
		err = ac.AppendBlock(aOneBlockHolder, block)
		if err != nil {
			ac.Close()
			return err
		}

		count, _ := block.TransactionCount()
		transactions += count

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
	println("Done Several Years")
	return nil
}
