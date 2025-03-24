package memfile

import (
	"bufio"
	"io"
	"os"
)

type appendOptimizedFile struct {
	file             *os.File
	bufferedWriter   *bufio.Writer // nil unless the last operation was an append
	nextAppendOffset int64         // valid if the last operation was an append
	totalSize        int64         // Includes stuff in the bufferedWriter
}

// Check that implements
var _ AppendableLookupFile = (*appendOptimizedFile)(nil)

func NewAppendOptimizedFile(file *os.File) (AppendableLookupFile, error) {
	result := appendOptimizedFile{}
	result.file = file
	result.bufferedWriter = nil // Until an append is done
	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}
	result.totalSize = stat.Size()
	return &result, nil
}

func (a *appendOptimizedFile) ReadAt(p []byte, off int64) (int, error) {
	err := a.flush()
	if err != nil {
		return 0, err
	}
	nRead, err := a.file.ReadAt(p, off)
	if err != nil {
		return nRead, err
	}
	return nRead, nil
}

func (a *appendOptimizedFile) WriteAt(p []byte, off int64) (n int, err error) {
	if a.bufferedWriter != nil {
		// Previous operation was an append
		if off == a.nextAppendOffset {
			// This WriteAt operation is also an append
			// So we just squirt it through the buffer
			nWritten, err := a.bufferedWriter.Write(p)
			if err != nil {
				// A failed append doesn't count as an append
				a.bufferedWriter = nil
				a.totalSize += int64(nWritten)
				return nWritten, err
			}
			a.nextAppendOffset += int64(nWritten)
			a.totalSize += int64(nWritten)
			return nWritten, nil
		} else {
			// This WriteAt operation is NOT an append, so
			// we need to flush previous appends first
			err := a.flush()
			if err != nil {
				return 0, err
			}
			// Do the WriteAt without a buffer
			nWritten, err := a.file.WriteAt(p, off)
			if off+int64(nWritten) > a.totalSize {
				a.totalSize = off + int64(nWritten)
			}
			return nWritten, err
		}
	} else {
		// Previous operation was not append. But what about this operation?
		if off == a.totalSize {
			// Yes it is an append, set up a buffer and use it
			a.file.Seek(0, io.SeekEnd)
			a.bufferedWriter = bufio.NewWriter(a.file)
			nWritten, err := a.bufferedWriter.Write(p)
			if err != nil {
				// A failed append doesn't count as an append
				a.totalSize += int64(nWritten)
				a.bufferedWriter = nil
				return nWritten, err
			}
			a.nextAppendOffset = off + int64(nWritten)
			a.totalSize += int64(nWritten)
			return nWritten, nil
		}
		// No it's not an append. Do it without a buffer
		nWritten, err := a.file.WriteAt(p, off)
		if off+int64(nWritten) > a.totalSize {
			a.totalSize = off + int64(nWritten)
		}
		return nWritten, err
	}
}

func (a *appendOptimizedFile) Close() error {
	er := a.flush()
	if er != nil {
		return er
	}
	return a.file.Close()
}

func (a *appendOptimizedFile) Size() int64 {
	return a.totalSize
}

func (a *appendOptimizedFile) flush() error {
	if a.bufferedWriter != nil {
		err := a.bufferedWriter.Flush()
		if err != nil {
			return err
		}
		a.bufferedWriter = nil
	}
	return nil
}

func (a *appendOptimizedFile) Sync() error {
	err := a.flush()
	if err != nil {
		return err
	}
	err = a.file.Sync()
	if err != nil {
		return err
	}
	return nil
}
