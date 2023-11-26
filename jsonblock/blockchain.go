package jsonblock

import (
	"errors"
	"github.com/KitchenMishap/pudding-shed/chainreadinterface"
	"github.com/KitchenMishap/pudding-shed/indexedhashes"
)

// OneBlockChain - We call it a one block chain, as one block is in memory at any point in time
type OneBlockChain struct {
	blockFetcher       IBlockJsonFetcher
	transLocationStore ITransLocatorStore
	currentJsonBytes   []byte
	currentBlock       *jsonBlockEssential
}

var TheOneBlockChain = OneBlockChain{
	blockFetcher:       &hardCodedBlockFetcher{},
	transLocationStore: CreateOpenTransLocationStore("Temp_Testing\\TransLocation"),
	currentJsonBytes:   nil,
	currentBlock:       nil,
}

func (obc *OneBlockChain) switchBlock(blockHeight int64) (*jsonBlockEssential, error) {
	if obc.currentBlock == nil || int64(obc.currentBlock.J_height) != blockHeight {
		bytes, err := obc.blockFetcher.FetchBlockJsonBytes(blockHeight)
		if err != nil {
			return nil, err
		}
		block, err := parseJsonBlock(bytes)
		if err != nil {
			return nil, err
		}

		// Remove coinbase txis (instead we model coinbase transactions DEFINED as having NO entries in vin)
		err = postJsonRemoveCoinbaseTxis(block)
		if err != nil {
			return nil, err
		}

		// Convert the hash strings to binary
		err = postJsonEncodeSha256s(block)
		if err != nil {
			return nil, err
		}

		// Convert bitcoin floats to satoshi ints
		postJsonCalculateSatoshis(block)

		// Gather the transaction hashes, as we'll need to look them up, from now and forevermore
		err = postJsonGatherTransHashes(block, obc.transLocationStore)
		if err != nil {
			return nil, err
		}

		// Store in each array element (tx, txi, txo), the element's index into that array
		postJsonArrayIndicesIntoElements(block)

		err = postJsonUpdateTransReferences(block, obc.transLocationStore)
		if err != nil {
			return nil, err
		}

		obc.currentJsonBytes = bytes
		obc.currentBlock = block
	}
	return obc.currentBlock, nil
}

func postJsonRemoveCoinbaseTxis(block *jsonBlockEssential) error {
	// Coinbase Txi's are a transaction's entry in vin[] which represent the concept of coinbase in Bitcoin Core's JSON
	// We detect them by way of absence of hash string

	// Go through transactions
	for nthTrans := range block.J_tx {
		transPtr := &block.J_tx[nthTrans]
		// If there's a coinbase Txi, it will be the only one in vin
		if len(transPtr.J_vin) == 1 {
			if transPtr.J_vin[0].J_txid == "" { // This is how we identify a "coinbase txi"
				transPtr.J_vin = transPtr.J_vin[:0] // Slice to an empty array. Better than nil as encodes to empty array in JSON
			}
		}
	}
	return nil
}

