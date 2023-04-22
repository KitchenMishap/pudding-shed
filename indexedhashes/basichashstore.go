package indexedhashes

import (
	"encoding/binary"
	"io"
)

type ReadWriteSeekCloser interface {
	io.ReadWriteSeeker
	io.Closer
}

type BasicHashStore struct {
	file ReadWriteSeekCloser
}

func NewBasicHashStore(file ReadWriteSeekCloser) *BasicHashStore {
	result := BasicHashStore{}
	result.file = file
	return &result
}

func (bhs *BasicHashStore) AppendHash(hash *Sha256) (int64, error) {
	bytecount, err := bhs.file.Seek(0, io.SeekEnd)
	if err != nil {
		return -1, err
	}
	_, err = bhs.file.Write(hash[0:32])
	if err != nil {
		return -1, err
	}
	return bytecount / 32, nil
}

// IndexOfHash This is a very slow naive implementation, and should only be used for testing
func (bhs *BasicHashStore) IndexOfHash(hash *Sha256) (int64, error) {
	_, err := bhs.file.Seek(0, io.SeekStart)
	if err != nil {
		return -1, err
	}
	var hashInFile Sha256
	index := int64(0)
	for {
		bytecount, err := bhs.file.Read(hashInFile[0:32])
		if bytecount == 0 || err != nil {
			return int64(-1), err
		}
		if hashInFile == *hash {
			return index, nil
		}
		index++
	}
}

func (bhs *BasicHashStore) GetHashAtIndex(index int64, hash *Sha256) error {
	_, err := bhs.file.Seek(32*index, io.SeekStart)
	if err != nil {
		return err
	}
	_, err = bhs.file.Read(hash[0:32])
	return err
}

func (bhs *BasicHashStore) CountHashes() (int64, error) {
	bytecount, err := bhs.file.Seek(0, io.SeekEnd)
	if err != nil {
		return -1, err
	}
	return bytecount / 32, nil
}

func (bhs *BasicHashStore) Close() error {
	err := bhs.file.Close()
	bhs.file = nil
	return err
}

func (bhs *BasicHashStore) WholeFileAsInt32() ([]uint32, error) {
	entries, err := bhs.CountHashes()
	if err != nil {
		println("Could not count hashes")
		return nil, err
	}
	println("Reading")
	raw := make([]byte, entries*32)
	bhs.file.Seek(0, 0)
	tot := int64(0)
	for {
		println("Reading chunk")
		n, err := bhs.file.Read(raw[tot : entries*32])
		if err != nil {
			println("Could not read raw from hashes file")
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
