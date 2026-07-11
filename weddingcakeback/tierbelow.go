package weddingcakeback

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/edsrzf/mmap-go"
)

// TierBelow represents a tier other than TierTop at a particular index.
// The tier known as TierBelow[0] is "under" TierTop, and TierBelow[1] is under that.
// Each TierBelow is comprised of up to 255 DonutForests.

// A TierBelow object is primarily concerned with reading from te tier.
// (For writing, a TierBelow is instead constructed on disk one DonutForest at a time by DonutForestWite)
type TierBelow struct {
	Folder                 string
	TierFolder             string
	TierIndex              byte
	LevelXXNodesMemoryMaps []mmap.MMap // These can be treated very much like byte slices
	JumpTableMemoryMap     mmap.MMap
	DonutForestsInfo       []DonutForestInfo
}

func NewTierBelow(folder string, tierIndex byte) *TierBelow {
	result := TierBelow{}
	result.Folder = folder

	result.TierIndex = tierIndex
	tierFolderName := fmt.Sprintf("Tier%d", tierIndex)
	result.TierFolder = filepath.Join(folder, tierFolderName)

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
	return nil
}

func (tb *TierBelow) Close() error {
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

	return nil
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
	}
	return nil
}

type DonutForestInfo struct {
	FirstGlobalPresentationIndex GlobalPiType
	Levels                       []DonutForestLevelSlices
}
type DonutForestLevelSlices struct {
	// These are slices into the mmap'ed files
	IndexBytes mmap.MMap
	NodesBytes mmap.MMap
}
