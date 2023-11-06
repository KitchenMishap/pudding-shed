package chainstorage

import (
	"github.com/KitchenMishap/pudding-shed/chainreadinterface"
	"github.com/KitchenMishap/pudding-shed/tinychain"
	"testing"
)

func TestCopyOfTinyChain_Helper(crc chainreadinterface.IBlockChain, t *testing.T) {
	tinychain.TestThirdTransaction_helper(crc, t)
	tinychain.TestLatestNextTransaction_helper(crc, t)
	tinychain.TestTransaction_helper(crc, t)
	tinychain.TestLatestPrevNextBlock_helper(crc, t)
	tinychain.TestLatestBlockNotGenesis_helper(crc, t)
	tinychain.TestLatestNextBlock_helper(crc, t)
	tinychain.TestGenesisTransaction_helper(crc, t)
	tinychain.TestGenesisNextParent_helper(crc, t)
	tinychain.TestGenesisParentInvalid_helper(crc, t)
	tinychain.TestGenesisBlock_helper(crc, t)
	tinychain.TestInvalidHandle_helper(crc, t)
	tinychain.TestGenesisHandle_helper(crc, t)
	hBlock0 := BlockHandle{HashHeight{height: 0, heightSpecified: true, hashSpecified: false}}
	hBlock00 := BlockHandle{HashHeight{height: 0, heightSpecified: true, hashSpecified: false}}
	hBlock1 := BlockHandle{HashHeight{height: 1, heightSpecified: true, hashSpecified: false}}
	tinychain.TestHashEquality_helper(hBlock0, hBlock00, hBlock1, t)
}
