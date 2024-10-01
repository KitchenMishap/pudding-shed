package main

import (
	"github.com/KitchenMishap/pudding-shed/chainstorage"
	"github.com/KitchenMishap/pudding-shed/indexedhashes"
)

func main() {
	chainstorage.PrevFirstTxo = -1
	chainstorage.PrevTrans = -1

	// err := jobs.SeveralYearsParallel(16, "delegated")
	// err := jobs.SeveralYearsParallel(4, "delegated")

	//err := justhashes.JustHashes("F:/Data/JustHashes", 863000)

	err := indexedhashes.CreateHashEmptyFiles()
	//err := indexedhashes.CreateHashIndexFiles()

	if err != nil {
		println(err.Error())
		println("There was an error :-O")
	}
	println("End of main()")
}
