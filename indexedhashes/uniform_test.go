package indexedhashes

import (
	"fmt"
	"os"
	"testing"
)

func TestUniform(t *testing.T) {

	const gigabytesMem = 1
	uhc, _, err := NewUniformHashStoreCreatorAndPreloaderFromFile(
		"F:/Data/UniformHashes", "Transactions", gigabytesMem)
	if err != nil {
		println(err.Error())
		println("There was an error :-O")
	} else {
		hs, err := uhc.openHashStorePrivate()
		if err != nil {
			println(err.Error())
			println("There was an error :-O")
		} else {
			success, err := hs.Test()
			if err != nil {
				println(err.Error())
				println("There was an error :-O")
			} else if success {
				println("success")
			} else {
				println("failure")
			}
		}
	}
}

func TestUniformSmallRoundTrip(t *testing.T) {
	const gigabytesMem = 1
	const dir = "Temp_Testing/SmallRoundTrip"
	const hashes = 10000

	fmt.Println("Deleting folder...")
	err := os.RemoveAll(dir)
	if err != nil {
		t.Error(err)
	}
	err = os.Mkdir(dir, 0777)
	if err != nil {
		t.Error(err)
	}

	fmt.Println("Creating hashstore...")
	uhc, _ := NewUniformHashStoreCreatorAndPreloader(dir, "CountingHashes", hashes/2, 2, gigabytesMem)
	err = uhc.CreateHashStore()
	if err != nil {
		t.Error(err)
	}

	fmt.Println("Writing hashes file...")
	file, err := os.Create(dir + "/CountingHashes/Hashes.hsh")
	if err != nil {
		t.Error(err)
	}
	for i := uint64(0); i < hashes; i++ {
		hash := HashOfInt(i)
		n, err := file.Write(hash[:])
		if err != nil {
			t.Error(err)
		}
		if n != 32 {
			t.Errorf("Wrote %d bytes, expected %d", n, 32)
		}
	}
	err = file.Close()
	if err != nil {
		t.Error(err)
	}

	fmt.Println("Indexing the hashes...")
	uhc2, mp, err := NewUniformHashStoreCreatorAndPreloaderFromFile(
		dir, "CountingHashes", gigabytesMem)
	if err != nil {
		t.Error(err)
	}
	err = mp.IndexTheHashes()
	if err != nil {
		t.Error(err)
	}

	fmt.Println("Testing the hashes...")
	hs, err := uhc2.openHashStorePrivate()
	if err != nil {
		t.Error(err)
	}
	success, err := hs.Test()
	if err != nil {
		t.Error(err)
	}
	if success == false {
		t.Error("The test failed!")
	}
	fmt.Println("End of test")
}
