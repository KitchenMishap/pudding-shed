package indexedhashes

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/KitchenMishap/pudding-shed/memfile"
	"github.com/KitchenMishap/pudding-shed/numberedfolders"
	"github.com/KitchenMishap/pudding-shed/wordfile"
	"io"
	"math"
	"os"
)

func NewUniformHashStoreCreatorAndPreloader(
	folder string, name string,
	hashCountEstimate int64, digitsPerFolder int,
	gigabytesMem int64) (HashStoreCreator, *MultipassPreloader) {
	creator := NewUniformHashStoreCreatorPrivate(hashCountEstimate,
		folder, name, digitsPerFolder)
	preloader := NewMultipassPreloader(creator, 1024*1024*1024*gigabytesMem)
	return creator, preloader
}

func NewUniformHashStoreCreatorAndPreloaderFromFile(
	folder string, name string, gigabytesMem int64) (*UniformHashStoreCreator, *MultipassPreloader, error) {
	creator := UniformHashStoreCreator{}
	creator.folder = folder
	creator.name = name
	sep := string(os.PathSeparator)
	paramsFilePath := folder + sep + name + sep + "Params.json"
	paramsFile, err := os.Open(paramsFilePath)
	if err != nil {
		return nil, nil, err
	}
	defer paramsFile.Close()
	bytes, err := io.ReadAll(paramsFile)
	if err != nil {
		return nil, nil, err
	}
	err = json.Unmarshal(bytes, &creator.params)
	if err != nil {
		return nil, nil, err
	}
	preloader := NewMultipassPreloader(&creator, 1024*1024*1024*gigabytesMem)
	return &creator, preloader, nil
}

func NewUniformHashStoreCreatorPrivate(hashCountEstimate int64,
	folder string, name string, digitsPerFolder int) *UniformHashStoreCreator {
	result := UniformHashStoreCreator{}
	result.folder = folder
	result.name = name
	result.params.DigitsPerFolder = digitsPerFolder

	// Poisson distribution for lambda = 75:
	// https://homepage.divms.uiowa.edu/~mbognar/applets/pois.html
	// P(X >= 102) = 0.00175
	// So for a 102-entry system, with targetEntriesPerFile = 75, 0.175% of entries will end up in overflow files
	// (102*40 is about 4096 with 16 bytes spare; 4096 is a typical hard drive allocation unit)
	const targetEntriesPerFile = 75
	targetNumFiles := hashCountEstimate / targetEntriesPerFile
	possibleHashes := math.Pow(2, 64) // Because taking 64 LS bits of hash
	result.params.HashDivider = uint64(possibleHashes / float64(targetNumFiles))

	sep := string(os.PathSeparator)
	fileParams, _ := os.Create(result.folderPath() + sep + "Params.json")
	defer fileParams.Close()
	byts, _ := json.Marshal(result.params)
	_, _ = fileParams.Write(byts)

	return &result
}

type UniformHashStoreParams struct {
	HashDivider     uint64 `json:"hashDivider"`
	DigitsPerFolder int    `json:"digitsPerFolder"`
}

type UniformHashStoreCreator struct {
	folder string
	name   string
	params UniformHashStoreParams
}

func (uc *UniformHashStoreCreator) folderPath() string {
	sep := string(os.PathSeparator)
	return uc.folder + sep + uc.name
}

func (uc *UniformHashStoreCreator) hashFilePath() string {
	sep := string(os.PathSeparator)
	return uc.folderPath() + sep + "Hashes.hsh"
}

func (uc *UniformHashStoreCreator) binStartsFilePath() string {
	sep := string(os.PathSeparator)
	return uc.folderPath() + sep + "BinStarts.bst"
}

func (uc *UniformHashStoreCreator) HashStoreExists() bool {
	// Hash store exists if folder exists and .hsh file exists
	_, err := os.Stat(uc.folderPath())
	if err != nil {
		return false
	}
	filePath := uc.hashFilePath()
	_, err = os.Stat(filePath)
	return err == nil
}

