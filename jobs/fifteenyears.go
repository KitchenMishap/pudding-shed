package jobs

import (
	"fmt"
	"github.com/KitchenMishap/pudding-shed/chainstorage"
	"github.com/KitchenMishap/pudding-shed/corereader"
	"github.com/KitchenMishap/pudding-shed/jsonblock"
	"github.com/KitchenMishap/pudding-shed/transactionindexing"
	"os"
	"runtime"
	"time"
)

func SeveralYearsPrimaries(years int, transactionIndexingMethod string) error {
	const path = "F:\\Data\\CurrentJob"
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
		lastBlock = 860530 // 09 Sep 2024
	}

	println("Removing previous files...")
	err := os.RemoveAll(path)
	if err != nil {
		return err
	}

	delegated := transactionIndexingMethod == "delegated"

	acc, err := chainstorage.NewConcreteAppendableChainCreator(path,
		[]string{"time", "mediantime", "difficulty", "strippedsize", "size", "weight"},
		[]string{"size", "vsize", "weight"},
		delegated, false, true)
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
			}
		}
		err = ac.AppendBlock(aOneBlockChain, block)
		if err != nil {
			ac.Close()
			return err
		}

		count, _ := block.TransactionCount()
		transactions += count

		hBlock, err = aOneBlockChain.NextBlock(hBlock)
		if err != nil {
			ac.Close()
			return err
		}
		block, err = aOneBlockChain.BlockInterface(hBlock)
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
