package jobs

import (
	"fmt"
	"github.com/KitchenMishap/pudding-shed/indexedhashes"
	"github.com/KitchenMishap/pudding-shed/tinychain"
	"os"
	"strconv"
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
	for i := uint64(0); i < 100000; i++ {
		hash := tinychain.HashOfInt(i)
		_, err := hs.AppendHash(&hash)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
	}
}

func TestMemLeak2(t *testing.T) {
	fmt.Println("Removing files from any previous run")
	os.RemoveAll("F:\\Data\\TestMemLeak")

	fmt.Println("Creating and appending files")
	for a := 0; a < 100; a++ {
		for b := 0; b < 100; b++ {
			dirPath := "F:\\Data\\TestMemLeak\\"
			dirPath += strconv.Itoa(a) + "\\"
			dirPath += strconv.Itoa(b)
			err := os.MkdirAll(dirPath, 0777)
			if err != nil {
				panic("Mkdir panic")
			}
			for c := 0; c < 50; c++ {
				path := dirPath + "\\" + strconv.Itoa(c) + ".txt"
				file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0755)
				if err != nil {
					panic("OpenFile panic:" + err.Error())
				}
				toAppend := [10]byte{} // ten zero bytes
				_, err = file.Write(toAppend[:])
				if err != nil {
					panic("Write panic")
				}
				err = file.Sync()
				if err != nil {
					panic("Sync panic")
				}
				err = file.Close()
				if err != nil {
					panic("Close panic")
				}
				file = nil
			}
		}
	}
}
