package chainstorage

import (
	"github.com/KitchenMishap/pudding-shed/chainreadinterface"
	"github.com/KitchenMishap/pudding-shed/wordfile"
)

type concreteHashesChain struct {
	blkHashList wordfile.HashFile
	trnHashList wordfile.HashFile
	adrHashList wordfile.HashFile
}

// Check that implements
var _ IAppendableHashesChain = (*concreteHashesChain)(nil)

func (chc *concreteHashesChain) AppendBlock(blockChain chainreadinterface.IBlockChain,
	// Store block hash
	hBlock chainreadinterface.IBlockHandle) error {
	block, err := blockChain.BlockInterface(hBlock)
	if err != nil {
		return err
	}
	if !block.HashSpecified() {
		panic("this function assumes block specifies a hash")
	}
	blkHash, err := block.Hash()
	if err != nil {
		return err
	}
	blkHeight, err := chc.blkHashList.AppendHash(blkHash)
	if err != nil {
		return err
	}
	if block.HeightSpecified() && hBlock.Height() != blkHeight {
		panic("cannot append a block out of sequence")
	}

	nTrans, err := block.TransactionCount()
	if err != nil {
		return err
	}
	if nTrans == 0 {
		panic("this code assumes at least one transaction per block")
		// Otherwise, not every entry in blkFirstTrans will be written
	}
	for t := int64(0); t < nTrans; t++ {
		// Store transaction hash
		hTrans, err := block.NthTransaction(t)
		if err != nil {
			return err
		}
		trans, err := blockChain.TransInterface(hTrans)
		if err != nil {
			return err
		}
		if !trans.HashSpecified() {
			panic("this code assumes transaction hash is specified")
		}
		transHash, err := trans.Hash()
		if err != nil {
			return err
		}
		_, err = chc.trnHashList.AppendHash(transHash)
		if err != nil {
			return err
		}

		nTxo, err := trans.TxoCount()
		if err != nil {
			return err
		}
		for o := int64(0); o < nTxo; o++ {
			// Store address hash of txo
			hTxo, err := trans.NthTxo(o)
			if err != nil {
				return err
			}
			txo, err := blockChain.TxoInterface(hTxo)
			if err != nil {
				return err
			}
			hAddress, err := txo.Address()
			if err != nil {
				return err
			}
			addr, err := blockChain.AddressInterface(hAddress)
			if err != nil {
				return err
			}
			if !addr.HashSpecified() {
				panic("this code assumes hash of address is specified")
			}
			addrHash := addr.Hash()
			_, err = chc.adrHashList.AppendHash(addrHash)
			if err != nil {
				return err
			}
		}
	}
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
