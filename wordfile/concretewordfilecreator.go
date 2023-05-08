package wordfile

import (
	"os"
	"path/filepath"
)

type ConcreteWordFileCreator struct {
	name     string
	folder   string
	wordSize int64
}

func NewConcreteWordFileCreator(name string, folder string, wordSize int64) ConcreteWordFileCreator {
	result := ConcreteWordFileCreator{}
	result.name = name
	result.folder = folder
	result.wordSize = wordSize
	return result
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
			return err
		}
	}

	fullName := filepath.Join(wfc.folder, wfc.name+".int")
	file, err := os.Create(fullName)
	if err != nil {
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

	result := NewWordFile(file, wfc.wordSize)
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
