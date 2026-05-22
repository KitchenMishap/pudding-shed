package indexedhashes3

import (
	"fmt"
	"os"

	"github.com/KitchenMishap/pudding-shed/wordfile"
)

type AdaptivePassPreloader struct {
	params          *HashIndexingParams
	folderPath      string
	binStartsFile   *os.File
	binNumsWordFile wordfile.WriterAtWord
	gbPerPass       int64
	overflowFiles   *overflowFiles
}

func NewAdaptivePassPreloader(params *HashIndexingParams, folderPath string,
	binNumsWordFile wordfile.WriterAtWord, gbMem int64) *AdaptivePassPreloader {
	result := AdaptivePassPreloader{}
	result.params = params
	result.folderPath = folderPath
	result.binStartsFile = nil
	result.binNumsWordFile = binNumsWordFile
	result.gbPerPass = gbMem
	result.overflowFiles = newOverflowFiles(folderPath, params)
	return &result
}

func (ap *AdaptivePassPreloader) IndexTheHashes(threads int) error {
	err := ap.createInitialFiles()
	if err != nil {
		return err
	}

	sline := fmt.Sprintf("%s: Running adaptive passes...\n", ap.folderPath)
	fmt.Print(sline)

	sep := string(os.PathSeparator)
	hashesFilepath := ap.folderPath + sep + "Hashes.hsh"

	var counts []binWorkInfo
	counts, err = stageOneCountBinWork(hashesFilepath, ap.params)

	err = stageTwoStageThreeHandleHashBins(hashesFilepath, ap.binNumsWordFile, ap.binStartsFile,
		ap.overflowFiles, ap.params, counts, ap.gbPerPass, threads)
	if err != nil {
		return err
	}
	err = ap.binNumsWordFile.Close()
	if err != nil {
		return err
	}
	fmt.Println("...Done passes")
	return nil
}

/*
func (mp *MultipassPreloader) TestTheHashes() error {
	{
		sep := string(os.PathSeparator)
		binStartsFile, err := os.Open(mp.folderPath + sep + "BinStarts.bes")
		if err != nil {
			return err
		}
		wfc := wordfile.NewConcreteWordFileCreator("BinNums", mp.folderPath, mp.params.bytesRoomForBinNum, false, false)
		hs, err := newHashStoreObject(mp.params, mp.folderPath, wfc, binStartsFile, false)
		if err != nil {
			return err
		}

		hashesFilepath := mp.folderPath + sep + "Hashes.hsh"
		hashesFile, err := os.Open(hashesFilepath)
		if err != nil {
			return err
		}
		defer func() { _ = hashesFile.Close() }()

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
}*/

func (ap *AdaptivePassPreloader) createInitialFiles() error {
	biggestBinNumPlusOne := ap.params.NumberOfBins()
	binStartsFileSize := biggestBinNumPlusOne * ap.params.EntriesInBinStart() * ap.params.BytesPerBinEntry()

	sep := string(os.PathSeparator)
	bsFilePath := ap.folderPath + sep + "BinStarts.bes"
	var err error
	ap.binStartsFile, err = os.Create(bsFilePath)
	if err != nil {
		return err
	}
	err = ap.binStartsFile.Truncate(binStartsFileSize)
	if err != nil {
		fmt.Println("Couldn't make a ", binStartsFileSize/1024/1024, "MB file")
		return err
	}

	bnFilePath := ap.folderPath + sep + "BinNums.int"
	f, err := os.Create(bnFilePath)
	if err != nil {
		return err
	}
	err = f.Close()
	if err != nil {
		return err
	}

	ofFolderPath := ap.folderPath + sep + "BinOverflows"
	err = os.MkdirAll(ofFolderPath, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}
