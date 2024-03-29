package chainstorage

import (
	"errors"
	"github.com/KitchenMishap/pudding-shed/indexedhashes"
	"github.com/KitchenMishap/pudding-shed/wordfile"
	"path"
	"slices"
)

type ConcreteAppendableChainCreator struct {
	blocksFolder                  string
	transactionsFolder            string
	transactionInputsFolder       string
	transactionOutputsFolder      string
	blockHashStoreCreator         *indexedhashes.ConcreteHashStoreCreator
	transactionHashStoreCreator   *indexedhashes.ConcreteHashStoreCreator
	blkFirstTransWordFileCreator  *wordfile.ConcreteWordFileCreator
	trnParentBlockWordFileCreator *wordfile.ConcreteWordFileCreator
	trnFirstTxiWordFileCreator    *wordfile.ConcreteWordFileCreator
	trnFirstTxoWordFileCreator    *wordfile.ConcreteWordFileCreator
	txiTxWordFileCreator          *wordfile.ConcreteWordFileCreator
	txiVoutWordFileCreator        *wordfile.ConcreteWordFileCreator
	txoSatsWordFileCreator        *wordfile.ConcreteWordFileCreator
	supportedBlkNeis              map[string]int
	supportedTrnNeis              map[string]int
}

func NewConcreteAppendableChainCreator(
	folder string) (*ConcreteAppendableChainCreator, error) {
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
	result.trnParentBlockWordFileCreator = wordfile.NewConcreteWordFileCreator("parentblock", result.transactionsFolder, 3)
	result.trnFirstTxiWordFileCreator = wordfile.NewConcreteWordFileCreator("firsttxi", result.transactionsFolder, 5)
	result.trnFirstTxoWordFileCreator = wordfile.NewConcreteWordFileCreator("firsttxo", result.transactionsFolder, 5)
	result.txiTxWordFileCreator = wordfile.NewConcreteWordFileCreator("tx", result.transactionInputsFolder, 4)
	result.txiVoutWordFileCreator = wordfile.NewConcreteWordFileCreator("vout", result.transactionInputsFolder, 4)
	result.txoSatsWordFileCreator = wordfile.NewConcreteWordFileCreator("value", result.transactionOutputsFolder, 8)

	result.supportedBlkNeis = map[string]int{
		"version":      4,
		"time":         4,
		"mediantime":   4,
		"nonce":        4,
		"difficulty":   8, // Actually real not int, but we truncate to integer part. We might still run out of bytes
		"strippedsize": 4,
		"size":         4,
		"weight":       4,
	}
	result.supportedTrnNeis = map[string]int{
		"version":  4,
		"size":     4,
		"vsize":    4,
		"weight":   4,
		"locktime": 4,
	}

	return &result, nil
}

func (cacc *ConcreteAppendableChainCreator) Exists() bool {
	return cacc.blockHashStoreCreator.HashStoreExists()
}

func (cacc *ConcreteAppendableChainCreator) Create(blkNeiNames []string, trnNeiNames []string) error {
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
	err = cacc.trnParentBlockWordFileCreator.CreateWordFile()
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

	for supportedName, size := range cacc.supportedBlkNeis {
		if slices.Contains(blkNeiNames, supportedName) {
			blkNonEssentialIntCreator := wordfile.NewConcreteWordFileCreator(supportedName, cacc.blocksFolder, int64(size))
			err = blkNonEssentialIntCreator.CreateWordFile()
			if err != nil {
				return err
			}
		}
	}
	for _, requestedName := range blkNeiNames {
		_, supported := cacc.supportedBlkNeis[requestedName]
		if !supported {
			return errors.New(requestedName + " is not a supported block NonEssentialInt")
		}
	}

	for supportedName, size := range cacc.supportedTrnNeis {
		if slices.Contains(trnNeiNames, supportedName) {
			trnNonEssentialIntCreator := wordfile.NewConcreteWordFileCreator(supportedName, cacc.transactionsFolder, int64(size))
			err = trnNonEssentialIntCreator.CreateWordFile()
			if err != nil {
				return err
			}
		}
	}
	for _, requestedName := range trnNeiNames {
		_, supported := cacc.supportedTrnNeis[requestedName]
		if !supported {
			return errors.New(requestedName + " is not a supported transaction NonEssentialInt")
		}
	}

	return nil
}

