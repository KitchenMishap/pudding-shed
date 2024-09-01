package intarrayarray

import (
	"github.com/KitchenMishap/pudding-shed/numberedfolders"
	"os"
)

type IntArrayMapStore struct {
	folder              string
	name                string
	numberedFolders     numberedfolders.NumberedFolders
	arrayCountPerFile   int64
	elementByteSize     int64
	latestFilepath      string        // "Latest" is the file corresponding to latest txo we have seen
	latestFileNumber    int64         // The index represented by the first entry in latest file
	latestIntArrayArray IntArrayArray // We keep latest in memory even when we need to go back to something older
	olderFilepath       string        // "Older" is a file we temporarily load as we need to append to old data
	olderIntArrayArray  IntArrayArray
}

func (ms *IntArrayMapStore) folderPath(folders string) string {
	sep := string(os.PathSeparator)
	if folders == "" {
		return ms.folder + sep + ms.name
	} else {
		return ms.folder + sep + ms.name + sep + folders
	}
}

func (ms *IntArrayMapStore) filePath(folders string, filename string) string {
	sep := string(os.PathSeparator)
	return ms.folderPath(folders) + sep + filename + ".iaa"
}

func (ms *IntArrayMapStore) GetArray(arrayKey int64) ([]int64, error) {
	err := ms.Sync()
	if err != nil {
		return []int64{}, err
	}

	// Find the folder and filename
	folders, filename, _ := ms.numberedFolders.NumberToFoldersAndFile(arrayKey)
	filepath := ms.filePath(folders, filename)
	// Load in the file (don't use currentIntArrayArray for reads
	// because we're not so likely to use the same file twice in a row for reads)
	intArrayArray := NewIntArrayArray(ms.arrayCountPerFile, ms.elementByteSize)
	err = intArrayArray.LoadFile(filepath)
	if err != nil {
		return []int64{}, err
	}

	// Get the array
	return intArrayArray.GetArray(arrayKey % ms.arrayCountPerFile), nil
}

func (ms *IntArrayMapStore) AppendToArray(arrayKey int64, value int64) error {
	// Find the folder and filename
	folders, filename, fileNumber := ms.numberedFolders.NumberToFoldersAndFile(arrayKey)
	folderPath := ms.folderPath(folders)
	filepath := ms.filePath(folders, filename)
	// Is it already loaded as the latest? (This is the most common case, and is quick to handle)
	if ms.latestFilepath == filepath {
		// Append to the array
		ms.latestIntArrayArray.AppendToArray(arrayKey%ms.arrayCountPerFile, value)
		return nil
	}
	// OK it's not the same as our latest. But is it even later?
	if fileNumber > ms.latestFileNumber {
		// Save the previous latest, if any
		if ms.latestFilepath != "" {
			ms.latestIntArrayArray.SaveFile(ms.latestFilepath)
		}
		// Load a "new" latest
		// Create the folder if necessary
		err := os.MkdirAll(folderPath, 0755)
		if err != nil {
			return err
		}
		// Make a note of the filename and number
		ms.latestFilepath = filepath
		ms.latestFileNumber = fileNumber
		err = ms.latestIntArrayArray.LoadFile(filepath)
		if err != nil {
			// We don't mind about file not found, happy to be empty
		}
		// Append to the array
		ms.latestIntArrayArray.AppendToArray(arrayKey%ms.arrayCountPerFile, value)
		return nil
	}
	// It's not the existing latest, or a new latest, so it must be older
	// We keep "latest" in memory, as we're very likely to use it again soon
	// But we'll temporarily use the "older" intarrayarray to deal with this older request
	// Is it the same as any existing "older" that we might continue to use?
	if ms.olderFilepath == filepath {
		// We can just append to it! That was convenient
		ms.olderIntArrayArray.AppendToArray(arrayKey%ms.arrayCountPerFile, value)
		return nil
	} else {
		// Save any previous "older"
		if ms.olderFilepath != "" {
			// Don't forget to create the folder... as it may not exist yet
			err := os.MkdirAll(folderPath, 0755)
			if err != nil {
				return err
			}

			err = ms.olderIntArrayArray.SaveFile(ms.olderFilepath)
			if err != nil {
				return err
			}
		}
		// We can now safely load the required file into "older"
		// Make a note of the filename
		ms.olderFilepath = filepath
		err := ms.olderIntArrayArray.LoadFile(filepath)
		if err != nil {
			// We don't mind a file not found here
		}
		// And append to it
		ms.olderIntArrayArray.AppendToArray(arrayKey%ms.arrayCountPerFile, value)
		return nil
	}
	// Note we're not saving anything unnecessarily yet
}

func (ms *IntArrayMapStore) FlushFile() error {
	if ms.olderFilepath != "" {
		err := ms.olderIntArrayArray.SaveFile(ms.olderFilepath)
		if err != nil {
			return err
		}
		ms.olderFilepath = ""
	}
	if ms.latestFilepath != "" {
		err := ms.latestIntArrayArray.SaveFile(ms.latestFilepath)
		if err != nil {
			return err
		}
		ms.latestFilepath = ""
		ms.latestFileNumber = 0
	}
	return nil
}

func (ms *IntArrayMapStore) Sync() error {
	return ms.FlushFile()
}
