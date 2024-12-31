package jobs

import (
	"fmt"
	"github.com/KitchenMishap/pudding-shed/chainstorage"
	"github.com/KitchenMishap/pudding-shed/concurrency"
	"github.com/KitchenMishap/pudding-shed/corereader"
	"github.com/KitchenMishap/pudding-shed/jsonblock"
	"github.com/KitchenMishap/pudding-shed/transactionindexing"
	"runtime"
	"runtime/debug"
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
	if err != nil {
		return err
	}
	jsonblock.PostJsonRemoveCoinbaseTxis(ppt.jsonBlock)
	jsonblock.PostJsonCalculateSatoshis(ppt.jsonBlock)
	err = jsonblock.PostJsonEncodeAddressHashes(ppt.jsonBlock)
	if err != nil {
		return err
	}
	err = jsonblock.PostJsonEncodeSha256s(ppt.jsonBlock)
	if err != nil {
		return err
	}
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

type preprocessTaskHashes struct {
	blockBytes []byte
	jsonBlock  *jsonblock.JsonBlockHashes
	err        error
	outChan    *chan *preprocessTaskHashes
}

func (ppt *preprocessTaskHashes) Process() error {
	var err error
	ppt.jsonBlock, err = jsonblock.ParseJsonBlockHashes(ppt.blockBytes)
	if err != nil {
		return err
	}
	err = jsonblock.PostJsonEncodeAddressHashes2(ppt.jsonBlock)
	if err != nil {
		return err
	}
	err = jsonblock.PostJsonEncodeSha256s2(ppt.jsonBlock)
	if err != nil {
		return err
	}
	return err
}
func (ppt *preprocessTaskHashes) SetError(err error) {
	ppt.err = err
}
func (ppt *preprocessTaskHashes) Done() {
	*ppt.outChan <- ppt
}
func (ppt *preprocessTaskHashes) GetError() error {
	return ppt.err
}

func PhaseOneParallel(lastBlock int64, hc chainstorage.IAppendableHashesChain) error {
	readerPool := corereader.NewPool(30)
	workerPool := concurrency.NewWorkerPool(30)
	statusMap := sync.Map{}

	// Going parallel now

	// These two are much further downstream. But we create them here as here is the only good place
	// for the cutoff valve when the sequencer becomes bloated!
	sequencedChan := make(chan *jsonblock.JsonBlockHashes)
	sequencer := concurrency.NewSequencerContainerHashes(0, 30, &sequencedChan)

	// A channel to receive []byte slices a block at a time
	haveReadChannel := make(chan *corereader.Task)

	// A go routine to squirt all the block requests through readerPool
	go func() {
		for height := int64(0); height <= lastBlock; height++ {
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

	preprocessedChan := make(chan *preprocessTaskHashes)

	// A go func to take things out of haveReadChannel and squirt them
	// through the worker pool
	go func() {
		for obj := range haveReadChannel {
			statusMap.Store("haveReadToWorker", "transferring")
			byts := obj.ResultBytes
			tsk := preprocessTaskHashes{}
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
	var aOneBlockHolder = jsonblock.CreateOneBlockHolderHashes()

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
	height := int64(hBlock.J_height)
	transactions := int64(0)
	fmt.Println()
	for height <= lastBlock {
		if height%1000 == 0 {
			t := time.Now()
			fmt.Print(t.Format("Mon Jan 2 15:04:05"), "Block ", height, " Transaction ", transactions, "\r")
			// The sync is just so we can see
			// file sizes in explorer during processing
			if height%10000 == 0 {
				err := hc.Sync()
				if err != nil {
					return err
				}
				runtime.GC()
				debug.FreeOSMemory()
			}
		}

		/*		i := 0
				statusMap.Range(func(key, value interface{}) bool {
					fmt.Printf("\t%v: %v\n", key, value)
					i++
					return true
				})*/

		err := hc.AppendHashes(hBlock)
		if err != nil {
			hc.Close()
			return err
		}

		count := int64(len(hBlock.J_tx))
		transactions += count

		/*		i = 0
				statusMap.Range(func(key, value interface{}) bool {
					fmt.Printf("\t%v: %v\n", key, value)
					i++
					return true
				})*/

		if height == lastBlock {
			height++
		} else {
			hBlock, err = aOneBlockHolder.NextBlock()
			if err != nil {
				hc.Close()
				return err
			}
			height = int64(hBlock.J_height)
		}
	}
	hc.Close()
	runtime.GC()
	debug.FreeOSMemory()
	fmt.Println("Done Several Years Parallel Phase 1")
	return nil
}

func PhaseThreeParallel(lastBlock int64,
	ac chainstorage.IAppendableChain,
	transactionIndexer transactionindexing.ITransactionIndexer) error {

	readerPool := corereader.NewPool(30)
	workerPool := concurrency.NewWorkerPool(30)

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
		for height := int64(0); height <= lastBlock; height++ {
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
	fmt.Println()
	for height <= lastBlock {
		if height%1000 == 0 {
			t := time.Now()
			fmt.Print(t.Format("Mon Jan 2 15:04:05"), "Block ", height, " Transaction ", transactions, "\r")
			// The sync is just so we can see
			// file sizes in explorer during processing
			if height%10000 == 0 {
				err := ac.Sync()
				if err != nil {
					return err
				}
				runtime.GC()
				debug.FreeOSMemory()
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

		if height == lastBlock {
			height++
		} else {
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
	}
	ac.Close()
	runtime.GC()
	debug.FreeOSMemory()
	fmt.Println("Done Several Years Parallel")
	return nil
}
