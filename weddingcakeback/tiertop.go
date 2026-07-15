package weddingcakeback

import (
	"encoding/binary"
	"io"
	"os"
	"path/filepath"

	"github.com/KitchenMishap/pudding-shed/memfile"
	"github.com/KitchenMishap/pudding-shed/wordfile"
)

// TierTopIndex is a type only used locally within a TierTop
// It's values are equal to GlobalPiType's minus firstGlobalPresentationIndex
// It is used to directly index into theSlice, and is also the value type of theMap.
// There is no "special value" to mean "no match"; there doesn't need to be.
type TierTopIndex int

// TierTop is the type of the top tier. It is different from the tiers below it, so has its own class
// Note that the next tier below this is indexed as ZERO (TierTop has no tier index!)
type TierTop struct {
	folder         string
	config         *CakeConfig
	nextTierConfig *TierBelowConfig // Within the CakeConfig
	readonly       bool
	underlyingFile *os.File
	hashesFile     *wordfile.HashFile
	// The "value" type (int) of theMap is an index into theSlice. It is neither a GlobalPiType nor a HashIndexIdType.
	theMap                       map[[32]byte]TierTopIndex // ToDo support hash sizes other than 32
	theSlice                     [][]byte                  // indexed by TierTopIndex values
	firstGlobalPresentationIndex GlobalPiType
	nextTier                     TierReadable
}

// Check that implements
var _ TierReadable = (*TierTop)(nil)

