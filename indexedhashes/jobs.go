package indexedhashes

import (
	"fmt"
	"github.com/KitchenMishap/pudding-shed/memblocker"
	"github.com/KitchenMishap/pudding-shed/memfile"
	"github.com/KitchenMishap/pudding-shed/wordfile"
	"os"
)

func CreateHashIndexFiles() error {
	creator := newUniformHashStoreCreatorPrivate(1000000, "E:/Data/Hashes", "BlockHashes", 2)
	hs, err := creator.openHashStorePrivate()
	if err != nil {
		return err
	}
	memBlocker := memblocker.NewMemBlocker(16000000000)
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

		if index%1000 == 0 {
			fmt.Printf("index: %d\n", index)
		}
	}
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
