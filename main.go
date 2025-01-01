package main

import (
	"fmt"
	"github.com/KitchenMishap/pudding-shed/chainstorage"
	"github.com/KitchenMishap/pudding-shed/jobs"
	"time"
)

func main() {
	t := time.Now()
	fmt.Println(t.Format("Mon Jan 2 15:04:05"))

	chainstorage.PrevFirstTxo = -1
	chainstorage.PrevTrans = -1

	err := jobs.SeveralYearsPrimaries(16, "delegated")

	// err := jobs.SeveralYearsParallel(16, "delegated")
	// err := jobs.SeveralYearsParallel(4, "delegated")

	//err := justhashes.JustHashes("F:/Data/JustHashes", 863000)

	//err := indexedhashes.CreateHashEmptyFiles()
	//err := indexedhashes.CreateHashIndexFiles()

	//const hashCountEstimate = 1000000000
	//const digitsPerFolder = 2
	//const gigabytesMem = 1

	//uhc, _ := indexedhashes.NewUniformHashStoreCreatorAndPreloader(
	//	"F:/Data/UniformHashes", "Transactions",
	//	hashCountEstimate, digitsPerFolder, gigabytesMem)
	//err := uhc.CreateHashStore()

	/*
		_, mp, err := indexedhashes.NewUniformHashStoreCreatorAndPreloaderFromFile(
			"F:/Data/UniformHashes", "Transactions", gigabytesMem)
		if err != nil {
			println(err.Error())
			println("There was an error :-O")
		} else {
			err = mp.IndexTheHashes()
		}
	*/

	if err != nil {
		println(err.Error())
		println("There was an error :-O")
	}
	fmt.Println("End of main()")
	t = time.Now()
	fmt.Println(t.Format("Mon Jan 2 15:04:05"))
}
