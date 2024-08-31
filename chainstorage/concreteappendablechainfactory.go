package chainstorage

import (
	"errors"
	"github.com/KitchenMishap/pudding-shed/chainreadinterface"
	"github.com/KitchenMishap/pudding-shed/indexedhashes"
	"github.com/KitchenMishap/pudding-shed/intarrayarray"
	"github.com/KitchenMishap/pudding-shed/transactionindexing"
	"github.com/KitchenMishap/pudding-shed/wordfile"
	"path"
	"slices"
)

type ConcreteAppendableChainCreator struct {
	blocksFolder                      string
	transactionsFolder                string
	transactionInputsFolder           string
	transactionOutputsFolder          string
	addressesFolder                   string
	parentsFolder                     string
	blockHashStoreCreator             *indexedhashes.ConcreteHashStoreCreator
	transactionHashStoreCreator       *indexedhashes.ConcreteHashStoreCreator
	addressHashStoreCreator           *indexedhashes.ConcreteHashStoreCreator
	blkFirstTransWordFileCreator      *wordfile.ConcreteWordFileCreator
	trnFirstTxiWordFileCreator        *wordfile.ConcreteWordFileCreator
	trnFirstTxoWordFileCreator        *wordfile.ConcreteWordFileCreator
	txiTxWordFileCreator              *wordfile.ConcreteWordFileCreator
	txiVoutWordFileCreator            *wordfile.ConcreteWordFileCreator
	txoSatsWordFileCreator            *wordfile.ConcreteWordFileCreator
	txoAddressWordFileCreator         *wordfile.ConcreteWordFileCreator
	txoSpentTxiWordFileCreator        *wordfile.ConcreteWordFileCreator
	addrFirstTxoWordFileCreator       *wordfile.ConcreteWordFileCreator
	parentBlockOfTransWordFileCreator *wordfile.ConcreteWordFileCreator
	parentTransOfTxiWordFileCreator   *wordfile.ConcreteWordFileCreator
	parentTransOfTxoWordFileCreator   *wordfile.ConcreteWordFileCreator
	addrAdditionalTxosIaaCreator      *intarrayarray.ConcreteMapStoreCreator

	supportedBlkNeis map[string]int
	supportedTrnNeis map[string]int

	requestedBlkNeis []string
	requestedTrnNeis []string

	transactionIndexingToBeDelegated bool
}

// Check that implements
var _ IAppendableChainFactory = (*ConcreteAppendableChainCreator)(nil)

