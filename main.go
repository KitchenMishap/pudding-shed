package main

import (
	"fmt"
	"github.com/KitchenMishap/pudding-shed/chainstorage"
	"github.com/KitchenMishap/pudding-shed/indexedhashes"
	"time"
)

func main() {
	t := time.Now()
	fmt.Println(t.Format("Mon Jan 2 15:04:05"))

	chainstorage.PrevFirstTxo = -1
	chainstorage.PrevTrans = -1

	// err := jobs.SeveralYearsParallel(16, "delegated")
	// err := jobs.SeveralYearsParallel(4, "delegated")

	//err := justhashes.JustHashes("F:/Data/JustHashes", 863000)

	//err := indexedhashes.CreateHashEmptyFiles()
	//err := indexedhashes.CreateHashIndexFiles()

	uhc := indexedhashes.NewUniformHashStoreCreatorPrivate(1000000000,
		"F:/Data/UniformHashes", "Transactions", 2)
	gigabyte := int64(1024 * 1024 * 1024)
	mp := indexedhashes.NewMultipassPreloader(uhc, 1*gigabyte, 0, 2)
	err := mp.IndexTheHashes()

	if err != nil {
		println(err.Error())
		println("There was an error :-O")
	}
	println("End of main()")
	t = time.Now()
	fmt.Println(t.Format("Mon Jan 2 15:04:05"))
}
