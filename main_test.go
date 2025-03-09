package main

import (
	"fmt"
	"github.com/KitchenMishap/pudding-shed/chainstorage"
	"github.com/KitchenMishap/pudding-shed/jobs"
	"testing"
	"time"
)

// We have a TestMain which mimics Main()... because you can only CPU profile a test!
func TestMain(m *testing.M) {
	t := time.Now()
	fmt.Println(t.Format("Mon Jan 2 15:04:05"))

	chainstorage.PrevFirstTxo = -1
	chainstorage.PrevTrans = -1

	err := jobs.SeveralYearsPrimaries(16, "delegated", false, false, true)

	if err != nil {
		println(err.Error())
		println("There was an error :-O")
	}
	fmt.Println("End of TestMain()")
	t = time.Now()
	fmt.Println(t.Format("Mon Jan 2 15:04:05"))
}
