package jobs

import (
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"sync"
	"time"

	"github.com/KitchenMishap/pudding-shed/chainstorage"
	"github.com/KitchenMishap/pudding-shed/concurrency"
	"github.com/KitchenMishap/pudding-shed/corereader2"
	"github.com/KitchenMishap/pudding-shed/indexedhashes"
	"github.com/KitchenMishap/pudding-shed/indexedhashes3"
	"github.com/KitchenMishap/pudding-shed/intrinsicobjectscri"
	"github.com/KitchenMishap/pudding-shed/jsonblock"
	"github.com/KitchenMishap/pudding-shed/transactionindexing"
)

func RunIntrinsic(path string, transactionIndexingMethod string, years int, threads int, gbMem int,
	doPhase1 bool, doPhase2 bool, doPhase3 bool, phase3BlockLimit int64) error {

	dateFormat := "Mon Jan 2 15:04:05"

	blocks := blocksEachYear[years]

	// ===============================
	// FIRST we just gather the hashes
	if doPhase1 {
		fmt.Println(time.Now().Format(dateFormat) + "\t\tRemoving previous files...")
		err := os.RemoveAll(path)
		if err != nil {
			return err
		}

		err = PhaseOneParallelIntrinsic(path, blocks, threads)
		if err != nil {
			return err
		}
	}

	// ==========================
	// SECOND we index the hashes
	if doPhase2 {
		phase := ""
		phaseName := ""
		phaseStart := time.Now()
		timeTaken := time.Now().Sub(time.Now())
		mins := timeTaken.Minutes()
		sTimeUpdate := ""

		phase = "2 of 3"
		phaseName = "Index the hashes"
		phaseStart = time.Now()
		fmt.Println(phaseStart.Format(dateFormat)+"\tPHASE ", phase, "\t"+phaseName+"...")
		sep := string(os.PathSeparator)
		var bParams *indexedhashes3.HashIndexingParams = nil
		var tParams *indexedhashes3.HashIndexingParams = nil
		var aParams *indexedhashes3.HashIndexingParams = nil
		if years == 2 {
			bParams = indexedhashes3.Sensible2YearsBlockHashParams()
			tParams = indexedhashes3.Sensible2YearsTransactionHashParams()
			aParams = indexedhashes3.Sensible2YearsAddressHashParams()
		} else {
			bParams = indexedhashes3.Sensible16YearsBlockHashParams()
			tParams = indexedhashes3.Sensible16YearsTransactionHashParams()
			aParams = indexedhashes3.Sensible16YearsAddressHashParams()
		}
		_, bpl, err := indexedhashes3.NewHashStoreCreatorAndPreloader(path, "Blocks"+sep+"Hashes", bParams, gbMem)
		if err != nil {
			return err
		}
		_, tpl, err := indexedhashes3.NewHashStoreCreatorAndPreloader(path, "Transactions"+sep+"Hashes", tParams, gbMem)
		if err != nil {
			return err
		}
		_, apl, err := indexedhashes3.NewHashStoreCreatorAndPreloader(path, "Addresses"+sep+"Hashes", aParams, gbMem)
		if err != nil {
			return err
		}

		stepStart := time.Now()
		err = bpl.IndexTheHashes()
		if err != nil {
			return err
		}
		runtime.GC()
		debug.FreeOSMemory()
		timeTaken = time.Now().Sub(stepStart)
		mins = timeTaken.Minutes()
		sTimeUpdate = fmt.Sprintf("%s\tBLOCKS STEP took %.1f mins", time.Now().Format(dateFormat), mins)
		fmt.Println(sTimeUpdate)

		stepStart = time.Now()
		err = tpl.IndexTheHashes()
		if err != nil {
			return err
		}
		runtime.GC()
		debug.FreeOSMemory()
		timeTaken = time.Now().Sub(stepStart)
		mins = timeTaken.Minutes()
		sTimeUpdate = fmt.Sprintf("%s\tTRANSACTIONS STEP took %.1f mins", time.Now().Format(dateFormat), mins)
		fmt.Println(sTimeUpdate)

		stepStart = time.Now()
		err = apl.IndexTheHashes()
		if err != nil {
			return err
		}
		runtime.GC()
		debug.FreeOSMemory()
		timeTaken = time.Now().Sub(stepStart)
		mins = timeTaken.Minutes()
		sTimeUpdate = fmt.Sprintf("%s\tADDRESSES STEP took %.1f mins", time.Now().Format(dateFormat), mins)
		fmt.Println(sTimeUpdate)

		timeTaken = time.Now().Sub(phaseStart)
		mins = timeTaken.Minutes()
		sTimeUpdate = fmt.Sprintf("%s\tPHASE %s took %.1f mins", time.Now().Format(dateFormat), phase, mins)
		fmt.Println(sTimeUpdate)

		/*
			fmt.Println("Testing Block hashes:")
			err = bpl.TestTheHashes()
			if err != nil {
				return err
			}
			fmt.Println("Testing Transaction hashes:")
			err = tpl.TestTheHashes()
			if err != nil {
				return err
			}
		*/
	}

	// =======================================================================
	// THIRD we go through the blockchain again, gathering main data this time
	if doPhase3 {
		phase := "3 of 3"
		phaseName := "Gather the blockchain"

		phaseStart := time.Now()
		fmt.Println(phaseStart.Format(dateFormat)+"\tPHASE ", phase, "\t"+phaseName+"...")

		lastBlock := blocksEachYear[years] - 1

		if phase3BlockLimit != 0 && phase3BlockLimit < lastBlock {
			lastBlock = phase3BlockLimit - 1
		}

		//transactionsCount := int64(0)
		transactionsTarget := transactionsEachYear[years]
		//lastTrans := int64(0)

		// We WILL need an indexer this time!
		delegated := transactionIndexingMethod == "delegated"

		acc, err := chainstorage.NewConcreteAppendableChainCreator(path,
			[]string{"time", "mediantime", "difficulty", "strippedsize", "size", "weight"},
			[]string{"size", "vsize", "weight"},
			delegated, false, true)
		if err != nil {
			return err
		}
		err = acc.CreateFromHashStores()
		if err != nil {
			return err
		}
		ac, ind, err := acc.OpenWithIndexer()
		if err != nil {
			return err
		}

		var transactionIndexer transactionindexing.ITransactionIndexer = nil
		if transactionIndexingMethod == "delegated" {
			transactionIndexer = ind
		} else if transactionIndexingMethod == "separate files" {
			transactionIndexer = jsonblock.CreateOpenTransactionIndexerFiles(path)
		} else {
			panic("incorrect parameter " + transactionIndexingMethod)
		}

		if parallel {
			limitedBlocks := blocks
			if phase3BlockLimit != 0 && phase3BlockLimit < limitedBlocks {
				limitedBlocks = phase3BlockLimit
			}

			err = PhaseThreeParallelIntrinsic(limitedBlocks, lastBlock, transactionsTarget, ac, transactionIndexer, threads)
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("serial mode not yet implemented")
		}

		timeTaken := time.Now().Sub(phaseStart)
		mins := timeTaken.Minutes()
		sTimeUpdate := fmt.Sprintf("%s\tPHASE %s took %.1f mins", time.Now().Format(dateFormat), phase, mins)
		fmt.Println(sTimeUpdate)
	}

	return nil
}

