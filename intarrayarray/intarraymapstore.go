package intarrayarray

import (
	"github.com/KitchenMishap/pudding-shed/numberedfolders"
	"os"
)

type IntArrayMapStore struct {
	folder               string
	name                 string
	numberedFolders      numberedfolders.NumberedFolders
	arrayCountPerFile    int64
	elementByteSize      int64
	currentFilepath      string
	currentIntArrayArray IntArrayArray
}

func (ms *IntArrayMapStore) folderPath(folders string) string {
	sep := string(os.PathSeparator)
	return ms.folder + sep + ms.name + sep + folders
}

func (ms *IntArrayMapStore) filePath(folders string, filename string) string {
	sep := string(os.PathSeparator)
	return ms.folderPath(folders) + sep + filename + ".iaa"
}

func (ms *IntArrayMapStore) GetArray(arrayKey int64) ([]int64, error) {
	// Find the folder and filename
	folders, filename := ms.numberedFolders.NumberToFoldersAndFile(arrayKey)
	filepath := ms.filePath(folders, filename)
	// Load in the file (don't use currentIntArrayArray for reads
	// because we're not so likely to use the same file twice in a row for reads)
	intArrayArray := NewIntArrayArray(ms.arrayCountPerFile, ms.elementByteSize)
	err := intArrayArray.LoadFile(filepath)
	if err != nil {
		return []int64{}, err
	}

	// Get the array
	return intArrayArray.GetArray(arrayKey % ms.arrayCountPerFile), nil
}

func (ms *IntArrayMapStore) AppendToArray(arrayKey int64, value int64) error {
	// Find the folder and filename
	folders, filename := ms.numberedFolders.NumberToFoldersAndFile(arrayKey)
	folderPath := ms.folderPath(folders)
	filepath := ms.filePath(folders, filename)
	// Is it already loaded?
	if ms.currentFilepath != filepath {
		// Save the existing one
		if ms.currentFilepath != "" {
			ms.currentIntArrayArray.SaveFile(ms.currentFilepath)
		}

		// Create the folder if necessary
		err := os.MkdirAll(folderPath, 0755)
		if err != nil {
			return err
		}

		// Load in the file
		ms.currentFilepath = filepath
		ms.currentIntArrayArray = NewIntArrayArray(ms.arrayCountPerFile, ms.elementByteSize)
		err = ms.currentIntArrayArray.LoadFile(filepath)
		if err != nil {
			// Don't care if file doesn't exist
		}
	}
	// The file we want (containing the array we wish to append to) is now loaded

	// Append to the array
	ms.currentIntArrayArray.AppendToArray(arrayKey%ms.arrayCountPerFile, value)

	return nil
}

func (ms *IntArrayMapStore) FlushFile() error {
	if ms.currentFilepath != "" {
		err := ms.currentIntArrayArray.SaveFile(ms.currentFilepath)
		if err != nil {
			return err
		}
		ms.currentFilepath = ""
	}
	return nil
}

func (ms *IntArrayMapStore) Sync() error {
	return ms.FlushFile()
}
