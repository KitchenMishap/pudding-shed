package indexedhashes

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/KitchenMishap/pudding-shed/numberedfolders"
	"github.com/KitchenMishap/pudding-shed/testpoints"
	"math"
	"os"
	"sync"
)

// The file formats for UniformHashStore are as follows.
// All file contents are raw binary.

// The file Hashes.hsh contains the raw binary hashes, 32 bytes each. This file is an input to MultipassPreloader.
// The hashes in this file (and elsewhere) are LittleEndian: LSBs first.
// For block hashes, which due to proof of work start with zeroes, these zeroes are at the END of each 32 bytes.
// This means that the first 8 bytes can be treated as pseudo-random, and can be used to generate
// addresses so that each index can go in a bin identified by the address.
// Multiple different hashes, so long as they correspond to the same address, therefore end up in the same bin.

// The file BinStarts.bst is a fixed size large structured file split sequentially into bins of a fixed size.
// The fixed bin size WAS intended to correspond to the file system's file allocation size, but this is no longer true.
// The number of bins is fixed and optimized based on a rough estimate (not a maximum) of the number of hashes expected.
// A bin starts with an 8 byte count of the number of indices in the bin. Despite the 8 bytes, this number
// is typically less than 100, but CAN be more than 102.
// Then 8 bytes reserved (zero). There are then 102 slots of 8 byte entries.
// Unused entries at the end are zeroed. So a bin is always 8(count) + 8(reserved) + 102*8(entries) = 832 bytes.
// An entry consists of an 8 byte index for the hash (ie the order the hash appears in Hashes.hsh).
// The hash itself is NO LONGER stored in an entry - instead use the index to refer back into the original Hashes.hsh.
// If the count specified in the first 8 bytes is more than 102, then there will be files for overflow in BinOverflows/.

// BinOverflows/ files are small files in a numbered folder structure based on the bin address.
// Each file is a multiple of 8 byte entries, the first entry in the file being slot 102, and with the
// same (index) format as above - with no count or reserved or zero padding (the file can be any multiple of 8)

// MultipassPreloader is a system that preloads a UniformHashStore fileset.
// To minimize memory usage, it reads the hashes file multiple times, each time
// concentrating on a different set of address bins (the set is known as a dumpster)
type MultipassPreloader struct {
	creator         *UniformHashStoreCreator
	bytesPerPass    int64 // A parameter of the method, but does not affect the files which are output
	binStartsFile   *os.File
	numberedFolders numberedfolders.NumberedFolders
}

func NewMultipassPreloader(creator *UniformHashStoreCreator, bytesPerPass int64) *MultipassPreloader {
	result := MultipassPreloader{}
	result.creator = creator
	result.bytesPerPass = bytesPerPass
	result.binStartsFile = nil
	result.numberedFolders = numberedfolders.NewNumberedFolders(0, creator.params.DigitsPerFolder)
	return &result
}

type indexHash struct {
	index int64
	hash  Sha256 // Hashes are stored whilst in mem, but not in file
}

type binStart struct {
	entryCount  int64
	reserved    int64
	indexHashes [102]indexHash
}

const binStartSize = 8 + 8 + 102*8

func (bs *binStart) ToBytes(bytes *[binStartSize]byte) {
	binary.LittleEndian.PutUint64(bytes[0:8], uint64(bs.entryCount))
	binary.LittleEndian.PutUint64(bytes[8:16], uint64(bs.reserved))
	for i := 0; i < 102; i++ {
		binary.LittleEndian.PutUint64(bytes[16+i*8:16+i*8+8], uint64(bs.indexHashes[i].index))
	}
}

func (bs *binStart) FromBytes(bytes *[binStartSize]byte) {
	bs.entryCount = int64(binary.LittleEndian.Uint64(bytes[0:8]))
	bs.reserved = int64(binary.LittleEndian.Uint64(bytes[8:16]))
	for i := 0; i < 102; i++ {
		bs.indexHashes[i].index = int64(binary.LittleEndian.Uint64(bytes[16+i*8 : 16+i*8+8]))
	}
}

func (bs *binStart) hashAlreadyInBinStart(hash Sha256) bool {
	entryCount := bs.entryCount
	if entryCount > 102 {
		entryCount = 102 // Not interested in the overflows
	}
	for i := int64(0); i < entryCount; i++ {
		if bs.indexHashes[i].hash == hash {
			return true
		}
	}
	return false
}

// SinglePassDetails holds details of one of the multiple passes
type SinglePassDetails struct {
	firstAddress          int64
	lastAddressPlusOne    int64
	binStarts             []binStart    // Fixed size
	overflows             [][]indexHash // Elements (arrays) grow
	duplicateRemovalMutex sync.Mutex
}

func NewSinglePassDetails(firstAddress int64, addressCount int64) *SinglePassDetails {
	result := SinglePassDetails{}
	result.firstAddress = firstAddress
	result.lastAddressPlusOne = firstAddress + addressCount
	result.binStarts = make([]binStart, addressCount)
	result.overflows = make([][]indexHash, addressCount)
	for i := int64(0); i < addressCount; i++ {
		result.overflows[i] = make([]indexHash, 0) // Each starts empty, with empty capacity
	}
	return &result
}

