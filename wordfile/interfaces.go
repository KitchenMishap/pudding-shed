package wordfile

import "io"

type ReaderAtWord interface {
	io.Closer
	ReadWordAt(off int64) (int64, error)
}

type WriterAtWord interface {
	io.Closer
	WriteWordAt(val int64, off int64) error
}

type WordCounter interface {
	CountWords() (int64, error)
}

type ReadWriteAtWordCounter interface {
	ReaderAtWord
	WriterAtWord
	WordCounter
	Sync() error
	WordSize() int64
}

type ReadAtWordCounter interface {
	ReaderAtWord
	WordCounter
	WordSize() int64
}

type WordFileCreator interface {
	WordFileExists() bool
	CreateWordFile() error
	OpenWordFile() (ReadWriteAtWordCounter, error)
	OpenWordFileReadOnly() (ReadAtWordCounter, error)
	CreateWordFileFilledZeros(count int64) error
}
