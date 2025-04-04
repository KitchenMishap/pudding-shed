package wordfile

import (
	"errors"
	"github.com/KitchenMishap/pudding-shed/memfile"
	"log"
	"os"
	"path/filepath"
)

type ConcreteWordFileCreator struct {
	name        string
	folder      string
	wordSize    int64
	memShadowed bool
}

func NewConcreteWordFileCreator(name string, folder string, wordSize int64, memShadowed bool) *ConcreteWordFileCreator {
	if wordSize == 0 {
		panic("must be at least one byte")
	}
	result := ConcreteWordFileCreator{}
	result.name = name
	result.folder = folder
	result.wordSize = wordSize
	result.memShadowed = memShadowed
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

	wordCount, err := wfc.countWords(file)
	if err != nil {
		return nil, err
	}

	var result ReadWriteAtWordCounter
	appendOptimizedFile, err := memfile.NewAppendOptimizedFile(file)
	if err != nil {
		return nil, err
	}
	if wfc.memShadowed {
		result, err = NewMemShadowedWordFile(appendOptimizedFile, wfc.wordSize, wordCount)
		if err != nil {
			return nil, err
		}
	} else {
		result = NewWordFile(appendOptimizedFile, wfc.wordSize, wordCount)
	}
	return result, nil
}
func (wfc *ConcreteWordFileCreator) OpenWordFileReadOnly() (ReadAtWordCounter, error) {
	fullName := filepath.Join(wfc.folder, wfc.name+".int")
	file, err := os.OpenFile(fullName, os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}

	wordCount, err := wfc.countWords(file)
	if err != nil {
		return nil, err
	}

	fileWithSize := memfile.NewFileWithSize(file)
	var result ReadAtWordCounter
	if wfc.memShadowed {
		result, err = NewMemShadowedWordFile(fileWithSize, wfc.wordSize, wordCount)
		if err != nil {
			return nil, err
		}
	} else {
		result = NewWordFile(fileWithSize, wfc.wordSize, wordCount)
	}
	return result, nil
}
func (wfc *ConcreteWordFileCreator) countWords(file *os.File) (int64, error) {
	fi, err := file.Stat()
	if err != nil {
		log.Println(err)
		log.Println("countWords(): Couldn't call file.Stat()")
		return 0, err
	}
	filesize := fi.Size()
	if filesize%wfc.wordSize != 0 {
		log.Println("countWords(): File is not a whole number of words")
		return 0, errors.New("file is not a whole number of words")
	}
	return filesize / wfc.wordSize, nil
}
func (wfc *ConcreteWordFileCreator) CreateWordFileFilledZeros(count int64) error {
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
	err = file.Truncate(count * wfc.wordSize)
	if err != nil {
		return err
	}

	defer file.Close()

	return nil
}
