package chainstorage

import (
	"github.com/KitchenMishap/pudding-shed/corereader"
	"github.com/KitchenMishap/pudding-shed/jsonblock"
	"os"
	"testing"
)
import "github.com/KitchenMishap/pudding-shed/tinychain"

func TestCopyTinyChain(t *testing.T) {
	err := os.RemoveAll("Temp_Testing/Chain")
	if err != nil {
		t.Fail()
	}

	acc, err := NewConcreteAppendableChainCreator("Temp_Testing/Chain")
	if err != nil {
		t.Fail()
	}
	err = acc.Create([]string{"size", "time"}, []string{"vsize", "version"})
	if err != nil {
		t.Fail()
	}
	ac, cac, err := acc.Open(false)
	if err != nil {
		t.Fail()
	}

	hBlock := tinychain.TheTinyChain.GenesisBlock()
	block, err := tinychain.TheTinyChain.BlockInterface(hBlock)
	if err != nil {
		t.Fail()
	}
	err = ac.AppendBlock(tinychain.TheTinyChain, block)
	if err != nil {
		t.Fail()
	}

	hBlock, err = tinychain.TheTinyChain.NextBlock(hBlock)
	if err != nil {
		t.Fail()
	}
	for !hBlock.IsInvalid() {
		block, err := tinychain.TheTinyChain.BlockInterface(hBlock)
		if err != nil {
			t.Fail()
		}
		err = ac.AppendBlock(tinychain.TheTinyChain, block)
		if err != nil {
			t.Fail()
		}
		hBlock, err = tinychain.TheTinyChain.NextBlock(hBlock)
		if err != nil {
			t.Fail()
		}
	}

	bc := ac.GetAsChainReadInterface()
	TestCopyOfTinyChain_Helper(bc, t)

	crc := cac.GetAsConcreteReadableChain()

	hBlock0 := BlockHandle{HashHeight{height: 0, heightSpecified: true, hashSpecified: false}, crc}
	hBlock00 := BlockHandle{HashHeight{height: 0, heightSpecified: true, hashSpecified: false}, crc}
	hBlock1 := BlockHandle{HashHeight{height: 1, heightSpecified: true, hashSpecified: false}, crc}
	tinychain.TestHashEquality_helper(&hBlock0, &hBlock00, &hBlock1, t)

	ac.Close()
}

func TestCopyJsonChain(t *testing.T) {
	err := os.RemoveAll("Temp_Testing\\JsonChain")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	acc, err := NewConcreteAppendableChainCreator("Temp_Testing\\JsonChain")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	err = acc.Create([]string{"time", "size"}, []string{"version", "vsize"})
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	ac, cac, err := acc.Open(true)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	aOneBlockChain := jsonblock.CreateOneBlockChain(&jsonblock.HardCodedBlockFetcher{}, cac.GetAsDelegatedTransactionIndexer())

	hBlock := aOneBlockChain.GenesisBlock()
	block, err := aOneBlockChain.BlockInterface(hBlock)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	err = ac.AppendBlock(aOneBlockChain, block)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	hBlock, err = aOneBlockChain.NextBlock(hBlock)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	for !hBlock.IsInvalid() {
		block, err := aOneBlockChain.BlockInterface(hBlock)
		if err != nil {
			t.Error(err)
			t.FailNow()
		}
		err = ac.AppendBlock(aOneBlockChain, block)
		if err != nil {
			t.Error(err)
			t.FailNow()
		}
		hBlock, err = aOneBlockChain.NextBlock(hBlock)
		if err != nil {
			t.Error(err)
			t.FailNow()
		}
	}

	bc := ac.GetAsChainReadInterface()
	TestCopyOfJsonRealChain_Helper(bc, t)

	ac.Close()
}

func TestCopyRealChain(t *testing.T) {
	err := os.RemoveAll("Temp_Testing\\RealChain")
	if err != nil {
		t.Error(err)
	}

	acc, err := NewConcreteAppendableChainCreator("Temp_Testing\\RealChain")
	if err != nil {
		t.Error(err)
	}
	err = acc.Create([]string{"time"}, []string{"vsize"})
	if err != nil {
		t.Error(err)
	}
	ac, cac, err := acc.Open(true)
	if err != nil {
		t.Error(err)
	}

	var aReader corereader.CoreReader
	var aOneBlockChain = jsonblock.CreateOneBlockChain(&aReader, cac.GetAsDelegatedTransactionIndexer())

	hBlock := aOneBlockChain.GenesisBlock()
	block, err := aOneBlockChain.BlockInterface(hBlock)
	if err != nil {
		t.Error(err)
	}
	err = ac.AppendBlock(aOneBlockChain, block)
	if err != nil {
		t.Error(err)
	}

	hBlock, err = aOneBlockChain.NextBlock(hBlock)
	if err != nil {
		t.Error(err)
	}
	for i := 0; i < 1000; i++ {
		if i%100 == 0 {
			println("Block ", i)
		}

		block, err := aOneBlockChain.BlockInterface(hBlock)
		if err != nil {
			t.Error(err)
		}
		err = ac.AppendBlock(aOneBlockChain, block)
		if err != nil {
			t.Error(err)
		}
		hBlock, err = aOneBlockChain.NextBlock(hBlock)
		if err != nil {
			t.Error(err)
		}
	}

	bc := ac.GetAsChainReadInterface()
	TestBlockNeiExistance_Helper(bc, "size", false, t)
	TestBlockNeiExistance_Helper(bc, "time", true, t)
	TestTransactionNeiExistance_Helper(bc, "version", false, t)
	TestTransactionNeiExistance_Helper(bc, "vsize", true, t)

	ac.Close()
}
