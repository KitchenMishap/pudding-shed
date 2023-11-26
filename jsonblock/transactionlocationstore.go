package jsonblock

import (
	"github.com/KitchenMishap/pudding-shed/indexedhashes"
	"github.com/KitchenMishap/pudding-shed/wordfile"
)

// transLocationStore implements ITransLocatorByHash
var _ ITransLocatorByHash = (*transLocationStore)(nil) // Check that implements

type transLocationStore struct {
	transHashStore  indexedhashes.HashReadWriter
	blockHeightFile wordfile.ReadWriteAtWordCounter
	nthTransFile    wordfile.ReadWriteAtWordCounter
}

// transLocation implements ITransIndicesPath
var _ ITransIndicesPath = (*transLocation)(nil) // Check that implements

type transLocation struct {
	blockHeight int64
	nthTrans    int64
}

func (tl *transLocation) BlockHeight() int64     { return tl.blockHeight }
func (tl *transLocation) NthTransInBlock() int64 { return tl.nthTrans }

// functions in TransactionLocatorStore to implement ITransactionLocatorByHash

func (tls *transLocationStore) GetTransIndicesPathByHash(sha256 indexedhashes.Sha256) (ITransIndicesPath, error) {
	transIndex, err := tls.transHashStore.IndexOfHash(&sha256)
	if err != nil {
		return nil, err
	}
	blockHeight, err := tls.blockHeightFile.ReadWordAt(transIndex)
	if err != nil {
		return nil, err
	}
	nthTrans, err := tls.nthTransFile.ReadWordAt(transIndex)
	if err != nil {
		return nil, err
	}

	location := transLocation{}
	location.blockHeight = blockHeight
	location.nthTrans = nthTrans
	return &location, nil
}

// functions in TransactionLocatorStore to implement ITransactionLocatorStore

func (tls *transLocationStore) StoreIndicesPathForHash(sha256 indexedhashes.Sha256, blockHeight int64, nthTransInBlock int64) error {
	// Is it already there?
	index, err := tls.transHashStore.IndexOfHash(&sha256)
	if err != nil {
		return err
	}
	if index != -1 {
		return nil // Already stored
	}

	transIndex, err := tls.transHashStore.AppendHash(&sha256)
	if err != nil {
		return err
	}
	err = tls.blockHeightFile.WriteWordAt(blockHeight, transIndex)
	if err != nil {
		return err
	}
	err = tls.nthTransFile.WriteWordAt(nthTransInBlock, transIndex)
	return err
}

func CreateOpenTransLocationStore(folderName string) ITransLocatorStore {
	obj := transLocationStore{}

	hashStoreCreator, err := indexedhashes.NewConcreteHashStoreCreator(
		"Transactions", folderName, 30, 4, 3)
	if err != nil {
		panic(err)
	}
	if !hashStoreCreator.HashStoreExists() {
		err := hashStoreCreator.CreateHashStore()
		if err != nil {
			panic(err)
		}
	}
	hashStore, err := hashStoreCreator.OpenHashStore()
	if err != nil {
		panic(err)
	}
	obj.transHashStore = hashStore

	wordFileCreator := wordfile.NewConcreteWordFileCreator("BlockHeights", folderName, 3)
	if !wordFileCreator.WordFileExists() {
		err := wordFileCreator.CreateWordFile()
		if err != nil {
			panic(err)
		}
	}
	wordFile, err := wordFileCreator.OpenWordFile()
	if err != nil {
		panic(err)
	}
	obj.blockHeightFile = wordFile

	// Assume maximum of 65536 transactions per block (equivalent 16 bytes per transaction, not going to happen!)
	wordFileCreator2 := wordfile.NewConcreteWordFileCreator("NthTrans", folderName, 2)
	if !wordFileCreator2.WordFileExists() {
		err := wordFileCreator2.CreateWordFile()
		if err != nil {
			panic(err)
		}
	}
	wordFile2, err := wordFileCreator2.OpenWordFile()
	if err != nil {
		panic(err)
	}
	obj.nthTransFile = wordFile2

	return &obj
}
