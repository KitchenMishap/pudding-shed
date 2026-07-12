package weddingcakeback

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/KitchenMishap/pudding-shed/memfile"
	"github.com/KitchenMishap/pudding-shed/wordfile"
	"github.com/edsrzf/mmap-go"
)

// TierBelow represents a tier other than TierTop at a particular index.
// The tier known as TierBelow[0] is "under" TierTop, and TierBelow[1] is under that.
// Each TierBelow is comprised of up to 255 DonutForests.

// A TierBelow object is primarily concerned with reading from the tier.
// (For writing, a TierBelow is instead constructed on disk one DonutForest at a time by DonutForestWite)
type TierBelow struct {
	Folder                 string
	TierFolder             string
	TierIndex              byte
	Config                 *CakeConfig
	LevelXXNodesMemoryMaps []mmap.MMap // These can be treated very much like byte slices
	JumpTableMemoryMap     mmap.MMap
	DonutForestsInfo       []DonutForestInfo
	underlyingFile         *os.File
	hashesFile             *wordfile.HashFile
	nextTier               TierReadable
}

// Check that implements
var _ TierReadable = (*TierBelow)(nil)

func NewTierBelow(folder string, tierIndex byte, config *CakeConfig) *TierBelow {
	result := TierBelow{}
	result.Folder = folder

	result.TierIndex = tierIndex
	tierFolderName := fmt.Sprintf("Tier%d", tierIndex)
	result.TierFolder = filepath.Join(folder, tierFolderName)

	result.Config = config

	return &result
}

func (tb *TierBelow) Open() error {
	// First we find all the LevelXX.bin files we can and mmap them into memory
	err := tb.mmapLevelFiles()
	if err != nil {
		return err
	}

	err = tb.mmapJumpTableFile()
	if err != nil {
		// Unmap the mmap'ed level files for a clean failure
		for _, memoryMap := range tb.LevelXXNodesMemoryMaps {
			_ = memoryMap.Unmap()
		}
		return err
	}

	err = tb.readInfoFile()
	if err != nil {
		// Unmap the mmap'ed level files for a clean failure
		for _, memoryMap := range tb.LevelXXNodesMemoryMaps {
			_ = memoryMap.Unmap()
		}
		// Unmap the mmap'ed jump table for a clean failure
		_ = tb.JumpTableMemoryMap.Unmap()
		return err
	}

	filePath := filepath.Join(tb.TierFolder, "Hashes.hsh")
	tb.underlyingFile, err = os.Open(filePath)
	if err != nil {
		// Unmap the mmap'ed level files for a clean failure
		for _, memoryMap := range tb.LevelXXNodesMemoryMaps {
			_ = memoryMap.Unmap()
		}
		// Unmap the mmap'ed jump table for a clean failure
		_ = tb.JumpTableMemoryMap.Unmap()
		return err
	}
	// Count the hashes
	stat, err := tb.underlyingFile.Stat()
	if err != nil {
		// Unmap the mmap'ed level files for a clean failure
		for _, memoryMap := range tb.LevelXXNodesMemoryMaps {
			_ = memoryMap.Unmap()
		}
		// Unmap the mmap'ed jump table for a clean failure
		_ = tb.JumpTableMemoryMap.Unmap()
		return err
	}
	hashesCount := stat.Size() / 32 // ToDo support other hash sizes

	aoFile, err := memfile.NewAppendOptimizedFile(tb.underlyingFile)
	if err != nil {
		// Unmap the mmap'ed level files for a clean failure
		for _, memoryMap := range tb.LevelXXNodesMemoryMaps {
			_ = memoryMap.Unmap()
		}
		// Unmap the mmap'ed jump table for a clean failure
		_ = tb.JumpTableMemoryMap.Unmap()
		return err
	}
	tb.hashesFile = wordfile.NewHashFile(aoFile, hashesCount)

	return nil
}

func (tb *TierBelow) Close() error {
	if tb.nextTier != nil {
		err := tb.nextTier.Close()
		if err != nil {
			return err
		}
	}
	// Unmap the mmap'ed files
	for _, memoryMap := range tb.LevelXXNodesMemoryMaps {
		err := memoryMap.Unmap()
		if err != nil {
			return err
		}
	}
	tb.LevelXXNodesMemoryMaps = tb.LevelXXNodesMemoryMaps[:0]
	err := tb.JumpTableMemoryMap.Unmap()
	if err != nil {
		return err
	}

	err = tb.hashesFile.Close()
	if err != nil {
		return err
	}

	return nil
}