func NewConcreteAppendableChainCreator(
	folder string, blkNeiNames []string, trnNeiNames []string,
	transactionIndexingToBeDelegated bool) (*ConcreteAppendableChainCreator, error) {
	result := ConcreteAppendableChainCreator{}

	result.blocksFolder = path.Join(folder, "Blocks")
	result.transactionsFolder = path.Join(folder, "Transactions")
	result.transactionInputsFolder = path.Join(folder, "TransactionInputs")
	result.transactionOutputsFolder = path.Join(folder, "TransactionOutputs")
	result.addressesFolder = path.Join(folder, "Addresses")
	result.parentsFolder = path.Join(folder, "Parents")

	//																			Count at 15 years:
	roomFor16milBlocks := int64(3) // 256^3 = 16,777,216 blocks					There were 824,197 blocks
	roomFor4bilTrans := int64(4)   // 256^4 = 4,294,967,296 transactions		There were 947,337,057 transactions
	roomFor1trilTxxs := int64(5)   // 256^5 = 1,099,511,627,776 txos or txis	There were 2,652,374,369 txos (including spent)
	roomFor1trilAddrs := int64(5)  //	,,			,,			 addresses		There must be fewer addresses than txos
	roomForAllSatoshis := int64(7) // 256^7 = 72,057,594,037,927,936 sats		There will be 2,100,000,000,000,000 sats

	// WARNING true will cost you 50GB memory!
	// And trash your SSD if you don't have it
	useMemFileForHashes := true
	useAppendOptimized := true

	var err error
	result.blockHashStoreCreator, err = indexedhashes.NewConcreteHashStoreCreator(
		"Blocks", result.blocksFolder, 30, roomFor16milBlocks, 3, useMemFileForHashes, useAppendOptimized)
	if err != nil {
		return nil, err
	}
	result.transactionHashStoreCreator, err = indexedhashes.NewConcreteHashStoreCreator(
		"Transactions", result.transactionsFolder, 30, roomFor4bilTrans, 3, useMemFileForHashes, useAppendOptimized)
	if err != nil {
		return nil, err
	}
	result.addressHashStoreCreator, err = indexedhashes.NewConcreteHashStoreCreator(
		"Addresses", result.addressesFolder, 32, roomFor1trilAddrs, 3, useMemFileForHashes, useAppendOptimized) // ToDo [  ] Calculate these parameters!
	if err != nil {
		return nil, err
	}
	appendOpt := true
	result.blkFirstTransWordFileCreator = wordfile.NewConcreteWordFileCreator("firsttrans", result.blocksFolder, roomFor4bilTrans, appendOpt)
	result.trnFirstTxiWordFileCreator = wordfile.NewConcreteWordFileCreator("firsttxi", result.transactionsFolder, roomFor1trilTxxs, appendOpt)
	result.trnFirstTxoWordFileCreator = wordfile.NewConcreteWordFileCreator("firsttxo", result.transactionsFolder, roomFor1trilTxxs, appendOpt)
	result.txiTxWordFileCreator = wordfile.NewConcreteWordFileCreator("tx", result.transactionInputsFolder, roomFor4bilTrans, appendOpt)
	result.txiVoutWordFileCreator = wordfile.NewConcreteWordFileCreator("vout", result.transactionInputsFolder, 4, appendOpt)
	result.txoSatsWordFileCreator = wordfile.NewConcreteWordFileCreator("value", result.transactionOutputsFolder, roomForAllSatoshis, appendOpt)
	result.txoAddressWordFileCreator = wordfile.NewConcreteWordFileCreator("address", result.transactionOutputsFolder, roomFor1trilAddrs, appendOpt)
	result.txoSpentTxiWordFileCreator = wordfile.NewConcreteWordFileCreator("spenttotxi", result.transactionOutputsFolder, roomFor1trilTxxs, appendOpt)
	result.addrFirstTxoWordFileCreator = wordfile.NewConcreteWordFileCreator("firsttxo", result.addressesFolder, roomFor1trilTxxs, appendOpt)
	result.parentBlockOfTransWordFileCreator = wordfile.NewConcreteWordFileCreator("parentblockoftrans", result.parentsFolder, roomFor4bilTrans, appendOpt)
	result.parentTransOfTxiWordFileCreator = wordfile.NewConcreteWordFileCreator("parenttransoftxi", result.parentsFolder, roomFor1trilTxxs, appendOpt)
	result.parentTransOfTxoWordFileCreator = wordfile.NewConcreteWordFileCreator("parenttransoftxo", result.parentsFolder, roomFor1trilTxxs, appendOpt)
	result.addrAdditionalTxosIaaCreator = intarrayarray.NewConcreteMapStoreCreator("additionaltxos", result.addressesFolder, 2, 2, roomFor1trilTxxs)
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

	result.requestedBlkNeis = blkNeiNames
	result.requestedTrnNeis = trnNeiNames

	result.transactionIndexingToBeDelegated = transactionIndexingToBeDelegated

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
	err = cacc.addressHashStoreCreator.CreateHashStore()
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
	err = cacc.txoAddressWordFileCreator.CreateWordFile()
	if err != nil {
		return err
	}
	err = cacc.txoSpentTxiWordFileCreator.CreateWordFile()
	if err != nil {
		return err
	}
	err = cacc.addrFirstTxoWordFileCreator.CreateWordFile()
	if err != nil {
		return err
	}
	err = cacc.parentBlockOfTransWordFileCreator.CreateWordFile()
	if err != nil {
		return err
	}
	err = cacc.parentTransOfTxiWordFileCreator.CreateWordFile()
	if err != nil {
		return err
	}
	err = cacc.parentTransOfTxoWordFileCreator.CreateWordFile()
	if err != nil {
		return err
	}
	err = cacc.addrAdditionalTxosIaaCreator.CreateMap()
	if err != nil {
		return err
	}

	for supportedName, size := range cacc.supportedBlkNeis {
		if slices.Contains(cacc.requestedBlkNeis, supportedName) {
			blkNonEssentialIntCreator := wordfile.NewConcreteWordFileCreator(supportedName, cacc.blocksFolder, int64(size), true)
			err = blkNonEssentialIntCreator.CreateWordFile()
			if err != nil {
				return err
			}
		}
	}
	for _, requestedName := range cacc.requestedBlkNeis {
		_, supported := cacc.supportedBlkNeis[requestedName]
		if !supported {
			return errors.New(requestedName + " is not a supported block NonEssentialInt")
		}
	}

	for supportedName, size := range cacc.supportedTrnNeis {
		if slices.Contains(cacc.requestedTrnNeis, supportedName) {
			trnNonEssentialIntCreator := wordfile.NewConcreteWordFileCreator(supportedName, cacc.transactionsFolder, int64(size), true)
			err = trnNonEssentialIntCreator.CreateWordFile()
			if err != nil {
				return err
			}
		}
	}
	for _, requestedName := range cacc.requestedTrnNeis {
		_, supported := cacc.supportedTrnNeis[requestedName]
		if !supported {
			return errors.New(requestedName + " is not a supported transaction NonEssentialInt")
		}
	}

	return nil
}

