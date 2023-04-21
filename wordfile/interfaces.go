package wordfile

type ReaderAtWord interface {
	ReadWordAt(off int64) (int64, error)
}

type WriterAtWord interface {
	WriteWordAt(off int64, val int64) error
}

type WordCounter interface {
	CountWords() (int64, error)
}
