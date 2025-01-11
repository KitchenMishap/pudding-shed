package indexedhashes3

import (
	"fmt"
	"os"
)

type MultipassPreloader struct {
	params        *HashIndexingParams
	folderPath    string
	binStartsFile *os.File
	bytesPerPass  int64
	overflowFiles *overflowFiles
}

func (mp *MultipassPreloader) IndexTheHashes() error {
	err := mp.createInitialFiles()
	if err != nil {
		return err
	}

	expectedEntriesPerBin := mp.params.HashCountEstimate() / mp.params.NumberOfBins()
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

		passDetails := newSinglePassDetails(firstBinNum, bins)
		err = passDetails.readIn(mp)
		if err != nil {
			return err
		}

		err := passDetails.writeFiles(mp)
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
		return err
	}

	ofFolderPath := mp.folderPath + sep + "BinOverflows"
	err = os.MkdirAll(ofFolderPath, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}