func NewTierTop(cakeFolder string, config *CakeConfig, readOnly bool) (*TierTop, error) {
	result := TierTop{}
	result.folder = cakeFolder
	result.config = config
	result.nextTierConfig = &config.TierBelowConfigs[0]
	result.nextTier = nil
	err := result.Open(readOnly)
	if err != nil {
		return nil, err
	}

	tierSoFar := BakingSourceTier(&result)

	// Try opening each next tier
	// An empty tier does NOT indicate no subsequent non-empty tiers
	var lastNonEmptyTier *TierBelow = nil
	// Check as many as are described in the config
	for i := byte(0); i < byte(len(config.TierBelowConfigs)); i++ {
		nextTier := NewTierBelow(cakeFolder, i, config)
		exists := nextTier.ExistsOnDisk()
		if exists {
			lastNonEmptyTier = nextTier
			err = nextTier.Open()
			if err != nil {
				return nil, err
			}
			tierSoFar.SetNextTier(nextTier)
			tierSoFar = nextTier
		} else {
			// An empty tier. Create it for now, and then later break the chain of trailing empty tiers
			nextTier.OpenAsEmptyTier()
			tierSoFar.SetNextTier(nextTier)
			tierSoFar = nextTier
		}
	}
	// Now break the chain of pointers, getting rid of the trailing empty tiers
	if lastNonEmptyTier == nil {
		// TierTop (and it may itself be empty) is the last non-empty tier
		result.nextTier = nil
	} else {
		lastNonEmptyTier.SetNextTier(nil)
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
	hashesCount := stat.Size() / int64(tt.config.HashLength)

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

	tt.theMap = make(map[Sha256]TierTopIndex, 65535)
	tt.theSlice = make([][]byte, hashesCount, 65535)

	for i := TierTopIndex(0); i < TierTopIndex(hashesCount); i++ {
		hash, err := tt.hashesFile.ReadHashAt(int64(i))
		if err != nil {
			return err
		}
		tt.theMap[hash] = i
		tt.theSlice[i] = hash[:]
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
	index, ok := tt.theMap[h]
	if !ok {
		return GlobalPiNoMatch, false, nil
	}
	return GlobalPiFromTierTopIndex(index, tt.firstGlobalPresentationIndex), true, nil
}

// TryGetHashAtIndex for TierReadable interface
func (tt *TierTop) TryGetHashAtIndex(globalIndex GlobalPiType, hash []byte) (bool, error) {
	if !GlobalPiWithinRange(globalIndex, tt.firstGlobalPresentationIndex, len(tt.theSlice)) {
		return false, nil // index is outside the range of this tier
	}
	sliceIndex := TierTopIndexFromGlobalPi(globalIndex, -tt.firstGlobalPresentationIndex)
	copy(hash, tt.theSlice[sliceIndex])
	return true, nil
}

// GetNextTier for TierReadable interface
func (tt *TierTop) GetNextTier() TierReadable {
	return tt.nextTier
}

func (tt *TierTop) CountHashes() (GlobalPiType, error) {
	return GlobalPiFromUint64(uint64(len(tt.theSlice))), nil
}

func (tt *TierTop) CloseAll() error {
	if tt.nextTier != nil {
		err := tt.nextTier.CloseAll()
		if err != nil {
			return err
		}
	}
	return tt.CloseThis()
}
func (tt *TierTop) CloseThis() error {
	if tt.hashesFile != nil {
		return tt.hashesFile.Close()
	}
	tt.hashesFile = nil
	return nil
}

func (tt *TierTop) AppendHash(hash []byte) (GlobalPiType, error) {
	index := len(tt.theSlice)
	if index >= 65535 {
		panic("TierTop only supports 65535 hashes (append would overflow)")
	}
	// We calculate the following BEFORE any writer.Write() below changes things
	resultIndex := GlobalPiFromTierTopIndex(TierTopIndex(index), tt.firstGlobalPresentationIndex)

	// Todo support hash sizes other than 32
	if len(hash) != 32 {
		panic("Currently only 32 byte hashes supported")
	}
	h := [32]byte{}
	copy(h[:], hash)

	tt.theMap[h] = TierTopIndex(index)
	tt.theSlice = append(tt.theSlice, hash)
	if tt.hashesFile == nil {
		panic("Hashes file not open")
	}
	_, err := tt.hashesFile.AppendHash(h)
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
	return tt.nextTierConfig.PrefixBytesCount
}
func (tt *TierTop) GetNextTierIndex() byte {
	// The index of the next tier after TierTop is (surprisingly) 0
	return 0
}
func (tt *TierTop) GetIndicesCount() uint64 {
	// This tier (TierTop) has no prefix bytes, so it has 256^0 = 1 indices
	return 1
}
func (tt *TierTop) GetHashesAtIndex(index uint64, offsetToUse GlobalPiType) []SingleTreeHash {
	if index != 0 {
		panic("TierZero.GetHashesForIndex() should only be called with index=0")
	}
	result := make([]SingleTreeHash, len(tt.theSlice))
	for i := range tt.theSlice {
		globalPi := GlobalPiFromTierTopIndex(TierTopIndex(i), offsetToUse)
		singleTreePi := SingleTreePiFromGlobalPi(globalPi, offsetToUse)
		result[i].PresentationIndex = singleTreePi
		result[i].SourceOffset = offsetToUse
		result[i].Hash = make([]byte, tt.config.HashLength) // Todo Yuk!
		copy(result[i].Hash, tt.theSlice[i][:])
	}
	return result
}
func (tt *TierTop) AppendHashesFile(hashesFile *os.File) error {
	if tt.hashesFile != nil {
		err := tt.hashesFile.Sync()
		if err != nil {
			return err
		}
	}
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
		panic("Cannot make empty if readonly")
	}

	// Count the hashes
	hashesCount := tt.hashesFile.CountHashes()

	// Close the hashes file
	err := tt.hashesFile.Close()
	if err != nil {
		return err
	}

	// Update the FirstPresentationIndex
	tt.firstGlobalPresentationIndex += GlobalPiFromUint64(uint64(hashesCount))

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
		err = tt.nextTier.CloseThis()
		if err != nil {
			return err
		}
		tt.nextTier = nil
	}

	return nil
}
func (tt *TierTop) SetNextTier(tierReadable TierReadable) {
	tt.nextTier = tierReadable
}
