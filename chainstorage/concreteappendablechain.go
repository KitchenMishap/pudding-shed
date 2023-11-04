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

func (cac concreteAppendableChain) AppendBlock(blockChain chainreadinterface.IBlockChain,
	hBlock chainreadinterface.IBlockHandle) error {
	block := blockChain.BlockInterface(hBlock)
	if !block.HashSpecified() {
		panic("this function assumes block specifies a hash")
	}
	blkHash := block.Hash()
	blkNum, err := cac.blkHashes.AppendHash(&blkHash)
	if err != nil {
		return err
	}
	if hBlock.HeightSpecified() && hBlock.Height() != blkNum {
		panic("cannot append a block out of sequence")
	}

	nTrans := block.TransactionCount()
	for t := int64(0); t < nTrans; t++ {
		hTrans := block.NthTransaction(t)
		trans := blockChain.TransInterface(hTrans)
		transNum, err := cac.appendTransaction(blockChain, trans)
		if err != nil {
			return err
		}
		if trans.HeightSpecified() && trans.Height() != transNum {
			panic("cannot append a transaction out of sequence")
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

func (cac concreteAppendableChain) appendTransaction(blockChain chainreadinterface.IBlockChain,
	hTrans chainreadinterface.ITransHandle) (int64, error) {
	trans := blockChain.TransInterface(hTrans)
	if !trans.HashSpecified() {
		panic("this function assumes that trans specifies a hash")
	}
	transHash := trans.Hash()
	transNum, err := cac.trnHashes.AppendHash(&transHash)
	if err != nil {
		return -1, err
	}
	if trans.HeightSpecified() && trans.Height() != transNum {
		panic("cannot append a transaction out of sequence")
	}

	nTxis := trans.TxiCount()
	for nTxi := int64(0); nTxi < nTxis; nTxi++ {
		hTxi := trans.NthTxi(nTxi)
		txi := blockChain.TxiInterface(hTxi)
		txiHeight, err := cac.appendTxi(txi)
		if err != nil {
			return -1, err
		}
		if txi.TxiHeightSpecified() && txi.TxiHeight() != txiHeight {
			panic("cannot append a txi out of sequence")
		}
		if nTxi == 0 {
			err := cac.trnFirstTxi.WriteWordAt(txiHeight, transNum)
			if err != nil {
				return -1, err
			}
		}
	}

	nTxos := trans.TxoCount()
	for nTxo := int64(0); nTxo < nTxos; nTxo++ {
		hTxo := trans.NthTxo(nTxo)
		txo := blockChain.TxoInterface(hTxo)
		txoHeight, err := cac.appendTxo(txo)
		if err != nil {
			return -1, err
		}
		if txo.TxoHeightSpecified() && txo.TxoHeight() != txoHeight {
			panic("cannot append a txo out of sequence")
		}
		if nTxo == 0 {
			err := cac.trnFirstTxo.WriteWordAt(txoHeight, transNum)
			if err != nil {
				return -1, err
			}
		}
	}

	return transNum, nil
}

func (cac concreteAppendableChain) appendTxi(txi chainreadinterface.ITxi) (int64, error) {
	sourceTxo := txi.SourceTxo()
	if !sourceTxo.ParentSpecified() {
		panic("this implementation assumes the txi's source txo specifies a parent transaction and index")
	}
	sourceTrans := sourceTxo.ParentTrans()
	sourceIndex := sourceTxo.ParentIndex()
	if !sourceTrans.HeightSpecified() {
		panic("this implementation assumes the source transaction specifies a height")
	}
	sourceTransHeight := sourceTrans.Height()
	txiHeight, err := cac.txiTx.CountWords()
	if err != nil {
		return -1, err
	}
	if txi.TxiHeightSpecified() && txi.TxiHeight() != txiHeight {
		panic("cannot append a txi out of sequence")
	}
	err = cac.txiTx.WriteWordAt(sourceTransHeight, txiHeight)
	if err != nil {
		return -1, err
	}
	err = cac.txiVout.WriteWordAt(sourceIndex, txiHeight)
	if err != nil {
		return -1, err
	}

	return txiHeight, nil
}

func (cac concreteAppendableChain) appendTxo(txo chainreadinterface.ITxo) (int64, error) {
	sats := txo.Satoshis()
	txoHeight, err := cac.txoSats.CountWords()
	if err != nil {
		return -1, err
	}
	if txo.TxoHeightSpecified() && txo.TxoHeight() != txoHeight {
		panic("cannot append a txo out of sequence")
	}
	err = cac.txoSats.WriteWordAt(sats, txoHeight)
	if err != nil {
		return -1, err
	}

	return txoHeight, nil
}

func (cac concreteAppendableChain) Close() error {
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