func (tb *TierBelow) TryIndexOfHash(hash []byte) (GlobalPiType, bool, error) {
	var flagsHashByteIndicesUnexamined uint64
	if tb.Config.HashLength == 64 {
		flagsHashByteIndicesUnexamined = 0xFFFFFFFFFFFFFFFF
	} else {
		flagsHashByteIndicesUnexamined = 1<<(tb.Config.HashLength) - 1
	}

	// First find an index into the jump table
	prefix := hash[:8]
	prefixBytesCount := tb.TierIndex
	multiplier := 1
	prefixIndex := 0
	flagBit := uint64(1)
	for prefixByte := byte(0); prefixByte < prefixBytesCount; prefixByte++ {
		prefixIndex += int(prefix[prefixByte]) * multiplier
		multiplier <<= 8
		flagsHashByteIndicesUnexamined ^= flagBit
		flagBit <<= 1
	}

	// Work out the size of each jump table
	nodeIdConfig := &tb.Config.TierBelowConfigs[tb.TierIndex].NodeIdConfig
	hashIndexIdConfig := &tb.Config.TierBelowConfigs[tb.TierIndex].HashIndexIdConfig
	reassuranceBytesCount := tb.Config.TierBelowConfigs[tb.TierIndex].ReassuranceBytesCount
	nodeIdSize := (*nodeIdConfig).StorageBytes()
	jumpTableEntries := 1 << (prefixBytesCount * 8) // = 256 ^ prefixBytesCount
	jumpTableSize := jumpTableEntries * nodeIdSize

	// Now iterate over all DonutForests
	donutForestsCount := len(tb.DonutForestsInfo)
	for donutForestIndex := range donutForestsCount {
		donutForestInfo := &tb.DonutForestsInfo[donutForestIndex]

		// Read root of SingleTree from jump table
		jumpTableByteOffset := donutForestIndex*jumpTableSize + prefixIndex*nodeIdSize
		singleTreeNodeId := (*nodeIdConfig).ReadID(tb.JumpTableMemoryMap[jumpTableByteOffset : jumpTableByteOffset+nodeIdSize])
		if singleTreeNodeId != 0 {
			level := prefixBytesCount
			hashIndexId := tb.recurseLookupHash(hash, level, singleTreeNodeId,
				flagsHashByteIndicesUnexamined, donutForestInfo,
				nodeIdConfig, hashIndexIdConfig, reassuranceBytesCount)

			if hashIndexId != HashIndexIdNoMatch {
				// Found a potential match
				// ToDo check against hashes file
				return GlobalPiType(hashIndexId-1) + tb.DonutForestsInfo[donutForestIndex].FirstGlobalPresentationIndex, true, nil
			}
		}
	}
	return GlobalPiNoMatch, false, nil
}

func (tb *TierBelow) TryGetHashAtIndex(index GlobalPiType, hash []byte) (bool, error) {
	// tb.hashesFile covers all the hashes contained in multiple DonutForest's of the tier.
	// We therefore adjust the index based on the first DonutForest
	localHashIndexId := index - tb.DonutForestsInfo[0].FirstGlobalPresentationIndex
	if localHashIndexId < 0 {
		return false, nil
	} else if localHashIndexId >= tb.hashesFile.CountHashes() {
		return false, nil
	}
	h, err := tb.hashesFile.ReadHashAt(localHashIndexId)
	if err != nil {
		return false, err
	}
	copy(hash, h[:])
	return true, nil
}

func (tb *TierBelow) GetNextTier() TierReadable {
	return tb.nextTier
}

