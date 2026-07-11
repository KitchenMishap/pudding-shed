package weddingcakeback

import (
	"encoding/binary"
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/KitchenMishap/pudding-shed/memfile"
	"github.com/KitchenMishap/pudding-shed/wordfile"
)

// There are some fiddly conversions in here between GlobalPiType and HashIndexIdType.
// HashIndexIdType's are used internally to a tier such as TierTop, and start at 1.
// GlobalPiType's have an offset (held in TierTop), and the first ever hash corresponds to a GlobalPiType of 0.

// TierTop is the type of the top tier. It is different from the tiers below it, so has its own class
// Note that the next tier below this is indexed as ZERO (TierTop has no tier index!)
type TierTop struct {
	folder         string
	readonly       bool
	underlyingFile *os.File
	hashesFile     *wordfile.HashFile
	// As with other tiers, theMap deals with HashIndexIdType which is not the same as GlobalPiType
	theMap map[Sha256]HashIndexIdType
	// Because HashIndexType's reserve zero to mean "no match", subtract 1 from a HashIndexId to index theSlice
	theSlice                     []Sha256
	firstGlobalPresentationIndex GlobalPiType
}

func (tt *TierTop) HashIndexIdFromGlobalPresentationIndex(global GlobalPiType) HashIndexIdType {
	// The value reserved for "no match"
	if global == GlobalPiNoMatch {
		return HashIndexIdNoMatch
	}
	// ch.FirstGlobalPresentationIndexOfChunk maps to hashIndexId 1
	hashIndexId := global - tt.firstGlobalPresentationIndex + 1
	if hashIndexId < 1 {
		panic("Global presentation index lower than first presentation index")
	}
	return HashIndexIdType(hashIndexId)
}

func (tt *TierTop) GlobalPresentationIndexFromHashIndexId(hashIndexId HashIndexIdType) GlobalPiType {
	// The value reserved for "no match"
	if hashIndexId == HashIndexIdNoMatch {
		return GlobalPiNoMatch
	}
	// HashIndexIdType 1 maps to ch.FirstGlobalPresentationIndexOfChunk
	return GlobalPiType(hashIndexId) - 1 + tt.firstGlobalPresentationIndex
}

func NewTierTop(folderPath string, readOnly bool) (*TierTop, error) {
	result := TierTop{}
	result.firstGlobalPresentationIndex = 0
	result.folder = folderPath
	filePath := filepath.Join(folderPath, "TierTop", "Hashes.hsh")
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

	firstPiFile, err := os.Open(filepath.Join(folderPath, "TierTop", "FirstPresentationIndex.bin"))
	if err != nil {
		return nil, err
	}
	defer func() { _ = firstPiFile.Close() }()
	var firstPi GlobalPiType
	err = binary.Read(firstPiFile, binary.LittleEndian, &firstPi)
	if err != nil {
		return nil, err
	}
	result.firstGlobalPresentationIndex = firstPi

	aoFile, err := memfile.NewAppendOptimizedFile(result.underlyingFile)
	if err != nil {
		return nil, err
	}
	result.hashesFile = wordfile.NewHashFile(aoFile, hashesCount)

	result.theMap = make(map[Sha256]HashIndexIdType, 65535)
	result.theSlice = make([]Sha256, hashesCount, 65535)

	// HashIndexId's start at 1
	for i := HashIndexIdType(1); i < HashIndexIdType(hashesCount+1); i++ {
		hash, err := result.hashesFile.ReadHashAt(int64(i - 1))
		if err != nil {
			return nil, err
		}
		result.theMap[hash] = i
		result.theSlice[i-1] = hash
	}
	return &result, nil
}

func (tt *TierTop) IndexOfHash(hash *Sha256) (GlobalPiType, error) {
	hashIndexId, ok := tt.theMap[*hash]
	if !ok {
		return GlobalPiNoMatch, nil
	}
	return tt.GlobalPresentationIndexFromHashIndexId(hashIndexId), nil
}

func (tt *TierTop) GetHashAtIndex(index int64, hash *Sha256) error {
	if index < int64(len(tt.theSlice)) {
		*hash = tt.theSlice[tt.HashIndexIdFromGlobalPresentationIndex(index)-1]
		return nil
	}
	return errors.New("index out of range")
}

func (tt *TierTop) CountHashes() (GlobalPiType, error) {
	return GlobalPiType(len(tt.theSlice)), nil
}

func (tt *TierTop) Close() error {
	return tt.hashesFile.Close()
}

func (tt *TierTop) AppendHash(hash *Sha256) (GlobalPiType, error) {
	index, err := tt.CountHashes()
	if err != nil {
		return -1, err
	}
	tt.theMap[*hash] = HashIndexIdType(index + 1)
	tt.theSlice = append(tt.theSlice, *hash)
	_, err = tt.hashesFile.AppendHash(*hash)
	if err != nil {
		return -1, err
	}
	return tt.GlobalPresentationIndexFromHashIndexId(HashIndexIdType(index + 1)), nil
}

func (tt *TierTop) Sync() error {
	return tt.hashesFile.Sync()
}

// Functions to implement as interface BakingSourceTier
// Check that implements
var _ BakingSourceTier = (*TierTop)(nil)

func (tt *TierTop) GetNextTierPrefixBytesCount() byte {
	// Next tier is TierBelow[0]
	// A DonutForest in TierBelow[0] has no prefix bytes (so it has 256^0 = 1 tree in the forest)
	return 0
}
func (tt *TierTop) GetNextTierIndex() byte {
	// The index of the next tier after TierTop is (surprisingly) 0
	return 0
}
func (tt *TierTop) GetIndicesCount() uint64 {
	// This tier (zero) has no prefix bytes, so it has 256^0 = 1 indices
	return 1
}
func (tt *TierTop) GetHashesAtIndex(index uint64, config *CakeConfig) []SingleTreeHash {
	if index != 0 {
		panic("TierZero.GetHashesForIndex() should only be called with index=0")
	}
	result := make([]SingleTreeHash, len(tt.theSlice))
	for i := range tt.theSlice {
		// The presentation indices here need to be "local" rather than "global"
		result[i].PresentationIndex = HashIndexIdType(i + 1)
		result[i].Hash = make([]byte, config.HashLength) // Todo Yuk!
		copy(result[i].Hash, tt.theSlice[i][:])
	}
	return result
}
func (tt *TierTop) AppendHashesFile(hashesFile *os.File) error {
	srcFilename := filepath.Join(tt.folder, "TierTop", "Hashes.hsh")
	srcFile, err := os.Open(srcFilename)
	if err != nil {
		return err
	}
	defer func() { _ = srcFile.Close() }()

	defer func() { _ = hashesFile.Close() }()

	// 3. Efficiently stream/copy the data from source to destination
	// io.Copy uses a small internal buffer, preventing high memory usage for large files
	_, err = io.Copy(hashesFile, srcFile)
	if err != nil {
		return err
	}

	return nil
}
func (tt *TierTop) GetFirstPresentationIndex() GlobalPiType {
	return tt.firstGlobalPresentationIndex
}
