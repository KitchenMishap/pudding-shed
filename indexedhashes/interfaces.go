package indexedhashes

type Sha256 [32]byte

type HashReader interface {
	IndexOfHash(hash *Sha256) (int64, error)
	GetHashAtIndex(index int64, hash *Sha256) error
	CountHashes() (int64, error)
	Close() error
	WholeFileAsInt32() ([]uint32, error)
}

type HashReadWriter interface {
	HashReader
	AppendHash(hash *Sha256) (int64, error)
}

type HashStoreCreator interface {
	HashStoreExists() bool
	CreateHashStore() error
	OpenHashStore() (HashReadWriter, error)
	OpenHashStoreReadOnly() (HashReader, error)
}
