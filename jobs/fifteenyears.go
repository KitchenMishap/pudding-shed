package jobs

import (
	"fmt"
	"github.com/KitchenMishap/pudding-shed/chainstorage"
	"github.com/KitchenMishap/pudding-shed/corereader"
	"github.com/KitchenMishap/pudding-shed/indexedhashes3"
	"github.com/KitchenMishap/pudding-shed/jsonblock"
	"github.com/KitchenMishap/pudding-shed/transactionindexing"
	"os"
	"runtime"
	"runtime/debug"
	"time"
)

const readerThreads = 25
const path = "E:\\NewParams888888"
const parallel = true

func SeveralYearsPrimaries(years int, transactionIndexingMethod string, doPhase1 bool, doPhase2 bool, doPhase3 bool) error {

	blocksEachYear := []int64{0, 32879, 100888, 160463, 215006, // Block heights at year 0..4
		278460, 337312, 391569, 446472, 502401, // years 5..9
		556874, 611138, 664332, 717044, 770225, // years 10..14
		824204, 877669, 888888, 970000, 1020000} // years 16..19 are Extrapolated estimates
	transactionsEachYear := []int64{0, 33100, 219927, 2134383, 10675169, // Transaction counts at year 0..4
		30358181, 55676798, 101532619, 184473419, 288705840, // years 5..9
		369960799, 489809099, 602400644, 699932865, 793116734, // years 10..14
		947356729, 1139137995, 1200000000, 1300000000, 1400000000} // years 15..19 (Estimates from 16)

	// Choose a number of blocks
	lastBlock := blocksEachYear[years] - 1
	transactionsCount := int64(0)
	transactionsTarget := transactionsEachYear[years]
	lastTrans := int64(0)

	dateFormat := "Mon Jan 2 15:04:05"

	var err error = nil

	if doPhase1 {
		fmt.Println(time.Now().Format(dateFormat) + "\t\tRemoving previous files...")
		err := os.RemoveAll(path)
		if err != nil {
			return err
		}
	}

	phase := ""
	phaseName := ""
	progressInterval := 5 * time.Second
	phaseStart := time.Now()
	nextProgress := time.Now()
	timeTaken := time.Now().Sub(time.Now())
	mins := timeTaken.Minutes()
	sTimeUpdate := ""

	// ===============================
	// FIRST we just gather the hashes
	if doPhase1 {
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
		nextProgress := phaseStart.Add(progressInterval)
		fmt.Println(phaseStart.Format(dateFormat)+"\tPHASE ", phase, "\t"+phaseName+"...")

		transactionsCount = int64(0)
		lastTrans = int64(0)

		if parallel {
			err = PhaseOneParallel(lastBlock, transactionsTarget, hc)
			if err != nil {
				return err
			}
		} else {
			// This object gets blocks from Bitcoin Core one by one, but just holding
			// the various hashes and nothing else
			aReaderForHashes := corereader.NewPool(readerThreads, false)
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

		timeTaken = time.Now().Sub(phaseStart)
		mins = timeTaken.Minutes()
		sTimeUpdate = fmt.Sprintf("%s\tPHASE %s took %.1f mins", time.Now().Format(dateFormat), phase, mins)
		fmt.Println(sTimeUpdate)

		_, transactionsCount, _, err = hc.CountHashes()
		hc.Close()
		if err != nil {
			return err
		}

		fmt.Println("For some reason I seem to want you to restart Bitcoin Core...")
		fmt.Println("Please restart Bitcoin Core and press ENTER here:")
		fmt.Scanln()
		fmt.Println("Thank you.")
	}

	// ==========================
	// SECOND we index the hashes
	if doPhase2 {
		phase = "2 of 3"
		phaseName = "Index the hashes"

		phaseStart = time.Now()
		nextProgress = phaseStart.Add(progressInterval)
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
		_, bpl, err := indexedhashes3.NewHashStoreCreatorAndPreloader(path, "Blocks"+sep+"Hashes", bParams)
		if err != nil {
			return err
		}
		_, tpl, err := indexedhashes3.NewHashStoreCreatorAndPreloader(path, "Transactions"+sep+"Hashes", tParams)
		if err != nil {
			return err
		}
		_, apl, err := indexedhashes3.NewHashStoreCreatorAndPreloader(path, "Addresses"+sep+"Hashes", aParams)
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

	// =======================================================================
	// THIRD we go through the blockchain again, gathering main data this time
	if doPhase3 {
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
			err = PhaseThreeParallel(lastBlock, transactionsTarget, ac, transactionIndexer)
			if err != nil {
				return err
			}
		} else {

			aReader := corereader.NewPool(readerThreads, true)
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
	}

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
