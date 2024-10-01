package indexedhashes

import (
	"encoding/binary"
	"errors"
	"github.com/KitchenMishap/pudding-shed/memblocker"
	"os"
	"sync"
)

func NewUniformHashPreLoader(us *UniformHashStore, blocker *memblocker.MemBlocker) *UniformHashPreLoader {
	result := UniformHashPreLoader{}
	result.uniform = us
	result.memStore.mapAddressToHashEntries = make(map[uint64]*hashEntriesList)
	result.cacheStore.mapAddressToHashEntries = make(map[uint64]*hashEntriesList)
	result.memStore.mapSizeGroups = make(map[int]map[uint64]bool)
	result.cacheStore.mapSizeGroups = make(map[int]map[uint64]bool)
	addresses := result.biggestAddress() + 1
	result.fileStore = newFileStoreBits(addresses)
	result.memBlocker = blocker
	return &result
}

type UniformHashPreLoader struct {
	uniform    *UniformHashStore
	memStore   entryStore
	cacheStore entryStore
	fileStore  *fileStoreBits
	memBlocker *memblocker.MemBlocker
}

type hashEntry struct {
	hash  *Sha256
	index uint64
}

type hashEntriesList struct {
	address     uint64
	hashEntries []hashEntry
}

func NewHashEntriesList(address uint64) *hashEntriesList {
	result := hashEntriesList{}
	result.address = address
	return &result
}

func (el *hashEntriesList) append(entry hashEntry) {
	el.hashEntries = append(el.hashEntries, entry)
}

// There are two of these with similar characteristics
type entryStore struct {
	lock                    sync.Mutex
	mapAddressToHashEntries map[uint64]*hashEntriesList // Indexed by address which is derived from hash
	mapSizeGroups           map[int]map[uint64]bool     // Outer key is size of list. Inner key is address. Bool is dummy
}

func (es *entryStore) addIfListed(address uint64, entry hashEntry) bool {
	es.lock.Lock()
	defer es.lock.Unlock()
	list, listed := es.mapAddressToHashEntries[address]
	if listed {
		prevSize := len(list.hashEntries)
		// Remove from map of list sizes at old size
		delete(es.mapSizeGroups[prevSize], address)

		newSize := prevSize + 1
		list.append(entry)
		// Add to map of list sizes at new size
		// But does "newSize" have a map yet?
		_, mapExists := es.mapSizeGroups[newSize]
		if !mapExists {
			es.mapSizeGroups[newSize] = make(map[uint64]bool)
		}
		es.mapSizeGroups[newSize][address] = true // true is dummy to make a set out of a map
		return true
	}
	return false
}

func (es *entryStore) addNewList(address uint64, list *hashEntriesList) {
	es.lock.Lock()
	defer es.lock.Unlock()
	size := len(list.hashEntries)
	es.mapAddressToHashEntries[address] = list
	// Add to map of list sizes at size
	// But does "size" have a map yet?
	_, mapExists := es.mapSizeGroups[size]
	if !mapExists {
		es.mapSizeGroups[size] = make(map[uint64]bool)
	}
	es.mapSizeGroups[size][address] = true // true is dummy to make a set out of a map
}

func (es *entryStore) biggestListSize() int {
	es.lock.Lock()
	defer es.lock.Unlock()
	biggest := 0
	for k, _ := range es.mapSizeGroups {
		if len(es.mapSizeGroups[k]) > 0 { // Shouldn't have to do this test!
			if k > biggest {
				biggest = k
			}
		}
	}
	return biggest
}

func (es *entryStore) extractAListOfSize(listSize int) *hashEntriesList {
	es.lock.Lock()
	defer es.lock.Unlock()
	//fmt.Println("extracting a list of size ", listSize)
	addrMap, exists := es.mapSizeGroups[listSize]
	if !exists {
		return nil
	}
	//fmt.Println("There are ", len(addrMap), " lists of that size ")
	if len(addrMap) == 0 {
		return nil
	}
	for address, _ := range addrMap {
		// Extract
		entriesList, exists := es.mapAddressToHashEntries[address]
		if !exists {
			panic("address in addrMap but not in mapAddressToHashEntries")
		}
		delete(es.mapAddressToHashEntries, address)
		if entriesList == nil {
			panic("Whoops")
		}
		// Also need to delete from the map based on count
		delete(es.mapSizeGroups[listSize], address)
		return entriesList
		// We only wanted the first one in the map
	}
	return nil
}

// The file store is used in combination with the files themselves.
// The object here merely records whether the files are empty
type fileStoreBits struct {
	allFilesEmpty    bool
	fileNotEmptyBits []uint64
}

func newFileStoreBits(addresses uint64) *fileStoreBits {
	words := addresses / uint64(64)
	spare := addresses % uint64(64)
	if spare > 0 {
		words++
	}
	result := fileStoreBits{}
	// Starts of with all bits zero, each zero meaning "file empty"
	result.fileNotEmptyBits = make([]uint64, words)
	// And an additional bool to summarize when all files are empty
	result.allFilesEmpty = true
	return &result
}

func (fs *fileStoreBits) isFileEmpty(address uint64) bool {
	if fs.allFilesEmpty {
		return true
	}
	word := address / uint64(64)
	bit := address % uint64(64)
	mask := uint64(1) << bit
	if fs.fileNotEmptyBits[word]&mask != 0 {
		return false
	}
	return true
}

