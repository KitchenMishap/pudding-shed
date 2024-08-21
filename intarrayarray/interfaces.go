package intarrayarray

// IntArrayMap maps uint64's to growing arrays of uint64s

type IntArrayMapStoreReadOnly interface {
	GetArray(arrayKey int64) []int64
}

type IntArrayMapStoreReadWrite interface {
	IntArrayMapStoreReadOnly
	AppendToArray(arrayKey int64, value int64)
	FlushFile()
}

type IntArrayMapStoreCreator interface {
	MapExists() bool
	CreateMap() error
	OpenMap() (IntArrayMapStoreReadWrite, error)
	OpenMapReadOnly() (IntArrayMapStoreReadOnly, error)
}
