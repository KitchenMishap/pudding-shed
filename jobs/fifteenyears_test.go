package jobs

import (
	"fmt"
	"github.com/KitchenMishap/pudding-shed/indexedhashes"
	"github.com/KitchenMishap/pudding-shed/tinychain"
	"testing"
)

func TestFiveYearsDelegated(t *testing.T) {
	SeveralYearsParallel(4, "delegated")
}

func TestOneYearUndelegated(t *testing.T) {
	SeveralYearsPrimaries(1, "separate files")
}

func TestMemoryLeak(t *testing.T) {
	cr := indexedhashes.NewUniformHashStoreCreator(1000000000, "F:\\Data\\TestMemLeak", "TestName", 2)
	err := cr.CreateHashStore()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	hs, err := cr.OpenHashStore()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	for i := uint64(0); i < 10000000; i++ {
		hash := tinychain.HashOfInt(i)
		_, err := hs.AppendHash(&hash)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
	}
}
