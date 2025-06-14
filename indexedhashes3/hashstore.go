package indexedhashes3

import (
	"github.com/KitchenMishap/pudding-shed/indexedhashes"
	"github.com/KitchenMishap/pudding-shed/wordfile"
	"os"
)

type hashStore struct {
	params          *HashIndexingParams
	overflowFiles   *overflowFiles
	binNumFileWrite wordfile.ReadWriteAtWordCounter
	binNumFileRead  wordfile.ReadAtWordCounter
	binStartsFile   *os.File
}

// Compiler check that implements
var _ indexedhashes.HashReader = (*hashStore)(nil)
var _ indexedhashes.HashReadWriter = (*hashStore)(nil)

func newHashStoreObject(params *HashIndexingParams, folderPath string,
	binNumWordFileCreator wordfile.WordFileCreator,
	binStartsFile *os.File) (*hashStore, error) {
	result := hashStore{}
	result.params = params
	result.overflowFiles = newOverflowFiles(folderPath, params)
	var err error
	result.binNumFileWrite, err = binNumWordFileCreator.OpenWordFile()
	result.binNumFileRead = result.binNumFileWrite
	if err != nil {
		return nil, err
	}
	result.binStartsFile = binStartsFile
	return &result, nil
}

func newHashStoreObjectReadOnly(params *HashIndexingParams, folderPath string,
	binNumWordFileCreator wordfile.WordFileCreator,
	binStartsFile *os.File) (*hashStore, error) {
	result := hashStore{}
	result.params = params
	result.overflowFiles = newOverflowFiles(folderPath, params)
	var err error
	result.binNumFileWrite = nil
	result.binNumFileRead, err = binNumWordFileCreator.OpenWordFileReadOnly()
	if err != nil {
		return nil, err
	}
	result.binStartsFile = binStartsFile
	return &result, nil
}

func (h hashStore) AppendHash(hash *indexedhashes.Sha256) (int64, error) {
	hash3 := Hash(*hash)

	// Split into truncated hash, bin number, and sort number
	trunc := hash3.toTruncatedHash()
	abbr := hash3.toAbbreviatedHash()
	bn := abbr.toBinNum(h.params)
	sn := abbr.toSortNum(h.params)

	// Store the binNumber to file
	hashesSoFar, err := h.binNumFileRead.CountWords()
	if err != nil {
		return -1, err
	}
	err = h.binNumFileWrite.WriteWordAt(int64(bn), hashesSoFar)
	if err != nil {
		return -1, err
	}

	theBin, err := loadBinFromFiles(bn, h.binStartsFile, h.overflowFiles, h.params)
	if err != nil {
		return -1, err
	}

	theBin.insertBinEntry(sn, hashIndex(hashesSoFar), trunc, h.params)
	err = saveBinToFiles(bn, theBin, h.binStartsFile, h.overflowFiles, h.params)
	if err != nil {
		return -1, err
	}

	return hashesSoFar, nil
}

func (h hashStore) IndexOfHash(hash *indexedhashes.Sha256) (int64, error) {
	hash3 := Hash(*hash)

	// Split into truncated hash, bin number, and sort number
	trunc := hash3.toTruncatedHash()
	abbr := hash3.toAbbreviatedHash()
	bn := abbr.toBinNum(h.params)
	sn := abbr.toSortNum(h.params)

	theBin, err := loadBinFromFiles(bn, h.binStartsFile, h.overflowFiles, h.params)
	if err != nil {
		return -1, err
	}

	index := theBin.lookupByHash(trunc, sn, h.params)
	return int64(index), nil
}

func (h hashStore) GetHashAtIndex(index int64, hash *indexedhashes.Sha256) error {
	bn, err := h.binNumFileRead.ReadWordAt(index)
	if err != nil {
		return err
	}

	theBin, err := loadBinFromFiles(binNum(bn), h.binStartsFile, h.overflowFiles, h.params)
	if err != nil {
		return err
	}

	hash3 := theBin.lookupByIndex(hashIndex(index), binNum(bn), h.params)
	*hash = indexedhashes.Sha256(*hash3)
	return nil
}

func (h hashStore) CountHashes() (int64, error) {
	count, err := h.binNumFileRead.CountWords()
	if err != nil {
		return -1, err
	}
	return count, nil
}

func (h hashStore) Close() error {
	// if binNumFileWrite exists, it is equal to binNumFileRead
	err := h.binNumFileRead.Close()
	if err != nil {
		return err
	}
	err = h.binStartsFile.Close()
	if err != nil {
		return err
	}
	return nil
}

func (h hashStore) Sync() error {
	err := h.binNumFileWrite.Sync()
	if err != nil {
		return err
	}
	err = h.binStartsFile.Sync()
	if err != nil {
		return err
	}
	return nil
}
