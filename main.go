package main

import (
	"fmt"
	"github.com/KitchenMishap/pudding-shed/chainstorage"
	"github.com/KitchenMishap/pudding-shed/jobs"
	"github.com/KitchenMishap/pudding-shed/wordfile"
	"time"
)

func main() {
	//																			Count at 15 years:
	//roomFor16milBlocks := int64(3) // 256^3 = 16,777,216 blocks					There were 824,204 blocks
	//roomFor4bilTrans := int64(4)   // 256^4 = 4,294,967,296 transactions		There were 947,337,057 transactions
	//roomFor1trilTxxs := int64(5)   // 256^5 = 1,099,511,627,776 txos or txis	There were 2,652,374,369 txos (including spent)
	//roomFor1trilAddrs := int64(5) //	,,			,,			 addresses		There must be fewer addresses than txos
	//roomForAllSatoshis := int64(7) // 256^7 = 72,057,594,037,927,936 sats		There will be 2,100,000,000,000,000 sats

	t := time.Now()
	fmt.Println(t.Format("Mon Jan 2 15:04:05"))

	chainstorage.PrevFirstTxo = -1
	chainstorage.PrevTrans = -1

	err := jobs.SeveralYearsPrimaries(17, "delegated", true, true, true)

	if err != nil {
		println(err.Error())
		println("There was an error :-O")
	}
	fmt.Println("End of main()")
	t = time.Now()
	fmt.Println(t.Format("Mon Jan 2 15:04:05"))

	wfc := wordfile.NewConcreteWordFileCreator("firsttxo", "F:\\Data\\TwoYear\\Addresses", 5, false)
	wf, err := wfc.OpenWordFileReadOnly()
	if err != nil {
		println(err.Error())
		return
	}
	val, err := wf.ReadWordAt(264249)
	if err != nil {
		println(err.Error())
		return
	}
	if val == 0 {
		panic("First txo of address 264249 should be non zero")
	}
	wf.Close()

}
