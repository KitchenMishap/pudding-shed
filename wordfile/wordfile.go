package wordfile

import (
	"encoding/binary"
	"github.com/KitchenMishap/pudding-shed/memfile"
	"log"
)

type WordFile struct {
	file      memfile.AppendableLookupFile
	wordSize  int64
	wordCount int64
}

func NewWordFile(file memfile.AppendableLookupFile, wordSize int64, wordCount int64) *WordFile {
	p := new(WordFile)
	p.file = file
	p.wordSize = wordSize
	p.wordCount = wordCount
	return p
}

func (wf *WordFile) ReadWordAt(off int64) (int64, error) {
	var intBytes [8]byte
	_, err := wf.file.ReadAt(intBytes[0:wf.wordSize], off*wf.wordSize)
	if err != nil {
		return -1, err
	}
	word := int64(binary.LittleEndian.Uint64(intBytes[0:8]))
	return word, nil
}

func (wf *WordFile) WriteWordAt(val int64, off int64) error {
	var intBytes [8]byte
	binary.LittleEndian.PutUint64(intBytes[0:8], uint64(val))
	_, err := wf.file.WriteAt(intBytes[0:wf.wordSize], off*wf.wordSize)
	if err != nil {
		return err
	}
	if off+1 > wf.wordCount {
		wf.wordCount = off + 1
	}
	return nil
}

func (wf *WordFile) CountWords() (words int64, err error) {
	return wf.wordCount, nil
}

func (wf *WordFile) Close() error {
	err := wf.file.Close()
	if err != nil {
		log.Println(err)
		log.Println("Close(): Could not call Close()")
	}
	return err
}

func (wf *WordFile) Sync() error {
	return wf.file.Sync()
}

func (wf *WordFile) WordSize() int64 {
	return wf.wordSize
}
