package intarrayarray

import (
	"github.com/KitchenMishap/pudding-shed/numberedfolders"
	"os"
)

type CachedIntArrayMapStore struct {
	// Settings
	folder                 string
	name                   string
	numberedFolders        numberedfolders.NumberedFolders
	arrayCountPerFile      int64
	elementByteSize        int64
	cacheElementCountLimit int64
	// Live
	cacheElementCount             int64
	lastAccessCounter             int64 // An increasing index, indexing time of last access
	mapFolderFilenameToArrayArray map[string]IntArrayArray
	mapFolderFilenameToLastAccess map[string]int64
	mapFolderFilenameToFolderPath map[string]string

	//	latestFilepath      string        // "Latest" is the file corresponding to latest txo we have seen
	//	latestFileNumber    int64         // The index represented by the first entry in latest file
	//	latestIntArrayArray IntArrayArray // We keep latest in memory even when we need to go back to something older
	//	olderFilepath       string        // "Older" is a file we temporarily load as we need to append to old data
	//	olderIntArrayArray  IntArrayArray
}

func (ms *CachedIntArrayMapStore) folderPath(folders string) string {
	sep := string(os.PathSeparator)
	if folders == "" {
		return ms.folder + sep + ms.name
	} else {
		return ms.folder + sep + ms.name + sep + folders
	}
}

func (ms *CachedIntArrayMapStore) filePath(folders string, filename string) string {
	sep := string(os.PathSeparator)
	return ms.folderPath(folders) + sep + filename + ".iaa"
}

func (ms *CachedIntArrayMapStore) filePath2(foldersFilename string) string {
	sep := string(os.PathSeparator)
	return ms.folder + sep + ms.name + sep + foldersFilename + ".iaa"
}

func (ms *CachedIntArrayMapStore) GetArray(arrayKey int64) ([]int64, error) {
	err := ms.Sync() // Expensive :-O
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

func (ms *CachedIntArrayMapStore) loadIntoCache(folders string, filename string) (*IntArrayArray, error) {
	sep := string(os.PathSeparator)
	var foldersFilename string
	if folders == "" {
		foldersFilename = filename
	} else {
		foldersFilename = folders + sep + filename
	}
	folderPath := ms.folderPath(folders)

	cached, ok := ms.mapFolderFilenameToArrayArray[foldersFilename]
	if ok {
		return &cached, nil // Already cached
	}

	// Newly cached
	cached = NewIntArrayArray(ms.arrayCountPerFile, ms.elementByteSize)
	filePath := ms.filePath(folders, filename)
	err := cached.LoadFile(filePath)
	if err != nil {
		// Don't mind error here, we just stay empty
	}
	ms.mapFolderFilenameToArrayArray[foldersFilename] = cached
	ms.mapFolderFilenameToFolderPath[foldersFilename] = folderPath
	ms.mapFolderFilenameToLastAccess[foldersFilename] = 0
	ms.cacheElementCountLimit += cached.elementTally

	for ms.cacheElementCount >= ms.cacheElementCountLimit {
		ms.trimCache(foldersFilename)
	}
	return &cached, nil
}

func (ms *CachedIntArrayMapStore) trimCache(excludeFoldersFilename string) error {
	oldest := ms.lastAccessCounter // Start at latest
	oldestFoldersFilename := ""

	for key, value := range ms.mapFolderFilenameToLastAccess {
		if value < oldest && key != excludeFoldersFilename {
			oldest = value
			oldestFoldersFilename = key
		}
	}

	cached := ms.mapFolderFilenameToArrayArray[oldestFoldersFilename]
	filepath := ms.filePath2(oldestFoldersFilename)
	folderPath := ms.mapFolderFilenameToFolderPath[oldestFoldersFilename]
	// Don't forget to create the folder... as it may not exist yet
	err := os.MkdirAll(folderPath, 0755)
	if err != nil {
		return err
	}
	err = cached.SaveFile(filepath)
	if err != nil {
		return err
	}
	ms.cacheElementCountLimit -= cached.elementTally

	delete(ms.mapFolderFilenameToArrayArray, oldestFoldersFilename)
	delete(ms.mapFolderFilenameToLastAccess, oldestFoldersFilename)
	delete(ms.mapFolderFilenameToFolderPath, oldestFoldersFilename)
	return nil
}

func (ms *CachedIntArrayMapStore) AppendToArray(arrayKey int64, value int64) error {
	// Find the folder and filename
	folders, filename, _ := ms.numberedFolders.NumberToFoldersAndFile(arrayKey)

	cached, err := ms.loadIntoCache(folders, filename)
	if err != nil {
		// We don't mind being empty
	}

	cached.AppendToArray(arrayKey%ms.arrayCountPerFile, value)
	ms.cacheElementCount++

	var foldersFilename string
	if folders == "" {
		foldersFilename = filename
	} else {
		sep := string(os.PathSeparator)
		foldersFilename = folders + sep + filename
	}
	ms.lastAccessCounter++
	ms.mapFolderFilenameToLastAccess[foldersFilename] = ms.lastAccessCounter

	return nil
}

func (ms *CachedIntArrayMapStore) FlushFile() error {
	for key, value := range ms.mapFolderFilenameToArrayArray {
		filepath := ms.filePath2(key)
		folderPath := ms.mapFolderFilenameToFolderPath[key]
		// Don't forget to create the folder... as it may not exist yet
		err := os.MkdirAll(folderPath, 0755)
		if err != nil {
			return err
		}

		err = value.SaveFile(filepath)
		if err != nil {
			return err
		}
	}
	ms.mapFolderFilenameToArrayArray = make(map[string]IntArrayArray)
	ms.mapFolderFilenameToLastAccess = make(map[string]int64)
	ms.mapFolderFilenameToFolderPath = make(map[string]string)
	ms.cacheElementCount = 0
	ms.lastAccessCounter = 0
	return nil
}

func (ms *CachedIntArrayMapStore) Sync() error {
	return ms.FlushFile()
}
