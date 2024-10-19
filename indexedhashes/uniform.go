package indexedhashes

import (
	"encoding/binary"
	"errors"
	"github.com/KitchenMishap/pudding-shed/memfile"
	"github.com/KitchenMishap/pudding-shed/numberedfolders"
	"github.com/KitchenMishap/pudding-shed/wordfile"
	"math"
	"os"
)

func NewUniformHashStoreCreator(hashCountEstimate int64,
	folder string, name string, digitsPerFolder int) HashStoreCreator {
	return NewUniformHashStoreCreatorPrivate(hashCountEstimate,
		folder, name, digitsPerFolder)
}

func NewUniformHashStoreCreatorPrivate(hashCountEstimate int64,
	folder string, name string, digitsPerFolder int) *UniformHashStoreCreator {
	result := UniformHashStoreCreator{}
	result.folder = folder
	result.name = name
	result.digitsPerFolder = digitsPerFolder

	// Poisson distribution for lambda = 75:
	// https://homepage.divms.uiowa.edu/~mbognar/applets/pois.html
	// P(X >= 102) = 0.00175
	// So for a 102-entry system, with targetEntriesPerFile = 75, 0.175% of entries will end up in overflow files
	// (102*40 is about 4096 with 16 bytes spare; 4096 is a typical hard drive allocation unit)
	const targetEntriesPerFile = 75
	targetNumFiles := hashCountEstimate / targetEntriesPerFile
	possibleHashes := math.Pow(2, 64) // Because taking 64 LS bits of hash
	result.hashDivider = uint64(possibleHashes / float64(targetNumFiles))

	return &result
}

type UniformHashStoreCreator struct {
	folder          string
	name            string
	digitsPerFolder int
	hashDivider     uint64
}

func (uc *UniformHashStoreCreator) folderPath() string {
	sep := string(os.PathSeparator)
	return uc.folder + sep + uc.name
}

func (uc *UniformHashStoreCreator) hashFilePath() string {
	sep := string(os.PathSeparator)
	return uc.folderPath() + sep + "Hashes.hsh"
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
	return err
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
	store.hashDivider = uc.hashDivider
	store.numberedFolders = numberedfolders.NewNumberedFolders(0, uc.digitsPerFolder)

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
}

func (us *UniformHashStore) folderPathFilePathFromFoldersFilename(folders string, filename string) (string, string) {
	sep := string(os.PathSeparator)
	if folders == "" {
		return us.folderPath,
			us.folderPath + sep + filename + ".idx"
	} else {
		return us.folderPath + sep + folders,
			us.folderPath + sep + folders + sep + filename + ".idx"
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
	file.Sync()
	file.Close()
	file = nil

	return newIndex, nil
}

func (us *UniformHashStore) IndexOfHash(hash *Sha256) (int64, error) {
	address := us.addressForHash(hash)
	_, filePath := us.folderPathFilePathForAddress(address)

	contents, err := os.ReadFile(filePath)
	if err != nil {
		return -1, nil
	} // Just means hash doesn't exists; not an error

	spare := len(contents) % 40
	if spare != 0 {
		return -1, errors.New("indices file is not a multiple of 40 bytes")
	}
	entries := len(contents) / 40

	for i := 0; i < entries; i++ {
		offset := i * 40
		byteOffset := int(-1)
		for byteOffset = 0; byteOffset < 32; byteOffset++ {
			if contents[offset+8+byteOffset] != hash[byteOffset] {
				// Not a match
				break
			}
			if byteOffset == 31 {
				// A match, read the index
				u64 := binary.LittleEndian.Uint64(contents[offset : offset+8])
				return int64(u64), nil
			}
		}
	}
	return -1, nil // No hash found, but not an error
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
	return us.hashFile.Close()
}

func (us *UniformHashStore) Sync() error {
	return us.hashFile.Sync()
}
