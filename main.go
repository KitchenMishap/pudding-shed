package main

import (
	"github.com/KitchenMishap/pudding-shed/artprojectacid"
	"os"
)

func main() {
	const lastBlock = 824196 // 15 Years of blockchain
	const dbDir = "F:\\Data\\SeveralYears"
	const opDir = "artprojectacid\\WorkingDir"

	if opDir != "" {
		err := os.MkdirAll(opDir, os.ModePerm)
		if err != nil {
			panic(err)
		}
	}

	artprojectacid.GatherBlocksToFile(dbDir, lastBlock, opDir+"\\acidblocks.json")
}
