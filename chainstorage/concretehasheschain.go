package chainstorage

import (
	"fmt"
	"github.com/KitchenMishap/pudding-shed/jsonblock"
	"github.com/KitchenMishap/pudding-shed/testpoints"
	"github.com/KitchenMishap/pudding-shed/wordfile"
)

type concreteHashesChain struct {
	blkHashList wordfile.HashFile
	trnHashList wordfile.HashFile
	adrHashList wordfile.HashFile
}

// Check that implements
var _ IAppendableHashesChain = (*concreteHashesChain)(nil)

func (chc *concreteHashesChain) AppendHashes(block *jsonblock.JsonBlockHashes) error {
	// === TestPoint ===
	if testpoints.TestPointBlockEnable && block.J_height == testpoints.TestPointBlock {
		fmt.Println("TESTPOINT: concreteHashesChain.AppendHashes(block height ", testpoints.TestPointBlock, ")")
	}

	blkHash := block.BlockHash()
	_, err := chc.blkHashList.AppendHash(blkHash)
	if err != nil {
		return err
	}

	nTrans := len(block.J_tx)
	if nTrans == 0 {
		panic("this code assumes at least one transaction per block")
		// Otherwise, not every entry in blkFirstTrans will be written
	}
	for t := 0; t < nTrans; t++ {
		// Store transaction hash
		trans := block.J_tx[t]
		transHash := trans.TransHash()
		_, err = chc.trnHashList.AppendHash(transHash)
		if err != nil {
			return err
		}

		nTxo := len(trans.J_vout)
		for o := 0; o < nTxo; o++ {
			// Store address hash of txo
			txo := trans.J_vout[o]
			addrHash := txo.AddrHash()
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
