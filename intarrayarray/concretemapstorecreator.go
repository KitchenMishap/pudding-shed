package intarrayarray

import (
	"github.com/KitchenMishap/pudding-shed/numberedfolders"
	"math"
	"os"
)

type ConcreteMapStoreCreator struct {
	name            string
	folder          string
	digitsPerFile   int64
	digitsPerFolder int64
	elementByteSize int64
	cached          bool
}

func NewConcreteMapStoreCreator(name string, folder string, digitsPerFile int64, digitsPerFolder int64, elementByteSize int64, cached bool) *ConcreteMapStoreCreator {
	result := ConcreteMapStoreCreator{}
	result.name = name
	result.folder = folder
	result.digitsPerFile = digitsPerFile
	result.digitsPerFolder = digitsPerFolder
	result.elementByteSize = elementByteSize
	result.cached = cached
	return &result
}

func (c *ConcreteMapStoreCreator) MapExists() bool {
	return true // ToDo [  ]
}

func (c *ConcreteMapStoreCreator) CreateMap() error {
	dir := c.folder + string(os.PathSeparator) + c.name
	os.RemoveAll(dir)
	os.MkdirAll(dir, os.ModePerm)
	return nil
}

func (c *ConcreteMapStoreCreator) OpenMap() (IntArrayMapStoreReadWrite, error) {
	if c.cached {
		result := CachedIntArrayMapStore{}
		result.folder = c.folder
		result.name = c.name
		result.arrayCountPerFile = int64(math.Pow10(int(c.digitsPerFile)))
		result.elementByteSize = c.elementByteSize
		result.numberedFolders = numberedfolders.NewNumberedFolders(int(c.digitsPerFile), int(c.digitsPerFolder))
		result.cacheElementCountLimit = 8000000000 / c.elementByteSize // 8GB
		return &result, nil

	} else {
		result := IntArrayMapStore{}
		result.folder = c.folder
		result.name = c.name
		result.arrayCountPerFile = int64(math.Pow10(int(c.digitsPerFile)))
		result.elementByteSize = c.elementByteSize
		result.numberedFolders = numberedfolders.NewNumberedFolders(int(c.digitsPerFile), int(c.digitsPerFolder))
		result.latestIntArrayArray = NewIntArrayArray(result.arrayCountPerFile, c.elementByteSize)
		result.olderIntArrayArray = NewIntArrayArray(result.arrayCountPerFile, c.elementByteSize)
		return &result, nil
	}
}

func (c *ConcreteMapStoreCreator) OpenMapReadOnly() (IntArrayMapStoreReadOnly, error) {
	result := IntArrayMapStore{}
	result.folder = c.folder
	result.name = c.name
	result.arrayCountPerFile = int64(math.Pow10(int(c.digitsPerFile)))
	result.elementByteSize = c.elementByteSize
	result.numberedFolders = numberedfolders.NewNumberedFolders(int(c.digitsPerFile), int(c.digitsPerFolder))
	result.latestIntArrayArray = NewIntArrayArray(result.arrayCountPerFile, c.elementByteSize)
	result.olderIntArrayArray = NewIntArrayArray(result.arrayCountPerFile, c.elementByteSize)
	return &result, nil
}
