package weddingcakeback

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/KitchenMishap/pudding-shed/memfile"
	"github.com/KitchenMishap/pudding-shed/wordfile"
)

type TierZero struct {
	folder                       string
	readonly                     bool
	underlyingFile               *os.File
	hashesFile                   *wordfile.HashFile
	theMap                       map[Sha256]GlobalPiType // These two are read into
	theSlice                     []Sha256                // memory in NewTierZero()
	firstGlobalPresentationIndex GlobalPiType
}

func NewTierZero(folderPath string, readOnly bool) (*TierZero, error) {
	result := TierZero{}
	result.firstGlobalPresentationIndex = 0
	result.folder = folderPath
	filePath := filepath.Join(folderPath, "Hashes.hsh")
	result.readonly = readOnly
	var err error
	if readOnly {
		result.underlyingFile, err = os.Open(filePath)
	} else {
		result.underlyingFile, err = os.OpenFile(filePath, os.O_RDWR|os.O_APPEND, 0)
	}
	if err != nil {
		return nil, err
	}
	// Count the hashes
	stat, err := result.underlyingFile.Stat()
	if err != nil {
		return nil, err
	}
	hashesCount := stat.Size() / 32

	aoFile, err := memfile.NewAppendOptimizedFile(result.underlyingFile)
	if err != nil {
		return nil, err
	}
	result.hashesFile = wordfile.NewHashFile(aoFile, hashesCount)

	result.theMap = make(map[Sha256]GlobalPiType, 65535)
	result.theSlice = make([]Sha256, hashesCount, 65535)

	for i := GlobalPiType(0); i < GlobalPiType(hashesCount); i++ {
		hash, err := result.hashesFile.ReadHashAt(int64(i))
		if err != nil {
			return nil, err
		}
		result.theMap[hash] = i
		result.theSlice[i] = hash
	}
	return &result, nil
}

func (tz *TierZero) IndexOfHash(hash *Sha256) (GlobalPiType, error) {
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

func (tz *TierZero) CountHashes() (GlobalPiType, error) {
	return GlobalPiType(len(tz.theSlice)), nil
}

func (tz *TierZero) Close() error {
	return tz.hashesFile.Close()
}

func (tz *TierZero) AppendHash(hash *Sha256) (GlobalPiType, error) {
	index, err := tz.CountHashes()
	if err != nil {
		return -1, err
	}
	tz.theMap[*hash] = index
	tz.theSlice = append(tz.theSlice, *hash)
	result, err := tz.hashesFile.AppendHash(*hash)
	if err != nil {
		return -1, err
	}
	return result, nil
}

func (tz *TierZero) Sync() error {
	return tz.hashesFile.Sync()
}
