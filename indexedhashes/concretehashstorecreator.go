package indexedhashes

import (
	"errors"
	"github.com/KitchenMishap/pudding-shed/memfile"
	"os"
	"path/filepath"
)

type ConcreteHashStoreCreator struct {
	name                        string
	folder                      string
	partialHashBitCount         int64
	entryByteCount              int64
	collisionsPerChunk          int64
	useMemFileForLookup         bool
	useAppendOptimizedForHashes bool
}

func NewConcreteHashStoreCreator(name string, folder string,
	partialHashBitCount int64, entryByteCount int64, collisionsPerChunk int64,
	useMemFileForLookup bool, useAppendOptimizedForHashes bool) (*ConcreteHashStoreCreator, error) {
	if entryByteCount > ZEROBUF {
		err := errors.New("hard coded ZEROBUF not big enough")
		return nil, err
	}

	result := ConcreteHashStoreCreator{}
	result.name = name
	result.folder = folder
	result.partialHashBitCount = partialHashBitCount
	result.entryByteCount = entryByteCount
	result.collisionsPerChunk = collisionsPerChunk
	result.useMemFileForLookup = useMemFileForLookup
	result.useAppendOptimizedForHashes = useAppendOptimizedForHashes
	return &result, nil
}

func (hsc *ConcreteHashStoreCreator) HashStoreExists() bool {
	// Check whether hashes file exists
	// We assume this would mean the lookup and collisions files also exist
	fullName := filepath.Join(hsc.folder, hsc.name+".hsh")
	fileHsh, err := os.Open(fullName)
	defer fileHsh.Close()
	if err != nil {
		// Doesn't exist.
		return false
	}
	return true
}

func (hsc *ConcreteHashStoreCreator) CreateHashStore() error {
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

	// Lookup file (.lkp) starts off as a full file of zeroes
	fullName = filepath.Join(hsc.folder, hsc.name+".lkp")
	lookupFile, err := os.Create(fullName)
	if err != nil {
		return err
	}
	defer lookupFile.Close()
	err = lookupFile.Truncate(hsc.entryByteCount << hsc.partialHashBitCount)
	if err != nil {
		return err
	}

	// Collisions file (.cls) starts off as an empty file
	fullName = filepath.Join(hsc.folder, hsc.name+".cls")
	collisionsFile, err := os.Create(fullName)
	if err != nil {
		return err
	}
	defer collisionsFile.Close()

	return nil
}

func (hsc *ConcreteHashStoreCreator) OpenHashStore() (HashReadWriter, error) {
	// Open the hashes file
	fullName := filepath.Join(hsc.folder, hsc.name+".hsh")
	hashesFileUnderlying, err := os.OpenFile(fullName, os.O_RDWR, 0)
	if err != nil {
		return nil, err
	}
	var hashesFile memfile.AppendableLookupFile
	if hsc.useAppendOptimizedForHashes {
		hashesFile = memfile.NewAppendOptimizedFile(hashesFileUnderlying)
	} else {
		hashesFile = hashesFileUnderlying
	}

	// Open the lookup file
	var lookupFile memfile.SparseLookupFile
	fullName = filepath.Join(hsc.folder, hsc.name+".lkp")
	if hsc.useMemFileForLookup {
		lookupFile, err = memfile.NewFixedSizeMemFile(fullName, 0)
	} else {
		lookupFile, err = os.OpenFile(fullName, os.O_RDWR, 0)
	}
	if err != nil {
		hashesFile.Close()
		return nil, err
	}

	// Open the collisions file
	fullName = filepath.Join(hsc.folder, hsc.name+".cls")
	collisionsFile, err := os.OpenFile(fullName, os.O_RDWR, 0)
	if err != nil {
		hashesFile.Close()
		lookupFile.Close()
		return nil, err
	}

	// Populate and return the HashReadWriter object
	// Create the BasicHashStore sub-object
	bhs := NewBasicHashStore(hashesFile)
	result := NewHashStore(hsc.partialHashBitCount, hsc.entryByteCount, hsc.collisionsPerChunk,
		bhs, lookupFile, collisionsFile)
	return result, nil
}
func (hsc *ConcreteHashStoreCreator) OpenHashStoreReadOnly() (HashReader, error) {
	// Open the hashes file
	fullName := filepath.Join(hsc.folder, hsc.name+".hsh")
	hashesFile, err := os.OpenFile(fullName, os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	// Open the lookup file
	fullName = filepath.Join(hsc.folder, hsc.name+".lkp")
	lookupFile, err := os.OpenFile(fullName, os.O_RDONLY, 0)
	if err != nil {
		hashesFile.Close()
		return nil, err
	}
	// Open the collisions file
	fullName = filepath.Join(hsc.folder, hsc.name+".cls")
	collisionsFile, err := os.OpenFile(fullName, os.O_RDONLY, 0)
	if err != nil {
		hashesFile.Close()
		lookupFile.Close()
		return nil, err
	}

	// Populate and return the HashReader object
	// Create the BasicHashStore sub-object
	bhs := NewBasicHashStore(hashesFile)
	result := NewHashStore(hsc.partialHashBitCount, hsc.entryByteCount, hsc.collisionsPerChunk,
		bhs, lookupFile, collisionsFile)
	return result, nil
}
