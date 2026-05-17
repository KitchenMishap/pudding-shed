package chainstorage

import (
	"fmt"

	"github.com/KitchenMishap/pudding-shed/chainhandleinterface"
	"github.com/KitchenMishap/pudding-shed/chainreadinterface"
	"github.com/KitchenMishap/pudding-shed/indexedhashes"
	"github.com/KitchenMishap/pudding-shed/testpoints"
	"github.com/KitchenMishap/pudding-shed/wordfile"
)

type concreteHashesChain struct {
	blkHashList wordfile.HashFile
	trnHashList wordfile.HashFile
	adrHashList wordfile.HashFile

	// These are re-used objects in the interest of the "zero allocations" aspiration
	za_blockReceiver       *chainhandleinterface.BlockReceiver
	za_transactionReceiver *chainhandleinterface.BitcoinCoreTransactionReceiver
	za_txiReceiver         *chainhandleinterface.TxiReceiver
	za_txoReceiver         *chainhandleinterface.TxoReceiver
}

// Check that implements
var _ IAppendableHashesChain = (*concreteHashesChain)(nil)

func (chc *concreteHashesChain) AppendHashesCri(chain chainreadinterface.IBlockChain,
	hBlock chainreadinterface.IBlockHandle, blockHeight int64) error {

	// === TestPoint ===
	if testpoints.TestPointBlockEnable && blockHeight == testpoints.TestPointBlock {
		fmt.Println("TESTPOINT: concreteHashesChain.AppendHashesBinary(block height ", testpoints.TestPointBlock, ")")
	}

	if !hBlock.HashSpecified() {
		panic("This fn assumes that block handle specifies hash")
	}

	blkHash, err := hBlock.Hash()
	if err != nil {
		return err
	}
	_, err = chc.blkHashList.AppendHash(blkHash)
	if err != nil {
		return err
	}

	blk, err := chain.BlockInterface(hBlock)
	if err != nil {
		return err
	}

	nTrans, err := blk.TransactionCount()
	if err != nil {
		return err
	}
	if nTrans == 0 {
		panic("this code assumes at least one transaction per block")
		// Otherwise, not every entry in blkFirstTrans will be written
	}
	for t := int64(0); t < nTrans; t++ {
		// Store transaction hash
		hTrans, err := blk.NthTransaction(t)
		if err != nil {
			return err
		}
		if !hTrans.HashSpecified() {
			panic("This fn assumes trans handle specifies txid as hash")
		}
		transHash, err := hTrans.Hash()
		if err != nil {
			return err
		}
		_, err = chc.trnHashList.AppendHash(transHash)
		if err != nil {
			return err
		}

		trans, err := chain.TransInterface(hTrans)
		if err != nil {
			return err
		}

		nTxo, err := trans.TxoCount()
		if err != nil {
			return err
		}
		for txoInd := range nTxo {
			hTxo, err := trans.NthTxo(txoInd)
			if err != nil {
				return err
			}
			txo, err := chain.TxoInterface(hTxo)
			if err != nil {
				return err
			}
			hAddr, err := txo.Address()
			if err != nil {
				return err
			}
			if !hAddr.HashSpecified() {
				panic("This fn assumes address handle specifies address hash")
			}
			addrHash := hAddr.Hash()
			_, err = chc.adrHashList.AppendHash(addrHash)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (chc *concreteHashesChain) AppendHashesChi(chain chainhandleinterface.IBlockChain,
	hBlock chainhandleinterface.BlockHandle) error {

	if chc.za_blockReceiver == nil {
		chc.za_blockReceiver = chainhandleinterface.NewBlockReceiver()
	}
	if chc.za_transactionReceiver == nil {
		chc.za_transactionReceiver = chainhandleinterface.NewBitcoinCoreTransactionReceiver()
	}
	if chc.za_txiReceiver == nil {
		chc.za_txiReceiver = chainhandleinterface.NewTxiReceiver()
	}
	if chc.za_txoReceiver == nil {
		chc.za_txoReceiver = chainhandleinterface.NewTxoReceiver()
	}

	err := chain.GetBlockInfo(hBlock, chc.za_blockReceiver)
	if err != nil {
		return err
	}

	// === TestPoint ===
	if testpoints.TestPointBlockEnable && chc.za_blockReceiver.BlockHeight == testpoints.TestPointBlock {
		fmt.Println("TESTPOINT: concreteHashesChain.AppendHashesBinary(block height ", testpoints.TestPointBlock, ")")
	}

	blkHash := chc.za_blockReceiver.BlockHash
	// Store block hash
	_, err = chc.blkHashList.AppendHash(blkHash)
	if err != nil {
		return err
	}

	nTrans := len(chc.za_blockReceiver.TransactionHandles)
	if nTrans == 0 {
		panic("this code assumes at least one transaction per block")
		// Otherwise, not every entry in blkFirstTrans will be written
	}

	for _, hTrans := range chc.za_blockReceiver.TransactionHandles {
		err = chain.GetTransactionInfo(hTrans, chc.za_transactionReceiver)
		if err != nil {
			return err
		}
		// Store transaction hash
		_, err = chc.trnHashList.AppendHash(chc.za_transactionReceiver.TransactionHash)
		if err != nil {
			return err
		}

		for _, hTxo := range chc.za_transactionReceiver.TxoHandles {
			err = chain.GetTxoInfo(hTxo, chc.za_txoReceiver)
			if err != nil {
				return err
			}
			// puddingHash3 (hash of ScriptPubKey bytes) is peculiar to pudding-shed software,
			// and is not generally known to bitcoiners
			puddingHash3 := indexedhashes.HashOfBytes(chc.za_txoReceiver.ScriptPubBytes)
			_, err = chc.adrHashList.AppendHash(puddingHash3)
		} // for txo
	} // for trans

	return nil
}

func (chc *concreteHashesChain) Close() {
	chc.blkHashList.Close()
	chc.trnHashList.Close()
	chc.adrHashList.Close()
}

func (chc *concreteHashesChain) Sync() error {
	err := chc.blkHashList.Sync()
	if err != nil {
		return err
	}
	err = chc.trnHashList.Sync()
	if err != nil {
		return err
	}
	err = chc.adrHashList.Sync()
	if err != nil {
		return err
	}
	return nil
}

func (chc *concreteHashesChain) CountHashes() (blocks int64, transactions int64, addresses int64, err error) {
	blocks = chc.blkHashList.CountHashes()
	transactions = chc.trnHashList.CountHashes()
	addresses = chc.adrHashList.CountHashes()
	return blocks, transactions, addresses, nil
}
