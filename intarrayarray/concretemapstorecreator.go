package intarrayarray

import (
	"github.com/KitchenMishap/pudding-shed/numberedfolders"
	"math"
)

type ConcreteMapStoreCreator struct {
	name            string
	folder          string
	digitsPerFile   int64
	digitsPerFolder int64
	elementByteSize int64
}

func NewConcreteMapStoreCreator(name string, folder string, digitsPerFile int64, digitsPerFolder int64, elementByteSize int64) *ConcreteMapStoreCreator {
	result := ConcreteMapStoreCreator{}
	result.name = name
	result.folder = folder
	result.digitsPerFile = digitsPerFile
	result.digitsPerFolder = digitsPerFolder
	result.elementByteSize = elementByteSize
	return &result
}

func (c *ConcreteMapStoreCreator) MapExists() bool {
	return true // ToDo [  ]
}

func (c *ConcreteMapStoreCreator) CreateMap() {
	// ToDo [  ]
}

func (c *ConcreteMapStoreCreator) OpenMap() IntArrayMapStoreReadWrite {
	result := IntArrayMapStore{}
	result.arrayCountPerFile = int64(math.Pow10(int(c.digitsPerFile)))
	result.elementByteSize = c.elementByteSize
	result.numberedFolders = numberedfolders.NewNumberedFolders(int(c.digitsPerFile), int(c.digitsPerFolder))
	return &result
}

func (c *ConcreteMapStoreCreator) OpenMapReadOnly() IntArrayMapStoreReadOnly {
	result := IntArrayMapStore{}
	result.arrayCountPerFile = int64(math.Pow10(int(c.digitsPerFile)))
	result.elementByteSize = c.elementByteSize
	result.numberedFolders = numberedfolders.NewNumberedFolders(int(c.digitsPerFile), int(c.digitsPerFolder))
	return &result
}
