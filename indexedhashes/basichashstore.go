package indexedhashes

import (
	"github.com/KitchenMishap/pudding-shed/wordfile"
	"io"
)

type ReadWriteSeekCloser interface {
	io.ReadWriteSeeker
	io.Closer
}

type BasicHashStore struct {
	hashFile *wordfile.HashFile
}

func NewBasicHashStore(hashFile *wordfile.HashFile) *BasicHashStore {
	result := BasicHashStore{}
	result.hashFile = hashFile
	return &result
}

func (bhs *BasicHashStore) AppendHash(hash *Sha256) (int64, error) {
	hashCount := bhs.hashFile.CountHashes()
	err := bhs.hashFile.WriteHashAt(*hash, hashCount)
	if err != nil {
		return -1, err
	}
	return hashCount, nil
}

// IndexOfHash This is a very slow naive implementation, and should only be used for testing
func (bhs *BasicHashStore) IndexOfHash(hash *Sha256) (int64, error) {
	hashCount := bhs.hashFile.CountHashes()
	for index := int64(0); index < hashCount; index++ {
		hashInFile, err := bhs.hashFile.ReadHashAt(index)
		if err != nil {
			return -1, err
		}
		if hashInFile == *hash {
			return index, nil
		}
	}
	return -1, nil
}

func (bhs *BasicHashStore) GetHashAtIndex(index int64, hash *Sha256) error {
	theHash, err := bhs.hashFile.ReadHashAt(index)
	*hash = theHash
	return err
}

func (bhs *BasicHashStore) CountHashes() (int64, error) {
	return bhs.hashFile.CountHashes(), nil
}

func (bhs *BasicHashStore) Close() error {
	err := bhs.hashFile.Close()
	if err != nil {
		return err
	}
	return nil
}

/*
func (bhs *BasicHashStore) WholeFileAsInt32() ([]uint32, error) {
	entries, err := bhs.CountHashes()
	if err != nil {
		log.Println("WholeFileAsInt32(): Could not call CountHashes()")
		return nil, err
	}
	println("Reading")
	raw := make([]byte, entries*32)
	_, err = bhs.file.Seek(0, 0)
	if err != nil {
		log.Println(err)
		log.Println("WholeFileAsInt32(): Could not call file.Seek()")
		return nil, err
	}
	tot := int64(0)
	for {
		println("Reading chunk")
		n, err := bhs.file.Read(raw[tot : entries*32])
		if err != nil {
			log.Println(err)
			log.Println("WholeFileAsInt32(): Could not call file.Read()")
			return nil, err
		}
		tot += int64(n)
		if tot == entries*32 {
			break
		}
	}

	println("Parsing")

	arr := make([]uint32, entries)
	for i := int64(0); i < entries; i++ {
		start := 32 * i
		var intBytes [4]byte
		// LittleEndian, and we want MSBs, so read from last 4 bytes
		intBytes[0] = raw[start+28]
		intBytes[1] = raw[start+29]
		intBytes[2] = raw[start+30]
		intBytes[3] = raw[start+31]
		arr[i] = binary.LittleEndian.Uint32(intBytes[0:4])
	}
	return arr, nil
}
*/

func (bhs *BasicHashStore) Sync() error {
	return bhs.hashFile.Sync()
}