func (uc *UniformHashStoreCreator) CreateHashStore() error {
	err := os.RemoveAll(uc.folderPath())
	if err != nil {
		return err
	}
	err = os.MkdirAll(uc.folderPath(), 0755)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(uc.hashFilePath(), os.O_CREATE, 0755)
	defer file.Close()
	if err != nil {
		return err
	}

	sep := string(os.PathSeparator)
	fileParams, err := os.OpenFile(uc.folderPath()+sep+"Params.json", os.O_CREATE, 0755)
	defer fileParams.Close()
	if err != nil {
		return err
	}
	byts, err := json.Marshal(uc.params)
	_, err = fileParams.Write(byts)
	if err != nil {
		return err
	}

	return nil
}

func (uc *UniformHashStoreCreator) OpenHashStore() (HashReadWriter, error) {
	return uc.openHashStorePrivate()
}

func (uc *UniformHashStoreCreator) openHashStorePrivate() (*UniformHashStore, error) {
	if !uc.HashStoreExists() {
		return nil, errors.New("Hash store (folder) does not exist")
	}
	store := UniformHashStore{}
	store.folderPath = uc.folderPath()
	store.hashDivider = uc.params.HashDivider
	store.numberedFolders = numberedfolders.NewNumberedFolders(0, uc.params.DigitsPerFolder)

	file, err := os.OpenFile(uc.hashFilePath(), os.O_RDWR, 0755)
	if err != nil {
		return nil, err
	}
	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}
	size := stat.Size()
	underlying, err := memfile.NewAppendOptimizedFile(file)
	if err != nil {
		return nil, err
	}
	store.hashFile = wordfile.NewHashFile(underlying, size/32)
	file, err = os.OpenFile(uc.binStartsFilePath(), os.O_RDWR, 0755)
	if err != nil {
		return nil, err
	}
	store.binStartsFile = file
	return &store, nil
}

func (uc *UniformHashStoreCreator) OpenHashStoreReadOnly() (HashReader, error) {
	store, err := uc.OpenHashStore()
	return store, err
}

type UniformHashStore struct {
	hashDivider     uint64
	folderPath      string
	numberedFolders numberedfolders.NumberedFolders
	hashFile        *wordfile.HashFile
	binStartsFile   *os.File
}

func (us *UniformHashStore) folderPathFilePathFromFoldersFilename(folders string, filename string) (string, string) {
	sep := string(os.PathSeparator)
	if folders == "" {
		return us.folderPath,
			us.folderPath + sep + "BinOverflows" + sep + filename + ".ovf"
	} else {
		return us.folderPath + sep + folders,
			us.folderPath + sep + "BinOverflows" + sep + folders + sep + filename + ".ovf"
	}
}

func (us *UniformHashStore) addressForHash(hash *Sha256) uint64 {
	hashLSBs := binary.LittleEndian.Uint64(hash[0:8])
	dividedHash := hashLSBs / us.hashDivider
	return dividedHash
}

func (us *UniformHashStore) folderPathFilePathForAddress(address uint64) (string, string) {
	folders, filename, _ := us.numberedFolders.NumberToFoldersAndFile(int64(address))
	return us.folderPathFilePathFromFoldersFilename(folders, filename)
}

func (us *UniformHashStore) AppendHash(hash *Sha256) (int64, error) {
	newIndex := us.hashFile.CountHashes()
	err := us.hashFile.WriteHashAt(*hash, newIndex)
	if err != nil {
		return -1, err
	}

	// Appending index followed by hash
	toAppend := [40]byte{}
	binary.LittleEndian.PutUint64(toAppend[0:8], uint64(newIndex))
	copy(toAppend[8:40], hash[0:32])

	address := us.addressForHash(hash)
	folderPath, filePath := us.folderPathFilePathForAddress(address)
	err = os.MkdirAll(folderPath, 0755)
	if err != nil {
		return -1, err
	}
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		return -1, err
	}
	_, err = file.Write(toAppend[:])
	if err != nil {
		return -1, err
	}
	err = file.Sync()
	if err != nil {
		return -1, err
	}
	err = file.Close()
	if err != nil {
		return -1, err
	}
	file = nil

	return newIndex, nil
}

