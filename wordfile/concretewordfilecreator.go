package wordfile

import (
	"github.com/KitchenMishap/pudding-shed/memfile"
	"log"
	"os"
	"path/filepath"
)

type ConcreteWordFileCreator struct {
	name           string
	folder         string
	wordSize       int64
	appendOptimize bool
}

func NewConcreteWordFileCreator(name string, folder string, wordSize int64, appendOptimize bool) *ConcreteWordFileCreator {
	result := ConcreteWordFileCreator{}
	result.name = name
	result.folder = folder
	result.wordSize = wordSize
	result.appendOptimize = appendOptimize
	return &result
}

func (wfc *ConcreteWordFileCreator) WordFileExists() bool {
	fullName := filepath.Join(wfc.folder, wfc.name+".int")
	file, err := os.Open(fullName)
	defer file.Close()
	if err != nil {
		// Doesn't exist.
		return false
	}
	return true
}

func (wfc *ConcreteWordFileCreator) CreateWordFile() error {
	// First create folder if necessary
	if wfc.folder != "" {
		err := os.MkdirAll(wfc.folder, os.ModePerm)
		if err != nil {
			log.Println(err)
			log.Println("CreateWordFile(): Could not call MkdirAll()")
			return err
		}
	}

	fullName := filepath.Join(wfc.folder, wfc.name+".int")
	file, err := os.Create(fullName)
	if err != nil {
		log.Println(err)
		log.Println("CreateWordFile(): Could not call os.Create()")
		return err
	}
	defer file.Close()

	return nil
}

func (wfc *ConcreteWordFileCreator) OpenWordFile() (ReadWriteAtWordCounter, error) {
	fullName := filepath.Join(wfc.folder, wfc.name+".int")
	file, err := os.OpenFile(fullName, os.O_RDWR, 0)
	if err != nil {
		return nil, err
	}

	var result ReadWriteAtWordCounter
	if wfc.appendOptimize {
		appendOptimizedFile := memfile.NewAppendOptimizedFile(file)
		result = NewWordFile(appendOptimizedFile, wfc.wordSize)
	} else {
		result = NewWordFile(file, wfc.wordSize)
	}
	return result, nil
}
func (wfc *ConcreteWordFileCreator) OpenWordFileReadOnly() (ReadAtWordCounter, error) {
	fullName := filepath.Join(wfc.folder, wfc.name+".int")
	file, err := os.OpenFile(fullName, os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}

	result := NewWordFile(file, wfc.wordSize)
	return result, nil
}