func postJsonEncodeSha256s(block *jsonBlockEssential) error {
	// First the block hash
	err := indexedhashes.HashHexToSha256(block.J_hash, &block.hash)
	if err != nil {
		return err
	}

	// Then the trans hashes
	for nthTrans := range block.J_tx {
		transPtr := &block.J_tx[nthTrans]
		err = indexedhashes.HashHexToSha256(transPtr.J_txid, &transPtr.txid)
		if err != nil {
			return err
		}
		// Then the references to trans hashes in the txis
		for nthTxi := range transPtr.J_vin {
			txiPtr := &transPtr.J_vin[nthTxi]
			err = indexedhashes.HashHexToSha256(txiPtr.J_txid, &transPtr.txid)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func postJsonCalculateSatoshis(block *jsonBlockEssential) {
	const satoshisPerBitcoin = float64(100_000_000)
	for nthTrans := range block.J_tx {
		transPtr := &block.J_tx[nthTrans]
		for nthTxo := range transPtr.J_vout {
			txoPtr := &transPtr.J_vout[nthTxo]
			txoPtr.satoshis = int64(satoshisPerBitcoin * txoPtr.J_value)
		}
	}
}

func postJsonGatherTransHashes(block *jsonBlockEssential, store ITransLocatorStore) error {
	for nthTrans := range block.J_tx {
		transPtr := &block.J_tx[nthTrans]
		err := store.StoreIndicesPathForHash(transPtr.txid, int64(block.J_height), int64(nthTrans))
		if err != nil {
			return err
		}
	}
	return nil
}

func postJsonArrayIndicesIntoElements(block *jsonBlockEssential) {
	// We copy Height into each element of the Tx array,
	// And place each element's index into each element
	for nth := range block.J_tx {
		txPtr := &block.J_tx[nth]
		txPtr.handle.blockHeight = int64(block.J_height)
		txPtr.handle.nthInBlock = int64(nth)
		// Then we go deeper, into the Tx[nth]'s vin[] and vout[]
		for vinIndex := range txPtr.J_vin {
			txPtr.J_vin[vinIndex].parentTrans.blockHeight = int64(block.J_height)
			txPtr.J_vin[vinIndex].parentTrans.nthInBlock = int64(nth)
			txPtr.J_vin[vinIndex].parentVIndex = int64(vinIndex)
		}
		for voutIndex := range txPtr.J_vout {
			txPtr.J_vout[voutIndex].parentTrans.blockHeight = int64(block.J_height)
			txPtr.J_vout[voutIndex].parentTrans.nthInBlock = int64(nth)
			txPtr.J_vout[voutIndex].parentVIndex = int64(voutIndex)
		}
	}
}

// postJsonUpdateTransReferences() uses the accrued transaction hash map to
// locate the txos (by way of block/nthTrans/vindex path indices) corresponding to each txi.
func postJsonUpdateTransReferences(block *jsonBlockEssential, transStore ITransLocatorStore) error {
	// Use the map to locate the txos (by indices path) referenced by trans hashes in the txis in this block
	for nthTrans := range block.J_tx {
		transPtr := &block.J_tx[nthTrans]
		for nthTxi := range transPtr.J_vin {
			txiPtr := &transPtr.J_vin[nthTxi]
			// Look up the path indices by source transaction hash
			transIndicesPath, err := transStore.GetTransIndicesPathByHash(txiPtr.txid)
			if err != nil {
				return err
			}
			// Store the path indices of the source txo, into the txi that referenced it
			txiPtr.sourceTrans.blockHeight = transIndicesPath.BlockHeight()
			txiPtr.sourceTrans.nthInBlock = transIndicesPath.NthTransInBlock()
			// txiPtr.J_vout is already there of course
		}
	}
	return nil
}

// Functions in jsonblock.OneBlockChain to implement chainreadinterface.IBlockTree as part of chainreadinterface.IBlockChain

func (obc *OneBlockChain) InvalidBlock() chainreadinterface.IBlockHandle {
	bh := BlockHandle{}
	bh.height = -1
	return &bh
}

func (obc *OneBlockChain) InvalidTrans() chainreadinterface.ITransHandle {
	th := TransHandle{}
	th.nthInBlock = -1
	return &th
}

func (obc *OneBlockChain) GenesisBlock() chainreadinterface.IBlockHandle {
	block, err := obc.switchBlock(0)
	if err != nil {
		panic("couldn't switch to genesis block")
	}
	return block
}

func (obc *OneBlockChain) ParentBlock(block chainreadinterface.IBlockHandle) chainreadinterface.IBlockHandle {
	if block.IsInvalid() {
		return obc.InvalidBlock()
	}
	if block.HeightSpecified() {
		if block.Height() == 0 {
			return obc.InvalidBlock()
		}
		prevBlock, err := obc.switchBlock(block.Height() - 1)
		if err != nil {
			panic(err)
		}
		return prevBlock
	} else {
		panic("this function expects block to have height specified")
	}
}

func (obc *OneBlockChain) GenesisTransaction() (chainreadinterface.ITransHandle, error) {
	genesisBlock, err := obc.switchBlock(0)
	if err != nil {
		return nil, err
	}
	trans, err := genesisBlock.NthTransaction(0)
	return trans, err
}

func (obc *OneBlockChain) PreviousTransaction(trans chainreadinterface.ITransHandle) chainreadinterface.ITransHandle {
	// Todo [  ] Replace panics with error returns
	if !trans.IndicesPathSpecified() {
		panic("this function assumes indices path specified for trans")
	}
	blockHeight, nthInBlock := trans.IndicesPath()
	if nthInBlock == 0 {
		if blockHeight == 0 {
			return obc.InvalidTrans()
		}
		// Load previous block
		block, err := obc.switchBlock(blockHeight - 1)
		if err != nil {
			panic(err)
		}
		transCount, err := block.TransactionCount()
		if err != nil {
			panic(err)
		}
		trans, err := block.NthTransaction(transCount - 1)
		if err != nil {
			panic(err)
		}
		return trans
	}
	nthInBlock -= 1
	// Load block in case not loaded
	block, err := obc.switchBlock(blockHeight)
	if err != nil {
		panic(err)
	}
	nthTrans, err := block.NthTransaction(nthInBlock)
	if err != nil {
		panic(err)
	}
	return nthTrans
}

func (obc *OneBlockChain) IsBlockTree() bool { return false } // This is a BlockChain not a full tree

func (obc *OneBlockChain) BlockInterface(handle chainreadinterface.IBlockHandle) (chainreadinterface.IBlock, error) {
	if !handle.HeightSpecified() {
		panic("this function assumes block handle specifies height")
	}
	res, err := obc.switchBlock(handle.Height())
	return res, err
}

func (obc *OneBlockChain) TransInterface(handle chainreadinterface.ITransHandle) (chainreadinterface.ITransaction, error) {
	if !handle.IndicesPathSpecified() {
		return nil, errors.New("this function assumes indices path is specified in ITransHandle")
	}
	blockHeight, nthInBlock := handle.IndicesPath()
	block, err := obc.switchBlock(blockHeight)
	if err != nil {
		return nil, err
	}

	trans := &block.J_tx[nthInBlock]
	return trans, nil
}

func (obc *OneBlockChain) TxiInterface(handle chainreadinterface.ITxiHandle) (chainreadinterface.ITxi, error) {
	if !handle.IndicesPathSpecified() {
		return nil, errors.New("this function assumes indices path is specified in ITxiHandle")
	}
	blockHeight, nthInBlock, vIndex := handle.IndicesPath()
	block, err := obc.switchBlock(blockHeight)
	if err != nil {
		return nil, err
	}

	txi := &block.J_tx[nthInBlock].J_vin[vIndex]
	return txi, nil
}

func (obc *OneBlockChain) TxoInterface(handle chainreadinterface.ITxoHandle) (chainreadinterface.ITxo, error) {
	if !handle.IndicesPathSpecified() {
		return nil, errors.New("this function assumes indices path is specified in ITxoHandle")
	}
	blockHeight, nthInBlock, vIndex := handle.IndicesPath()
	block, err := obc.switchBlock(blockHeight)
	if err != nil {
		return nil, err
	}

	txo := &block.J_tx[nthInBlock].J_vout[vIndex]
	return txo, nil
}

// Functions in jsonblock.OneBlockChain to implement chainreadinterface.IBlockChain

func (obc *OneBlockChain) LatestBlock() (chainreadinterface.IBlockHandle, error) {
	count, err := obc.blockFetcher.CountBlocks()
	if err != nil {
		return nil, err
	}
	block, err := obc.switchBlock(count - 1)
	if err != nil {
		return nil, err
	}
	return block, nil
}

func (obc *OneBlockChain) NextBlock(bh chainreadinterface.IBlockHandle) (chainreadinterface.IBlockHandle, error) {
	blockHeight := bh.Height()
	count, err := obc.blockFetcher.CountBlocks()
	if err != nil {
		return nil, err
	}
	if blockHeight+1 >= count {
		return obc.InvalidBlock(), nil
	}

	block, err := obc.switchBlock(blockHeight + 1)
	if err != nil {
		return nil, err
	}
	return block, nil
}

func (obc *OneBlockChain) LatestTransaction() (chainreadinterface.ITransHandle, error) {
	count, err := obc.blockFetcher.CountBlocks()
	if err != nil {
		return nil, err
	}
	block, err := obc.switchBlock(count - 1)
	if err != nil {
		return nil, err
	}

	transCount, err := block.TransactionCount()
	if err != nil {
		return nil, err
	}
	handle := TransHandle{
		blockHeight: int64(block.J_height),
		nthInBlock:  transCount - 1,
	}
	return &handle, nil
}

func (obc *OneBlockChain) NextTransaction(transHandle chainreadinterface.ITransHandle) (chainreadinterface.ITransHandle, error) {
	if !transHandle.IndicesPathSpecified() {
		return nil, errors.New("this function assumes ITransHandle specifies indices path")
	}
	blockHeight, nthInBlock := transHandle.IndicesPath()
	block, err := obc.switchBlock(blockHeight)
	if err != nil {
		return nil, err
	}

	blockTransCount, err := block.TransactionCount()
	if err != nil {
		return nil, err
	}
	if nthInBlock+1 >= blockTransCount {
		// Next block
		nextBlock, err := obc.NextBlock(block)
		if err != nil {
			return nil, err
		}
		if nextBlock.IsInvalid() {
			return obc.InvalidTrans(), nil
		}
		// First trans in next block
		obc.switchBlock(nextBlock.Height())
		return &obc.currentBlock.J_tx[0], nil
	}
	return &obc.currentBlock.J_tx[nthInBlock+1], nil
}