func (tb *TierBelow) recurseLookupHash(hash []byte, levelNum byte,
	nodeIdWithinLevel NodeIdType, flagsHashByteIndicesUnexamined uint64,
	donutForestInfo *DonutForestInfo,
	nodeIdConfig *NByteIdConfig[NodeIdType], hashIndexIdConfig *NByteIdConfig[HashIndexIdType],
	reassuranceBytesCount byte) HashIndexIdType {

	// Look at the node we were directed to
	var node donutForestNode
	donutForestInfo.Levels[levelNum].ExtractNode(nodeIdWithinLevel, &node, nodeIdConfig)

	// Have we reached a leaf node?
	isLeaf, reassuranceBytes, hashIndexId := node.detailsIfLeaf(reassuranceBytesCount, hashIndexIdConfig)
	if isLeaf {
		// Now we need to check some bytes of our hash... the "next" ones that haven't been examined yet...
		// against the reassurance bytes specified by the leaf
		mask := flagsHashByteIndicesUnexamined
		byteToMaybeExamine := 0
		reassuranceByteCounter := 0
		for mask != 0 && reassuranceByteCounter < len(reassuranceBytes) {
			if mask&uint64(1) == 1 {
				// Yes we need to examine it
				match := hash[byteToMaybeExamine] == reassuranceBytes[reassuranceByteCounter]
				if !match {
					// This leaf was our only shot at a match, but the reassurance bytes were not reassuring!
					//fmt.Printf("Leaf failed reassurance bytes match at level %d\n", levelNum)
					return 0 // Not a match
				}
				// A byte matched... now lets see if there are any more reassurance bytes available...
				reassuranceByteCounter++
			}
			// Either we didn't need to examine byteToExamine (it had already been examined),
			// or it matched. In either case, we may have more reassurance bytes to check...
			// so keep looking...
			mask >>= 1
			byteToMaybeExamine++
		}
		// All the available reassurance bytes matched... It's a potential match!
		// It is for the caller to double check the ENTIRE hash, now that we've identified a unique strong match
		// with the data available
		return hashIndexId
	}
	// Not a leaf.
	// This node is instructing us to dig deeper, by examining another byte in the hash.
	byteIndexToExamine, mediumSlots, tinySlots := node.hashByteIndexToExamine(nodeIdConfig)
	// Examine the specified byte
	byteThatWasFound := hash[byteIndexToExamine]
	// See if that byte get's us to another node at the next level...
	nextNodeId := node.nextLevelNodeId(byteThatWasFound, mediumSlots, tinySlots, nodeIdConfig)
	if nextNodeId == 0 {
		return HashIndexIdNoMatch // A dead end. No match. The hash definitely isn't in this DonutForest.
	}
	// Make a note that this byte index of the hash has now been examined
	bitMask := uint64(1) << byteIndexToExamine
	if flagsHashByteIndicesUnexamined&bitMask == 0 {
		panic("We examined the same byte of the hash twice")
	}
	flagsHashByteIndicesUnexamined ^= bitMask // Clear the bit to say it is no longer unexamined

	// Go deeper, with our new node id at the next level...
	return tb.recurseLookupHash(hash, levelNum+1, nextNodeId, flagsHashByteIndicesUnexamined,
		donutForestInfo, nodeIdConfig, hashIndexIdConfig, reassuranceBytesCount)
}