func (cacc *ConcreteAppendableChainCreator) Open(transactionIndexingToBeDelegated bool) (IAppendableChain,
	*concreteAppendableChain, error) {
	result := concreteAppendableChain{}
	result.transactionIndexingIsDelegated = transactionIndexingToBeDelegated

	var err error
	result.blkHashes, err = cacc.blockHashStoreCreator.OpenHashStore()
	if err != nil {
		return nil, nil, err
	}
	result.trnHashes, err = cacc.transactionHashStoreCreator.OpenHashStore()
	if err != nil {
		result.blkHashes.Close()
		return nil, nil, err
	}
	result.blkFirstTrans, err = cacc.blkFirstTransWordFileCreator.OpenWordFile()
	if err != nil {
		result.blkHashes.Close()
		result.trnHashes.Close()
		return nil, nil, err
	}
	result.trnParentBlock, err = cacc.trnParentBlockWordFileCreator.OpenWordFile()
	if err != nil {
		result.blkHashes.Close()
		result.trnHashes.Close()
		result.blkFirstTrans.Close()
		return nil, nil, err
	}
	result.trnFirstTxi, err = cacc.trnFirstTxiWordFileCreator.OpenWordFile()
	if err != nil {
		result.blkHashes.Close()
		result.trnHashes.Close()
		result.blkFirstTrans.Close()
		result.trnParentBlock.Close()
		return nil, nil, err
	}
	result.trnFirstTxo, err = cacc.trnFirstTxoWordFileCreator.OpenWordFile()
	if err != nil {
		result.blkHashes.Close()
		result.trnHashes.Close()
		result.blkFirstTrans.Close()
		result.trnParentBlock.Close()
		result.trnFirstTxi.Close()
		return nil, nil, err
	}
	result.txiTx, err = cacc.txiTxWordFileCreator.OpenWordFile()
	if err != nil {
		result.blkHashes.Close()
		result.trnHashes.Close()
		result.blkFirstTrans.Close()
		result.trnParentBlock.Close()
		result.trnFirstTxi.Close()
		result.trnFirstTxo.Close()
		return nil, nil, err
	}
	result.txiVout, err = cacc.txiVoutWordFileCreator.OpenWordFile()
	if err != nil {
		result.blkHashes.Close()
		result.trnHashes.Close()
		result.blkFirstTrans.Close()
		result.trnParentBlock.Close()
		result.trnFirstTxi.Close()
		result.trnFirstTxo.Close()
		result.txiTx.Close()
		return nil, nil, err
	}
	result.txoSats, err = cacc.txoSatsWordFileCreator.OpenWordFile()
	if err != nil {
		result.blkHashes.Close()
		result.trnHashes.Close()
		result.blkFirstTrans.Close()
		result.trnParentBlock.Close()
		result.trnFirstTxi.Close()
		result.trnFirstTxo.Close()
		result.txiTx.Close()
		result.txiVout.Close()
		return nil, nil, err
	}

	result.blkNonEssentialInts = make(map[string]wordfile.ReadWriteAtWordCounter)
	// We try to open each of the supported block NonEssentialInt wordfiles.
	// However they do not need to exist, and if they're not there we don't error.
	for supportedName, size := range cacc.supportedBlkNeis {
		blkNonEssentialIntCreator := wordfile.NewConcreteWordFileCreator(supportedName, cacc.blocksFolder, int64(size))
		wfile, err := blkNonEssentialIntCreator.OpenWordFile()
		if err == nil {
			result.blkNonEssentialInts[supportedName] = wfile
		}
	}

	result.trnNonEssentialInts = make(map[string]wordfile.ReadWriteAtWordCounter)
	// We try to open each of the supported transaction NonEssentialInt wordfiles.
	// However they do not need to exist, and if they're not there we don't error.
	for supportedName, size := range cacc.supportedTrnNeis {
		trnNonEssentialIntCreator := wordfile.NewConcreteWordFileCreator(supportedName, cacc.transactionsFolder, int64(size))
		wfile, err := trnNonEssentialIntCreator.OpenWordFile()
		if err == nil {
			result.trnNonEssentialInts[supportedName] = wfile
		}
	}

	return &result, &result, nil
}
