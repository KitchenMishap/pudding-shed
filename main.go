package main

import "github.com/KitchenMishap/pudding-shed/artprojectacid"

func main() {
	lastBlock := 824196 // 15 Years of blockchain

	artprojectacid.GatherBlocksToFile("F:\\Data\\SeveralYears", lastBlock, "F:\\Data\\AcidBlocks.json")
}
