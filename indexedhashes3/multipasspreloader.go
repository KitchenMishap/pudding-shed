package indexedhashes3

import (
	"fmt"
	"github.com/KitchenMishap/pudding-shed/wordfile"
	"os"
)

type MultipassPreloader struct {
	params          *HashIndexingParams
	folderPath      string
	binStartsFile   *os.File
	binNumsWordFile wordfile.WriterAtWord
	bytesPerPass    int64
	overflowFiles   *overflowFiles
}

func NewMultipassPreloader(params *HashIndexingParams, folderPath string, binNumsWordFile wordfile.WriterAtWord) *MultipassPreloader {
	result := MultipassPreloader{}
	result.params = params
	result.folderPath = folderPath
	result.binStartsFile = nil
	result.binNumsWordFile = binNumsWordFile
	result.bytesPerPass = 1024 * 1024 * 1024 // One Gigabyte (memory)
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
