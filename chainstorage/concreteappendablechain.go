package chainstorage

import (
	"errors"
	"github.com/KitchenMishap/pudding-shed/chainreadinterface"
	"github.com/KitchenMishap/pudding-shed/indexedhashes"
	"github.com/KitchenMishap/pudding-shed/wordfile"
)

type concreteAppendableChain struct {
	blkFirstTrans       wordfile.ReadWriteAtWordCounter
	blkHashes           indexedhashes.HashReadWriter
	trnHashes           indexedhashes.HashReadWriter
	trnFirstTxi         wordfile.ReadWriteAtWordCounter
	trnFirstTxo         wordfile.ReadWriteAtWordCounter
	txiTx               wordfile.ReadWriteAtWordCounter
	txiVout             wordfile.ReadWriteAtWordCounter
	txoSats             wordfile.ReadWriteAtWordCounter
	blkNonEssentialInts map[string]wordfile.ReadWriteAtWordCounter
	trnNonEssentialInts map[string]wordfile.ReadWriteAtWordCounter
}

func (cac *concreteAppendableChain) GetAsConcreteReadableChain() *concreteReadableChain {
	result := concreteReadableChain{
		blkFirstTrans: cac.blkFirstTrans,
		blkHashes:     cac.blkHashes,
		trnHashes:     cac.trnHashes,
		trnFirstTxi:   cac.trnFirstTxi,
		trnFirstTxo:   cac.trnFirstTxo,
		txiTx:         cac.txiTx,
		txiVout:       cac.txiVout,
		txoSats:       cac.txoSats,
	}
	result.blkNonEssentialInts = make(map[string]wordfile.ReadAtWordCounter)
	result.trnNonEssentialInts = make(map[string]wordfile.ReadAtWordCounter)
	for k, v := range cac.blkNonEssentialInts {
		result.blkNonEssentialInts[k] = v
	}
	for k, v := range cac.trnNonEssentialInts {
		result.trnNonEssentialInts[k] = v
	}

	return &result
}

func (cac *concreteAppendableChain) GetAsChainReadInterface() chainreadinterface.IBlockChain {
	return cac.GetAsConcreteReadableChain()
}

func (cac *concreteAppendableChain) AppendBlock(blockChain chainreadinterface.IBlockChain,
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
	blkNum, err := cac.blkHashes.AppendHash(&blkHash)
	if err != nil {
		return err
	}
	if hBlock.HeightSpecified() && hBlock.Height() != blkNum {
		panic("cannot append a block out of sequence")
	}

	for name, wfile := range cac.blkNonEssentialInts {
		themap, err := block.NonEssentialInts()
		if err != nil {
			return err
		}
		val, success := (*themap)[name]
		if !success {
			return errors.New("could not read block non-essential int named: " + name)
		}
		err = wfile.WriteWordAt(val, blkNum)
		if err != nil {
			return err
		}
	}

	// First the "frames" of the transactions
	nTrans, err := block.TransactionCount()
	if err != nil {
		return err
	}
	if nTrans == 0 {
		panic("this code assumes at least one transaction per block")
		// Otherwise, not every entry in blkFirstTrans will be written
	}
	for t := int64(0); t < nTrans; t++ {
		hTrans, err := block.NthTransaction(t)
		if err != nil {
			return err
		}
		transNum, err := cac.appendTransactionFrame(blockChain, hTrans)
		if err != nil {
			return err
		}
		if hTrans.HeightSpecified() && hTrans.Height() != transNum {
			panic("cannot append a transaction out of sequence")
		}
		if t == 0 {
			err = cac.blkFirstTrans.WriteWordAt(transNum, blkNum)
			if err != nil {
				return err
			}
		}
	}

	// Then the "contents" of the transactions
	nTrans, err = block.TransactionCount()
	if err != nil {
		return err
	}
	if nTrans == 0 {
		panic("this code assumes at least one transaction per block")
		// Otherwise, not every entry in blkFirstTrans will be written
	}
	for t := int64(0); t < nTrans; t++ {
		hTrans, err := block.NthTransaction(t)
		if err != nil {
			return err
		}
		_, err = cac.appendTransactionContents(blockChain, hTrans)
	}

	return nil
}

func (cac *concreteAppendableChain) appendTransactionFrame(blockChain chainreadinterface.IBlockChain,
	hTrans chainreadinterface.ITransHandle) (int64, error) {
	trans, err := blockChain.TransInterface(hTrans)
	if err != nil {
		return -1, err
	}
	if !trans.HashSpecified() {
		panic("this function assumes that trans specifies a hash")
	}
	transHash, err := trans.Hash()
	if err != nil {
		return -1, err
	}
	transNum, err := cac.trnHashes.AppendHash(&transHash)
	if err != nil {
		return -1, err
	}
	if trans.HeightSpecified() && trans.Height() != transNum {
		panic("cannot append a transaction out of sequence")
	}

	return transNum, nil
}

