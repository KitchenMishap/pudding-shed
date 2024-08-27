package jobs

import (
	"fmt"
	"github.com/KitchenMishap/pudding-shed/chainstorage"
	"github.com/KitchenMishap/pudding-shed/corereader"
	"github.com/KitchenMishap/pudding-shed/jsonblock"
	"github.com/KitchenMishap/pudding-shed/transactionindexing"
	"os"
	"time"
)

func SeveralYearsPrimaries(years int, transactionIndexingMethod string) error {
	const path = "F:\\Data\\858000AddressesCswParents"
	lastBlock := int64(10000) // Default
	if years == 1 {
		lastBlock = 33000
	} else if years == 2 {
		lastBlock = 66000
	} else if years == 15 {
		lastBlock = 824196
	} else if years == 16 {
		lastBlock = 858000 // 23 Aug 2024
	}

	err := os.RemoveAll(path)
	if err != nil {
		return err
	}

	delegated := transactionIndexingMethod == "delegated"

	acc, err := chainstorage.NewConcreteAppendableChainCreator(path)
	if err != nil {
		return err
	}
	err = acc.Create([]string{"time", "mediantime", "difficulty", "strippedsize", "size", "weight"},
		[]string{"size", "vsize", "weight"})
	if err != nil {
		return err
	}
	ac, cac, err := acc.Open(delegated)
	if err != nil {
		return err
	}

	var transactionIndexer transactionindexing.ITransactionIndexer = nil
	if transactionIndexingMethod == "delegated" {
		transactionIndexer = cac.GetAsDelegatedTransactionIndexer()
	} else if transactionIndexingMethod == "separate files" {
		transactionIndexer = jsonblock.CreateOpenTransactionIndexerFiles(path)
	} else {
		panic("incorrect parameter " + transactionIndexingMethod)
	}

	var aReader corereader.CoreReader
	var aOneBlockChain = jsonblock.CreateOneBlockChain(&aReader, transactionIndexer)

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
			err := ac.Sync()
			if err != nil {
				return err
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
