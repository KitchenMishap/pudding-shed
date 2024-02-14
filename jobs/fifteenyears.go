package jobs

import (
	"fmt"
	"github.com/KitchenMishap/pudding-shed/chainstorage"
	"github.com/KitchenMishap/pudding-shed/corereader"
	"github.com/KitchenMishap/pudding-shed/jsonblock"
	"os"
	"time"
)

func FifteenYearsPrimaries() error {
	const path = "F:\\Data\\FifteenYears"
	const lastBlock = int64(824196)

	err := os.RemoveAll(path)
	if err != nil {
		return err
	}

	var aReader corereader.CoreReader
	var aOneBlockChain = jsonblock.CreateOneBlockChain(&aReader, path)

	acc, err := chainstorage.NewConcreteAppendableChainCreator(path)
	if err != nil {
		return err
	}
	err = acc.Create([]string{"time", "mediantime", "difficulty", "strippedsize", "size", "weight"},
		[]string{"size", "vsize", "weight"})
	if err != nil {
		return err
	}
	ac, _, err := acc.Open()
	if err != nil {
		return err
	}

	hBlock := aOneBlockChain.GenesisBlock()
	block, err := aOneBlockChain.BlockInterface(hBlock)
	if err != nil {
		return err
	}
	height := block.Height()
	for height <= lastBlock {
		if height%1000 == 0 {
			t := time.Now()
			fmt.Println(t.Format("Mon Jan 2 15:04:05"))
			fmt.Println("Block ", height)
		}
		err = ac.AppendBlock(aOneBlockChain, block)
		if err != nil {
			ac.Close()
			return err
		}

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
	err = ac.Close()
	if err != nil {
		return err
	}
	println("Done Fifteen Years")
	return nil
}
