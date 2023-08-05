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
	block := tinychain.TheTinyChain.BlockInterface(hBlock)
	err = ac.AppendBlock(tinychain.TheHandles, tinychain.TheTinyChain, block)
	if err != nil {
		t.Fail()
	}

	hBlock = tinychain.TheTinyChain.NextBlock(hBlock)
	for !hBlock.IsInvalid() {
		block := tinychain.TheTinyChain.BlockInterface(hBlock)
		err = ac.AppendBlock(tinychain.TheHandles, tinychain.TheTinyChain, block)
		if err != nil {
			t.Fail()
		}
		hBlock = tinychain.TheTinyChain.NextBlock(hBlock)
	}

	err = ac.Close()
	if err != nil {
		t.Fail()
	}
}
