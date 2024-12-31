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

const readerThreads = 25
const path = "F:\\Data\\CurrentJob"
const parallel = true

func SeveralYearsPrimaries(years int, transactionIndexingMethod string) error {

	blocksEachYear := []int64{0, 32879, 100888, 160463, 215006, // Block heights at year 0..4
		278460, 337312, 391569, 446472, 502401, // years 5..9
		556874, 611138, 664332, 717044, 770225, // years 10..14
		824204, 870000, 920000, 970000, 1020000} // years 16..19 are Extrapolated estimates
	transactionsEachYear := []int64{0, 33100, 219927, 2134383, 10675169, // Transaction counts at year 0..4
		30358181, 55676798, 101532619, 184473419, 288705840, // years 5..9
		369960799, 489809099, 602400644, 699932865, 793116734, // years 10..14
		947356729, 1136000000, 1200000000, 1300000000, 1400000000} // years 15..19 (Estimates from 16)

	// Choose a number of blocks
	lastBlock := blocksEachYear[years] - 1

	dateFormat := "Mon Jan 2 15:04:05"

	fmt.Println(time.Now().Format(dateFormat) + "\t\tRemoving previous files...")
	err := os.RemoveAll(path)
	if err != nil {
		return err
	}

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

	progressInterval := 5 * time.Second

	phaseStart := time.Now()
	nextProgress := phaseStart.Add(progressInterval)
	fmt.Println(phaseStart.Format(dateFormat)+"\tPHASE ", phase, "\t"+phaseName+"...")

	transactionsCount := int64(0)
	transactionsTarget := transactionsEachYear[years]
	lastTrans := int64(0)

	if parallel {
		err = PhaseOneParallel(lastBlock, hc)
		if err != nil {
			return err
		}
	} else {
		// This object gets blocks from Bitcoin Core one by one, but just holding
		// the various hashes and nothing else
		aReaderForHashes := corereader.NewPool(readerThreads)
		var aOneBlockChainForHashes = jsonblock.CreateHashesBlockChain(aReaderForHashes)

		for height := int64(0); height <= lastBlock; height++ {
			for year := 0; year <= 19; year++ {
				if height == blocksEachYear[year] {
					fmt.Printf("\nBlocks: %d, transactions: %d\n", height, transactionsCount)
				}
			}
			if time.Now().Compare(nextProgress) > 0 || height == lastBlock {
				nextProgress = time.Now().Add(progressInterval)
				sLine := progressString(dateFormat, transactionsCount, transactionsCount-lastTrans, transactionsTarget, "transactions", progressInterval)
				fmt.Print(sLine + "\r")
				lastTrans = transactionsCount
			}
			if height%1000 == 0 {
				err = hc.Sync() // Sync is just so we can see file sizes increase in Explorer
				if err != nil {
					return err
				}
				runtime.GC()
				debug.FreeOSMemory()
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
			_, transactionsCount, _, err = hc.CountHashes()
		}
		fmt.Println()
	}

	timeTaken := time.Now().Sub(phaseStart)
	mins := timeTaken.Minutes()
	sTimeUpdate := fmt.Sprintf("%s\tPHASE %s took %.1f mins", time.Now().Format(dateFormat), phase, mins)
	fmt.Println(sTimeUpdate)

	_, transactionsCount, _, err = hc.CountHashes()
	hc.Close()
	if err != nil {
		return err
	}

	// ==========================
	// SECOND we index the hashes
	phase = "2 of 3"
	phaseName = "Index the hashes"

	// Don't panic... when these estimates are exceeded as the blockchain gets bigger, there is no major problem!
	blocksCountEstimate := int64(1000000)   // There were 824,024 blocks after 15 years
	transCountEstimate := int64(1000000000) // There were 947,337,057 transactions after 15 years
	addrsCountEstimate := int64(3000000000) // There were 2,652,374,369 txos after 15 years, so there were fewer addresses

	phaseStart = time.Now()
	nextProgress = phaseStart.Add(progressInterval)
	fmt.Println(phaseStart.Format(dateFormat)+"\tPHASE ", phase, "\t"+phaseName+"...")
	sep := string(os.PathSeparator)
	_, bpl := indexedhashes.NewUniformHashStoreCreatorAndPreloader(path, "Blocks"+sep+"Hashes", blocksCountEstimate, 2, 1)
	_, tpl := indexedhashes.NewUniformHashStoreCreatorAndPreloader(path, "Transactions"+sep+"Hashes", transCountEstimate, 2, 1)
	_, apl := indexedhashes.NewUniformHashStoreCreatorAndPreloader(path, "Addresses"+sep+"Hashes", addrsCountEstimate, 2, 1)
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

	if parallel {
		err = PhaseThreeParallel(lastBlock, ac, transactionIndexer)
		if err != nil {
			return err
		}
	} else {

		aReader := corereader.NewPool(25)
		var aOneBlockChain = jsonblock.CreateOneBlockChain(aReader, transactionIndexer)

		hBlock := aOneBlockChain.GenesisBlock()
		block, err := aOneBlockChain.BlockInterface(hBlock)
		if err != nil {
			return err
		}
		height := block.Height()
		transactions := int64(0)
		for height <= lastBlock {
			if time.Now().Compare(nextProgress) > 0 || height == lastBlock {
				nextProgress = time.Now().Add(progressInterval)
				sLine := progressString(dateFormat, transactions, transactions-lastTrans, transactionsTarget, "transactions", progressInterval)
				fmt.Print(sLine + "\r")
				lastTrans = transactions
			}
			if height%10000 == 0 {
				// The sync is just so we can see
				// file sizes in explorer during processing
				err := ac.Sync()
				if err != nil {
					return err
				}
				runtime.GC()
				debug.FreeOSMemory()
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
	}

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
