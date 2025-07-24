package indexedhashes3

import (
	"errors"
	"fmt"
	"github.com/KitchenMishap/pudding-shed/indexedhashes"
	"github.com/KitchenMishap/pudding-shed/wordfile"
	"os"
	"strconv"
)

type MultipassPreloader struct {
	params          *HashIndexingParams
	folderPath      string
	binStartsFile   *os.File
	binNumsWordFile wordfile.WriterAtWord
	bytesPerPass    int64
	overflowFiles   *overflowFiles
}

func NewMultipassPreloader(params *HashIndexingParams, folderPath string, binNumsWordFile wordfile.WriterAtWord, gbMem int) *MultipassPreloader {
	result := MultipassPreloader{}
	result.params = params
	result.folderPath = folderPath
	result.binStartsFile = nil
	result.binNumsWordFile = binNumsWordFile
	result.bytesPerPass = int64(gbMem * 1024 * 1024 * 1024)
	result.overflowFiles = newOverflowFiles(folderPath, params)
	return &result
}

func (mp *MultipassPreloader) IndexTheHashes() error {
	err := mp.createInitialFiles()
	if err != nil {
		return err
	}

	expectedEntriesPerBin := mp.params.HashCountEstimate() / mp.params.NumberOfBins()
	if expectedEntriesPerBin == 0 {
		expectedEntriesPerBin = 1
	}
	binsPerPass := mp.bytesPerPass / (mp.params.BytesPerBinEntry() * expectedEntriesPerBin)
	passes := 1 + mp.params.NumberOfBins()/binsPerPass

	for pass := int64(0); pass < passes; pass++ {
		sline := "\r" + fmt.Sprintf("%s: Pass %d of %d", mp.folderPath, pass, passes)
		fmt.Print(sline)
		firstBinNum := pass * binsPerPass
		bins := binsPerPass
		if firstBinNum+bins > mp.params.NumberOfBins() {
			bins = mp.params.NumberOfBins() - firstBinNum
		}

		passDetails := newSinglePassDetails(firstBinNum, bins, mp.binNumsWordFile)
		err = passDetails.readIn(mp)
		if err != nil {
			return err
		}

		//passDetails.checkThereAreNonEmptyBins()

		err := passDetails.writeFiles(mp)
		if err != nil {
			return err
		}
	}
	fmt.Println()
	err = mp.binNumsWordFile.Close()
	if err != nil {
		return err
	}
	err = mp.binStartsFile.Close()
	if err != nil {
		return err
	}

	return nil
}

func (mp *MultipassPreloader) TestTheHashes() error {
	{
		sep := string(os.PathSeparator)
		binStartsFile, err := os.Open(mp.folderPath + sep + "BinStarts.bes")
		if err != nil {
			return err
		}
		wfc := wordfile.NewConcreteWordFileCreator("BinNums", mp.folderPath, mp.params.bytesRoomForBinNum, false)
		hs, err := newHashStoreObject(mp.params, mp.folderPath, wfc, binStartsFile)
		if err != nil {
			return err
		}

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
				err := mp.testOneHash(hs, hashIndex, hash)
				if err != nil {
					return err
				}
				hashIndex++
			}
			nBytes, _ = hashesFile.Read(chunk)
		}
		return nil
	}
}

func (mp *MultipassPreloader) testOneHash(hs indexedhashes.HashReader, hashIndex int64, hash indexedhashes.Sha256) error {
	foundHash := indexedhashes.Sha256{}
	err := hs.GetHashAtIndex(hashIndex, &foundHash)
	if err != nil {
		return err
	}
	foundIndex, err := hs.IndexOfHash(&hash)
	if err != nil {
		return err
	}
	if foundHash != hash {
		fmt.Println("found hash does not match hash at index " + strconv.FormatInt(hashIndex, 10))
	}
	if foundIndex == -1 {
		return errors.New("hash not found at index " + strconv.FormatInt(hashIndex, 10))
	}
	if foundIndex != hashIndex {
		fmt.Println("hash at index " + strconv.FormatInt(hashIndex, 10) + " found at " + strconv.FormatInt(foundIndex, 10) + " instead")
	}
	return nil
}

func (mp *MultipassPreloader) createInitialFiles() error {
	biggestBinNumPlusOne := mp.params.NumberOfBins()
	binStartsFileSize := biggestBinNumPlusOne * mp.params.EntriesInBinStart() * mp.params.BytesPerBinEntry()

	sep := string(os.PathSeparator)
	bsFilePath := mp.folderPath + sep + "BinStarts.bes"
	var err error
	mp.binStartsFile, err = os.Create(bsFilePath)
	if err != nil {
		return err
	}
	err = mp.binStartsFile.Truncate(binStartsFileSize)
	if err != nil {
		fmt.Println("Couldn't make a ", binStartsFileSize/1024/1024, "MB file")
		return err
	}

	bnFilePath := mp.folderPath + sep + "BinNums.int"
	f, err := os.Create(bnFilePath)
	if err != nil {
		return err
	}
	err = f.Close()
	if err != nil {
		return err
	}

	ofFolderPath := mp.folderPath + sep + "BinOverflows"
	err = os.MkdirAll(ofFolderPath, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}
