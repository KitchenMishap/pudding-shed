package chainstorage

import (
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
	err = acc.Create()
	if err != nil {
		t.Fail()
	}
	ac, cac, err := acc.Open()
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

	err = ac.Close()
	if err != nil {
		t.Fail()
	}
}
