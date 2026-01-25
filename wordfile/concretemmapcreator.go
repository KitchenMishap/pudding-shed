package wordfile

import (
	"errors"
	"github.com/KitchenMishap/pudding-shed/memfile"
	"log"
	"os"
	"path/filepath"
)

type ConcreteMmapWordFileCreator struct {
	name     string
	folder   string
	wordSize int64
}

// Check that implements
var _ WordFileCreator = (*ConcreteMmapWordFileCreator)(nil)

func NewConcreteMmapWordFileCreator(name string, folder string, wordSize int64) *ConcreteMmapWordFileCreator {
	if wordSize == 0 {
		panic("must be at least one byte")
	}
	result := ConcreteMmapWordFileCreator{}
	result.name = name
	result.folder = folder
	result.wordSize = wordSize
	return &result
}

func (wfc *ConcreteMmapWordFileCreator) WordFileExists() bool {
	fullName := filepath.Join(wfc.folder, wfc.name+".int")
	file, err := os.Open(fullName)
	defer file.Close()
	if err != nil {
		// Doesn't exist.
		return false
	}
	return true
}

func (wfc *ConcreteMmapWordFileCreator) CreateWordFile() error {
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

func (wfc *ConcreteMmapWordFileCreator) OpenWordFileReadOnly() (ReadAtWordCounter, error) {
	fullName := filepath.Join(wfc.folder, wfc.name+".int")

	// 1. Standard OS Open
	file, err := os.Open(fullName)
	if err != nil {
		return nil, err
	}
	wordCount, err := wfc.countWords(file)
	if err != nil {
		file.Close()
		return nil, err
	}
	m, err := memfile.NewMmapReadOnly(file)
	if err != nil {
		file.Close()
		return nil, err
	}

	return NewWordFileEx(m, wfc.wordSize, wordCount), nil
}

func (wfc *ConcreteMmapWordFileCreator) OpenWordFile() (ReadWriteAtWordCounter, error) {
	return nil, errors.New("OpenWordFile() not implemented. Use OpenWordFileReadOnly()")
}

func (wfc *ConcreteMmapWordFileCreator) CreateWordFileFilledZeros(count int64) error {
	return errors.New("CreateWordFileFilledZeros() not implemented. Read only interface")
}

func (wfc *ConcreteMmapWordFileCreator) countWords(file *os.File) (int64, error) {
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
