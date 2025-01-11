package indexedhashes3

import (
	"errors"
	"fmt"
	"github.com/KitchenMishap/pudding-shed/testpoints"
	"github.com/KitchenMishap/pudding-shed/wordfile"
	"os"
)

// singlePassDetails holds details of one of the multiple passes
type singlePassDetails struct {
	firstBinNum       int64
	lastBinNumPlusOne int64
	binNumsWordFile   wordfile.WriterAtWord
	bins              []bin
}

func newSinglePassDetails(firstBinNum int64, binsCount int64, binNumsWordFile wordfile.WriterAtWord) *singlePassDetails {
	result := singlePassDetails{}
	result.firstBinNum = firstBinNum
	result.lastBinNumPlusOne = firstBinNum + binsCount
	if firstBinNum == 0 {
		result.binNumsWordFile = binNumsWordFile
	}
	result.bins = make([]bin, binsCount)
	for i := int64(0); i < binsCount; i++ {
		result.bins[i] = newEmptyBin()
	}
	return &result
}

func (spd *singlePassDetails) readIn(mp *MultipassPreloader) error {
	sep := string(os.PathSeparator)
	hashesFilepath := mp.folderPath + sep + "Hashes.hsh"
	hashesFile, err := os.Open(hashesFilepath)
	if err != nil {
		return err
	}
	defer hashesFile.Close()

	hashIndex := int64(0)
	hash := [32]byte{}
	chunk := make([]byte, 4096) // We read up to 4096 bytes at a time
	nBytes, _ := hashesFile.Read(chunk)
	for nBytes > 0 {
		if nBytes%32 != 0 {
			return errors.New("invalid hash file length")
		}
		hashCount := nBytes / 32
		for index := 0; index < hashCount; index++ {
			copy(hash[:], chunk[index*32:index*32+32])
			err := spd.dealWithOneHash(hashIndex, &hash, mp)
			if err != nil {
				return err
			}
			hashIndex++
		}
		nBytes, _ = hashesFile.Read(chunk)
	}
	return nil
}

func (spd *singlePassDetails) dealWithOneHash(hi int64, hash *[32]byte, mp *MultipassPreloader) error {
	// === TestPoint ===
	// TestPoint for when inserting nth hash (but hit for nth block, nth transaction, and nth address)
	if testpoints.TestPointBlockEnable && hi == testpoints.TestPointBlock {
		fmt.Println("TESTPOINT: SinglePassDetails.dealWithOneHash(index = ", testpoints.TestPointBlock, ")")
	}

	hash3 := Hash(*hash)
	abbr := hash3.toAbbreviatedHash()
	bn := abbr.toBinNum(mp.params)

	if spd.firstBinNum == 0 {
		// First pass, store the binNum in a wordfile
		err := spd.binNumsWordFile.WriteWordAt(int64(bn), hi)
		if err != nil {
			return err
		}
	}

	// This single pass only deals with a certain range of bin numbers
	if int64(bn) < spd.firstBinNum || int64(bn) >= spd.lastBinNumPlusOne {
		return nil
	}

	th := hash3.toTruncatedHash()
	sn := abbr.toSortNum(mp.params)

	passBinNumber := int64(bn) - spd.firstBinNum
	theBin := spd.bins[passBinNumber]

	// Is it in the bin already?
	if theBin.lookupByHash(&th, sn, mp.params) != -1 {
		return nil
	}

	theBin.insertBinEntry(sn, hashIndex(hi), &th, mp.params)
	return nil
}

func (spd *singlePassDetails) writeFiles(mp *MultipassPreloader) error {
	for index, element := range spd.bins {
		bn := spd.firstBinNum + int64(index)
		err := saveBinToFiles(binNum(bn), element, mp.binStartsFile, mp.overflowFiles, mp.params)
		if err != nil {
			return err
		}
	}
	return nil
}
