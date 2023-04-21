package wordfile

import (
	"encoding/binary"
	"errors"
	"os"
)

type WordFile struct {
	file     *os.File
	wordSize int64
}

func NewWordFile(file *os.File, wordSize int64) *WordFile {
	p := new(WordFile)
	p.file = file
	p.wordSize = wordSize
	return p
}

func (wf *WordFile) ReadWordAt(off int64) (word int64, err error) {
	word = 0
	err = nil
	var intBytes [8]byte
	_, err = wf.file.ReadAt(intBytes[0:wf.wordSize], off*wf.wordSize)
	word = int64(binary.LittleEndian.Uint64(intBytes[0:8]))
	return
}

func (wf *WordFile) WriteWordAt(val int64, off int64) error {
	var intBytes [8]byte
	binary.LittleEndian.PutUint64(intBytes[0:8], uint64(val))
	_, err := wf.file.WriteAt(intBytes[0:wf.wordSize], off*wf.wordSize)
	return err
}

func (wf *WordFile) CountWords() (words int64, err error) {
	fi, err := wf.file.Stat()
	if err != nil {
		return 0, err
	}
	filesize := fi.Size()
	if filesize%wf.wordSize != 0 {
		return 0, errors.New("file is not a whole number of words")
	}
	return filesize / wf.wordSize, nil
}
