package chainstorage

import (
	"github.com/KitchenMishap/pudding-shed/chainreadinterface"
	"github.com/KitchenMishap/pudding-shed/tinychain"
	"testing"
)

func TestCopyOfTinyChain_Helper(bc chainreadinterface.IBlockChain, t *testing.T) {
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
