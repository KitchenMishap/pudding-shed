package memfile

import (
	"errors"
	"github.com/edsrzf/mmap-go"
	"io"
	"os"
)

type MmapReadOnly struct {
	data mmap.MMap
	file *os.File
}

// Check that implements
var _ io.ReaderAt = (*MmapReadOnly)(nil)
var _ io.Closer = (*MmapReadOnly)(nil)
var _ LookupFile = (*MmapReadOnly)(nil)

func NewMmapReadOnly(f *os.File) (*MmapReadOnly, error) {
	// Map as read-only. 0 is the flag for default behaviour.
	m, err := mmap.Map(f, mmap.RDONLY, 0)
	if err != nil {
		return nil, err
	}
	return &MmapReadOnly{
		data: m,
		file: f,
	}, nil
}

func (m *MmapReadOnly) ReadAt(p []byte, off int64) (n int, err error) {
	if off >= int64(len(m.data)) {
		return 0, io.EOF
	}
	n = copy(p, m.data[off:])
	if n < len(p) {
		return n, io.EOF
	}
	return n, nil
}

func (m *MmapReadOnly) ReadAll() ([]byte, error) {
	return m.data, nil
}

func (m *MmapReadOnly) Sync() error {
	return m.data.Flush()
}

func (m *MmapReadOnly) Close() error {
	// 1. Unmap the memory first
	err := m.data.Unmap()
	// 2. Close thee underlying file
	if fErr := m.file.Close(); fErr != nil {
		return fErr
	}
	return err
}

func (m *MmapReadOnly) Size() int64 { return int64(len(m.data)) }

func (m *MmapReadOnly) WriteAt(p []byte, off int64) (n int, err error) {
	return 0, errors.New("WriteAt() not implemented in MmapReadOnly")
}