func (us *UniformHashStore) IndexOfHash(hash *Sha256) (int64, error) {
	address := us.addressForHash(hash)

	// First look in BinStarts.bst
	offset := address * binStartSize
	byts := [binStartSize]byte{}
	_, err := us.binStartsFile.ReadAt(byts[:], int64(offset))
	if err != nil {
		fmt.Println("Error when: IndexOfHash() reading from ", offset, " of binStartsFile")
		return -1, err
	}
	bin := binStart{}
	bin.FromBytes(&byts)

	entries := bin.entryCount
	if entries > 102 {
		entries = 102 // Only the first 102 are in the binstarts file
	}
	for i := int64(0); i < entries; i++ {
		rumouredIndex := bin.indexHashes[i].index
		rumouredHash := Sha256{}
		err := us.GetHashAtIndex(rumouredIndex, &rumouredHash)
		if err != nil {
			return -1, err
		}
		if *hash == rumouredHash {
			return rumouredIndex, nil
		}
	}

	// Second look in overflow file
	if bin.entryCount > 102 {
		// Look in the overflow files
		_, filepath := us.folderPathFilePathForAddress(address)
		overflowFile, err := os.ReadFile(filepath)
		if err != nil {
			return -1, err
		}
		for i := int64(0); i < bin.entryCount-102; i++ {
			rumouredIndex := int64(binary.LittleEndian.Uint64(overflowFile[i*8 : i*8+8]))
			rumouredHash := Sha256{}
			err := us.GetHashAtIndex(rumouredIndex, &rumouredHash)
			if err != nil {
				return -1, err
			}
			if *hash == rumouredHash {
				return rumouredIndex, nil
			}
		}
	}

	return -1, nil
}

func (us *UniformHashStore) GetHashAtIndex(index int64, hash *Sha256) error {
	var err error
	*hash, err = us.hashFile.ReadHashAt(index)
	return err
}

func (us *UniformHashStore) CountHashes() (int64, error) {
	return us.hashFile.CountHashes(), nil
}

func (us *UniformHashStore) Close() error {
	err := us.hashFile.Close()
	if err != nil {
		return err
	}
	err = us.binStartsFile.Close()
	if err != nil {
		return err
	}
	return nil
}

func (us *UniformHashStore) Sync() error {
	return us.hashFile.Sync()
}

func (us *UniformHashStore) Test() (bool, error) {
	count, err := us.CountHashes()
	if err != nil {
		return false, err
	}
	// Percentage tested: 67.10946496250655
	// Repeated hash at  729158472 ,  -1
	// Hash mismatch at  729158472 ,  -1
	//for i := int64(729158472); i < count; i++ {
	for i := int64(0); i < count; i++ {
		hash := Sha256{}
		err := us.GetHashAtIndex(i, &hash)
		if err != nil {
			return false, err
		}

		index, err := us.IndexOfHash(&hash)
		if err != nil {
			return false, err
		}
		// Because we support repeated hashes, index might not be equal to i
		// But mention it anyway
		if index != i {
			fmt.Println("Repeated hash at ", i, ", ", index)
		}
		hash2 := Sha256{}
		err = us.GetHashAtIndex(index, &hash2)
		for j := 0; j < 32; j++ {
			if hash[j] != hash2[j] {
				fmt.Println("Hash mismatch at ", i, ", ", index)
				return false, nil
			}
		}

		if i%(count/100) == 0 {
			percent := 100.0 * float64(i) / float64(count)
			fmt.Println("Percentage tested:", percent)
		}
	}
	return true, nil
}
