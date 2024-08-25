package jsonblock

import (
	"errors"
	"github.com/KitchenMishap/pudding-shed/indexedhashes"
	"github.com/KitchenMishap/pudding-shed/transactionindexing"
	"github.com/KitchenMishap/pudding-shed/wordfile"
)

// transactionIndexerFiles implements ITransactionIndexer using its own files
var _ transactionindexing.ITransactionIndexer = (*transactionIndexerFiles)(nil) // Check that implements

type transactionIndexerFiles struct {
	transHashStore   indexedhashes.HashReadWriter
	transParentBlock wordfile.ReadWriteAtWordCounter
	blockFirstTrans  wordfile.ReadWriteAtWordCounter
}

// Functions in transactionIndexerFiles to implement ITransactionIndexer

func (tif *transactionIndexerFiles) StoreTransHashToHeight(sha256 *indexedhashes.Sha256, transHeight int64) error {
	transIndex, err := tif.transHashStore.AppendHash(sha256)
	if err != nil {
		return err
	}
	if transIndex != transHeight {
		return errors.New("must not store transaction hashes out of sequence")
	}
	return nil
}
func (tif *transactionIndexerFiles) StoreTransHeightToParentBlock(transHeight int64, parentBlockHeight int64) error {
	return tif.transParentBlock.WriteWordAt(parentBlockHeight, transHeight)
}
func (tif *transactionIndexerFiles) StoreBlockHeightToFirstTrans(blockHeight int64, firstTrans int64) error {
	return tif.blockFirstTrans.WriteWordAt(firstTrans, blockHeight)
}
func (tif *transactionIndexerFiles) RetrieveTransHashToHeight(sha256 *indexedhashes.Sha256) (int64, error) {
	return tif.transHashStore.IndexOfHash(sha256)
}
func (tif *transactionIndexerFiles) RetrieveTransHeightToParentBlock(transHeight int64) (int64, error) {
	return tif.transParentBlock.ReadWordAt(transHeight)
}
func (tif *transactionIndexerFiles) RetrieveBlockHeightToFirstTrans(blockHeight int64) (int64, error) {
	return tif.blockFirstTrans.ReadWordAt(blockHeight)
}

func CreateOpenTransactionIndexerFiles(folderName string) transactionindexing.ITransactionIndexer {
	obj := transactionIndexerFiles{}

	hashStoreCreator, err := indexedhashes.NewConcreteHashStoreCreator(
		"Transactions", folderName, 30, 4, 3, false)
	if err != nil {
		panic(err)
	}
	err = hashStoreCreator.CreateHashStore()
	if err != nil {
		panic(err)
	}
	hashStore, err := hashStoreCreator.OpenHashStore()
	if err != nil {
		panic(err)
	}
	obj.transHashStore = hashStore

	wordFileCreator := wordfile.NewConcreteWordFileCreator("transParentBlock", folderName, 3)
	err = wordFileCreator.CreateWordFile()
	if err != nil {
		panic(err)
	}
	wordFile, err := wordFileCreator.OpenWordFile()
	if err != nil {
		panic(err)
	}
	obj.transParentBlock = wordFile

	wordFileCreator2 := wordfile.NewConcreteWordFileCreator("blkFirstTran", folderName, 5)
	err = wordFileCreator2.CreateWordFile()
	if err != nil {
		panic(err)
	}
	wordFile2, err := wordFileCreator2.OpenWordFile()
	if err != nil {
		panic(err)
	}
	obj.blockFirstTrans = wordFile2

	return &obj
}

func (tif *transactionIndexerFiles) Close() {
	tif.transHashStore.Close()
	tif.transParentBlock.Close()
	tif.blockFirstTrans.Close()
}
