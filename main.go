package main

import (
	"github.com/KitchenMishap/pudding-shed/chainstorage"
	"github.com/KitchenMishap/pudding-shed/jobs"
)

func main() {
	chainstorage.PrevFirstTxo = -1
	chainstorage.PrevTrans = -1

	//	err := jobs.SeveralYearsParallel(16, "delegated")
	err := jobs.SeveralYearsParallel(7, "delegated")
	if err != nil {
		println(err.Error())
		println("There was an error :-O")
	}
	println("End of main()")
}
