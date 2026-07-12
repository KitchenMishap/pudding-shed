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
	config         *CakeConfig
	readonly       bool
	underlyingFile *os.File
	hashesFile     *wordfile.HashFile
	// As with other tiers, theMap deals with HashIndexIdType which is not the same as GlobalPiType
	theMap map[[32]byte]HashIndexIdType // ToDo support hash sizes other than 32
	// Because HashIndexType's reserve zero to mean "no match", subtract 1 from a HashIndexId to index theSlice
	theSlice                     [][]byte
	firstGlobalPresentationIndex GlobalPiType
	nextTier                     TierReadable
}

// Check that implements
var _ TierReadable = (*TierTop)(nil)

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

func NewTierTop(cakeFolder string, config *CakeConfig, readOnly bool) (*TierTop, error) {
	result := TierTop{}
	result.folder = cakeFolder
	result.config = config
	result.nextTier = nil
	err := result.Open(readOnly)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (tt *TierTop) Open(readOnly bool) error {
	filePath := filepath.Join(tt.folder, "TierTop", "Hashes.hsh")
	tt.readonly = readOnly
	var err error
	if readOnly {
		tt.underlyingFile, err = os.Open(filePath)
	} else {
		tt.underlyingFile, err = os.OpenFile(filePath, os.O_RDWR|os.O_APPEND, 0)
	}
	if err != nil {
		return err
	}
	// Count the hashes
	stat, err := tt.underlyingFile.Stat()
	if err != nil {
		return err
	}
	hashesCount := stat.Size() / 32 // ToDo support other hash sizes

	if hashesCount > 65535 {
		panic("TierTop only supports 65535 hashes (more found in file)")
	}

	firstPiFile, err := os.Open(filepath.Join(tt.folder, "TierTop", "FirstPresentationIndex.bin"))
	if err != nil {
		return err
	}
	defer func() { _ = firstPiFile.Close() }()
	var firstPi GlobalPiType
	err = binary.Read(firstPiFile, binary.LittleEndian, &firstPi)
	if err != nil {
		return err
	}
	tt.firstGlobalPresentationIndex = firstPi

	aoFile, err := memfile.NewAppendOptimizedFile(tt.underlyingFile)
	if err != nil {
		return err
	}
	tt.hashesFile = wordfile.NewHashFile(aoFile, hashesCount)

	tt.theMap = make(map[Sha256]HashIndexIdType, 65535)
	tt.theSlice = make([][]byte, hashesCount, 65535)

	// HashIndexId's start at 1
	for i := HashIndexIdType(1); i < HashIndexIdType(hashesCount+1); i++ {
		hash, err := tt.hashesFile.ReadHashAt(int64(i - 1))
		if err != nil {
			return err
		}
		tt.theMap[hash] = i
		tt.theSlice[i-1] = hash[:]
	}
	return nil
}

// TryIndexOfHash for TierReadable interface
func (tt *TierTop) TryIndexOfHash(hash []byte) (GlobalPiType, bool, error) {
	// ToDo support hash sizes other than 32
	if len(hash) != 32 {
		panic("hash size not supported")
	}
	h := [32]byte{}
	copy(h[:], hash)
	hashIndexId, ok := tt.theMap[h]
	if !ok {
		return GlobalPiNoMatch, false, nil
	}
	return tt.GlobalPresentationIndexFromHashIndexId(hashIndexId), true, nil
}

// TryGetHashAtIndex for TierReadable interface
func (tt *TierTop) TryGetHashAtIndex(index GlobalPiType, hash []byte) (bool, error) {
	localIndex := tt.HashIndexIdFromGlobalPresentationIndex(index) - 1
	if localIndex < 0 || localIndex >= HashIndexIdType(len(tt.theSlice)) {
		return false, nil
	}
	copy(hash, tt.theSlice[localIndex])
	return true, nil
}

// GetNextTier for TierReadable interface
func (tt *TierTop) GetNextTier() TierReadable {
	return tt.nextTier
}

func (tt *TierTop) CountHashes() (GlobalPiType, error) {
	return GlobalPiType(len(tt.theSlice)), nil
}

func (tt *TierTop) Close() error {
	if tt.nextTier != nil {
		err := tt.nextTier.Close()
		if err != nil {
			return err
		}
	}
	if tt.hashesFile != nil {
		return tt.hashesFile.Close()
	}
	return nil
}

func (tt *TierTop) AppendHash(hash []byte) (GlobalPiType, error) {
	index, err := tt.CountHashes()
	if index >= 65535 {
		panic("TierTop only supports 65535 hashes (append would overflow)")
	}
	if err != nil {
		return -1, err
	}
	// We calculate the following BEFORE any writer.Write() below changes things
	resultIndex := tt.GlobalPresentationIndexFromHashIndexId(HashIndexIdType(index + 1))

	// Todo support hash sizes other than 32
	if len(hash) != 32 {
		panic("Currently only 32 byte hashes supported")
	}
	h := [32]byte{}
	copy(h[:], hash)

	tt.theMap[h] = HashIndexIdType(index + 1)
	tt.theSlice = append(tt.theSlice, hash)
	_, err = tt.hashesFile.AppendHash(h)
	if err != nil {
		return -1, err
	}

	// Index was the count of  hashes before the append. If it was 65534, we now have 65535
	// If reached capacity, trigger a baking
	if index == 65534 {
		writer := NewDonutForestWrite(tt, tt.config)
		err = writer.Write(tt.folder)
		if err != nil {
			return -1, err
		}
	}

	return resultIndex, nil
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
func (tt *TierTop) MakeEmptyAfterBaking() error {
	// We add hashesCount to the FirstPresentationIndex, delete the files, then use
	// TierTopCreator to recreate fresh files. We then call Open().
	// This is to re-use as much code as possible.

	if tt.readonly {
		return errors.New("Cannot make empty if readonly")
	}

	// Count the hashes
	hashesCount := tt.hashesFile.CountHashes()

	// Close the hashes file
	err := tt.hashesFile.Close()
	if err != nil {
		return err
	}

	// Update the FirstPresentationIndex
	tt.firstGlobalPresentationIndex += hashesCount

	// Delete the hashes file
	err = os.Remove(filepath.Join(tt.folder, "TierTop", "Hashes.hsh"))
	if err != nil {
		return err
	}
	// Delete the FirstPresentationIndex.bin file
	err = os.Remove(filepath.Join(tt.folder, "TierTop", "FirstPresentationIndex.bin"))
	if err != nil {
		return err
	}

	// Create files again
	creator := NewTierTopCreator(tt.folder, tt.config)
	err = creator.Create(tt.firstGlobalPresentationIndex)
	if err != nil {
		return err
	}

	err = tt.Open(tt.readonly)
	if err != nil {
		return err
	}

	// We were emptying this tier after it was baked into a DonutForest of a subsequent tier.
	// Therefore we need to close the subsequent tier, ready to be opened in its new configuration.
	if tt.nextTier == nil {
		// The new DonutForest was just created in a brand new tier; there is no previously opened tier to close
	} else {
		tt.nextTier.Close()
		tt.nextTier = nil
	}

	return nil
}
func (tt *TierTop) SetNextTier(tierReadable TierReadable) {
	tt.nextTier = tierReadable
}
