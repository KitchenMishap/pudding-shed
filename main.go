package main

import (
	"flag"
	"fmt"
	"github.com/KitchenMishap/pudding-shed/chainstorage"
	"github.com/KitchenMishap/pudding-shed/jobs"
	"strconv"
	"time"
)

// Comment
func main() {
	//																			Count at 15 years:
	//roomFor16milBlocks := int64(3) // 256^3 = 16,777,216 blocks					There were 824,204 blocks
	//roomFor4bilTrans := int64(4)   // 256^4 = 4,294,967,296 transactions		There were 947,337,057 transactions
	//roomFor1trilTxxs := int64(5)   // 256^5 = 1,099,511,627,776 txos or txis	There were 2,652,374,369 txos (including spent)
	//roomFor1trilAddrs := int64(5) //	,,			,,			 addresses		There must be fewer addresses than txos
	//roomForAllSatoshis := int64(7) // 256^7 = 72,057,594,037,927,936 sats		There will be 2,100,000,000,000,000 sats

	var sDirFlag = flag.String("Dir", "", "Directory to create and store data")
	var nGbFlag = flag.Int("Gb", 1, "Number of gigabytes memory available to Phase 2")
	var nThreadsFlag = flag.Int("Threads", 4, "Number of threads to run")
	flag.Parse()

	if *sDirFlag == "" {
		fmt.Println("Please provide a directory to store data with the -Dir= commandline flag")
		return
	}

	fmt.Println("Starting PuddingShed:")
	fmt.Println("Dir=" + *sDirFlag)
	fmt.Println("Gb=" + strconv.Itoa(*nGbFlag))
	fmt.Println("Threads=" + strconv.Itoa(*nThreadsFlag))

	t := time.Now()
	fmt.Println(t.Format("Mon Jan 2 15:04:05"))

	chainstorage.PrevFirstTxo = -1
	chainstorage.PrevTrans = -1

	err := jobs.SeveralYearsPrimaries(2, "delegated",
		true, true, true, *sDirFlag, *nGbFlag, *nThreadsFlag)

	if err != nil {
		println(err.Error())
		println("There was an error :-O")
	}
	fmt.Println("End of main()")
	t = time.Now()
	fmt.Println(t.Format("Mon Jan 2 15:04:05"))
}
