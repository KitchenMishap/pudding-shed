package indexedhashes3

import "os"

type MultipassPreloader struct {
	params        *HashIndexingParams
	folderPath    string
	binStartsFile *os.File
}

func (mp *MultipassPreloader) CreateInitialFiles() error {
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
