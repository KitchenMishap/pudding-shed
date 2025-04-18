// Records
// Records are a way of arranging for multiple different sized words to be near eachother in a file, such that they
// can be read and written to a file together quickly. Records contain a number of named words, each having (sperately)
// a choice of 1,2,3,4,5,6,7, or 8 bytes, such that all possible values can be expressed in a uint64.
// A RecordsDescriptor is used to define the names, offsets, and bytesizes of the words within a record.
// Records do not contain other records, only numeric words.

package records

import "io"

type RecordDescriptor interface {
	AppendWordDescription(name string, wordSize int)
	RecordSize() int
	FieldDescriptor(name string) (offset int, byteSize int, err error)
}

type ReaderAtRecord interface {
	io.Closer
	ReadRecordAt(off int64) (Record, error)
}

type WriterAtRecord interface {
	io.Closer
	WriteRecordAt(val Record, off int64) error
}

type RecordCounter interface {
	CountRecords() (int64, error)
}

type ReadWriteAtRecordCounter interface {
	ReaderAtRecord
	WriterAtRecord
	RecordCounter
	Sync() error
	RecordSize() int
}

type ReadAtRecordCounter interface {
	ReaderAtRecord
	RecordCounter
	RecordSize() int
}

// Factory
type RecordFileCreator interface {
	RecordFileExists() bool
	CreateRecordFile() error
	OpenRecordFile() (ReadWriteAtRecordCounter, error)
	OpenRecordFileReadOnly() (ReadAtRecordCounter, error)
	CreateRecordFileFilledZeros(count int64) error
}
