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

	println("Removing previous files...")
	err := os.RemoveAll(path)
	if err != nil {
		return err
	}

	// FIRST we just gather the hashes

	hcc := chainstorage.NewConcreteHashesChainCreator(path)
	err = hcc.Create()
	if err != nil {
		return err
	}
	hc, err := hcc.Open()
	if err != nil {
		return err
	}

	aReaderForHashes := corereader.NewPool(10)
	// We don't need an indexer when just gathering hashes
	var aOneBlockChainForHashes = jsonblock.CreateOneBlockChain(aReaderForHashes, nil)

	hBlock := aOneBlockChainForHashes.GenesisBlock()
	block, err := aOneBlockChainForHashes.BlockInterface(hBlock)
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
				err := hc.Sync()
				if err != nil {
					return err
				}
				runtime.GC()
				debug.FreeOSMemory()
			}
		}
		err = hc.AppendBlock(aOneBlockChainForHashes, block)
		if err != nil {
			hc.Close()
			return err
		}

		count, _ := block.TransactionCount()
		transactions += count

		hBlock, err = aOneBlockChainForHashes.NextBlock(hBlock)
		if err != nil {
			hc.Close()
			return err
		}
		block, err = aOneBlockChainForHashes.BlockInterface(hBlock)
		if err != nil {
			hc.Close()
			return err
		}
		height = block.Height()
	}
	blocksCount, transactionsCount, addressesCount, err := hc.CountHashes()
	if err != nil {
		return err
	}
	hc.Close()

	// SECOND we index the hashes

	sep := string(os.PathSeparator)
	_, bpl := indexedhashes.NewUniformHashStoreCreatorAndPreloader(path, "Blocks"+sep+"Hashes", blocksCount, 2, 1)
	_, tpl := indexedhashes.NewUniformHashStoreCreatorAndPreloader(path, "Transactions"+sep+"Hashes", transactionsCount, 2, 1)
	_, apl := indexedhashes.NewUniformHashStoreCreatorAndPreloader(path, "Addresses"+sep+"Hashes", addressesCount, 2, 1)
	fmt.Println("Indexing the blocks...")
	err = bpl.IndexTheHashes()
	fmt.Println("Indexing the blocks...")
	err = tpl.IndexTheHashes()
	fmt.Println("Indexing the blocks...")
	err = apl.IndexTheHashes()
	fmt.Println("Done indexing")

	// THIRD we go through the blockchain again, gathering main data this time

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

	hBlock = aOneBlockChain.GenesisBlock()
	block, err = aOneBlockChain.BlockInterface(hBlock)
	if err != nil {
		return err
	}
	height = block.Height()
	transactions = int64(0)
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
	ac.Close()
	if transactionIndexingMethod != "delegated" {
		transactionIndexer.Close()
	}
	runtime.GC()
	debug.FreeOSMemory()
	println("Done Several Years")
	return nil
}
