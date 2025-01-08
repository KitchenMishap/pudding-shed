package main

import (
	"fmt"
	"github.com/KitchenMishap/pudding-shed/chainstorage"
	"github.com/KitchenMishap/pudding-shed/indexedhashes3"
	"github.com/KitchenMishap/pudding-shed/jobs"
	"time"
)

func main() {

	//																			Count at 15 years:
	roomFor16milBlocks := int64(3) // 256^3 = 16,777,216 blocks					There were 824,204 blocks
	roomFor4bilTrans := int64(4)   // 256^4 = 4,294,967,296 transactions		There were 947,337,057 transactions
	//roomFor1trilTxxs := int64(5)   // 256^5 = 1,099,511,627,776 txos or txis	There were 2,652,374,369 txos (including spent)
	roomFor1trilAddrs := int64(5) //	,,			,,			 addresses		There must be fewer addresses than txos
	//roomForAllSatoshis := int64(7) // 256^7 = 72,057,594,037,927,936 sats		There will be 2,100,000,000,000,000 sats

	percentOverflows := 0.01
	indexedhashes3.OptimizeAndInitializeParams("Blocks", 1000000, roomFor16milBlocks, 37164, percentOverflows)
	indexedhashes3.OptimizeAndInitializeParams("Transactions", 1000000000, roomFor4bilTrans, 36382, percentOverflows)
	indexedhashes3.OptimizeAndInitializeParams("Addresses", 3000000000, roomFor1trilAddrs, 55057, percentOverflows)
	fmt.Println("Blocks")
	indexedhashes3.IterateSmallestSize(1000000, roomFor16milBlocks, percentOverflows, 37000, 38000, 1)
	fmt.Println("Trans")
	indexedhashes3.IterateSmallestSize(1000000000, roomFor4bilTrans, percentOverflows, 36000, 37000, 1)
	fmt.Println("Addrs")
	indexedhashes3.IterateSmallestSize(3000000000, roomFor1trilAddrs, percentOverflows, 55000, 56000, 1)

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
