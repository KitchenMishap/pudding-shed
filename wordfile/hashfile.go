package wordfile

import (
	"github.com/KitchenMishap/pudding-shed/memfile"
)

type HashFile struct {
	file      memfile.AppendableLookupFile
	hashCount int64

	scratch []byte
}

func NewHashFile(file memfile.AppendableLookupFile, hashCount int64) *HashFile {
	p := new(HashFile)
	p.file = file
	p.hashCount = hashCount
	p.scratch = make([]byte, 32)
	return p
}

func (wf *HashFile) ReadHashAt(off int64) ([32]byte, error) {
	// Using scracth avoids MASSES of tiny heap allocations (slice headers)
	_, err := wf.file.ReadAt(wf.scratch, off*32)
	var hashBytes [32]byte
	if err != nil {
		return hashBytes, err
	}
	copy(hashBytes[:], wf.scratch)
	return hashBytes, nil
}

func (wf *HashFile) WriteHashAt(val [32]byte, off int64) error {
	// Using scratch avoids MASSES of tiny heap allocations (slice headers)
	copy(wf.scratch[:], val[:])
	_, err := wf.file.WriteAt(wf.scratch, off*32)
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
