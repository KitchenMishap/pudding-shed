package jobs

import (
	"fmt"
	"math"
	"os"
	"runtime"
	"runtime/debug"
	"time"

	"github.com/KitchenMishap/pudding-shed/chainreadinterface"
	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"github.com/KitchenMishap/pudding-shed/chainstorage"
	"github.com/KitchenMishap/pudding-shed/concurrency"
	"github.com/KitchenMishap/pudding-shed/corereader2"
	"github.com/KitchenMishap/pudding-shed/indexedhashes"
	"github.com/KitchenMishap/pudding-shed/indexedhashes3"
	"github.com/KitchenMishap/pudding-shed/intrinsicobjectscri"
	"github.com/KitchenMishap/pudding-shed/jsonblock"
	"github.com/KitchenMishap/pudding-shed/transactionindexing"
)

func RunIntrinsic(path string, useJson bool, transactionIndexingMethod string, years int, threads int, gbMem int,
	doPhase1 bool, doPhase2 bool, doPhase3 bool, phase3BlockLimit int64, isTest bool) error {

	var backslashR string
	if isTest {
		backslashR = "\n"
	} else {
		backslashR = "\r"
	}

	dateFormat := "Mon Jan 2 15:04:05"

	blocks := blocksEachYear[years]
	transactions := transactionsEachYear[years]

	// ===============================
	// FIRST we just gather the hashes
	if doPhase1 {
		fmt.Println(time.Now().Format(dateFormat) + "\t\tRemoving previous files...")
		err := os.RemoveAll(path)
		if err != nil {
			return err
		}

		err = PhaseOneParallelIntrinsic(path, useJson, blocks, transactions, threads, backslashR)
		if err != nil {
			return err
		}
		fmt.Println("")
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
		err = bpl.IndexTheHashes(threads)
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
		err = tpl.IndexTheHashes(threads)
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
		err = apl.IndexTheHashes(threads)
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

			err = PhaseThreeParallelIntrinsic(limitedBlocks, lastBlock, useJson, transactionsTarget, ac, transactionIndexer, threads, backslashR)
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

func PhaseOneParallelIntrinsic(path string, useJson bool, blocks int64, transactions int64, threads int, backslashR string) error {
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
	// They won't be ordered any more when they come out.
	go func() {
		for hash := range orderedBlockHashesChannel {
			seq.WaitForNotBloated() // Nobble the flow if it's bloated further downstream
			nobbledBlockHashesChannel <- hash
		}
		close(nobbledBlockHashesChannel)
	}()

	corereader2.GetAndParseBlocks(nobbledBlockHashesChannel, useJson, seq.InChan, threads)

	chain := intrinsicobjectscri.CreateOneBlockHolder(nil)
	go func() {
		for numberedBlock := range orderedNumberedBlockChannel {
			chain.InChan <- &numberedBlock.Block
		}
		close(chain.InChan)
	}()

	progress := NewProgress(0, transactions, 30,
		"transactions", "Gather Hashes", backslashR)

	blockHeight := int64(0)
	transCount := int64(0)
	hBlock := chain.GenesisBlock()
	for !hBlock.IsInvalid() {
		var blk chainreadinterface.IBlock
		blk, err = chain.BlockInterface(hBlock)
		if err != nil {
			return err
		}
		var transactionsInBlock int64
		transactionsInBlock, err = blk.TransactionCount()
		if err != nil {
			return err
		}

		err = hc.AppendHashesCri(chain, hBlock, blockHeight)
		if err != nil {
			return err
		}
		hBlock, err = chain.NextBlock(hBlock)
		if err != nil {
			return err
		}
		blockHeight++
		transCount += transactionsInBlock

		progress.Update(transCount)
	}

	hc.Close()
	return nil
}

func PhaseThreeParallelIntrinsic(blocks int64, lastBlock int64, useJson bool,
	transactionsTarget int64, ac chainstorage.IAppendableChain,
	transactionIndexer transactionindexing.ITransactionIndexer, threads int, backslashR string) error {

	transactionsCount := int64(0)

	// Now we start messing with channels and sequencers

	orderedBlockHashesChannel := make(chan indexedhashes.Sha256)
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
	// They won't be ordered any more when they come out.
	go func() {
		for hash := range orderedBlockHashesChannel {
			seq.WaitForNotBloated() // Nobble the flow if it's bloated further downstream
			nobbledBlockHashesChannel <- hash
		}
		close(nobbledBlockHashesChannel)
	}()

	corereader2.GetAndParseBlocks(nobbledBlockHashesChannel, useJson, seq.InChan, threads)

	// Final destination for the sequenced blocks
	var aOneBlockHolder = intrinsicobjectscri.CreateOneBlockHolder(transactionIndexer)

	// Squirt the sequenced blocks into the holder
	go func() {
		for b := range orderedNumberedBlockChannel {
			aOneBlockHolder.InChan <- &b.Block
		}
		close(aOneBlockHolder.InChan)
	}()

	progress := NewProgress(0, transactionsTarget, 30,
		"transactions", "gather blockchain", backslashR)

	hBlock := aOneBlockHolder.GenesisBlock()
	fmt.Println()
	for !hBlock.IsInvalid() {
		block, err := aOneBlockHolder.BlockInterface(hBlock)
		if err != nil {
			ac.Close()
			return err
		}
		height := block.Height()

		// The sync is just so we can see
		// file sizes in explorer during processing
		if height%10000 == 0 {
			err := ac.Sync()
			if err != nil {
				return err
			}
		}

		err = ac.AppendBlock(aOneBlockHolder, block)
		if err != nil {
			ac.Close()
			return err
		}

		var count int64
		count, err = block.TransactionCount()
		transactionsCount += count
		progress.Update(transactionsCount)

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

type Progress struct {
	startItem        int64
	endItem          int64
	intervalSeconds  int
	itemsName        string
	taskName         string
	firstUpdateGiven bool

	checkpointItem int64
	checkpointTime time.Time
	backslashR     string
	printer        *message.Printer
}

func NewProgress(startItem int64, endItem int64, intervalSeconds int, itemsName string, taskName string, backslashR string) *Progress {
	result := Progress{}
	result.startItem = startItem
	result.endItem = endItem
	result.intervalSeconds = intervalSeconds
	result.itemsName = itemsName
	result.taskName = taskName
	result.firstUpdateGiven = false

	// Presume progress now is startItems
	result.checkpointItem = startItem
	result.checkpointTime = time.Now()
	result.printer = message.NewPrinter(language.English)
	result.backslashR = backslashR
	return &result
}

func (p *Progress) Update(currentItem int64) {
	const dateFormat = "Mon Jan 2 15:04:05"

	now := time.Now()
	sDate := now.Format(dateFormat)
	if !p.firstUpdateGiven {
		fmt.Printf("%s Start of task \"%s\"\n", sDate, p.taskName)
		p.firstUpdateGiven = true
	}
	sinceCheckpoint := now.Sub(p.checkpointTime)
	if sinceCheckpoint > time.Duration(p.intervalSeconds)*time.Second {
		intervalItems := currentItem - p.checkpointItem
		intervalSeconds := sinceCheckpoint / time.Second
		itemsPerSecond := float64(intervalItems) / float64(intervalSeconds)
		secondsLeft := float64(p.endItem-currentItem) / itemsPerSecond
		prediction := now.Add(time.Duration(secondsLeft) * time.Second)
		msg := p.printer.Sprintf("%s: %d of %d %s done (%.2f%%), expect completion of \"%s\" at %s (rate %d)",
			sDate, currentItem-p.startItem, p.endItem-p.startItem, p.itemsName,
			float64(100)*float64(currentItem-p.startItem)/float64(p.endItem-p.startItem),
			p.taskName, prediction.Format(dateFormat), int(math.Round(itemsPerSecond)))
		fmt.Printf("%s%s          ", p.backslashR, msg)
		err := os.Stdout.Sync()
		if err != nil {
			panic(err)
		}

		p.checkpointTime = now
		p.checkpointItem = currentItem
	}
}
