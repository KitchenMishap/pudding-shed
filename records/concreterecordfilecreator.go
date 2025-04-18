package records

import (
	"errors"
	"github.com/KitchenMishap/pudding-shed/memfile"
	"log"
	"os"
	"path/filepath"
)

type ConcreteRecordFileCreator struct {
	name   string
	folder string
	rd     RecordDescriptor
}

func NewConcreteRecordFileCreator(name string, folder string, rd RecordDescriptor) *ConcreteRecordFileCreator {
	if rd.RecordSize() == 0 {
		panic("must be at least one byte")
	}
	result := ConcreteRecordFileCreator{}
	result.name = name
	result.folder = folder
	result.rd = rd
	return &result
}

func (rfc *ConcreteRecordFileCreator) RecordFileExists() bool {
	fullName := filepath.Join(rfc.folder, rfc.name+".rec")
	file, err := os.Open(fullName)
	defer file.Close()
	if err != nil {
		// Doesn't exist.
		return false
	}
	return true
}

func (rfc *ConcreteRecordFileCreator) CreateRecordFile() error {
	// First create folder if necessary
	if rfc.folder != "" {
		err := os.MkdirAll(rfc.folder, os.ModePerm)
		if err != nil {
			log.Println(err)
			log.Println("CreateRecordFile(): Could not call MkdirAll()")
			return err
		}
	}

	fullName := filepath.Join(rfc.folder, rfc.name+".rec")
	file, err := os.Create(fullName)
	if err != nil {
		log.Println(err)
		log.Println("CreateWordFile(): Could not call os.Create()")
		return err
	}
	defer file.Close()

	return nil
}

func (rfc *ConcreteRecordFileCreator) OpenRecordFile() (ReadWriteAtRecordCounter, error) {
	fullName := filepath.Join(rfc.folder, rfc.name+".rec")
	file, err := os.OpenFile(fullName, os.O_RDWR, 0)
	if err != nil {
		return nil, err
	}

	recordCount, err := rfc.countRecords(file)
	if err != nil {
		return nil, err
	}

	var result ReadWriteAtRecordCounter
	appendOptimizedFile, err := memfile.NewAppendOptimizedFile(file)
	if err != nil {
		return nil, err
	}
	result = NewRecordFile(appendOptimizedFile, rfc.rd, recordCount)

	return result, nil
}
func (rfc *ConcreteRecordFileCreator) OpenRecordFileReadOnly() (ReadAtRecordCounter, error) {
	fullName := filepath.Join(rfc.folder, rfc.name+".rec")
	file, err := os.OpenFile(fullName, os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}

	wordCount, err := rfc.countRecords(file)
	if err != nil {
		return nil, err
	}

	fileWithSize := memfile.NewFileWithSize(file)
	var result ReadAtRecordCounter
	result = NewRecordFile(fileWithSize, rfc.rd, wordCount)

	return result, nil
}
func (rfc *ConcreteRecordFileCreator) countRecords(file *os.File) (int64, error) {
	fi, err := file.Stat()
	if err != nil {
		log.Println(err)
		log.Println("countRecords(): Couldn't call file.Stat()")
		return 0, err
	}
	filesize := fi.Size()
	if filesize%int64(rfc.rd.RecordSize()) != 0 {
		log.Println("countWords(): File is not a whole number of records")
		return 0, errors.New("file is not a whole number of records")
	}
	return filesize / int64(rfc.rd.RecordSize()), nil
}
func (rfc *ConcreteRecordFileCreator) CreateRecordFileFilledZeros(count int64) error {
	// First create folder if necessary
	if rfc.folder != "" {
		err := os.MkdirAll(rfc.folder, os.ModePerm)
		if err != nil {
			return err
		}
	}

	fullName := filepath.Join(rfc.folder, rfc.name+".rec")
	file, err := os.Create(fullName)
	if err != nil {
		return err
	}
	err = file.Truncate(count * int64(rfc.rd.RecordSize()))
	if err != nil {
		return err
	}

	defer file.Close()

	return nil
}
