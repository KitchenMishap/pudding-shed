package indexedhashes

import (
	"fmt"
	"github.com/KitchenMishap/pudding-shed/memblocker"
	"github.com/KitchenMishap/pudding-shed/memfile"
	"github.com/KitchenMishap/pudding-shed/wordfile"
	"os"
	"runtime"
	"time"
)

func CreateHashIndexFiles() error {
	creator := newUniformHashStoreCreatorPrivate(1000000, "E:/Data/Hashes", "BlockHashes", 2)
	hs, err := creator.openHashStorePrivate()
	if err != nil {
		return err
	}
	memBlocker := memblocker.NewMemBlocker(8000000000)
	preloader := NewUniformHashPreLoader(hs, memBlocker)
	filename := "E:/Data/Hashes/BlockHashes/Hashes.hsh"
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	ao, err := memfile.NewAppendOptimizedFile(file)
	if err != nil {
		return err
	}
	hashFile := wordfile.NewHashFile(ao, 863000)

	// Values for monitoring and printing
	heapGb := 0.0
	hashesSent := 0
	fileWrites := 0
	allDone := false

	go func() {
		// Reporting
		for !allDone {
			t := time.Now()
			fmt.Println("Time is ", t.Format("2006-01-02 15:04:05"))
			fmt.Println("Heap GB:", heapGb, " hashes sent:", hashesSent, " file writes:", fileWrites)
			time.Sleep(3 * time.Second)
		}
	}()

	allHashesSent := false

	doneChan := make(chan bool)

	go func() {
		numDone := 0
		// Continue while not all hashes are sent, or some files were recently output
		for !allHashesSent || numDone > 0 {
			var err error
			numDone, err = preloader.extractSomeEntriesStoresToFiles(1)
			if err != nil {
				panic(err)
			}
			fileWrites += numDone
			runtime.GC()
			memBlocker.StartFreeingMem()
			memBlocker.MemoryWasFreed()
		}
		doneChan <- true
	}()

	for index := int64(0); index < 863000; index++ {
		hash := Sha256{}
		hash, err = hashFile.ReadHashAt(index)
		if err != nil {
			return err
		}
		entry := hashEntry{}
		entry.index = uint64(index)
		entry.hash = &hash
		preloader.delegateEntryToStores(entry, preloader.dividedAddressForHash(&hash))
		heapGb = float64(memBlocker.LastHeapSize()/100000) / 10000.0

		hashesSent++
	}
	allHashesSent = true
	fmt.Println("Done sending hashes.")
	fmt.Println("Waiting for writing to end...")
	_ = <-doneChan
	fmt.Println("...Done.")
	allDone = true
	time.Sleep(5 * time.Second) // Allow reporting to finish
	return nil
}

func CreateHashEmptyFiles() error {
	creator := newUniformHashStoreCreatorPrivate(1000000, "E:/Data/Hashes", "BlockHashes", 2)
	err := creator.CreateHashStore()
	if err != nil {
		return err
	}
	hs, err := creator.openHashStorePrivate()
	if err != nil {
		return err
	}
	preloader := UniformHashPreLoader{}
	preloader.uniform = hs
	err = preloader.createEmptyFiles()
	return err
}