// Files never go from "not empty" to "empty"
func (fs *fileStoreBits) setFileNotEmpty(address uint64) {
	fs.allFilesEmpty = false
	word := address / uint64(64)
	bit := address % uint64(64)
	mask := uint64(1) << bit
	fs.fileNotEmptyBits[word] |= mask
}

func (pl *UniformHashPreLoader) loadHashEntriesListIfFileExists(address uint64) (*hashEntriesList, error) {
	if pl.fileStore.isFileEmpty(address) {
		return nil, nil
	}
	_, filePath := pl.uniform.folderPathFilePathForAddress(address)
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	if len(bytes)%40 != 0 {
		return nil, errors.New("expect filesize multiple of 40")
	}
	result := hashEntriesList{}
	result.address = address
	result.hashEntries = make([]hashEntry, len(bytes)/40)
	for i := 0; i < len(bytes)/40; i++ {
		// Index followed by hash
		index := binary.LittleEndian.Uint64(bytes[i*40 : i*40+8])
		hash := Sha256{}
		copy(hash[:], bytes[i*40+8:i*40+40])
		result.hashEntries[i].index = index
		result.hashEntries[i].hash = &hash
	}
	return &result, nil
}

// createEmptyFiles is a LONG operation suitable for a goroutine
func (pl *UniformHashPreLoader) createEmptyFiles() error {
	top := pl.biggestAddress()
	// We iterate backwards to see progress easily in file manager
	for i := int64(top); i >= 0; i-- {
		folders, filename, _ := pl.uniform.numberedFolders.NumberToFoldersAndFile(int64(i))
		folderPath, filePath := pl.uniform.folderPathFilePathFromFoldersFilename(folders, filename)
		err := os.MkdirAll(folderPath, 0755)
		if err != nil {
			return err
		}
		file, err := os.Create(filePath)
		if err != nil {
			return err
		}
		err = file.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func (pl *UniformHashPreLoader) biggestAddress() uint64 {
	var hash Sha256 = [32]byte{255, 255, 255, 255, 255, 255, 255, 255,
		255, 255, 255, 255, 255, 255, 255, 255,
		255, 255, 255, 255, 255, 255, 255, 255,
		255, 255, 255, 255, 255, 255, 255, 255}
	return pl.dividedAddressForHash(&hash)
}

func (pl *UniformHashPreLoader) dividedAddressForHash(hash *Sha256) uint64 {
	hashLSBs := binary.LittleEndian.Uint64(hash[0:8])
	dividedHash := hashLSBs / pl.uniform.hashDivider
	return dividedHash
}

func (pl *UniformHashPreLoader) delegateEntryToStores(entry hashEntry, address uint64) error {
	//pl.memBlocker.WaitForSpareMemory()	 CALLER HAS TO DO THIS NOW!
	// Is there an entry for address in Mem?
	if pl.memStore.addIfListed(address, entry) {
		// We've put it in mem with its friends
	} else if pl.cacheStore.addIfListed(address, entry) {
		// We've put it in a cache of a file with its friends
	} else {
		list, err := pl.loadHashEntriesListIfFileExists(address)
		if err != nil {
			return err
		}
		if list != nil {
			// Entry list found in file. Add this to the list and cache the file
			list.append(entry)
			pl.cacheStore.addNewList(address, list)
		} else {
			// Not in a file. Put it in mem as a new list
			newList := NewHashEntriesList(address)
			newList.append(entry)
			pl.memStore.addNewList(address, newList)
		}
	}
	return nil
}

func (pl *UniformHashPreLoader) extractSomeEntriesStoresToFiles(count int) (int, error) {
	for i := 0; i < count; i++ {
		achieved, err := pl.extractEntryFromStoresToFile()
		if err != nil {
			return i, err
		}
		if !achieved {
			return i, nil
		} // No point carrying on if we couldn't do any
	}
	return count, nil
}

func (pl *UniformHashPreLoader) extractEntryFromStoresToFile() (bool, error) {
	// Decide the store that has the biggest available list
	biggestInMem := pl.memStore.biggestListSize()
	biggestInCache := pl.cacheStore.biggestListSize()
	if biggestInMem > biggestInCache {
		//fmt.Println("Finding the list of advertised size ", biggestInMem)
		aBigList := pl.memStore.extractAListOfSize(biggestInMem)
		if aBigList == nil {
			//fmt.Println("Couldn't find it!")
			return false, nil
		}
		err := pl.sendToFile(aBigList)
		if err == nil {
			return true, nil
		} else {
			return false, err
		}
	} else if biggestInCache > 0 {
		aBigList := pl.cacheStore.extractAListOfSize(biggestInCache)
		if aBigList == nil {
			return false, nil
		}
		err := pl.sendToFile(aBigList)
		if err == nil {
			return true, nil
		} else {
			return false, err
		}
	} else {
		return false, nil
	}
}

func (pl *UniformHashPreLoader) sendToFile(hl *hashEntriesList) error {
	// Construct file into memory first
	count := len(hl.hashEntries)
	bytes := make([]byte, 40*count)
	for i := 0; i < count; i++ {
		binary.LittleEndian.PutUint64(bytes[40*i:40*i+8], hl.hashEntries[i].index)
		copy(bytes[40*i+8:40*i+40], hl.hashEntries[i].hash[:])
	}

	_, filePath := pl.uniform.folderPathFilePathForAddress(hl.address)
	err := os.WriteFile(filePath, bytes, 0755)
	if err != nil {
		return err
	}
	//fmt.Println("Written file: ", filePath)
	pl.fileStore.setFileNotEmpty(hl.address)
	return nil
}