func (cacc *ConcreteAppendableChainCreator) Open() (IAppendableChain, error) {
	cac, err := cacc.openPrivate()
	return cac, err
}

func (cacc *ConcreteAppendableChainCreator) OpenWithIndexer() (IAppendableChain, transactionindexing.ITransactionIndexer, error) {
	cac, err := cacc.openPrivate()
	return cac, cac, err
}

func (cacc *ConcreteAppendableChainCreator) openPrivate() (*concreteAppendableChain, error) {
	result := concreteAppendableChain{}
	result.transactionIndexingIsDelegated = cacc.transactionIndexingToBeDelegated

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
	result.addrHashes, err = cacc.addressHashStoreCreator.OpenHashStore()
	if err != nil {
		result.blkHashes.Close()
		result.trnHashes.Close()
		return nil, err
	}
	result.blkFirstTrans, err = cacc.blkFirstTransWordFileCreator.OpenWordFile()
	if err != nil {
		result.blkHashes.Close()
		result.trnHashes.Close()
		result.addrHashes.Close()
		return nil, err
	}
	result.trnFirstTxi, err = cacc.trnFirstTxiWordFileCreator.OpenWordFile()
	if err != nil {
		result.blkHashes.Close()
		result.trnHashes.Close()
		result.addrHashes.Close()
		result.blkFirstTrans.Close()
		return nil, err
	}
	result.trnFirstTxo, err = cacc.trnFirstTxoWordFileCreator.OpenWordFile()
	if err != nil {
		result.blkHashes.Close()
		result.trnHashes.Close()
		result.addrHashes.Close()
		result.blkFirstTrans.Close()
		result.trnFirstTxi.Close()
		return nil, err
	}
	result.txiTx, err = cacc.txiTxWordFileCreator.OpenWordFile()
	if err != nil {
		result.blkHashes.Close()
		result.trnHashes.Close()
		result.addrHashes.Close()
		result.blkFirstTrans.Close()
		result.trnFirstTxi.Close()
		result.trnFirstTxo.Close()
		return nil, err
	}
	result.txiVout, err = cacc.txiVoutWordFileCreator.OpenWordFile()
	if err != nil {
		result.blkHashes.Close()
		result.trnHashes.Close()
		result.addrHashes.Close()
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
		result.addrHashes.Close()
		result.blkFirstTrans.Close()
		result.trnFirstTxi.Close()
		result.trnFirstTxo.Close()
		result.txiTx.Close()
		result.txiVout.Close()
		return nil, err
	}
	result.txoAddress, err = cacc.txoAddressWordFileCreator.OpenWordFile()
	if err != nil {
		result.blkHashes.Close()
		result.trnHashes.Close()
		result.addrHashes.Close()
		result.blkFirstTrans.Close()
		result.trnFirstTxi.Close()
		result.trnFirstTxo.Close()
		result.txiTx.Close()
		result.txiVout.Close()
		result.txoSats.Close()
		return nil, err
	}
	result.txoSpentTxi, err = cacc.txoSpentTxiWordFileCreator.OpenWordFile()
	if err != nil {
		result.blkHashes.Close()
		result.trnHashes.Close()
		result.addrHashes.Close()
		result.blkFirstTrans.Close()
		result.trnFirstTxi.Close()
		result.trnFirstTxo.Close()
		result.txiTx.Close()
		result.txiVout.Close()
		result.txoSats.Close()
		result.txoAddress.Close()
		return nil, err
	}
	result.addrFirstTxo, err = cacc.addrFirstTxoWordFileCreator.OpenWordFile()
	if err != nil {
		result.blkHashes.Close()
		result.trnHashes.Close()
		result.addrHashes.Close()
		result.blkFirstTrans.Close()
		result.trnFirstTxi.Close()
		result.trnFirstTxo.Close()
		result.txiTx.Close()
		result.txiVout.Close()
		result.txoSats.Close()
		result.txoAddress.Close()
		result.txoSpentTxi.Close()
		return nil, err
	}

	result.parentBlockOfTrans, err = cacc.parentBlockOfTransWordFileCreator.OpenWordFile()
	if err != nil {
		result.blkHashes.Close()
		result.trnHashes.Close()
		result.addrHashes.Close()
		result.blkFirstTrans.Close()
		result.trnFirstTxi.Close()
		result.trnFirstTxo.Close()
		result.txiTx.Close()
		result.txiVout.Close()
		result.txoSats.Close()
		result.txoAddress.Close()
		result.txoSpentTxi.Close()
		result.addrFirstTxo.Close()
		return nil, err
	}
	result.parentTransOfTxi, err = cacc.parentTransOfTxiWordFileCreator.OpenWordFile()
	if err != nil {
		result.blkHashes.Close()
		result.trnHashes.Close()
		result.addrHashes.Close()
		result.blkFirstTrans.Close()
		result.trnFirstTxi.Close()
		result.trnFirstTxo.Close()
		result.txiTx.Close()
		result.txiVout.Close()
		result.txoSats.Close()
		result.txoAddress.Close()
		result.txoSpentTxi.Close()
		result.addrFirstTxo.Close()
		result.parentBlockOfTrans.Close()
		return nil, err
	}
	result.parentTransOfTxo, err = cacc.parentTransOfTxoWordFileCreator.OpenWordFile()
	if err != nil {
		result.blkHashes.Close()
		result.trnHashes.Close()
		result.addrHashes.Close()
		result.blkFirstTrans.Close()
		result.trnFirstTxi.Close()
		result.trnFirstTxo.Close()
		result.txiTx.Close()
		result.txiVout.Close()
		result.txoSats.Close()
		result.txoAddress.Close()
		result.txoSpentTxi.Close()
		result.addrFirstTxo.Close()
		result.parentBlockOfTrans.Close()
		result.parentTransOfTxi.Close()
		return nil, err
	}

	result.addrAdditionalTxos, err = cacc.addrAdditionalTxosIaaCreator.OpenMap()
	if err != nil {
		result.blkHashes.Close()
		result.trnHashes.Close()
		result.addrHashes.Close()
		result.blkFirstTrans.Close()
		result.trnFirstTxi.Close()
		result.trnFirstTxo.Close()
		result.txiTx.Close()
		result.txiVout.Close()
		result.txoSats.Close()
		result.txoAddress.Close()
		result.txoSpentTxi.Close()
		result.addrFirstTxo.Close()
		result.parentBlockOfTrans.Close()
		result.parentTransOfTxi.Close()
		result.parentTransOfTxo.Close()
		return nil, err
	}

	result.blkNonEssentialInts = make(map[string]wordfile.ReadWriteAtWordCounter)
	// We try to open each of the supported block NonEssentialInt wordfiles.
	// However they do not need to exist, and if they're not there we don't error.
	for supportedName, size := range cacc.supportedBlkNeis {
		blkNonEssentialIntCreator := wordfile.NewConcreteWordFileCreator(supportedName, cacc.blocksFolder, int64(size), true)
		wfile, err := blkNonEssentialIntCreator.OpenWordFile()
		if err == nil {
			result.blkNonEssentialInts[supportedName] = wfile
		}
	}

	result.trnNonEssentialInts = make(map[string]wordfile.ReadWriteAtWordCounter)
	// We try to open each of the supported transaction NonEssentialInt wordfiles.
	// However they do not need to exist, and if they're not there we don't error.
	for supportedName, size := range cacc.supportedTrnNeis {
		trnNonEssentialIntCreator := wordfile.NewConcreteWordFileCreator(supportedName, cacc.transactionsFolder, int64(size), true)
		wfile, err := trnNonEssentialIntCreator.OpenWordFile()
		if err == nil {
			result.trnNonEssentialInts[supportedName] = wfile
		}
	}

	return &result, nil
}

