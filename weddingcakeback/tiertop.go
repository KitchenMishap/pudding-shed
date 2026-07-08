package weddingcakeback

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/KitchenMishap/pudding-shed/memfile"
	"github.com/KitchenMishap/pudding-shed/wordfile"
)

// TierTop is the type of the top tier. It is different from the tiers below it, so has its own class
// Note that the next tier below this is indexed as ZERO (TierTop has no tier index!)
type TierTop struct {
	folder                       string
	readonly                     bool
	underlyingFile               *os.File
	hashesFile                   *wordfile.HashFile
	theMap                       map[Sha256]GlobalPiType // These two are read into
	theSlice                     []Sha256                // memory in NewTierZero()
	firstGlobalPresentationIndex GlobalPiType
}

func NewTierTop(folderPath string, readOnly bool) (*TierTop, error) {
	result := TierTop{}
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

	for i := GlobalPiType(0); i < hashesCount; i++ {
		hash, err := result.hashesFile.ReadHashAt(int64(i))
		if err != nil {
			return nil, err
		}
		result.theMap[hash] = i
		result.theSlice[i] = hash
	}
	return &result, nil
}

func (tz *TierTop) IndexOfHash(hash *Sha256) (GlobalPiType, error) {
	index, ok := tz.theMap[*hash]
	if !ok {
		return -1, nil
	}
	return index, nil
}

func (tz *TierTop) GetHashAtIndex(index int64, hash *Sha256) error {
	if index < int64(len(tz.theSlice)) {
		*hash = tz.theSlice[index]
		return nil
	}
	return errors.New("index out of range")
}

func (tz *TierTop) CountHashes() (GlobalPiType, error) {
	return GlobalPiType(len(tz.theSlice)), nil
}

func (tz *TierTop) Close() error {
	return tz.hashesFile.Close()
}

func (tz *TierTop) AppendHash(hash *Sha256) (GlobalPiType, error) {
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

func (tz *TierTop) Sync() error {
	return tz.hashesFile.Sync()
}

// Functions to implement as interface BakingSourceTier
// Check that implements
var _ BakingSourceTier = (*TierTop)(nil)

func (tz *TierTop) GetNextTierPrefixBytesCount() byte {
	// Next tier is TierBelow[0]
	// A DonutForest in TierBelow[0] has no prefix bytes (so it has 256^0 = 1 tree in the forest)
	return 0
}
func (tz *TierTop) GetNextTierIndex() byte {
	// The index of the next tier after TierTop is (surprisingly) 0
	return 0
}
func (tz *TierTop) GetIndicesCount() uint64 {
	// This tier (zero) has no prefix bytes, so it has 256^0 = 1 indices
	return 1
}
func (tz *TierTop) GetHashesAtIndex(index uint64, config *CakeConfig) []SingleTreeHash {
	if index != 0 {
		panic("TierZero.GetHashesForIndex() should only be called with index=0")
	}
	result := make([]SingleTreeHash, len(tz.theSlice))
	for i := range tz.theSlice {
		result[i].PresentationIndex = tz.firstGlobalPresentationIndex + GlobalPiType(i)
		result[i].Hash = make([]byte, config.HashLength) // Todo Yuk!
		copy(result[i].Hash, tz.theSlice[i][:])
	}
	return result
}
