package indexedhashes

import (
	"github.com/KitchenMishap/pudding-shed/memfile"
	"github.com/KitchenMishap/pudding-shed/wordfile"
	"os"
	"path/filepath"
)

type ConcreteHashStoreCreatorMemory struct {
	name   string
	folder string
}

func NewConcreteHashStoreCreatorMemory(name string, folder string) (*ConcreteHashStoreCreatorMemory, error) {
	result := ConcreteHashStoreCreatorMemory{}
	result.name = name
	result.folder = folder
	return &result, nil
}

func (hsc *ConcreteHashStoreCreatorMemory) HashStoreExists() bool {
	// Check whether hashes file exists
	fullName := filepath.Join(hsc.folder, hsc.name+".hsh")
	fileHsh, err := os.Open(fullName)
	defer fileHsh.Close()
	if err != nil {
		// Doesn't exist.
		return false
	}
	return true
}

func (hsc *ConcreteHashStoreCreatorMemory) CreateHashStore() error {
	// First create folder if necessary
	err := os.MkdirAll(hsc.folder, os.ModePerm)
	if err != nil {
		return err
	}

	// Hashes file (.hsh) starts off as an empty file
	fullName := filepath.Join(hsc.folder, hsc.name+".hsh")
	hashesFile, err := os.Create(fullName)
	if err != nil {
		return err
	}
	defer hashesFile.Close()

	return nil
}

func (hsc *ConcreteHashStoreCreatorMemory) OpenHashStore() (HashReadWriter, error) {
	// Open the hashes file
	fullName := filepath.Join(hsc.folder, hsc.name+".hsh")
	hashesFileUnderlying, err := os.OpenFile(fullName, os.O_RDWR, 0)
	if err != nil {
		return nil, err
	}
	// Count the hashes
	stat, err := hashesFileUnderlying.Stat()
	if err != nil {
		return nil, err
	}
	hashesCount := stat.Size() / 32
	var hashesFile memfile.AppendableLookupFile
	hashesFile = hashesFileUnderlying

	// Populate and return the HashReadWriter object
	// Create the BasicHashStore sub-object
	hashFile := wordfile.NewHashFile(hashesFile, hashesCount)
	result := NewMemoryIndexedHashes(hashFile)
	return result, nil
}
func (hsc *ConcreteHashStoreCreatorMemory) OpenHashStoreReadOnly() (HashReader, error) {
	// Open the hashes file
	fullName := filepath.Join(hsc.folder, hsc.name+".hsh")
	hashesFile, err := os.OpenFile(fullName, os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	// Count the hashes
	stat, err := hashesFile.Stat()
	if err != nil {
		return nil, err
	}
	hashesCount := stat.Size() / 32

	// Populate and return the HashReader object
	// Create the BasicHashStore sub-object
	hashFile := wordfile.NewHashFile(hashesFile, hashesCount)
	result := NewMemoryIndexedHashes(hashFile)
	return result, nil
}
