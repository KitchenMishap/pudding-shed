package memfile

import (
	"errors"
	"os"
)

type fixedSizeMemFile struct {
	filename string
	perm     os.FileMode
	buffer   []byte
}

// Check that implements
var _ SparseLookupFile = (*fixedSizeMemFile)(nil)

// NewFixedSizeMemFile Opens the file and keeps a copy in memory
func NewFixedSizeMemFile(filename string, perm os.FileMode) (SparseLookupFile, error) {
	buffer, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	result := fixedSizeMemFile{filename: filename, perm: perm, buffer: buffer}
	return result, nil
}

func (f fixedSizeMemFile) ReadAt(p []byte, off int64) (n int, err error) {
	copied := copy(p, f.buffer[off:off+int64(len(p))])
	if copied == len(p) {
		return copied, nil
	}
	return copied, errors.New("memfile.fixedSizeMemFile.ReadAt(): Couldn't read requested bytes from mem")
}

func (f fixedSizeMemFile) WriteAt(p []byte, off int64) (n int, err error) {
	copied := copy(f.buffer[off:off+int64(len(p))], p)
	if copied == len(p) {
		return copied, nil
	}
	return copied, errors.New("memfile.fixedSizeMemFile.WriteAt(): Couldn't write requested bytes to mem")
}

func (f fixedSizeMemFile) Close() error {
	return os.WriteFile(f.filename, f.buffer, f.perm)
}