func (cacc *ConcreteAppendableChainCreator) OpenReadOnly() (chainreadinterface.IBlockChain, chainreadinterface.IHandleCreator, error) {
	concreteChain, err := cacc.openReadOnlyPrivate()
	return concreteChain, concreteChain, err
}

func (cacc *ConcreteAppendableChainCreator) openReadOnlyPrivate() (*concreteReadableChain, error) {
	result := concreteReadableChain{}

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
	result.addrHashes, err = cacc.addressHashStoreCreator.OpenHashStore()
	if err != nil {
		result.blkHashes.Close()
		result.trnHashes.Close()
		return nil, err
	}
	result.blkFirstTrans, err = cacc.blkFirstTransWordFileCreator.OpenWordFile()
	if err != nil {
		result.blkHashes.Close()
		result.trnHashes.Close()
		result.addrHashes.Close()
		return nil, err
	}
	result.trnFirstTxi, err = cacc.trnFirstTxiWordFileCreator.OpenWordFile()
	if err != nil {
		result.blkHashes.Close()
		result.trnHashes.Close()
		result.addrHashes.Close()
		result.blkFirstTrans.Close()
		return nil, err
	}
	result.trnFirstTxo, err = cacc.trnFirstTxoWordFileCreator.OpenWordFile()
	if err != nil {
		result.blkHashes.Close()
		result.trnHashes.Close()
		result.addrHashes.Close()
		result.blkFirstTrans.Close()
		result.trnFirstTxi.Close()
		return nil, err
	}
	result.txiTx, err = cacc.txiTxWordFileCreator.OpenWordFile()
	if err != nil {
		result.blkHashes.Close()
		result.trnHashes.Close()
		result.addrHashes.Close()
		result.blkFirstTrans.Close()
		result.trnFirstTxi.Close()
		result.trnFirstTxo.Close()
		return nil, err
	}
	result.txiVout, err = cacc.txiVoutWordFileCreator.OpenWordFile()
	if err != nil {
		result.blkHashes.Close()
		result.trnHashes.Close()
		result.addrHashes.Close()
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
		result.addrHashes.Close()
		result.blkFirstTrans.Close()
		result.trnFirstTxi.Close()
		result.trnFirstTxo.Close()
		result.txiTx.Close()
		result.txiVout.Close()
		return nil, err
	}
	result.txoAddress, err = cacc.txoAddressWordFileCreator.OpenWordFile()
	if err != nil {
		result.blkHashes.Close()
		result.trnHashes.Close()
		result.addrHashes.Close()
		result.blkFirstTrans.Close()
		result.trnFirstTxi.Close()
		result.trnFirstTxo.Close()
		result.txiTx.Close()
		result.txiVout.Close()
		result.txoSats.Close()
		return nil, err
	}
	result.txoSpentTxi, err = cacc.txoSpentTxiWordFileCreator.OpenWordFile()
	if err != nil {
		result.blkHashes.Close()
		result.trnHashes.Close()
		result.addrHashes.Close()
		result.blkFirstTrans.Close()
		result.trnFirstTxi.Close()
		result.trnFirstTxo.Close()
		result.txiTx.Close()
		result.txiVout.Close()
		result.txoSats.Close()
		result.txoAddress.Close()
		return nil, err
	}
	result.addrFirstTxo, err = cacc.addrFirstTxoWordFileCreator.OpenWordFile()
	if err != nil {
		result.blkHashes.Close()
		result.trnHashes.Close()
		result.addrHashes.Close()
		result.blkFirstTrans.Close()
		result.trnFirstTxi.Close()
		result.trnFirstTxo.Close()
		result.txiTx.Close()
		result.txiVout.Close()
		result.txoSats.Close()
		result.txoAddress.Close()
		result.txoSpentTxi.Close()
		return nil, err
	}

	result.parentBlockOfTrans, err = cacc.parentBlockOfTransWordFileCreator.OpenWordFile()
	if err != nil {
		result.blkHashes.Close()
		result.trnHashes.Close()
		result.addrHashes.Close()
		result.blkFirstTrans.Close()
		result.trnFirstTxi.Close()
		result.trnFirstTxo.Close()
		result.txiTx.Close()
		result.txiVout.Close()
		result.txoSats.Close()
		result.txoAddress.Close()
		result.txoSpentTxi.Close()
		result.addrFirstTxo.Close()
		return nil, err
	}
	result.parentTransOfTxi, err = cacc.parentTransOfTxiWordFileCreator.OpenWordFile()
	if err != nil {
		result.blkHashes.Close()
		result.trnHashes.Close()
		result.addrHashes.Close()
		result.blkFirstTrans.Close()
		result.trnFirstTxi.Close()
		result.trnFirstTxo.Close()
		result.txiTx.Close()
		result.txiVout.Close()
		result.txoSats.Close()
		result.txoAddress.Close()
		result.txoSpentTxi.Close()
		result.addrFirstTxo.Close()
		result.parentBlockOfTrans.Close()
		return nil, err
	}
	result.parentTransOfTxo, err = cacc.parentTransOfTxoWordFileCreator.OpenWordFile()
	if err != nil {
		result.blkHashes.Close()
		result.trnHashes.Close()
		result.addrHashes.Close()
		result.blkFirstTrans.Close()
		result.trnFirstTxi.Close()
		result.trnFirstTxo.Close()
		result.txiTx.Close()
		result.txiVout.Close()
		result.txoSats.Close()
		result.txoAddress.Close()
		result.txoSpentTxi.Close()
		result.addrFirstTxo.Close()
		result.parentBlockOfTrans.Close()
		result.parentTransOfTxi.Close()
		return nil, err
	}

	result.addrAdditionalTxos, err = cacc.addrAdditionalTxosIaaCreator.OpenMap()
	if err != nil {
		result.blkHashes.Close()
		result.trnHashes.Close()
		result.addrHashes.Close()
		result.blkFirstTrans.Close()
		result.trnFirstTxi.Close()
		result.trnFirstTxo.Close()
		result.txiTx.Close()
		result.txiVout.Close()
		result.txoSats.Close()
		result.txoAddress.Close()
		result.txoSpentTxi.Close()
		result.addrFirstTxo.Close()
		result.parentBlockOfTrans.Close()
		result.parentTransOfTxi.Close()
		result.parentTransOfTxo.Close()
		return nil, err
	}

	result.blkNonEssentialInts = make(map[string]wordfile.ReadAtWordCounter)
	// We try to open each of the supported block NonEssentialInt wordfiles.
	// However they do not need to exist, and if they're not there we don't error.
	for supportedName, size := range cacc.supportedBlkNeis {
		blkNonEssentialIntCreator := wordfile.NewConcreteWordFileCreator(supportedName, cacc.blocksFolder, int64(size), true)
		wfile, err := blkNonEssentialIntCreator.OpenWordFile()
		if err == nil {
			result.blkNonEssentialInts[supportedName] = wfile
		}
	}

	result.trnNonEssentialInts = make(map[string]wordfile.ReadAtWordCounter)
	// We try to open each of the supported transaction NonEssentialInt wordfiles.
	// However they do not need to exist, and if they're not there we don't error.
	for supportedName, size := range cacc.supportedTrnNeis {
		trnNonEssentialIntCreator := wordfile.NewConcreteWordFileCreator(supportedName, cacc.transactionsFolder, int64(size), true)
		wfile, err := trnNonEssentialIntCreator.OpenWordFile()
		if err == nil {
			result.trnNonEssentialInts[supportedName] = wfile
		}
	}

	return &result, nil
}
