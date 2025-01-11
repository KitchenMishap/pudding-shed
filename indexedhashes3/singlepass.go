package indexedhashes3

import (
	"errors"
	"fmt"
	"github.com/KitchenMishap/pudding-shed/testpoints"
	"os"
)

// singlePassDetails holds details of one of the multiple passes
type singlePassDetails struct {
	firstBinNum       int64
	lastBinNumPlusOne int64
	bins              []bin
}

func newSinglePassDetails(firstBinNum int64, binsCount int64) *singlePassDetails {
	result := singlePassDetails{}
	result.firstBinNum = firstBinNum
	result.lastBinNumPlusOne = firstBinNum + binsCount
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
			spd.dealWithOneHash(hashIndex, &hash, mp)
			hashIndex++
		}
		nBytes, _ = hashesFile.Read(chunk)
	}
	return nil
}

func (spd *singlePassDetails) dealWithOneHash(hi int64, hash *[32]byte, mp *MultipassPreloader) {
	// === TestPoint ===
	// TestPoint for when inserting nth hash (but hit for nth block, nth transaction, and nth address)
	if testpoints.TestPointBlockEnable && hi == testpoints.TestPointBlock {
		fmt.Println("TESTPOINT: SinglePassDetails.dealWithOneHash(index = ", testpoints.TestPointBlock, ")")
	}

	hash3 := Hash(*hash)
	abbr := hash3.toAbbreviatedHash()
	bn := abbr.toBinNum(mp.params)
	// This single pass only deals with a certain range of bin numbers
	if int64(bn) < spd.firstBinNum || int64(bn) >= spd.lastBinNumPlusOne {
		return
	}

	th := hash3.toTruncatedHash()
	sn := abbr.toSortNum(mp.params)

	passBinNumber := int64(bn) - spd.firstBinNum
	theBin := spd.bins[passBinNumber]

	// Is it in the bin already?
	if theBin.lookupByHash(&th, sn, mp.params) != -1 {
		return
	}

	theBin.insertBinEntry(sn, hashIndex(hi), &th, mp.params)
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