func (spd *SinglePassDetails) ReadIn(mp *MultipassPreloader) error {
	hashesFile, err := os.Open(mp.creator.hashFilePath())
	if err != nil {
		return err
	}
	defer hashesFile.Close()
	ih := indexHash{}
	ih.index = 0

	chunk := make([]byte, 4096) // We read up to 4096 bytes at a time

	nBytes, _ := hashesFile.Read(chunk)
	for nBytes > 0 {
		if nBytes%32 != 0 {
			return errors.New("invalid hash file length")
		}
		hashCount := nBytes / 32
		for index := 0; index < hashCount; index++ {
			copy(ih.hash[:], chunk[index*32:index*32+32])
			spd.dealWithOneHash(&ih, mp)
			ih.index++
		}
		nBytes, _ = hashesFile.Read(chunk)
	}
	return nil
}

func (spd *SinglePassDetails) dealWithOneHash(ih *indexHash, mp *MultipassPreloader) {
	// === TestPoint ===
	// TestPoint for when inserting nth hash (but hit for nth block, nth transaction, and nth address)
	if testpoints.TestPointBlockEnable && ih.index == testpoints.TestPointBlock {
		fmt.Println("TESTPOINT: SinglePassDetails.dealWithOneHash(index = ", testpoints.TestPointBlock, ")")
	}

	LSWord := binary.LittleEndian.Uint64(ih.hash[0:8])
	address := int64(LSWord / mp.creator.params.HashDivider)
	if address < spd.firstAddress || address >= spd.lastAddressPlusOne {
		return // Not interested. Will be dealt with in a different pass.
	}
	passAddress := address - spd.firstAddress
	bin := &spd.binStarts[passAddress]

	// Is it in the binstart already?
	if bin.hashAlreadyInBinStart(ih.hash) {
		return // Ignore all but first occurrence of a hash
	}

	// Is it in the overflows already?
	overflows := &spd.overflows[passAddress]
	for i := 0; i < len(*overflows); i++ {
		if (*overflows)[i].hash == ih.hash {
			return // Ignore all but first occurrence of a hash
		}
	}

	countSoFar := bin.entryCount
	bin.entryCount++
	if countSoFar < 102 {
		// Goes in binStarts
		bin.indexHashes[countSoFar] = *ih
	} else {
		// Goes in overflows
		spd.overflows[passAddress] = append(*overflows, *ih)
	}
}

func (spd *SinglePassDetails) writeIntoBinStartsFile(mp *MultipassPreloader) error {
	chunk := [binStartSize]byte{}
	for index, element := range spd.binStarts {
		binAddress := spd.firstAddress + int64(index)
		fileByte := binStartSize * binAddress
		element.ToBytes(&chunk)
		_, err := mp.binStartsFile.WriteAt(chunk[:], int64(fileByte))
		if err != nil {
			return err
		}
	}
	return nil
}

func (spd *SinglePassDetails) writeOverflowFiles(mp *MultipassPreloader) error {
	sep := string(os.PathSeparator)
	for index, element := range spd.overflows {
		binAddress := spd.firstAddress + int64(index)
		if len(element) > 0 {
			folder, filename, _ := mp.numberedFolders.NumberToFoldersAndFile(binAddress)
			folderPath := mp.creator.folderPath() + sep + "BinOverflows" + sep + folder
			err := os.MkdirAll(folderPath, os.ModePerm)
			if err != nil {
				return err
			}
			file, err := os.Create(folderPath + sep + filename + ".ovf")
			if err != nil {
				return err
			}
			defer file.Close()
			for i := 0; i < len(element); i++ {
				ih := element[i]
				bytes := make([]byte, 8)
				binary.LittleEndian.PutUint64(bytes[0:8], uint64(ih.index))
				_, err = file.Write(bytes)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func BiggestAddressPlusOne(hashDivider uint64) int64 {
	return int64(math.Pow(2, 64)/float64(hashDivider) + 0.5) // Round to nearest!
}

func (mp *MultipassPreloader) CreateInitialFiles() error {
	biggestAddressPlusOne := BiggestAddressPlusOne(mp.creator.params.HashDivider)
	bsFileSize := int64(binStartSize) * biggestAddressPlusOne

	sep := string(os.PathSeparator)
	bsFilePath := mp.creator.folderPath() + sep + "BinStarts.bst"
	var err error
	mp.binStartsFile, err = os.Create(bsFilePath)
	if err != nil {
		return err
	}
	err = mp.binStartsFile.Truncate(int64(bsFileSize))
	if err != nil {
		return err
	}

	ofFolderPath := mp.creator.folderPath() + sep + "BinOverflows"
	err = os.MkdirAll(ofFolderPath, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

func (mp *MultipassPreloader) IndexTheHashes() error {
	err := mp.CreateInitialFiles()
	if err != nil {
		return err
	}

	biggestAddressPlusOne := BiggestAddressPlusOne(mp.creator.params.HashDivider)
	bytesPerBin := int64(binStartSize)
	binsPerPass := mp.bytesPerPass / bytesPerBin
	addressesPerPass := binsPerPass
	passes := 1 + biggestAddressPlusOne/addressesPerPass

	for pass := int64(0); pass < passes; pass++ {
		sline := "\r" + fmt.Sprintf("%s: Pass %d of %d", mp.creator.folderPath(), pass, passes)
		fmt.Print(sline)
		firstAddress := pass * addressesPerPass
		addresses := addressesPerPass
		if firstAddress+addresses > biggestAddressPlusOne {
			addresses = biggestAddressPlusOne - firstAddress
		}
		passDetails := NewSinglePassDetails(firstAddress, addresses)
		err = passDetails.ReadIn(mp)
		if err != nil {
			return err
		}
		err = passDetails.writeIntoBinStartsFile(mp)
		if err != nil {
			return err
		}
		err = passDetails.writeOverflowFiles(mp)
		if err != nil {
			return err
		}
	}
	fmt.Println()
	err = mp.binStartsFile.Close()
	if err != nil {
		return err
	}

	return nil
}