func PhaseOneParallelIntrinsic(path string, blocks int64, threads int) error {
	hcc := chainstorage.NewConcreteHashesChainCreator(path)
	err := hcc.Create()
	if err != nil {
		return err
	}
	hc, err := hcc.Open()
	if err != nil {
		return err
	}

	// Now we start messing with channels and sequencers

	orderedBlockHashesChannel := make(chan indexedhashes.Sha256)
	fmt.Println("Starting to stream block hashes")
	go func() {
		err := corereader2.StreamBlockHashesFromGenesis(blocks, orderedBlockHashesChannel)
		if err != nil {
			panic(err) // ToDo
		}
	}()

	// These are used further downstream, but we need seq back here in order to ask if it's bloated
	orderedNumberedBlockChannel := make(chan *corereader2.HeightNumberedBlock)
	seq := concurrency.NewSequencerContainer[*corereader2.HeightNumberedBlock](0, 100, orderedNumberedBlockChannel)

	nobbledBlockHashesChannel := make(chan indexedhashes.Sha256)
	fmt.Println("Starting to pass hashes to GetBlocksBinary")
	// They won't be ordered any more when they come out.
	go func() {
		for hash := range orderedBlockHashesChannel {
			seq.WaitForNotBloated() // Nobble the flow if it's bloated further downstream
			nobbledBlockHashesChannel <- hash
		}
		close(nobbledBlockHashesChannel)
	}()

	fmt.Println("Starting get and parse blocks")
	corereader2.GetAndParseBlocksBinary(nobbledBlockHashesChannel, seq.InChan, threads)

	fmt.Println("Starting output")
	chain := intrinsicobjectscri.CreateOneBlockHolder(nil)
	go func() {
		for numberedBlock := range orderedNumberedBlockChannel {
			chain.InChan <- &numberedBlock.Block
		}
		close(chain.InChan)
	}()

	blockHeight := int64(0)
	hBlock := chain.GenesisBlock()
	for !hBlock.IsInvalid() {
		err = hc.AppendHashesCri(chain, hBlock, blockHeight)
		if err != nil {
			return err
		}
		hBlock, err = chain.NextBlock(hBlock)
		if err != nil {
			return err
		}
		blockHeight++
	}

	hc.Close()
	return nil
}

