package transactionindexing

import "github.com/KitchenMishap/pudding-shed/indexedhashes"

type ITransactionIndexer interface {
	StoreTransHashToHeight(sha256 *indexedhashes.Sha256, transHeight int64) error
	StoreTransHeightToParentBlock(transHeight int64, parentBlockHeight int64) error
	StoreBlockHeightToFirstTrans(blockHeight int64, firstTrans int64) error
	RetrieveTransHashToHeight(sha256 *indexedhashes.Sha256) (int64, error)
	RetrieveTransHeightToParentBlock(transHeight int64) (int64, error)
	RetrieveBlockHeightToFirstTrans(blockHeight int64) (int64, error)
	Close()
}
