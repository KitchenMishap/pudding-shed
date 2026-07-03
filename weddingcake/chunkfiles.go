package weddingcake

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/edsrzf/mmap-go"
)

// ChunkFilesRead and ChunkFilesWrite are objects that handle the files representing a set
// of chunks in a hash indexing store.

type ChunkFilesWrite struct {
	folder string
}
type ChunkFilesRead struct {
	folder            string
	chunks            []Chunk
	levelXXMemoryMaps []mmap.MMap // These can be treated very much like byte slices
}

func NewChunkFilesWrite(folder string) *ChunkFilesWrite {
	return &ChunkFilesWrite{folder: folder}
}
func NewChunkFilesRead(folder string) *ChunkFilesRead {
	return &ChunkFilesRead{folder: folder}
}

func (cfw *ChunkFilesWrite) AppendChunk(chunk *Chunk) error {
	// We append some details to ChunksInfo.bin,
	// and the indexBytes and nodesBytes to the various LayerXX.bin files.
	// ChunksInfo.bin is presumed to exist (an empty file designates an empty store)
	// whereas each of the LayerXX.bin files may need to be created here
	fullName := filepath.Join(cfw.folder, "ChunksInfo.bin")
	file, err := os.OpenFile(fullName, os.O_WRONLY|os.O_APPEND, 0)
	if err != nil {
		return err
	}
	bytes := make([]byte, 8+1, 8+1+len(chunk.ChunkLevels)*(4+4))
	// Field A (per chunk): FirstGlobalPresentationIndexOfChunk
	binary.LittleEndian.PutUint64(bytes[:8], uint64(chunk.FirstGlobalPresentationIndexOfChunk))
	// Field B (per chunk): Number of levels for chunk
	bytes[8] = byte(len(chunk.ChunkLevels))
	// Field C & D (per level): count of index bytes and count of nodes bytes for this chunk at this level
	indexBytesCountBytes := [8]byte{}
	nodesBytesCountBytes := [8]byte{}
	for level := 0; level < len(chunk.ChunkLevels); level++ {
		binary.LittleEndian.PutUint64(indexBytesCountBytes[:], uint64(len(chunk.ChunkLevels[level].IndexBytes)))
		binary.LittleEndian.PutUint64(nodesBytesCountBytes[:], uint64(len(chunk.ChunkLevels[level].NodeBytes)))
		bytes = append(bytes, indexBytesCountBytes[:]...)
		bytes = append(bytes, nodesBytesCountBytes[:]...)
	}
	_, err = file.Write(bytes)
	if err != nil {
		_ = file.Close()
		return err
	}
	err = file.Close()
	if err != nil {
		return err
	}

	for level := 0; level < len(chunk.ChunkLevels); level++ {
		filename := fmt.Sprintf("Level%02X.bin", level)
		fullName = filepath.Join(cfw.folder, filename)
		// This level may not have been created for previous chunks, so we need the O_CREATE flag.
		// 0644 grants read/write permissions to the owner and read-only to others
		file, err = os.OpenFile(fullName, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
		if err != nil {
			return err
		}
		_, err = file.Write(chunk.ChunkLevels[level].IndexBytes)
		if err != nil {
			_ = file.Close()
			return err
		}
		_, err = file.Write(chunk.ChunkLevels[level].NodeBytes)
		if err != nil {
			_ = file.Close()
			return err
		}
		err = file.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

func (cfr *ChunkFilesRead) ReadAndMmap(config StoreConfig) error {
	// First we find all the LevelXX.bin files we can and mmap them into memory
	err := cfr.mmapLevelFiles()
	if err != nil {
		return err
	}

	// Then we read the ChunksInfo.bin file
	fullName := filepath.Join(cfr.folder, "ChunksInfo.bin")
	file, err := os.Open(fullName)
	if err != nil {
		return err
	}
	bytes, err := io.ReadAll(file)
	if err != nil {
		_ = file.Close()
		return err
	}
	err = file.Close()
	if err != nil {
		return err
	}

	// For each level, we need (inside this function) to keep track of how many bytes have been accounted for
	levelAccountedBytes := [65]uint64{} // Max 65 levels (initially zero)

	if len(bytes) == 0 {
		// Zero length ChunksInfo.bin means no chunks
		if len(cfr.levelXXMemoryMaps) > 0 {
			panic("ChunksInfo.bin empty but LevelXX.bin files exist")
		}
		cfr.chunks = make([]Chunk, 0)
	}
	offset := 0
	chunkIndex := 0
	for offset < len(bytes) {
		// Because there are bytes left, we have a first (or another) chunk
		cfr.chunks = append(cfr.chunks, Chunk{Config: config})
		// Field A (per chunk): 8 bytes "FirstGlobalPresentationIndexOfChunk"
		if len(bytes)-offset < 8 {
			_ = file.Close()
			return errors.New("ChunksInfo.bin file format error : Missing FirstPI")
		}
		cfr.chunks[chunkIndex].FirstGlobalPresentationIndexOfChunk = GlobalPiType(binary.LittleEndian.Uint64(bytes[offset : offset+8]))
		offset += 8
		// Field B (per chunk): 1 byte levels count
		if len(bytes)-offset < 1 {
			_ = file.Close()
			return errors.New("ChunksInfo.bin file format error : Missing levels count")
		}
		levelsCount := bytes[offset]
		offset += 1
		if len(cfr.levelXXMemoryMaps) < int(levelsCount) {
			panic("Fewer LevelXX files than specified in ChunksInfo.bin")
		}
		cfr.chunks[chunkIndex].ChunkLevels = make([]ChunkLevel, levelsCount)
		for level := byte(0); level < levelsCount; level++ {
			// For each level present:
			// Field C (per level per chunk): Four bytes length of indexBytes
			if len(bytes)-offset < 8+8 {
				_ = file.Close()
				return errors.New("ChunksInfo.bin file format error : Missing index and/or nodes byte counts")
			}
			indexBytesLength := binary.LittleEndian.Uint64(bytes[offset : offset+8])
			offset += 8
			// Field D (per level per chunk): Four bytes length of nodesBytes
			nodesBytesLength := binary.LittleEndian.Uint64(bytes[offset : offset+8])
			offset += 8
			// We combine these with levelAccountedBytes[] to determine slices into the mmap'd LevelXX.bin files
			indexSlice := cfr.levelXXMemoryMaps[level][levelAccountedBytes[level] : levelAccountedBytes[level]+indexBytesLength]
			levelAccountedBytes[level] += indexBytesLength
			nodesSlice := cfr.levelXXMemoryMaps[level][levelAccountedBytes[level] : levelAccountedBytes[level]+nodesBytesLength]
			levelAccountedBytes[level] += nodesBytesLength
			cfr.chunks[chunkIndex].ChunkLevels[level].IndexBytes = indexSlice
			cfr.chunks[chunkIndex].ChunkLevels[level].NodeBytes = nodesSlice
		}
	}
	return nil
}

func (cfr *ChunkFilesRead) mmapLevelFiles() error {
	// We mmap all the sequential "LevelXX.bin" files we can find, for hex XX = 00 to 40 inclusive.
	// (00 to 40 corresponds to 65 potential levels, 64 for the max byte length of supported hash,
	// plus another level for any final leaf nodes.
	// There will be no gaps, so once one is not found we stop looking.
	// Note that this function leaves the files open and mmap'd

	if len(cfr.levelXXMemoryMaps) > 0 {
		panic("Cannot mmap level files before un-mmap'ing previous")
	}
	for levelNum := 0; levelNum <= 0x40; levelNum++ {
		filename := fmt.Sprintf("Level%02X.bin", levelNum)
		fullName := filepath.Join(cfr.folder, filename)

		// First we try to open the file
		file, err := os.Open(fullName)
		if err != nil {
			break
		} // This file is not found; will try no more

		// If it exists, we will mmap it
		memoryMap, err := mmap.Map(file, mmap.RDONLY, 0)
		if err != nil {
			_ = file.Close()
			return err
		}
		cfr.levelXXMemoryMaps = append(cfr.levelXXMemoryMaps, memoryMap)

		// Now (even though still mmap'ed) we can and should close the file
		err = file.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func (cfr *ChunkFilesRead) UnMmapLevelFiles() error {
	for _, memoryMap := range cfr.levelXXMemoryMaps {
		err := memoryMap.Unmap()
		if err != nil {
			return err
		}
	}
	cfr.levelXXMemoryMaps = cfr.levelXXMemoryMaps[:0]
	return nil
}
