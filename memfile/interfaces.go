package memfile

import "io"

// SparseLookupFile is a file-like array-like object that may be optimized for the following scenario.
// The file is fixed size, starts life full of zeros, and has array elements each a fixed number of bytes.
// Elements are written or read one at a time.
// By referring to io interfaces, we can arrange for os.File to be a naive non-optimized implementation.
type SparseLookupFile interface {
	io.ReaderAt
	io.WriterAt
	io.Closer
}
