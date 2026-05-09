package jobs

import (
	"fmt"
	"os"
	"time"

	"github.com/KitchenMishap/pudding-shed/chainstorage"
	"github.com/KitchenMishap/pudding-shed/concurrency"
	"github.com/KitchenMishap/pudding-shed/indexedhashes"
	"github.com/KitchenMishap/pudding-shed/indexedhashes3"
	"github.com/KitchenMishap/pudding-shed/intrinsicobjects"
)

func RunIntrinsic(path string, years int, blocks int64, gbMem int, doPhase1 bool, doPhase2 bool, doPhase3 bool) error {
	dateFormat := "Mon Jan 2 15:04:05"

	// ===============================
	// FIRST we just gather the hashes
	if doPhase1 {
		fmt.Println(time.Now().Format(dateFormat) + "\t\tRemoving previous files...")
		err := os.RemoveAll(path)
		if err != nil {
			return err
		}

		err = PhaseOneParallelIntrinsic(path, blocks)
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
		err = bpl.IndexTheHashes()
		if err != nil {
			return err
		}
		err = tpl.IndexTheHashes()
		if err != nil {
			return err
		}
		err = apl.IndexTheHashes()
		if err != nil {
			return err
		}

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
	return nil
}

func PhaseOneParallelIntrinsic(path string, blocks int64) error {
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
		err := intrinsicobjects.StreamBlockHashesFromGenesis(blocks, orderedBlockHashesChannel)
		if err != nil {
			panic(err) // ToDo
		}
	}()

	// These are used further downstream, but we need seq back here in order to ask if it's bloated
	orderedNumberedBlockChannel := make(chan *intrinsicobjects.HeightNumberedBlock)
	seq := concurrency.NewSequencerContainer[*intrinsicobjects.HeightNumberedBlock](0, 100, orderedNumberedBlockChannel)

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
	intrinsicobjects.GetAndParseBlocks(nobbledBlockHashesChannel, seq.InChan, 10) // ToDo Threads low for now

	fmt.Println("Starting output")
	for numberedBlock := range orderedNumberedBlockChannel {
		if numberedBlock.BlockHeight%10_000 == 0 {
			fmt.Printf("Block %d parsed\n", numberedBlock.BlockHeight)
		}
		err = hc.AppendHashesIntrinsic(&numberedBlock.Block, numberedBlock.BlockHeight)
	}
	hc.Close()
	return nil
}
