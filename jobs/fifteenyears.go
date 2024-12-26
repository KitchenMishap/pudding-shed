package jobs

import (
	"fmt"
	"github.com/KitchenMishap/pudding-shed/chainstorage"
	"github.com/KitchenMishap/pudding-shed/corereader"
	"github.com/KitchenMishap/pudding-shed/indexedhashes"
	"github.com/KitchenMishap/pudding-shed/jsonblock"
	"github.com/KitchenMishap/pudding-shed/transactionindexing"
	"os"
	"runtime"
	"runtime/debug"
	"time"
)

func SeveralYearsPrimaries(years int, transactionIndexingMethod string) error {
	const path = "F:\\Data\\CurrentJob"

	// Choose a number of blocks
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
	} else if years == 7 {
		lastBlock = 391000 // 7 years = mnemonic for 100 million transactions
	} else if years == 15 {
		lastBlock = 840000
	} else if years == 16 {
		lastBlock = 876441 // 26 Dec 2024
	}

	dateFormat := "Mon Jan 2 15:04:05"

	fmt.Println(time.Now().Format(dateFormat) + "\t\tRemoving previous files...")
	err := os.RemoveAll(path)
	if err != nil {
		return err
	}

	// This object gets blocks from Bitcoin Core one by one, but just holding
	// the various hashes and nothing else
	aReaderForHashes := corereader.NewPool(10)
	var aOneBlockChainForHashes = jsonblock.CreateHashesBlockChain(aReaderForHashes)

	// ===============================
	// FIRST we just gather the hashes
	phase := "1 of 3"
	phaseName := "Gather the hashes"

	hcc := chainstorage.NewConcreteHashesChainCreator(path)
	err = hcc.Create()
	if err != nil {
		return err
	}
	hc, err := hcc.Open()
	if err != nil {
		return err
	}

	phaseStart := time.Now()
	fmt.Println(phaseStart.Format(dateFormat)+"\tPHASE ", phase, "\t"+phaseName+"...")
	lastTime := time.Now()
	for height := int64(0); height <= lastBlock; height++ {
		if height%1000 == 0 {
			t := time.Now()
			sline := progressString(dateFormat, height, 1000, lastBlock, "blocks", time.Now().Sub(lastTime))
			fmt.Print(sline + "\r")
			lastTime = t
			if height%10000 == 0 {
				err = hc.Sync() // Sync is just so we can see file sizes increase in Explorer
				if err != nil {
					return err
				}
				runtime.GC()
				debug.FreeOSMemory()
			}
		}
		block, err := aOneBlockChainForHashes.SwitchBlock(height)
		if err != nil {
			hc.Close()
			return err
		}

		err = hc.AppendHashes(block)
		if err != nil {
			return err
		}
	}
	fmt.Println()

	timeTaken := time.Now().Sub(phaseStart)
	mins := timeTaken.Minutes()
	sTimeUpdate := fmt.Sprintf("%s\tPHASE %s took %.1f mins", time.Now().Format(dateFormat), phase, mins)
	fmt.Println(sTimeUpdate)

	blocksCount, transactionsCount, addressesCount, err := hc.CountHashes()
	hc.Close()
	if err != nil {
		return err
	}

	// ==========================
	// SECOND we index the hashes
	phase = "2 of 3"
	phaseName = "Index the hashes"

	phaseStart = time.Now()
	fmt.Println(phaseStart.Format(dateFormat)+"\tPHASE ", phase, "\t"+phaseName+"...")
	sep := string(os.PathSeparator)
	_, bpl := indexedhashes.NewUniformHashStoreCreatorAndPreloader(path, "Blocks"+sep+"Hashes", blocksCount, 2, 1)
	_, tpl := indexedhashes.NewUniformHashStoreCreatorAndPreloader(path, "Transactions"+sep+"Hashes", transactionsCount, 2, 1)
	_, apl := indexedhashes.NewUniformHashStoreCreatorAndPreloader(path, "Addresses"+sep+"Hashes", addressesCount, 2, 1)
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

	// =======================================================================
	// THIRD we go through the blockchain again, gathering main data this time
	phase = "3 of 3"
	phaseName = "Gather the blockchain"

	phaseStart = time.Now()
	fmt.Println(phaseStart.Format(dateFormat)+"\tPHASE ", phase, "\t"+phaseName+"...")

	// We WILL need an indexer this time!
	delegated := transactionIndexingMethod == "delegated"

	acc, err := chainstorage.NewConcreteAppendableChainCreator(path,
		[]string{"time", "mediantime", "difficulty", "strippedsize", "size", "weight"},
		[]string{"size", "vsize", "weight"},
		delegated)
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

	aReader := corereader.NewPool(10)
	var aOneBlockChain = jsonblock.CreateOneBlockChain(aReader, transactionIndexer)

	hBlock := aOneBlockChain.GenesisBlock()
	block, err := aOneBlockChain.BlockInterface(hBlock)
	if err != nil {
		return err
	}
	height := block.Height()
	transactions := int64(0)
	for height <= lastBlock {
		if height%1000 == 0 {
			t := time.Now()
			sline := progressString(dateFormat, height, 1000, lastBlock, "blocks", time.Now().Sub(lastTime))
			fmt.Print(sline + "\r")
			lastTime = t
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
		err = ac.AppendBlock(aOneBlockChain, block)
		if err != nil {
			fmt.Println("Error when: AppendBlock(", block.Height(), ")")
			ac.Close()
			return err
		}

		count, _ := block.TransactionCount()
		transactions += count

		// If we were to call NextBlock on the last block we're interested in, there would be an attempt
		// to process transactions whose inputs come from transactions we haven't indexed.
		if height == lastBlock {
			height++ // Just increment so that the for loop ends
		} else {
			hBlock, err = aOneBlockChain.NextBlock(hBlock)
			if err != nil {
				ac.Close()
				return err
			}
			block, err = aOneBlockChain.BlockInterface(hBlock)
			if err != nil {
				ac.Close()
				fmt.Println("BlockInterface(", hBlock.Height(), ")")
				return err
			}
			height = block.Height()
		}
	}
	fmt.Println()
	ac.Close()
	if transactionIndexingMethod != "delegated" {
		transactionIndexer.Close()
	}
	runtime.GC()
	debug.FreeOSMemory()

	timeTaken = time.Now().Sub(phaseStart)
	mins = timeTaken.Minutes()
	sTimeUpdate = fmt.Sprintf("%s\tPHASE %s took %.1f mins", time.Now().Format(dateFormat), phase, mins)
	fmt.Println(sTimeUpdate)

	fmt.Println("Done Several Years")
	return nil
}

func progressString(dateFormat string,
	index int64, step int64, target int64, units string,
	lastDuration time.Duration) string {
	sDate := time.Now().Format(dateFormat)
	sProgress := fmt.Sprintf("%d %s of %d (%.1f%%)", index, units, target, 100.0*float32(index)/float32(target))
	seconds := lastDuration.Seconds() * float64(target-index) / float64(step)
	complete := time.Now().Add(time.Duration(seconds) * time.Second)
	sComplete := complete.Format(dateFormat)
	sExpected := fmt.Sprintf("Expect PHASE completion at %s", sComplete)
	result := sDate + ": " + sProgress + "; " + sExpected
	return result
}
