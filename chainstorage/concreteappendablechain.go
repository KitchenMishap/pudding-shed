package chainstorage

import (
	"errors"
	"fmt"
	"github.com/KitchenMishap/pudding-shed/chainreadinterface"
	"github.com/KitchenMishap/pudding-shed/indexedhashes"
	"github.com/KitchenMishap/pudding-shed/intarrayarray"
	"github.com/KitchenMishap/pudding-shed/testpoints"
	"github.com/KitchenMishap/pudding-shed/transactionindexing"
	"github.com/KitchenMishap/pudding-shed/wordfile"
)

var PrevFirstTxo int64
var PrevTrans int64

type concreteAppendableChain struct {
	blkHashes           indexedhashes.HashReadWriter
	trnFirstTxi         wordfile.ReadWriteAtWordCounter
	trnFirstTxo         wordfile.ReadWriteAtWordCounter
	txiTx               wordfile.ReadWriteAtWordCounter
	txiVout             wordfile.ReadWriteAtWordCounter
	txoSats             wordfile.ReadWriteAtWordCounter
	txoAddress          wordfile.ReadWriteAtWordCounter
	txoSpentTxi         wordfile.ReadWriteAtWordCounter
	blkNonEssentialInts map[string]wordfile.ReadWriteAtWordCounter
	trnNonEssentialInts map[string]wordfile.ReadWriteAtWordCounter
	addrHashes          indexedhashes.HashReadWriter
	addrFirstTxo        wordfile.ReadWriteAtWordCounter
	addrAdditionalTxos  intarrayarray.IntArrayMapStoreReadWrite
	parentTransOfTxi    wordfile.ReadWriteAtWordCounter
	parentTransOfTxo    wordfile.ReadWriteAtWordCounter

	transactionIndexingIsDelegated bool
	// The following must only be written to by this class if the above is false
	// (if true, they are written to by an external actor via IDelegatedTransactionIndexing)
	blkFirstTrans      wordfile.ReadWriteAtWordCounter
	trnHashes          indexedhashes.HashReadWriter
	parentBlockOfTrans wordfile.ReadWriteAtWordCounter
}

// Check that implements
var _ IAppendableChain = (*concreteAppendableChain)(nil)

