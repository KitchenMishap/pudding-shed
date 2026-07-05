package weddingcake

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/KitchenMishap/pudding-shed/memfile"
	"github.com/KitchenMishap/pudding-shed/wordfile"
)

type TierZero struct {
	filepath   string
	hashesFile *wordfile.HashFile
	theMap     map[Sha256]int64 // These two are read into
	theSlice   []Sha256         // memory in NewTierZero()
}

// Check that implements
var _ LegacyHashReadWriter = (*TierZero)(nil)

func NewTierZero(filepath string, readOnly bool) (*TierZero, error) {
	result := TierZero{}
	result.filepath = filepath
	var file *os.File
	var err error
	if readOnly {
		file, err = os.Open(filepath)
	} else {
		file, err = os.OpenFile(result.filepath, os.O_RDWR|os.O_APPEND, 0)
	}
	if err != nil {
		return nil, err
	}
	// Count the hashes
	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}
	hashesCount := stat.Size() / 32

	aoFile, err := memfile.NewAppendOptimizedFile(file)
	if err != nil {
		return nil, err
	}
	result.hashesFile = wordfile.NewHashFile(aoFile, hashesCount)

	result.theMap = make(map[Sha256]int64, hashesCount)
	result.theSlice = make([]Sha256, hashesCount)

	for i := int64(0); i < hashesCount; i++ {
		hash, err := result.hashesFile.ReadHashAt(i)
		if err != nil {
			return nil, err
		}
		result.theMap[hash] = i
		result.theSlice[i] = hash
	}
	return &result, nil
}

func (tz *TierZero) IndexOfHash(hash *Sha256) (int64, error) {
	index, ok := tz.theMap[*hash]
	if !ok {
		return -1, nil
	}
	return index, nil
}

func (tz *TierZero) GetHashAtIndex(index int64, hash *Sha256) error {
	if index < int64(len(tz.theSlice)) {
		*hash = tz.theSlice[index]
		return nil
	}
	return errors.New("index out of range")
}

func (tz *TierZero) CountHashes() (int64, error) {
	return int64(len(tz.theSlice)), nil
}

func (tz *TierZero) Close() error {
	return tz.hashesFile.Close()
}

func (tz *TierZero) AppendHash(hash *Sha256) (int64, error) {
	index, err := tz.CountHashes()
	if err != nil {
		return -1, err
	}
	tz.theMap[*hash] = index
	tz.theSlice = append(tz.theSlice, *hash)
	return tz.hashesFile.AppendHash(*hash)
}

func (tz *TierZero) Sync() error {
	return tz.hashesFile.Sync()
}

type TierZeroCreator struct {
	folder string
}

// Check that implements
var _ LegacyHashStoreCreator = (*TierZeroCreator)(nil)

func NewTierZeroCreator(folder string) *TierZeroCreator {
	result := TierZeroCreator{}
	result.folder = folder
	return &result
}

func (tzc *TierZeroCreator) HashStoreExists() bool {
	// Based on existence of <folder>/Tier0/Hashes.bin
	filePath := filepath.Join(tzc.folder, "Tier0", "Hashes.hsh")
	file, err := os.Open(filePath)
	if err == nil {
		defer func() { _ = file.Close() }()
		return true
	}
	return false
}

func (tzc *TierZeroCreator) CreateHashStore() error {
	folderPath := filepath.Join(tzc.folder, "Tier0")
	err := os.MkdirAll(folderPath, 0777)
	if err != nil {
		return err
	}
	filePath := filepath.Join(tzc.folder, "Tier0", "Hashes.hsh")
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()
	return nil
}

func (tzc *TierZeroCreator) OpenHashStore() (HashReadWriter[*Sha256], error) {
	filePath := filepath.Join(tzc.folder, "Tier0", "Hashes.hsh")
	tierZero, err := NewTierZero(filePath, false)
	if err != nil {
		return nil, err
	}
	return tierZero, nil
}

func (tzc *TierZeroCreator) OpenHashStoreReadOnly() (HashReader[*Sha256], error) {
	filePath := filepath.Join(tzc.folder, "Tier0", "Hashes.hsh")
	tierZero, err := NewTierZero(filePath, true)
	if err != nil {
		return nil, err
	}
	return tierZero, nil
}
