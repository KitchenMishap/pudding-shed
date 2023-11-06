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
	ac, err := acc.Open()
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

	crc := ac.GetAsChainReadInterface()
	TestCopyOfTinyChain_Helper(crc, t)

	err = ac.Close()
	if err != nil {
		t.Fail()
	}
}
