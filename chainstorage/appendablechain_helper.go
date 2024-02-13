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

func TestBlockNeiExistance_Helper(bc chainreadinterface.IBlockChain, neiName string, neiExpectedExistance bool, t *testing.T) {
	genesisHandle := bc.GenesisBlock()
	block, err := bc.BlockInterface(genesisHandle)
	if err != nil {
		t.FailNow()
	}
	nonEssentialInts, err := block.NonEssentialInts()
	if err != nil {
		t.FailNow()
	}
	_, neiExists := (*nonEssentialInts)[neiName]
	if neiExists != neiExpectedExistance {
		if neiExpectedExistance {
			t.Error("expected to find non-essential int " + neiName + " in block")
		} else {
			t.Error("didn't expect to find non-essential int " + neiName + " in block")
		}
	}
}

func TestTransactionNeiExistance_Helper(bc chainreadinterface.IBlockChain, neiName string, neiExpectedExistance bool, t *testing.T) {
	genesisHandle, err := bc.GenesisTransaction()
	if err != nil {
		t.FailNow()
	}
	trans, err := bc.TransInterface(genesisHandle)
	if err != nil {
		t.FailNow()
	}
	nonEssentialInts, err := trans.NonEssentialInts()
	if err != nil {
		t.FailNow()
	}
	_, neiExists := (*nonEssentialInts)[neiName]
	if neiExists != neiExpectedExistance {
		if neiExpectedExistance {
			t.Error("expected to find non-essential int " + neiName + " in block")
		} else {
			t.Error("didn't expect to find non-essential int " + neiName + " in block")
		}
	}
}