func (cac *concreteAppendableChain) GetAsConcreteReadableChain() *concreteReadableChain {
	result := concreteReadableChain{
		blkFirstTrans:      cac.blkFirstTrans,
		blkHashes:          cac.blkHashes,
		trnHashes:          cac.trnHashes,
		addrHashes:         cac.addrHashes,
		trnFirstTxi:        cac.trnFirstTxi,
		trnFirstTxo:        cac.trnFirstTxo,
		txiTx:              cac.txiTx,
		txiVout:            cac.txiVout,
		txoSats:            cac.txoSats,
		txoAddress:         cac.txoAddress,
		txoSpentTxi:        cac.txoSpentTxi,
		addrFirstTxo:       cac.addrFirstTxo,
		addrAdditionalTxos: cac.addrAdditionalTxos,
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

	// === TestPoint ===
	if testpoints.TestPointBlockEnable && hBlock.HeightSpecified() && hBlock.Height() == testpoints.TestPointBlock {
		fmt.Println("TESTPOINT: concreteAppendableChain.AppendBlock(", testpoints.TestPointBlock, ")")
	}

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

	// We now assume that all hashes have already been stored and indexed
	//blkNum, err := cac.blkHashes.AppendHash(&blkHash)
	blkNum, err := cac.blkHashes.IndexOfHash(&blkHash)
	if err != nil {
		return err
	}
	if blkNum == -1 {
		return errors.New("must be able to find index of block hash")
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
	firstTransOfBlock := int64(-1)
	for t := int64(0); t < nTrans; t++ {
		hTrans, err := block.NthTransaction(t)
		if err != nil {
			return err
		}
		transNum, err := cac.appendTransactionFrame(blockChain, blkNum, hTrans)
		if t == 0 {
			// We will need this soon, as duplicate transaction hashes
			// in blocks 91812 and 91842 mean we can't look up by hash!
			firstTransOfBlock = transNum
		}
		if err != nil {
			return err
		}
		if hTrans.HeightSpecified() && hTrans.Height() != transNum {
			panic("cannot append a transaction out of sequence")
		}
		if t == 0 && !cac.transactionIndexingIsDelegated {
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
		_, err = cac.appendTransactionContents(blockChain, hTrans, firstTransOfBlock+t)
	}

	return nil
}

func (cac *concreteAppendableChain) appendTransactionFrame(blockChain chainreadinterface.IBlockChain,
	blkNum int64, hTrans chainreadinterface.ITransHandle) (int64, error) {
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
	if !cac.transactionIndexingIsDelegated {
		// We now assume that all hashes have already been stored and indexed
		//transNum, err := cac.trnHashes.AppendHash(&transHash)
		transNum, err := cac.trnHashes.IndexOfHash(&transHash)
		if err != nil {
			return -1, err
		}
		if trans.HeightSpecified() && trans.Height() != transNum {
			panic("cannot append a transaction out of sequence")
		}
		err = cac.parentBlockOfTrans.WriteWordAt(blkNum, transNum)
		if err != nil {
			return -1, err
		}
		return transNum, nil
	}
	if hTrans.HeightSpecified() {
		return hTrans.Height(), nil
	} else if hTrans.IndicesPathSpecified() {
		blkFirstTrans, err := cac.RetrieveBlockHeightToFirstTrans(blkNum)
		if err != nil {
			return -1, err
		}
		_, nthTrans := hTrans.IndicesPath()
		transNum := blkFirstTrans + nthTrans
		return transNum, nil
	} else if hTrans.HashSpecified() {
		_, err := hTrans.Hash()
		if err != nil {
			return -1, err
		}
		panic("Can't do the following due to duplicate hashes in chain!")
		//transNum, err := cac.RetrieveTransHashToHeight(&sha256)
		//if err != nil {
		//	return -1, err
		//}
		//return transNum, nil
	} else {
		return -1, errors.New("no way to find transaction height")
	}
}

func (cac *concreteAppendableChain) appendTransactionContents(blockChain chainreadinterface.IBlockChain,
	hTrans chainreadinterface.ITransHandle, transHeight int64) (int64, error) {
	trans, err := blockChain.TransInterface(hTrans)
	if err != nil {
		return -1, err
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

	// Txis can (we hope) be written in parallel to txos. We therefore branch off a goroutine for the txis
	myChan := make(chan error)
	go func() {
		nTxis, err := trans.TxiCount()
		if err != nil {
			myChan <- err
			return
		}
		// We MUST write to the trnFirstTxi file, REGARDLESS of whether there ARE any Txis
		putativeTxiHeight, err := cac.txiTx.CountWords() // Count all the txis by counting the txiTx field file
		if err != nil {
			myChan <- err
			return
		}
		err = cac.trnFirstTxi.WriteWordAt(putativeTxiHeight, transHeight)
		if err != nil {
			myChan <- err
			return
		}
		for nTxi := int64(0); nTxi < nTxis; nTxi++ {
			hTxi, err := trans.NthTxi(nTxi)
			if err != nil {
				myChan <- err
				return
			}
			txi, err := blockChain.TxiInterface(hTxi)
			if err != nil {
				myChan <- err
				return
			}
			txiHeight, err := cac.appendTxi(txi, transHeight)
			if err != nil {
				myChan <- err
				return
			}
			if txi.TxiHeightSpecified() && txi.TxiHeight() != txiHeight {
				panic("cannot append a txi out of sequence")
			}
		}
		myChan <- nil
	}()

	nTxos, err := trans.TxoCount()
	if err != nil {
		return -1, err
	}
	// We MUST write to the trnFirstTxo file, REGARDLESS of whether there ARE any Txos
	putativeTxoHeight, err := cac.txoSats.CountWords() // Count all the txos by counting the txoSats field file
	if err != nil {
		return -1, err
	}
	if PrevFirstTxo != -1 {
		if putativeTxoHeight < PrevFirstTxo {
			panic("txo going backwards")
		}
		if transHeight < PrevTrans {
			panic("trans going backwards")
		}
	}
	err = cac.trnFirstTxo.WriteWordAt(putativeTxoHeight, transHeight)
	PrevFirstTxo = putativeTxoHeight
	PrevTrans = transHeight

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
		txoHeight, err := cac.appendTxo(blockChain, txo, transHeight)
		if err != nil {
			return -1, err
		}
		if txo.TxoHeightSpecified() && txo.TxoHeight() != txoHeight {
			panic("cannot append a txo out of sequence")
		}
	}

	// Wait for the txis to be done
	err = <-myChan
	if err != nil {
		return -1, err
	}

	return transHeight, nil
}

func (cac *concreteAppendableChain) appendTxi(txi chainreadinterface.ITxi, transIndex int64) (int64, error) {
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
	err = cac.parentTransOfTxi.WriteWordAt(transIndex, txiHeight)
	if err != nil {
		return -1, err
	}

	/* This was causing errors. DO AWAY WITH IT FOR NOW
	// This is a txi, so a txo has been spent to it
	// We need to tell the txo
	// First, find the txo
	firstTxoOfTrans, err := cac.trnFirstTxo.ReadWordAt(sourceTransHeight)
	if err != nil {
		return -1, err
	}
	txoHeight := firstTxoOfTrans + sourceIndex
	// Then write to the txo
	err = cac.txoSpentTxi.WriteWordAt(txiHeight, txoHeight)
	if err != nil {
		return -1, err
	}*/

	return txiHeight, nil
}

func (cac *concreteAppendableChain) appendTxo(blockChain chainreadinterface.IBlockChain, txo chainreadinterface.ITxo, transIndex int64) (int64, error) {
	// Check we are storing txos in sequence
	txoHeight, err := cac.txoSats.CountWords()
	if err != nil {
		return -1, err
	}
	if txo.TxoHeightSpecified() && txo.TxoHeight() != txoHeight {
		panic("cannot append a txo out of sequence")
	}

	// Store the sats
	sats, err := txo.Satoshis()
	if err != nil {
		return -1, err
	}
	err = cac.txoSats.WriteWordAt(sats, txoHeight)
	if err != nil {
		return -1, err
	}
	err = cac.parentTransOfTxo.WriteWordAt(transIndex, txoHeight)
	if err != nil {
		return -1, err
	}

	txoAddressHandle, err := txo.Address()
	if err != nil {
		return -1, err
	}
	txoAddress, err := blockChain.AddressInterface(txoAddressHandle)
	if err != nil {
		return -1, err
	}
	txoAddressHash := txoAddress.Hash()

	// We now assume that hashes of addresses are already stored and indexed.

	// To know whether we've not encountered an address before (when we DO know
	// that its hash has been stored and indexed), we'll have to check the size
	// of the addrFirstTxo file.

	// We'll first have to get the height of the address, using its hash
	addressHeight, err := cac.addrHashes.IndexOfHash(&txoAddressHash)
	if err != nil {
		return -1, err
	}
	if addressHeight == -1 {
		return -1, errors.New("hash of address should already be known")
	}
	fileCount, err := cac.addrFirstTxo.CountWords()
	if err != nil {
		return -1, err
	}
	addressEncountered := (fileCount > addressHeight)

	if addressEncountered {
		abc := 123
		abc++ // Breakpoint here to find first re-used address in the blockchain
		// Transaction f4184fc596403b9d638783cf57adfe4c75c605f6356fbc91338530e9831e9e16
		// contains the first address re-use in the blockchain. It is transaction index 171, and is in block 170.
		// It is a block reward going to two txos. txoHeight is 172, and addressHeight is 9 (the previous occurrence
		// of the address). It is NOT TRUE that "From here on, addressHeight should always be less than txoHeight."
		// It is not true, because the initial list of address hashes is not a list of unique hashes.
	}

	// If we've not encountered an address, this is the first txo that uses it
	if !addressEncountered {
		err = cac.addrFirstTxo.WriteWordAt(txoHeight, addressHeight)
		if err != nil {
			return -1, err
		}
	} else {
		// We've seen this address before. Add this txo to the address's list of additional txos
		err = cac.addrAdditionalTxos.AppendToArray(addressHeight, txoHeight)
		if err != nil {
			return -1, err
		}
	}

	// This txo needs to reference the address height
	err = cac.txoAddress.WriteWordAt(addressHeight, txoHeight)
	if err != nil {
		return -1, err
	}

	// This txo is so far unspent
	err = cac.txoSpentTxi.WriteWordAt(0, txoHeight)
	if err != nil {
		return -1, err
	}

	return txoHeight, nil
}

func (cac *concreteAppendableChain) Close() {
	cac.blkHashes.Close()
	cac.trnHashes.Close()
	cac.addrHashes.Close()
	cac.blkFirstTrans.Close()
	cac.trnFirstTxi.Close()
	cac.trnFirstTxo.Close()
	cac.txiTx.Close()
	cac.txiVout.Close()
	cac.txoSats.Close()
	cac.txoSpentTxi.Close()
	cac.txoAddress.Close()
	cac.addrFirstTxo.Close()
	cac.addrAdditionalTxos.FlushFile()
	cac.parentBlockOfTrans.Close()
	cac.parentTransOfTxi.Close()
	cac.parentTransOfTxo.Close()
}

func (cac *concreteAppendableChain) GetAsDelegatedTransactionIndexer() transactionindexing.ITransactionIndexer {
	if !cac.transactionIndexingIsDelegated {
		panic("This call assumes transaction indexing is delegated")
	}
	return cac
}

// Functions to implement concreteAppendableChain as an IDelegatedTrasactionIndexing
func (cac *concreteAppendableChain) StoreTransHashToHeight(sha256 *indexedhashes.Sha256, transHeight int64) error {
	// All hashes are now presumed already stored and indexed. Nothing to do.
	return nil
}
func (cac *concreteAppendableChain) StoreTransHeightToParentBlock(transHeight int64, parentBlockHeight int64) error {
	return cac.parentBlockOfTrans.WriteWordAt(parentBlockHeight, transHeight)
}
func (cac *concreteAppendableChain) StoreBlockHeightToFirstTrans(blockHeight int64, firstTrans int64) error {
	return cac.blkFirstTrans.WriteWordAt(firstTrans, blockHeight)
}
func (cac *concreteAppendableChain) RetrieveTransHashToHeight(sha256 *indexedhashes.Sha256) (int64, error) {
	// Note: This isn't as simple as it sounds... There are two identical transactions in blocks 91812 and
	// 91842 with identical hashes! We won't get both in this case of course.
	height, err := cac.trnHashes.IndexOfHash(sha256)
	if err != nil {
		fmt.Println("Error when: RetrieveTransHashToHeight(hash)")
	}
	return height, err
}
func (cac *concreteAppendableChain) RetrieveTransHeightToParentBlock(transHeight int64) (int64, error) {
	return cac.parentBlockOfTrans.ReadWordAt(transHeight)
}
func (cac *concreteAppendableChain) RetrieveBlockHeightToFirstTrans(blockHeight int64) (int64, error) {
	return cac.blkFirstTrans.ReadWordAt(blockHeight)
}
func (cac *concreteAppendableChain) Sync() error {
	err := cac.blkHashes.Sync()
	if err != nil {
		return err
	}
	err = cac.trnHashes.Sync()
	if err != nil {
		return err
	}
	err = cac.addrHashes.Sync()
	if err != nil {
		return err
	}
	err = cac.blkFirstTrans.Sync()
	if err != nil {
		return err
	}
	err = cac.trnFirstTxi.Sync()
	if err != nil {
		return err
	}
	err = cac.trnFirstTxo.Sync()
	if err != nil {
		return err
	}
	err = cac.txiTx.Sync()
	if err != nil {
		return err
	}
	err = cac.txiVout.Sync()
	if err != nil {
		return err
	}
	err = cac.txoSats.Sync()
	if err != nil {
		return err
	}
	err = cac.txoSpentTxi.Sync()
	if err != nil {
		return err
	}
	err = cac.txoAddress.Sync()
	if err != nil {
		return err
	}
	err = cac.addrFirstTxo.Sync()
	if err != nil {
		return err
	}
	err = cac.addrAdditionalTxos.Sync()
	if err != nil {
		return err
	}
	err = cac.parentBlockOfTrans.Sync()
	if err != nil {
		return err
	}
	err = cac.parentTransOfTxi.Sync()
	if err != nil {
		return err
	}
	err = cac.parentTransOfTxo.Sync()
	if err != nil {
		return err
	}
	for _, v := range cac.blkNonEssentialInts {
		err = v.Sync()
		if err != nil {
			return err
		}
	}
	for _, v := range cac.trnNonEssentialInts {
		err = v.Sync()
		if err != nil {
			return err
		}
	}

	return nil
}

func (cac *concreteAppendableChain) SelfTestTransHashes() error {
	return nil
	//return cac.trnHashes.SelfTest()
}

func (cac *concreteAppendableChain) CountHashes() (blocks int64, transactions int64, addresses int64, err error) {
	blocks, err = cac.blkHashes.CountHashes()
	if err != nil {
		return -1, -1, -1, err
	}
	transactions, err = cac.trnHashes.CountHashes()
	if err != nil {
		return -1, -1, -1, err
	}
	addresses, err = cac.addrHashes.CountHashes()
	if err != nil {
		return -1, -1, -1, err
	}
	return blocks, transactions, addresses, nil
}
