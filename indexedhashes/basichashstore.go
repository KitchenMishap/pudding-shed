package indexedhashes

import (
	"github.com/KitchenMishap/pudding-shed/memfile"
	"io"
	"log"
)

type ReadWriteSeekCloser interface {
	io.ReadWriteSeeker
	io.Closer
}

type BasicHashStore struct {
	file memfile.LookupFileWithSize
}

func NewBasicHashStore(file memfile.LookupFileWithSize) *BasicHashStore {
	result := BasicHashStore{}
	result.file = file
	return &result
}

func (bhs *BasicHashStore) AppendHash(hash *Sha256) (int64, error) {
	fi, err := bhs.file.Stat()
	if err != nil {
		return -1, err
	}
	bytecount := fi.Size()
	_, err = bhs.file.WriteAt(hash[0:32], bytecount)
	if err != nil {
		log.Println(err)
		log.Println("AppendHash(): Could not call file.WriteAt()")
		return -1, err
	}
	return bytecount / 32, nil
}

/*
// IndexOfHash This is a very slow naive implementation, and should only be used for testing

	func (bhs *BasicHashStore) IndexOfHash(hash *Sha256) (int64, error) {
		_, err := bhs.file.Seek(0, io.SeekStart)
		if err != nil {
			log.Println(err)
			log.Println("IndexOfHash(): Could not call file.Seek()")
			return -1, err
		}
		var hashInFile Sha256
		index := int64(0)
		for {
			bytecount, err := bhs.file.Read(hashInFile[0:32])
			if bytecount == 0 || err != nil {
				log.Println(err)
				log.Println("IndexOfHash(): file.Read() did not read any bytes")
				return int64(-1), err
			}
			if hashInFile == *hash {
				return index, nil
			}
			index++
		}
	}
*/
func (bhs *BasicHashStore) GetHashAtIndex(index int64, hash *Sha256) error {
	_, err := bhs.file.ReadAt(hash[0:32], 32*index)
	if err != nil {
		log.Println(err)
		log.Println("GetHashAtIndex(): Could not call file.Read()")
	}
	return err
}

func (bhs *BasicHashStore) CountHashes() (int64, error) {
	fi, err := bhs.file.Stat()
	if err != nil {
		return -1, err
	}
	bytecount := fi.Size()
	return bytecount / 32, nil
}

func (bhs *BasicHashStore) Close() error {
	err := bhs.file.Close()
	if err != nil {
		log.Println(err)
		log.Println("Close(): Could not call file.Close()")
		return err
	}
	bhs.file = nil
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
	return bhs.file.Sync()
}
