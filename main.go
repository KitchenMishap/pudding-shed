package main

import (
	"github.com/KitchenMishap/pudding-shed/chainstorage"
	"github.com/KitchenMishap/pudding-shed/jobs"
)

func main() {
	chainstorage.PrevFirstTxo = -1
	chainstorage.PrevTrans = -1

	//	err := jobs.SeveralYearsPrimaries(16, "delegated")
	err := jobs.SeveralYearsPrimaries(3, "delegated")
	if err != nil {
		println(err.Error())
		println("There was an error :-O")
	}
	println("End of main()")
}
