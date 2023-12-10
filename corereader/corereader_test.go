package corereader

import (
	"encoding/json"
	"github.com/KitchenMishap/pudding-shed/jsonblock"
	"testing"
)

func Test_GetBlockHeight0Hash(t *testing.T) {
	var reader CoreReader
	hashHex, err := reader.getHashHexByHeight(0)
	if err != nil {
		t.Error(err)
	}
	if hashHex != "000000000019d6689c085ae165831e934ff763ae46a2a6c172b3f1b60a8ce26f" {
		t.Error("hash of block height 0 is wrong")
	}
}

func Test_LotsOfBlocks(t *testing.T) {
	var reader CoreReader
	blocks, err := reader.CountBlocks()
	if err != nil {
		t.Error(err)
	}

	if blocks < 800_000 {
		t.Error("chain should contain more than 800,000 blocks")
	}
}

func Test_ParseGenesis(t *testing.T) {
	var reader CoreReader
	jsonBytes, err := reader.FetchBlockJsonBytes(0)
	if err != nil {
		t.Error(err)
	}

	var jsonBlock jsonblock.JsonBlockEssential
	err = json.Unmarshal(jsonBytes, &jsonBlock)
	if err != nil {
		t.Error(err)
	}
}
