package indexedhashes

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/KitchenMishap/pudding-shed/numberedfolders"
	"math"
	"os"
	"sync"
)

// The file formats for UniformHashStore2 are as follows.
// All file contents are raw binary.

// The file Hashes.hsh contains the raw binary hashes, 32 bytes each. This file is an input to MultipassPreloader.
// The hashes in this file (and elsewhere) are LittleEndian: LSBs first.
// For block hashes, which due to proof of work start with zeroes, these zeroes are at the END of each 32 bytes.
// This means that the first 8 bytes can be treated as pseudo-random, and can be used to generate
// addresses so that each (hash, index) pair can go in a bin identified by the address.
// Multiple different hashes, so long as they correspond to the same address, therefore end up in the same bin.

// The file BinStarts.bst is a fixed size large structured file split sequentially into bins of a fixed size.
// The fixed bin size is intended to correspond to the file system's file allocation size (usually 4096 bytes).
// The number of bins is fixed and optimized based on a rough estimate (not a maximum) of the number of hashes expected.
// A bin starts with an 8 byte count of the number of hashes in the bin. Despite the 8 bytes, this number
// is typically less than 100, but CAN be more than 102.
// Then 8 bytes reserved (zero). There are then 102 slots of 40 byte entries.
// Unused entries at the end are zeroed. So a bin is always 8(count) + 8(reserved) + 102*40(entries) = 4096 bytes.
// A 40 byte entry consists of an 8 byte index for the hash (ie the order the hash appears in Hashes.hsh), followed
// by the 32 byte hash.
// If the count specified in the first 8 bytes is more than 102, then there will be files for overflow in BinOverflows/.

// BinOverflows/ files are small files in a numbered folder structure based on the bin address.
// Each file is a multiple of 40 byte entries, the first entry in the file being slot 102, and with the
// same (index, hash) format as above - with no count or reserved or zero padding (the file can be any multiple of 40)

// MultipassPreloader is a system that preloads a UniformHashStore2 fileset.
// To minimize memory usage, it reads the hashes file multiple times, each time
// concentrating on a different set of address bins (the set is known as a dumpster)
type MultipassPreloader struct {
	creator         *UniformHashStoreCreator
	bytesPerPass    int64 // A parameter of the method, but does not affect the files which are output
	binStartsFile   *os.File
	numberedFolders numberedfolders.NumberedFolders
}

func NewMultipassPreloader(creator *UniformHashStoreCreator, bytesPerPass int64, fileDigits int, folderDigits int) *MultipassPreloader {
	result := MultipassPreloader{}
	result.creator = creator
	result.bytesPerPass = bytesPerPass
	result.binStartsFile = nil
	result.numberedFolders = numberedfolders.NewNumberedFolders(fileDigits, folderDigits)
	return &result
}

type indexHash struct {
	index uint64
	hash  Sha256
}

type binStart struct {
	entryCount  int64
	reserved    int64
	indexHashes [102]indexHash
}

func (bs *binStart) ToBytes(bytes *[4096]byte) {
	binary.LittleEndian.PutUint64(bytes[0:8], uint64(bs.entryCount))
	binary.LittleEndian.PutUint64(bytes[8:16], uint64(bs.reserved))
	for i := 0; i < 102; i++ {
		binary.LittleEndian.PutUint64(bytes[16+i*40:16+i*40+8], bs.indexHashes[i].index)
		copy(bytes[16+8+i*40:16+8+i*40+32], bs.indexHashes[i].hash[:])
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
		}
		nBytes, _ = hashesFile.Read(chunk)
	}
	return nil
}

func (spd *SinglePassDetails) dealWithOneHash(ih *indexHash, mp *MultipassPreloader) {
	LSWord := binary.LittleEndian.Uint64(ih.hash[0:8])
	address := int64(LSWord / mp.creator.hashDivider)
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
	chunk := [4096]byte{}
	for index, element := range spd.binStarts {
		binAddress := spd.firstAddress + int64(index)
		fileByte := 4096 * binAddress
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
			file, err := os.Create(folderPath + sep + filename)
			if err != nil {
				return err
			}
			defer file.Close()
			for i := 0; i < len(element); i++ {
				ih := element[i]
				bytes := make([]byte, 40)
				binary.LittleEndian.PutUint64(bytes[0:8], ih.index)
				copy(bytes[8:8+32], ih.hash[:])
				_, err = file.Write(bytes)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (mp *MultipassPreloader) CreateInitialFiles() error {
	biggestAddressPlusOne := uint64(math.Pow(2, 64) / float64(mp.creator.hashDivider))
	bsFileSize := uint64(4096) * biggestAddressPlusOne

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
	fmt.Println("Creating initial files")
	err := mp.CreateInitialFiles()
	if err != nil {
		return err
	}

	biggestAddressPlusOne := int64(math.Pow(2, 64) / float64(mp.creator.hashDivider))
	bytesPerBin := int64(4096)
	binsPerPass := mp.bytesPerPass / bytesPerBin
	addressesPerPass := binsPerPass
	passes := 1 + biggestAddressPlusOne/addressesPerPass

	for pass := int64(0); pass < passes; pass++ {
		fmt.Println("Pass ", pass, " of ", passes)
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
		fmt.Println("Writing to BinStarts file")
		err = passDetails.writeIntoBinStartsFile(mp)
		if err != nil {
			return err
		}
		fmt.Println("Writing to BinOverflows file")
		err = passDetails.writeOverflowFiles(mp)
		if err != nil {
			return err
		}
	}
	fmt.Println("Done")

	return nil
}
