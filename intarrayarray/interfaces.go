package intarrayarray

// IntArrayMap maps uint64's to growing arrays of uint64s

type IntArrayMapStoreReadOnly interface {
	GetArray(arrayKey int64) ([]int64, error)
}

type IntArrayMapStoreReadWrite interface {
	IntArrayMapStoreReadOnly
	AppendToArray(arrayKey int64, value int64) error
	FlushFile() error
	Sync() error
}

type IntArrayMapStoreCreator interface {
	MapExists() bool
	CreateMap() error
	OpenMap() (IntArrayMapStoreReadWrite, error)
	OpenMapReadOnly() (IntArrayMapStoreReadOnly, error)
}
