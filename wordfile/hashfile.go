package wordfile

import (
	"github.com/KitchenMishap/pudding-shed/memfile"
)

type HashFile struct {
	file      memfile.AppendableLookupFile
	hashCount int64
}

func NewHashFile(file memfile.AppendableLookupFile, hashCount int64) *HashFile {
	p := new(HashFile)
	p.file = file
	p.hashCount = hashCount
	return p
}

func (wf *HashFile) ReadHashAt(off int64) ([32]byte, error) {
	var hashBytes [32]byte
	_, err := wf.file.ReadAt(hashBytes[0:32], off*32)
	if err != nil {
		return hashBytes, err
	}
	return hashBytes, nil
}

func (wf *HashFile) WriteHashAt(val [32]byte, off int64) error {
	_, err := wf.file.WriteAt(val[0:32], off*32)
	if err != nil {
		return err
	}
	if off+1 > wf.hashCount {
		wf.hashCount = off + 1
	}
	return err
}

func (wf *HashFile) CountHashes() (hashes int64) {
	return wf.hashCount
}

func (wf *HashFile) Close() error {
	err := wf.file.Close()
	if err != nil {
		return err
	}
	return err
}

func (wf *HashFile) Sync() error {
	return wf.file.Sync()
}

func (wf *HashFile) AppendHash(val [32]byte) (int64, error) {
	off := wf.hashCount
	err := wf.WriteHashAt(val, off)
	if err != nil {
		return -1, err
	}
	return off, nil
}
