package chainstorage

import (
	"errors"
	"github.com/KitchenMishap/pudding-shed/indexedhashes"
	"github.com/KitchenMishap/pudding-shed/wordfile"
	"path"
)

type ConcreteAppendableChainCreator struct {
	blocksFolder                 string
	transactionsFolder           string
	transactionInputsFolder      string
	transactionOutputsFolder     string
	blockHashStoreCreator        *indexedhashes.ConcreteHashStoreCreator
	transactionHashStoreCreator  *indexedhashes.ConcreteHashStoreCreator
	blkFirstTransWordFileCreator *wordfile.ConcreteWordFileCreator
	trnFirstTxiWordFileCreator   *wordfile.ConcreteWordFileCreator
	trnFirstTxoWordFileCreator   *wordfile.ConcreteWordFileCreator
	txiTxWordFileCreator         *wordfile.ConcreteWordFileCreator
	txiVoutWordFileCreator       *wordfile.ConcreteWordFileCreator
	txoSatsWordFileCreator       *wordfile.ConcreteWordFileCreator
}

func NewConcreteAppendableChainCreator(folder string) (*ConcreteAppendableChainCreator, error) {
	result := ConcreteAppendableChainCreator{}

	result.blocksFolder = path.Join(folder, "Blocks")
	result.transactionsFolder = path.Join(folder, "Transactions")
	result.transactionInputsFolder = path.Join(folder, "TransactionInputs")
	result.transactionOutputsFolder = path.Join(folder, "TransactionOutputs")
	var err error
	result.blockHashStoreCreator, err = indexedhashes.NewConcreteHashStoreCreator(
		"Blocks", result.blocksFolder, 30, 3, 3)
	if err != nil {
		return nil, err
	}
	result.transactionHashStoreCreator, err = indexedhashes.NewConcreteHashStoreCreator(
		"Transactions", result.transactionsFolder, 30, 4, 3)
	if err != nil {
		return nil, err
	}
	result.blkFirstTransWordFileCreator = wordfile.NewConcreteWordFileCreator("firsttrans", result.blocksFolder, 5)
	result.trnFirstTxiWordFileCreator = wordfile.NewConcreteWordFileCreator("firsttxi", result.transactionsFolder, 5)
	result.trnFirstTxoWordFileCreator = wordfile.NewConcreteWordFileCreator("firsttxo", result.transactionsFolder, 5)
	result.txiTxWordFileCreator = wordfile.NewConcreteWordFileCreator("tx", result.transactionInputsFolder, 4)
	result.txiVoutWordFileCreator = wordfile.NewConcreteWordFileCreator("vout", result.transactionInputsFolder, 4)
	result.txoSatsWordFileCreator = wordfile.NewConcreteWordFileCreator("value", result.transactionOutputsFolder, 8)
	return &result, nil
}

func (cacc *ConcreteAppendableChainCreator) Exists() bool {
	return cacc.blockHashStoreCreator.HashStoreExists()
}

func (cacc *ConcreteAppendableChainCreator) Create() error {
	if cacc.Exists() {
		return errors.New("AppendableChain already exists")
	}
	err := cacc.blockHashStoreCreator.CreateHashStore()
	if err != nil {
		return err
	}
	err = cacc.transactionHashStoreCreator.CreateHashStore()
	if err != nil {
		return err
	}
	err = cacc.blkFirstTransWordFileCreator.CreateWordFile()
	if err != nil {
		return err
	}
	err = cacc.trnFirstTxiWordFileCreator.CreateWordFile()
	if err != nil {
		return err
	}
	err = cacc.trnFirstTxoWordFileCreator.CreateWordFile()
	if err != nil {
		return err
	}
	err = cacc.txiTxWordFileCreator.CreateWordFile()
	if err != nil {
		return err
	}
	err = cacc.txiVoutWordFileCreator.CreateWordFile()
	if err != nil {
		return err
	}
	err = cacc.txoSatsWordFileCreator.CreateWordFile()
	if err != nil {
		return err
	}
	return nil
}

func (cacc *ConcreteAppendableChainCreator) Open() (IAppendableChain, error) {
	result := concreteAppendableChain{}
	var err error
	result.blkHashes, err = cacc.blockHashStoreCreator.OpenHashStore()
	if err != nil {
		return nil, err
	}
	result.trnHashes, err = cacc.transactionHashStoreCreator.OpenHashStore()
	if err != nil {
		result.blkHashes.Close()
		return nil, err
	}
	result.blkFirstTrans, err = cacc.blkFirstTransWordFileCreator.OpenWordFile()
	if err != nil {
		result.blkHashes.Close()
		result.trnHashes.Close()
		return nil, err
	}
	result.trnFirstTxi, err = cacc.trnFirstTxiWordFileCreator.OpenWordFile()
	if err != nil {
		result.blkHashes.Close()
		result.trnHashes.Close()
		result.blkFirstTrans.Close()
		return nil, err
	}
	result.trnFirstTxo, err = cacc.trnFirstTxoWordFileCreator.OpenWordFile()
	if err != nil {
		result.blkHashes.Close()
		result.trnHashes.Close()
		result.blkFirstTrans.Close()
		result.trnFirstTxi.Close()
		return nil, err
	}
	result.txiTx, err = cacc.txiTxWordFileCreator.OpenWordFile()
	if err != nil {
		result.blkHashes.Close()
		result.trnHashes.Close()
		result.blkFirstTrans.Close()
		result.trnFirstTxi.Close()
		result.trnFirstTxo.Close()
		return nil, err
	}
	result.txiVout, err = cacc.txiVoutWordFileCreator.OpenWordFile()
	if err != nil {
		result.blkHashes.Close()
		result.trnHashes.Close()
		result.blkFirstTrans.Close()
		result.trnFirstTxi.Close()
		result.trnFirstTxo.Close()
		result.txiTx.Close()
		return nil, err
	}
	result.txoSats, err = cacc.txoSatsWordFileCreator.OpenWordFile()
	if err != nil {
		result.blkHashes.Close()
		result.trnHashes.Close()
		result.blkFirstTrans.Close()
		result.trnFirstTxi.Close()
		result.trnFirstTxo.Close()
		result.txiTx.Close()
		result.txiVout.Close()
		return nil, err
	}
	return &result, nil
}
