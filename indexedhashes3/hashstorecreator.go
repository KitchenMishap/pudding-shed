package indexedhashes3

import (
	"encoding/json"
	"github.com/KitchenMishap/pudding-shed/indexedhashes"
	"github.com/KitchenMishap/pudding-shed/wordfile"
	"io"
	"os"
)

type ConcreteHashStoreCreator struct {
	params                *HashIndexingParams
	folderPath            string
	binNumWordFileCreator wordfile.WordFileCreator
}

// Check that implements
var _ indexedhashes.HashStoreCreator = (*ConcreteHashStoreCreator)(nil)

func NewHashStoreCreatorAndPreloader(folder string, name string,
	params *HashIndexingParams) (indexedhashes.HashStoreCreator, *MultipassPreloader, error) {
	sep := string(os.PathSeparator)
	hashStoreFolderPath := folder + sep + name

	fileParams, _ := os.Create(hashStoreFolderPath + sep + "HashIndexingParams.json")
	defer fileParams.Close()
	byts, _ := json.Marshal(params)
	_, _ = fileParams.Write(byts)

	creator := NewHashStoreCreatorPrivate(params, hashStoreFolderPath)
	filename := folder + sep + name + sep + "BinNums.int"
	file, err := os.Create(filename)
	if err != nil {
		return nil, nil, err
	}
	wordFile := wordfile.NewWordFile(file, params.BytesRoomForBinNum(), 0)
	preloader := NewMultipassPreloader(params, hashStoreFolderPath, wordFile)
	return creator, preloader, nil
}

func NewHashStoreCreatorFromFile(
	folder string, name string) (indexedhashes.HashStoreCreator, error) {
	sep := string(os.PathSeparator)
	paramsFilePath := folder + sep + name + sep + "HashIndexingParams.json"
	paramsFile, err := os.Open(paramsFilePath)
	if err != nil {
		return nil, err
	}
	defer paramsFile.Close()
	bytes, err := io.ReadAll(paramsFile)
	if err != nil {
		return nil, err
	}
	params := HashIndexingParams{}
	err = json.Unmarshal(bytes, &params)
	if err != nil {
		return nil, err
	}
	params.calculateDerivedValues()

	creator := NewHashStoreCreatorPrivate(&params, folder+sep+name)
	return creator, nil
}

func NewHashStoreCreatorPrivate(params *HashIndexingParams, folderPath string) *ConcreteHashStoreCreator {
	result := ConcreteHashStoreCreator{}
	result.params = params
	result.folderPath = folderPath
	result.binNumWordFileCreator = wordfile.NewConcreteWordFileCreator(
		"BinNums", folderPath, params.BytesRoomForBinNum(), false)
	return &result
}

func (c *ConcreteHashStoreCreator) HashStoreExists() bool {
	// Hash store exists if binNumsFile exists
	return c.binNumWordFileCreator.WordFileExists()
}

func (c *ConcreteHashStoreCreator) CreateHashStore() error {
	// Create the parameters json file

	// Create the binNumsFile
	err := c.binNumWordFileCreator.CreateWordFile()
	if err != nil {
		return err
	}
	// Create the BinStarts file and truncate to the required size
	sep := string(os.PathSeparator)
	binStartsFilename := c.folderPath + sep + "BinStarts.bes"
	binStartsFile, err := os.Create(binStartsFilename)
	if err != nil {
		return err
	}
	err = binStartsFile.Truncate(c.params.BytesPerBinEntry() * c.params.EntriesInBinStart())
	if err != nil {
		return err
	}
	return nil
}

func (c *ConcreteHashStoreCreator) OpenHashStore() (indexedhashes.HashReadWriter, error) {
	sep := string(os.PathSeparator)
	binStartsFilename := c.folderPath + sep + "BinStarts.bes"
	binStartsFile, err := os.OpenFile(binStartsFilename, os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	hashStore, err := newHashStoreObject(c.params, c.folderPath, c.binNumWordFileCreator, binStartsFile)
	if err != nil {
		return nil, err
	}
	return hashStore, nil
}

func (c *ConcreteHashStoreCreator) OpenHashStoreReadOnly() (indexedhashes.HashReader, error) {
	sep := string(os.PathSeparator)
	binStartsFilename := c.folderPath + sep + "BinStarts.bes"
	binStartsFile, err := os.OpenFile(binStartsFilename, os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}
	hashStore, err := newHashStoreObjectReadOnly(c.params, c.folderPath, c.binNumWordFileCreator, binStartsFile)
	if err != nil {
		return nil, err
	}
	return hashStore, nil
}
