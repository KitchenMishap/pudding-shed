package main

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/KitchenMishap/pudding-shed/chainstorage"
	"github.com/KitchenMishap/pudding-shed/jobs"
)

// We have a TestMain which mimics Main()... because you can only CPU profile a test!
func Test_Main(t *testing.T) {
	tt := time.Now()
	fmt.Println(tt.Format("Mon Jan 2 15:04:05"))

	path := "E:\\Data\\TestMain"
	gb := 16
	threads := 10
	years := 2
	fmt.Println("Dir=" + path)
	fmt.Println("Yrs=" + strconv.Itoa(years))
	fmt.Println("Gb=" + strconv.Itoa(gb))
	fmt.Println("Threads=" + strconv.Itoa(threads))

	chainstorage.PrevFirstTxo = -1
	chainstorage.PrevTrans = -1

	err := jobs.SeveralYearsPrimaries(years, "delegated", true, true, true,
		path, gb, threads)

	if err != nil {
		println(err.Error())
		println("There was an error :-O")
		t.Error(err)
	}
	fmt.Println("End of TestMain()")
	tt = time.Now()
	fmt.Println(tt.Format("Mon Jan 2 15:04:05"))
}
