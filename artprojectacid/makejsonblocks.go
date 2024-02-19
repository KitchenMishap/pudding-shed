package artprojectacid

import (
	"encoding/json"
	"github.com/KitchenMishap/pudding-shed/chainreadinterface"
	"github.com/KitchenMishap/pudding-shed/chainstorage"
	"os"
)

func makeBlock(ibc chainreadinterface.IBlockChain, ibh chainreadinterface.IBlockHandle) JsonBlockType {
	res := JsonBlockType{}
	block, err := ibc.BlockInterface(ibh)
	if err != nil {
		panic(err)
	}

	nei, err := block.NonEssentialInts()
	if err != nil {
		panic(err)
	}

	res.Height = int(block.Height())
	res.MedianTime = int((*nei)["mediantime"])
	res.SizeBytes = int((*nei)["size"])
	res.ColourByte0, res.ColourByte1, res.ColourByte2 = colourBytes(ibc, ibh)

	return res
}

func gatherBlocks(ibc chainreadinterface.IBlockChain, lastBlock int) *JsonBlockArray {
	arrayObject := JsonBlockArray{}
	arrayObject.Blocks = make([]JsonBlockType, lastBlock+1)

	hBlock := ibc.GenesisBlock()
	for height := 0; height <= lastBlock; height++ {
		arrayObject.Blocks[height] = makeBlock(ibc, hBlock)

		var err error
		hBlock, err = ibc.NextBlock(hBlock)
		if err != nil {
			panic(err)
		}
	}

	return &arrayObject
}

func GatherBlocksToFile(blockchainFolder string, lastBlock int, filename string) {
	creator, err := chainstorage.NewConcreteAppendableChainCreator(blockchainFolder)
	if err != nil {
		panic(err)
	}
	writeablechain, _, err := creator.Open(false)
	if err != nil {
		panic(err)
	}
	chain := writeablechain.GetAsChainReadInterface()

	arrayObject := gatherBlocks(chain, lastBlock)
	theBytes, err := json.MarshalIndent(arrayObject, "", "    ")
	if err != nil {
		panic(err)
	}
	permissions := os.FileMode(0644)
	err = os.WriteFile(filename, theBytes, permissions)
	if err != nil {
		panic(err)
	}
}
