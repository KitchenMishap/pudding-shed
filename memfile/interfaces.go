package memfile

import (
	"io"
	"os"
)

// LookupFile is a file-like array-like object.
// The file has array elements each a fixed number of bytes.
// Elements are written or read one at a time.
// By referring to io interfaces, we can arrange for os.File to be a naive non-optimized implementation.
type LookupFile interface {
	io.ReaderAt
	io.WriterAt
	io.Closer
	Sync() error
}

// SparseLookupFile is a LookupFile that may be optimized for the following scenario.
// The file is fixed size, and starts life full of zeros.
// Elements are written or read one at a time.
type SparseLookupFile interface {
	LookupFile
}

type LookupFileWithSize interface {
	LookupFile
	Stat() (os.FileInfo, error)
}

// AppendableLookupFile is a LookupFile that may be optimized for the following scenario.
// The file typically grows by aggressive appending, one element at a time.
type AppendableLookupFile interface {
	LookupFileWithSize
}
