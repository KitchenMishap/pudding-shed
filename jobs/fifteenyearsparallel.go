package jobs

import (
	"fmt"
	"github.com/KitchenMishap/pudding-shed/chainstorage"
	"github.com/KitchenMishap/pudding-shed/concurrency"
	"github.com/KitchenMishap/pudding-shed/corereader"
	"github.com/KitchenMishap/pudding-shed/jsonblock"
	"github.com/KitchenMishap/pudding-shed/transactionindexing"
	"os"
	"sync"
	"time"
)

const BLOATLIMIT = 10

type preprocessTask struct {
	blockBytes []byte
	jsonBlock  *jsonblock.JsonBlockEssential
	err        error
	outChan    *chan *preprocessTask
}

func (ppt *preprocessTask) Process() error {
	var err error
	// We seem to get errors when running heavily concurrently.
	// Lets try some retries to shake things up.
	for retries := 0; retries < 5; retries++ {
		ppt.jsonBlock, err = jsonblock.ParseJsonBlock(ppt.blockBytes)
		if err == nil {
			if retries != 0 {
				fmt.Println("Succeeded json parse on retry ", retries)
			}
			break
		} else {
			if retries == 0 {
				fmt.Println("\nJson parse error:", err.Error())
				fmt.Println("Failed json parse on first try... retrying")
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
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

	// We seem to get errors when running heavily concurrently.
	// Lets try some retries to shake things up.
	for retries := 0; retries < 5; retries++ {
		ppt.jsonBlock, err = jsonblock.ParseJsonBlockHashes(ppt.blockBytes)
		if err == nil {
			if retries != 0 {
				fmt.Println("Succeeded json parse on retry ", retries)
			}
			break
		} else {
			if retries == 0 {
				fmt.Println("\nJson parse error:", err.Error())
				fmt.Println("Failed json parse on first try... retrying")
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
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

func PhaseOneParallel(lastBlock int64, transactionsTarget int64, hc chainstorage.IAppendableHashesChain, threads int) error {
	dateFormat := "Mon Jan 2 15:04:05"
	phase := "1 of 3"
	phaseName := "Gather the hashes"
	progressInterval := 5 * time.Second
	phaseStart := time.Now()
	nextProgress := phaseStart.Add(progressInterval)
	fmt.Println(phaseStart.Format(dateFormat)+"\tPHASE ", phase, "\t"+phaseName+"...")
	transactionsCount := int64(0)
	lastTrans := int64(0)

	readerPool := corereader.NewPool(threads, false)
	//workerPool := concurrency.NewWorkerPool(THREADS)
	workerPool := concurrency.NewWorkerPool(1)
	statusMap := sync.Map{}

	// Going parallel now

	// These two are much further downstream. But we create them here as here is the only good place
	// for the cutoff valve when the sequencer becomes bloated!
	sequencedChan := make(chan *jsonblock.JsonBlockHashes)
	sequencer := concurrency.NewSequencerContainerHashes(0, BLOATLIMIT, &sequencedChan)

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
		fmt.Println("\nSquirted all requests into reader pool...")
		statusMap.Store("startCoreReader", "FINISHING...")
		close(readerPool.InChan)
		readerPool.Flush()
		close(haveReadChannel)
		statusMap.Store("startCoreReader", "FINISHED")
		fmt.Println("All block requests submitted...")
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
			err := obj.ResultErr
			if err != nil {
				fmt.Println()
				fmt.Println(err.Error())
				fmt.Println("Requesting again...")
				height := obj.BlockHeight
				task := corereader.NewTask(height, &haveReadChannel)
				readerPool.InChan <- task
			} else {
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
		}
		statusMap.Store("haveReadToWorker", "FINISHING...")
		close(workerPool.InChan)
		statusMap.Store("haveReadToWorker", "FINISHED")
		fmt.Println("All blocks passed on to worker for parsing...")
	}()

	// Preprocessed blocks will now come through preprocessedChan
	// We now need to squirt them through a SequencerContainer to
	// get them back in the right order

	// Squirt them into the sequencer
	go func() {
		statusMap.Store("preprocessedToSequencer", "Starting")
		for blk := range preprocessedChan {
			if blk.err != nil {
				fmt.Println()
				fmt.Println(blk.err.Error())
				file, _ := os.Create("ErrorBlock.json")
				file.Write(blk.blockBytes)
				file.Close()
				panic("error preprocessing block, block written to ErrorBlock.json")
			} else {
				statusMap.Store("preprocessedToSequencer", "Squirting")
				sequencer.InChan <- blk.jsonBlock
				statusMap.Store("preprocessedToSequencer", "Squirted")
			}
		}
		statusMap.Store("preprocessedToSequencer", "FINISHING...")
		close(sequencer.InChan)
		statusMap.Store("preprocessedToSequencer", "FINISHED")
		fmt.Println("All blocks passed on to sequencer...")
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
		fmt.Println("All blocks passed out of sequencer...")
		statusMap.Store("sequencedIntoHolder", "FINISHING...")
		close(aOneBlockHolder.InChan)
		statusMap.Store("sequencedIntoHolder", "FINISHED")
		fmt.Println("All blocks passed on to holder...")
	}()

	hBlock := aOneBlockHolder.GenesisBlock()
	height := int64(hBlock.J_height)
	fmt.Println()
	for height <= lastBlock {
		if time.Now().Compare(nextProgress) > 0 || height == lastBlock {
			nextProgress = time.Now().Add(progressInterval)
			sLine := progressString(dateFormat, transactionsCount, transactionsCount-lastTrans, transactionsTarget, "transactions", progressInterval)
			fmt.Print(sLine + "\r")
			lastTrans = transactionsCount
		}

		// The sync is just so we can see
		// file sizes in explorer during processing
		if height%10000 == 0 {
			err := hc.Sync()
			if err != nil {
				return err
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
		_, transactionsCount, _, err = hc.CountHashes()

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
	fmt.Println()
	hc.Close()
	fmt.Println("Done Several Years Parallel Phase 1")
	return nil
}

func PhaseThreeParallel(lastBlock int64, transactionsTarget int64,
	ac chainstorage.IAppendableChain,
	transactionIndexer transactionindexing.ITransactionIndexer, threads int) error {

	dateFormat := "Mon Jan 2 15:04:05"
	phase := "1 of 3"
	phaseName := "Gather the hashes"
	progressInterval := 5 * time.Second
	phaseStart := time.Now()
	nextProgress := phaseStart.Add(progressInterval)
	fmt.Println(phaseStart.Format(dateFormat)+"\tPHASE ", phase, "\t"+phaseName+"...")
	transactionsCount := int64(0)
	lastTrans := int64(0)

	readerPool := corereader.NewPool(threads, true)
	//workerPool := concurrency.NewWorkerPool(THREADS)
	workerPool := concurrency.NewWorkerPool(1)

	statusMap := sync.Map{}

	// Going parallel now

	// These two are much further downstream. But we create them here as here is the only good place
	// for the cutoff valve when the sequencer becomes bloated!
	sequencedChan := make(chan *jsonblock.JsonBlockEssential)
	sequencer := concurrency.NewSequencerContainer(0, BLOATLIMIT, &sequencedChan)

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
			err := obj.ResultErr
			if err != nil {
				fmt.Println()
				fmt.Println(err.Error())
				fmt.Println("Requesting again...")
				height := obj.BlockHeight
				task := corereader.NewTask(height, &haveReadChannel)
				readerPool.InChan <- task
			} else {
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
			if blk.err != nil {
				fmt.Println()
				fmt.Println(blk.err.Error())
				file, _ := os.Create("ErrorBlock.json")
				file.Write(blk.blockBytes)
				file.Close()
				panic("error preprocessing block, block written to ErrorBlock.json")
			} else {
				statusMap.Store("preprocessedToSequencer", "Squirting")
				sequencer.InChan <- blk.jsonBlock
				statusMap.Store("preprocessedToSequencer", "Squirted")
			}
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
	fmt.Println()
	for height <= lastBlock {
		if time.Now().Compare(nextProgress) > 0 || height == lastBlock {
			nextProgress = time.Now().Add(progressInterval)
			sLine := progressString(dateFormat, transactionsCount, transactionsCount-lastTrans, transactionsTarget, "transactions", progressInterval)
			fmt.Print(sLine + "\r")
			lastTrans = transactionsCount
		}
		// The sync is just so we can see
		// file sizes in explorer during processing
		if height%10000 == 0 {
			err := ac.Sync()
			if err != nil {
				return err
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
		transactionsCount += count

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
	fmt.Println()
	ac.Close()
	fmt.Println("Done Several Years Parallel Phase 3")
	return nil
}