func PhaseThreeParallelIntrinsic(blocks int64, lastBlock int64, transactionsTarget int64, ac chainstorage.IAppendableChain,
	transactionIndexer transactionindexing.ITransactionIndexer, threads int) error {

	dateFormat := "Mon Jan 2 15:04:05"
	progressInterval := 5 * time.Second
	phaseStart := time.Now()
	nextProgress := phaseStart.Add(progressInterval)
	transactionsCount := int64(0)
	lastTrans := int64(0)

	statusMap := sync.Map{}

	// Now we start messing with channels and sequencers

	orderedBlockHashesChannel := make(chan indexedhashes.Sha256)
	fmt.Println("Starting to stream block hashes")
	go func() {
		err := corereader2.StreamBlockHashesFromGenesis(blocks, orderedBlockHashesChannel)
		if err != nil {
			panic(err) // ToDo
		}
	}()

	// These are used further downstream, but we need seq back here in order to ask if it's bloated
	orderedNumberedBlockChannel := make(chan *corereader2.HeightNumberedBlock)
	seq := concurrency.NewSequencerContainer[*corereader2.HeightNumberedBlock](0, 100, orderedNumberedBlockChannel)

	nobbledBlockHashesChannel := make(chan indexedhashes.Sha256)
	fmt.Println("Starting to pass hashes to GetBlocksBinary")
	// They won't be ordered any more when they come out.
	go func() {
		for hash := range orderedBlockHashesChannel {
			seq.WaitForNotBloated() // Nobble the flow if it's bloated further downstream
			nobbledBlockHashesChannel <- hash
		}
		close(nobbledBlockHashesChannel)
	}()

	fmt.Println("Starting get and parse blocks")
	corereader2.GetAndParseBlocksBinary(nobbledBlockHashesChannel, seq.InChan, threads)

	// Final destination for the sequenced blocks
	var aOneBlockHolder = intrinsicobjectscri.CreateOneBlockHolder(transactionIndexer)

	// Squirt the sequenced blocks into the holder
	go func() {
		statusMap.Store("sequencedIntoHolder", "Starting")
		for b := range orderedNumberedBlockChannel {
			statusMap.Store("sequencedIntoHolder", "Squirting")
			aOneBlockHolder.InChan <- &b.Block
			statusMap.Store("sequencedIntoHolder", "Squirted")
		}
		statusMap.Store("sequencedIntoHolder", "FINISHING...")
		close(aOneBlockHolder.InChan)
		statusMap.Store("sequencedIntoHolder", "FINISHED")
	}()

	hBlock := aOneBlockHolder.GenesisBlock()
	fmt.Println()
	for !hBlock.IsInvalid() {
		block, err := aOneBlockHolder.BlockInterface(hBlock)
		if err != nil {
			ac.Close()
			return err
		}
		height := block.Height()
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

		hBlock, err = aOneBlockHolder.NextBlock(hBlock)
		if err != nil {
			ac.Close()
			return err
		}
	}
	fmt.Println()
	ac.Close()
	return nil
}