func (tb *TierBelow) mmapLevelFiles() error {
	// We mmap all the sequential "LevelXXNodes.bin" files we can find, for hex XX = 00 to 40 inclusive.
	// (00 to 40 corresponds to 65 potential levels, 64 for the max byte length of supported hash,
	// plus another level for any final leaf nodes.
	// Note that this function leaves the files mmap'd

	// Due to the "jump table" "forest of trees" structure of DonutForests, nodes might not be found
	// at the earlier levels. So the pattern is "empty levels", followed by "non-empty levels" followed
	// by "empty levels" again.

	nonEmptyLevelFound := false
	if len(tb.LevelXXNodesMemoryMaps) > 0 {
		panic("Cannot mmap level files before un-mmap'ing previous")
	}
	for levelNum := 0; levelNum <= 0x40; levelNum++ {
		filename := fmt.Sprintf("Level%02XNodes.bin", levelNum)
		fullName := filepath.Join(tb.TierFolder, filename)

		// First we try to open the file
		file, err := os.Open(fullName)
		if err != nil {
			// Empty level found
			if nonEmptyLevelFound {
				break
			} // We've seen non-empty levels followed by this empty level, so stop looking
			tb.LevelXXNodesMemoryMaps = append(tb.LevelXXNodesMemoryMaps, nil)
		} else {
			nonEmptyLevelFound = true

			// If it exists, we will mmap it
			memoryMap, err := mmap.Map(file, mmap.RDONLY, 0)
			if err != nil {
				_ = file.Close()
				return err
			}
			tb.LevelXXNodesMemoryMaps = append(tb.LevelXXNodesMemoryMaps, memoryMap)

			// Now (even though still mmap'ed) we can and should close the file
			err = file.Close()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (tb *TierBelow) mmapJumpTableFile() error {
	fullName := filepath.Join(tb.TierFolder, "DonutForestsJumpTables.bin")

	// First we try to open the file
	file, err := os.Open(fullName)
	if err != nil {
		return err
	}
	tb.JumpTableMemoryMap, err = mmap.Map(file, mmap.RDONLY, 0)
	// Now (even though still mmap'ed) we can and should close the file
	if err != nil {
		_ = file.Close()
		return err
	}
	err = file.Close()
	if err != nil {
		return err
	}
	return nil
}

func (tb *TierBelow) readInfoFile() error {
	fullName := filepath.Join(tb.TierFolder, "DonutForestsInfo.bin")
	file, err := os.Open(fullName)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()
	bytes, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	// For each level, we need (inside this function) to keep track of how many bytes have been accounted for
	levelAccountedBytes := [65]uint64{} // Max 65 levels (initially zero)

	if len(bytes) == 0 {
		// Zero length file means no DonutForests
		if len(tb.LevelXXNodesMemoryMaps) > 0 {
			panic("DonutForestsInfo.bin empty but LevelXXNodes.bin files exist")
		}
		tb.DonutForestsInfo = make([]DonutForestInfo, 0)
		return nil
	}
	tb.DonutForestsInfo = make([]DonutForestInfo, 0, 255)
	offset := 0
	donutForestIndex := 0
	for offset < len(bytes) {
		// Because there are bytes left, we have a first (or another) DonutForest
		tb.DonutForestsInfo = append(tb.DonutForestsInfo, DonutForestInfo{})
		// Field A (per DonutForest): 8 bytes "FirstGlobalPresentationIndexOfChunk"
		if len(bytes)-offset < 8 {
			_ = file.Close()
			return errors.New("DonutForestsInfo.bin file format error : Missing FirstPI")
		}
		tb.DonutForestsInfo[donutForestIndex].FirstGlobalPresentationIndex =
			GlobalPiType(binary.LittleEndian.Uint64(bytes[offset : offset+8]))
		offset += 8
		// Field B (per DonutForest): 1 byte levels count
		if len(bytes)-offset < 1 {
			_ = file.Close()
			return errors.New("DonutForestsInfo.bin file format error : Missing levels count")
		}
		levelsCount := bytes[offset]
		offset += 1
		if len(tb.LevelXXNodesMemoryMaps) < int(levelsCount) {
			panic("Fewer LevelXX files than specified in DonutForestsInfo.bin")
		}
		tb.DonutForestsInfo[donutForestIndex].Levels = make([]DonutForestLevelSlices, levelsCount)
		for level := byte(0); level < levelsCount; level++ {
			// For each level present:
			// Field C (per level per chunk): 8 bytes length of indexBytes
			if len(bytes)-offset < 8+8 {
				_ = file.Close()
				return errors.New("DonutForestsInfo.bin file format error : Missing index and/or nodes byte counts")
			}
			indexBytesLength := binary.LittleEndian.Uint64(bytes[offset : offset+8])
			offset += 8
			// Field D (per level per chunk): Four bytes length of nodesBytes
			nodesBytesLength := binary.LittleEndian.Uint64(bytes[offset : offset+8])
			offset += 8
			// We combine these with levelAccountedBytes[] to determine slices into the mmap'd LevelXX.bin files
			indexSlice := tb.LevelXXNodesMemoryMaps[level][levelAccountedBytes[level] : levelAccountedBytes[level]+indexBytesLength]
			levelAccountedBytes[level] += indexBytesLength
			nodesSlice := tb.LevelXXNodesMemoryMaps[level][levelAccountedBytes[level] : levelAccountedBytes[level]+nodesBytesLength]
			levelAccountedBytes[level] += nodesBytesLength
			tb.DonutForestsInfo[donutForestIndex].Levels[level].IndexBytes = indexSlice
			tb.DonutForestsInfo[donutForestIndex].Levels[level].NodesBytes = nodesSlice
		}
		donutForestIndex++
	}
	return nil
}