func (cac *concreteAppendableChain) appendTransactionContents(blockChain chainreadinterface.IBlockChain,
	hTrans chainreadinterface.ITransHandle) (int64, error) {
	trans, err := blockChain.TransInterface(hTrans)
	if err != nil {
		return -1, err
	}

	// The transaction height might only be known on account of the chain we are appending to
	transHeight := int64(-1)
	if trans.HeightSpecified() {
		transHeight = trans.Height()
	} else {
		transHash, err := trans.Hash()
		if err != nil {
			return -1, err
		}
		transHeight, err = cac.trnHashes.IndexOfHash(&transHash)
		if err != nil {
			return -1, err
		}
	}

	for name, wfile := range cac.trnNonEssentialInts {
		themap, err := trans.NonEssentialInts()
		if err != nil {
			return -1, err
		}
		val, success := (*themap)[name]
		if !success {
			return -1, errors.New("could not read non-essential int named " + name)
		}
		err = wfile.WriteWordAt(val, transHeight)
		if err != nil {
			return -1, err
		}
	}

	nTxis, err := trans.TxiCount()
	if err != nil {
		return -1, err
	}
	// We MUST write to the trnFirstTxi file, REGARDLESS of whether there ARE any Txis
	putativeTxiHeight, err := cac.txiTx.CountWords() // Count all the txis by counting the txiTx field file
	if err != nil {
		return -1, err
	}
	err = cac.trnFirstTxi.WriteWordAt(putativeTxiHeight, transHeight)
	if err != nil {
		return -1, err
	}
	for nTxi := int64(0); nTxi < nTxis; nTxi++ {
		hTxi, err := trans.NthTxi(nTxi)
		if err != nil {
			return -1, err
		}
		txi, err := blockChain.TxiInterface(hTxi)
		if err != nil {
			return -1, err
		}
		txiHeight, err := cac.appendTxi(txi)
		if err != nil {
			return -1, err
		}
		if txi.TxiHeightSpecified() && txi.TxiHeight() != txiHeight {
			panic("cannot append a txi out of sequence")
		}
	}

	nTxos, err := trans.TxoCount()
	if err != nil {
		return -1, err
	}
	// We MUST write to the trnFirstTxo file, REGARDLESS of whether there ARE any Txos
	putativeTxoHeight, err := cac.txoSats.CountWords() // Count all the txis by counting the txiSats field file
	if err != nil {
		return -1, err
	}
	err = cac.trnFirstTxo.WriteWordAt(putativeTxoHeight, transHeight)
	if err != nil {
		return -1, err
	}
	for nTxo := int64(0); nTxo < nTxos; nTxo++ {
		hTxo, err := trans.NthTxo(nTxo)
		if err != nil {
			return -1, err
		}
		txo, err := blockChain.TxoInterface(hTxo)
		if err != nil {
			return -1, err
		}
		txoHeight, err := cac.appendTxo(txo)
		if err != nil {
			return -1, err
		}
		if txo.TxoHeightSpecified() && txo.TxoHeight() != txoHeight {
			panic("cannot append a txo out of sequence")
		}
	}

	return transHeight, nil
}

func (cac *concreteAppendableChain) appendTxi(txi chainreadinterface.ITxi) (int64, error) {
	// Note that we have no concept of a "coinbase txi". Instead we of course have coinbase transactions,
	// but these are DEFINED as having no txis. This is in contrast to Bitcoin Core's JSON format.

	// Txi's therefore always have a source txo

	sourceTxo, err := txi.SourceTxo()
	if err != nil {
		return -1, err
	}
	if !sourceTxo.ParentSpecified() {
		panic("this implementation assumes the txi's source txo specifies a parent transaction and index")
	}
	sourceTrans := sourceTxo.ParentTrans()
	sourceIndex := sourceTxo.ParentIndex()
	// sourceTrans is a transaction in the source chain
	// For a concreteAppendableChain, we need to store the sourceTransHeight
	// But the source chain (for example, Bitcoin Core) might not have heights for transactions
	// And furthermore, it might not even have hashes for transactions
	sourceTransHeight := int64(-1)
	// We try the following order:
	// (a) sourceTransHeight directly specified
	// (b) sourceTrans specified by indices path (block height and nthTransInBlock)
	// (c) sourceTrans specified hash
	if sourceTrans.HeightSpecified() {
		sourceTransHeight = sourceTrans.Height()
	} else if sourceTrans.IndicesPathSpecified() {
		sourceTransBlockHeight, sourceTransNthTransInBlock := sourceTrans.IndicesPath()
		// Get the trans height of the first trans in the relevant block,
		// ironically this comes from the chain we're building
		firstTransHeightInSourceTransBlock, err := cac.blkFirstTrans.ReadWordAt(sourceTransBlockHeight)
		if err != nil {
			return -1, err
		}
		sourceTransHeight = firstTransHeightInSourceTransBlock + sourceTransNthTransInBlock
	} else if sourceTrans.HashSpecified() {
		// We'll need to use the hash to determine the source transaction height, as the source hash is all we've got
		sourceTransHash, err := sourceTrans.Hash()
		if err != nil {
			return -1, err
		}
		// To determine the height of the transaction (from the hash),
		// ironically we'll have to use the chain we're appending to
		sourceTransHeight, err = cac.trnHashes.IndexOfHash(&sourceTransHash)
		if err != nil {
			return -1, err
		}
	} else {
		panic("source trans must be specified somehow")
	}
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

func (cac *concreteAppendableChain) appendTxo(txo chainreadinterface.ITxo) (int64, error) {
	sats, err := txo.Satoshis()
	if err != nil {
		return -1, err
	}
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
