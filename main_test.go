package main

import (
	"github.com/KitchenMishap/pudding-shed/chainstorage"
	"github.com/KitchenMishap/pudding-shed/jobs"
	"testing"
)

func TestMain(t *testing.T) {
	//	err := jobs.SeveralYearsPrimaries(16, "delegated")
	err := jobs.SeveralYearsPrimaries(4, "delegated")
	if err != nil {
		println(err.Error())
		println("There was an error :-O")
	}
	println("End of main()")
}

func TestTransactionHashes(t *testing.T) {
	err := jobs.SeveralYearsPrimaries(3, "delegated")
	if err != nil {
		println(err.Error())
		println("There was an error :-O")
		t.Error()
	}

	println("Main part of test...")
	println("Note that this test highlights the genuine issue that")
	println("the coinbase transactions of blocks 91812 and 91842")
	println("have identical hashes! (identical transactions, BIP37)")
	folder := "F:/Data/CurrentJob"
	creator, _ := chainstorage.NewConcreteAppendableChainCreator(folder, []string{}, []string{}, false)
	cac, err := creator.Open()
	if err != nil {
		println(err.Error())
		t.Error()
	}
	err = cac.SelfTestTransHashes()
	if err != nil {
		println(err.Error())
		t.Error()
	}
}
