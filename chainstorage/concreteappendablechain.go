package chainstorage

import (
	"github.com/KitchenMishap/pudding-shed/chainreadinterface"
	"github.com/KitchenMishap/pudding-shed/indexedhashes"
	"github.com/KitchenMishap/pudding-shed/wordfile"
)

type concreteAppendableChain struct {
	blkFirstTrans wordfile.ReadWriteAtWordCounter
	blkHashes     indexedhashes.HashReadWriter
	trnHashes     indexedhashes.HashReadWriter
	trnFirstTxi   wordfile.ReadWriteAtWordCounter
	trnFirstTxo   wordfile.ReadWriteAtWordCounter
	txiTx         wordfile.ReadWriteAtWordCounter
	txiVout       wordfile.ReadWriteAtWordCounter
	txoSats       wordfile.ReadWriteAtWordCounter
}

func (cac *concreteAppendableChain) AppendBlock(handleHandler chainreadinterface.IHandles,
	blockChain chainreadinterface.IBlockChain,
	block chainreadinterface.IBlock) error {
	hBlock := block.BlockHandle()
	blkHash := handleHandler.HashFromHBlock(hBlock)
	blkNum, err := cac.blkHashes.AppendHash(&blkHash)
	if err != nil {
		return err
	}

	nTrans := block.TransactionCount()
	for t := int64(0); t < nTrans; t++ {
		hTrans := block.NthTransactionHandle(t)
		trans := blockChain.TransactionInterface(hTrans)
		transNum, err := cac.appendTransaction(handleHandler, trans)
		if err != nil {
			return err
		}
		if t == 0 {
			err = cac.blkFirstTrans.WriteWordAt(transNum, blkNum)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (cac *concreteAppendableChain) appendTransaction(handleHandler chainreadinterface.IHandles,
	trans chainreadinterface.ITransaction) (int64, error) {
	hTrans := trans.TransactionHandle()
	transHash := handleHandler.HashFromHTransaction(hTrans)
	transNum, err := cac.trnHashes.AppendHash(&transHash)
	if err != nil {
		return -1, err
	}

	nTxis := trans.TxiCount()
	for nTxi := int64(0); nTxi < nTxis; nTxi++ {
		iTxi := trans.NthTxiInterface(nTxi)
		nTxi, err := cac.appendTxi(handleHandler, iTxi)
		if err != nil {
			return -1, err
		}
		if nTxi == 0 {
			err := cac.trnFirstTxi.WriteWordAt(nTxis, transNum)
			if err != nil {
				return -1, err
			}
		}
	}
	nTxos := trans.TxoCount()
	for nTxo := int64(0); nTxo < nTxos; nTxo++ {
		iTxo := trans.NthTxoInterface(nTxo)
		nTxo, err := cac.appendTxo(iTxo)
		if err != nil {
			return -1, err
		}
		if nTxo == 0 {
			err := cac.trnFirstTxo.WriteWordAt(nTxos, transNum)
			if err != nil {
				return -1, err
			}
		}
	}

	return transNum, nil
}

func (cac *concreteAppendableChain) appendTxi(iHandles chainreadinterface.IHandles,
	iTxi chainreadinterface.ITxi) (int64, error) {
	hSourceTrans := iTxi.SourceTransaction()
	transHeight := iHandles.HeightFromHTransaction(hSourceTrans)
	sourceIndex := iTxi.SourceIndex()
	nTxi, err := cac.txiTx.CountWords()
	if err != nil {
		return -1, err
	}
	err = cac.txiTx.WriteWordAt(transHeight, nTxi)
	if err != nil {
		return -1, err
	}
	err = cac.txiVout.WriteWordAt(sourceIndex, nTxi)
	if err != nil {
		return -1, err
	}
	return nTxi, nil
}

func (cac *concreteAppendableChain) appendTxo(iTxo chainreadinterface.ITxo) (int64, error) {
	sats := iTxo.Satoshis()
	nTxo, err := cac.txoSats.CountWords()
	if err != nil {
		return -1, err
	}
	err = cac.txoSats.WriteWordAt(sats, nTxo)
	if err != nil {
		return -1, err
	}
	return nTxo, nil
}

func (cac *concreteAppendableChain) Close() error {
	err := cac.blkHashes.Close()
	if err != nil {
		return err
	}
	err = cac.trnHashes.Close()
	if err != nil {
		return err
	}
	err = cac.blkFirstTrans.Close()
	if err != nil {
		return err
	}
	err = cac.trnFirstTxi.Close()
	if err != nil {
		return err
	}
	err = cac.trnFirstTxo.Close()
	if err != nil {
		return err
	}
	err = cac.txiTx.Close()
	if err != nil {
		return err
	}
	err = cac.txiVout.Close()
	if err != nil {
		return err
	}
	err = cac.txoSats.Close()
	if err != nil {
		return err
	}
	return nil
}
