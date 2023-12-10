package chainstorage

import (
	"github.com/KitchenMishap/pudding-shed/chainreadinterface"
	"github.com/KitchenMishap/pudding-shed/jsonblock"
	"github.com/KitchenMishap/pudding-shed/tinychain"
	"testing"
)

func TestCopyOfTinyChain_Helper(bc chainreadinterface.IBlockChain, t *testing.T) {
	tinychain.TestBlock2Trans2_helper(bc, t)
	tinychain.TestThirdTransaction_helper(bc, t)
	tinychain.TestLatestNextTransaction_helper(bc, t)
	tinychain.TestTransaction_helper(bc, t)
	tinychain.TestLatestPrevNextBlock_helper(bc, t)
	tinychain.TestLatestBlockNotGenesis_helper(bc, t)
	tinychain.TestLatestNextBlock_helper(bc, t)
	tinychain.TestGenesisTransaction_helper(bc, t)
	tinychain.TestGenesisNextParent_helper(bc, t)
	tinychain.TestGenesisParentInvalid_helper(bc, t)
	tinychain.TestGenesisBlock_helper(bc, t)
	tinychain.TestInvalidHandle_helper(bc, t)
	tinychain.TestGenesisHandle_helper(bc, t)
}

func TestCopyOfJsonRealChain_Helper(bc chainreadinterface.IBlockChain, t *testing.T) {
	tinychain.TestLatestNextTransaction_helper(bc, t) // These tinychain tests also apply to JsonRealChain
	tinychain.TestLatestPrevNextBlock_helper(bc, t)
	tinychain.TestLatestBlockNotGenesis_helper(bc, t)
	tinychain.TestLatestNextBlock_helper(bc, t)
	tinychain.TestGenesisTransaction_helper(bc, t)
	tinychain.TestGenesisNextParent_helper(bc, t)
	tinychain.TestGenesisParentInvalid_helper(bc, t)
	tinychain.TestGenesisBlock_helper(bc, t)
	tinychain.TestInvalidHandle_helper(bc, t)
	tinychain.TestGenesisHandle_helper(bc, t)
	jsonblock.TestJustFiveCoinbaseBlocks_helper(bc, t)
}
