package indexedhashes

import (
	"github.com/KitchenMishap/pudding-shed/wordfile"
)

type MemoryIndexedHashes struct {
	theMap   map[Sha256]int64
	theStore *BasicHashStore
	count    int64
}

func NewMemoryIndexedHashes(hashFile *wordfile.HashFile) *MemoryIndexedHashes {
	result := MemoryIndexedHashes{}
	result.theMap = make(map[Sha256]int64)
	result.theStore = NewBasicHashStore(hashFile)
	return &result
}

func (mih *MemoryIndexedHashes) AppendHash(hash *Sha256) (int64, error) {
	index, err := mih.theStore.AppendHash(hash)
	if err != nil {
		return -1, err
	}
	mih.theMap[*hash] = mih.count
	mih.count++
	return index, nil
}

func (mih *MemoryIndexedHashes) Sync() error {
	return mih.theStore.Sync()
}

func (mih *MemoryIndexedHashes) IndexOfHash(hash *Sha256) (int64, error) {
	val, ok := mih.theMap[*hash]
	if ok {
		return val, nil
	} else {
		return -1, nil // Not a proper error
	}
}

func (mih *MemoryIndexedHashes) GetHashAtIndex(index int64, hash *Sha256) error {
	return mih.theStore.GetHashAtIndex(index, hash)
}

func (mih *MemoryIndexedHashes) CountHashes() (int64, error) {
	return mih.count, nil
}

func (mih *MemoryIndexedHashes) Close() error {
	return mih.theStore.Close()
}
