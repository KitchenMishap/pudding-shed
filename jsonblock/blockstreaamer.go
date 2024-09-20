package jsonblock

import (
	"errors"
	"fmt"
	"github.com/KitchenMishap/pudding-shed/chainreadinterface"
	"github.com/KitchenMishap/pudding-shed/transactionindexing"
)

// OneBlockHolder provides an IBlockchain restricted to accessing blocks in sequence.
// GenesisBlock() must be called first, and only once. At this point OneBlockHolder takes a
// JsonBlockEssential* from its InChan and presumes it to be the genesis block.
// Subsequent calls to NextBlock must be in sequence, and causes another JsonBlockEssentials* to be
// taken from InChan. Out of sequence blocks will generate errors.
type OneBlockHolder struct {
	InChan                   chan *JsonBlockEssential
	transactionIndexer       transactionindexing.ITransactionIndexer
	currentBlock             *JsonBlockEssential
	latestBlockVisited       int64
	latestTransactionVisited int64
}

func CreateOneBlockHolder(
	indexer transactionindexing.ITransactionIndexer,
) *OneBlockHolder {
	res := OneBlockHolder{
		InChan:                   make(chan *JsonBlockEssential),
		transactionIndexer:       indexer,
		currentBlock:             nil,
		latestBlockVisited:       -1,
		latestTransactionVisited: -1,
	}
	return &res
}

// Functions in jsonblock.OneBlockChain to implement chainreadinterface.IBlockTree as part of chainreadinterface.IBlockChain

func (obh *OneBlockHolder) InvalidBlock() chainreadinterface.IBlockHandle {
	bh := BlockHandle{}
	bh.height = -1
	return &bh
}

func (obh *OneBlockHolder) InvalidTrans() chainreadinterface.ITransHandle {
	th := TransHandle{}
	th.nthInBlock = -1
	return &th
}

func (obh *OneBlockHolder) GenesisBlock() chainreadinterface.IBlockHandle {
	if obh.latestBlockVisited != -1 {
		panic("OneBlockHolder: Can only visit Genesis block once")
	}
	fmt.Println("OneBlockHolder: Attempting to receive genesis block")
	obh.currentBlock = <-obh.InChan
	fmt.Println("OneBlockHolder: Received genesis block")

	// There's some processing to be done on the block, non-parallel
	obh.PostJsonGatherTransHashes(obh.currentBlock)
	PostJsonArrayIndicesIntoElements(obh.currentBlock)
	obh.PostJsonUpdateTransReferences(obh.currentBlock)
	PostJsonGatherNonEssentialInts(obh.currentBlock)

	if obh.currentBlock == nil {
		panic("OneBlockHolder: First block was nil")
	}
	if obh.currentBlock.J_height != 0 {
		panic("OneBlockHolder: First block was not height zero")
	}
	obh.latestBlockVisited = 0
	return obh.currentBlock
}

func (obh *OneBlockHolder) ParentBlock(block chainreadinterface.IBlockHandle) chainreadinterface.IBlockHandle {
	panic("OneBlockHolder: ParentBlock() not supported")
}

func (obh *OneBlockHolder) GenesisTransaction() (chainreadinterface.ITransHandle, error) {
	panic("OneBlockHolder: GenesisTransaction() not supported")
}

func (obh *OneBlockHolder) PreviousTransaction(trans chainreadinterface.ITransHandle) chainreadinterface.ITransHandle {
	panic("OneBlockHolder: PreviousTransaction() not supported")
}

func (obh *OneBlockHolder) IsBlockTree() bool { return false } // This is a BlockChain not a full tree

func (obh *OneBlockHolder) BlockInterface(handle chainreadinterface.IBlockHandle) (chainreadinterface.IBlock, error) {
	if !handle.HeightSpecified() {
		panic("OneBlockHolder: only supports BlockInterface() by height")
	}
	if handle.Height() != obh.latestBlockVisited {
		panic("OneBlockHolder: block at this height not loaded")
	}
	return obh.currentBlock, nil
}

func (obh *OneBlockHolder) TransInterface(handle chainreadinterface.ITransHandle) (chainreadinterface.ITransaction, error) {
	if !handle.IndicesPathSpecified() {
		return nil, errors.New("this function assumes indices path is specified in ITransHandle")
	}
	blockHeight, nthInBlock := handle.IndicesPath()
	if blockHeight != obh.latestBlockVisited {
		panic("TransInterface requested for different block")
	}

	trans := &obh.currentBlock.J_tx[nthInBlock]
	return trans, nil
}

func (obh *OneBlockHolder) TxiInterface(handle chainreadinterface.ITxiHandle) (chainreadinterface.ITxi, error) {
	if !handle.IndicesPathSpecified() {
		return nil, errors.New("this function assumes indices path is specified in ITxiHandle")
	}
	blockHeight, nthInBlock, vIndex := handle.IndicesPath()
	if blockHeight != obh.latestBlockVisited {
		panic("TxiInterface requested for different block")
	}

	txi := &obh.currentBlock.J_tx[nthInBlock].J_vin[vIndex]
	return txi, nil
}

func (obh *OneBlockHolder) TxoInterface(handle chainreadinterface.ITxoHandle) (chainreadinterface.ITxo, error) {
	if !handle.IndicesPathSpecified() {
		return nil, errors.New("this function assumes indices path is specified in ITxiHandle")
	}
	blockHeight, nthInBlock, vIndex := handle.IndicesPath()
	if blockHeight != obh.latestBlockVisited {
		panic("TxoInterface requested for different block")
	}

	txo := &obh.currentBlock.J_tx[nthInBlock].J_vout[vIndex]
	return txo, nil
}

func (obh *OneBlockHolder) AddressInterface(handle chainreadinterface.IAddressHandle) (chainreadinterface.IAddress, error) {
	// jsonblock.AddressHandle sneakily supports chainreadinterface.IAddress with limited functionality, so
	// we use one of those
	if handle.HashSpecified() {
		result := AddressHandle{}
		result.hash = handle.Hash()
		return &result, nil
	} else {
		return nil, errors.New("jsonblock.OneBlockChain.AddressInterface(): This code depends on the address handle specifying a hash")
	}
}

// Functions in jsonblock.OneBlockChain to implement chainreadinterface.IBlockChain

func (obh *OneBlockHolder) LatestBlock() (chainreadinterface.IBlockHandle, error) {
	panic("OneBlockHolder: LatestBlock() not supported")
}

func (obh *OneBlockHolder) NextBlock(bh chainreadinterface.IBlockHandle) (chainreadinterface.IBlockHandle, error) {
	if bh.HeightSpecified() {
		if bh.Height() == obh.latestBlockVisited {
			fmt.Println("OneBlockHolder: Attempting to receive a block")
			obh.currentBlock = <-obh.InChan
			fmt.Println("OneBlockHolder: Received a block")

			// There's some processing to be done on the block, non-parallel
			obh.PostJsonGatherTransHashes(obh.currentBlock)
			PostJsonArrayIndicesIntoElements(obh.currentBlock)
			obh.PostJsonUpdateTransReferences(obh.currentBlock)
			PostJsonGatherNonEssentialInts(obh.currentBlock)

			obh.latestBlockVisited = int64(obh.currentBlock.J_height)
			return obh.currentBlock, nil
		} else {
			panic("Block out of sequence")
		}
	} else {
		panic("Height must be specified")
	}
}

func (obh *OneBlockHolder) LatestTransaction() (chainreadinterface.ITransHandle, error) {
	panic("OneBlockHolder: LatestTransaction() not supported")
}

func (obh *OneBlockHolder) NextTransaction(transHandle chainreadinterface.ITransHandle) (chainreadinterface.ITransHandle, error) {
	panic("OneBlockHolder: NextTransaction() not supported (but probably could be)")
}

func (obh *OneBlockHolder) PostJsonGatherTransHashes(block *JsonBlockEssential) error {
	blockHeight := int64(block.J_height)
	firstTransHeight := obh.latestTransactionVisited + 1
	err := obh.transactionIndexer.StoreBlockHeightToFirstTrans(blockHeight, firstTransHeight)
	if err != nil {
		return err
	}
	transHeight := obh.latestTransactionVisited
	for nthTrans := range block.J_tx {
		transHeight++
		err = obh.transactionIndexer.StoreTransHeightToParentBlock(transHeight, blockHeight)
		if err != nil {
			return err
		}
		transPtr := &block.J_tx[nthTrans]
		err = obh.transactionIndexer.StoreTransHashToHeight(&transPtr.txid, transHeight)
		if err != nil {
			return err
		}
	}
	obh.latestTransactionVisited = transHeight
	return nil
}

// postJsonUpdateTransReferences() uses the accrued transaction hash map to
// locate the txos (by way of block/nthTrans/vindex path indices) corresponding to each txi.
func (obh *OneBlockHolder) PostJsonUpdateTransReferences(block *JsonBlockEssential) error {
	// Use the map to locate the txos (by indices path) referenced by trans hashes in the txis in this block
	for nthTrans := range block.J_tx {
		transPtr := &block.J_tx[nthTrans]
		for nthTxi := range transPtr.J_vin {
			txiPtr := &transPtr.J_vin[nthTxi]

			// Look up the path indices by source transaction hash
			sourceTransHeight, err := obh.transactionIndexer.RetrieveTransHashToHeight(&txiPtr.txid)
			if err != nil {
				return err
			}
			sourceBlockHeight, err := obh.transactionIndexer.RetrieveTransHeightToParentBlock(sourceTransHeight)
			if err != nil {
				return err
			}
			sourceBlockFirstTrans, err := obh.transactionIndexer.RetrieveBlockHeightToFirstTrans(sourceBlockHeight)
			if err != nil {
				return err
			}
			sourceNthTrans := sourceTransHeight - sourceBlockFirstTrans

			// Store the path indices of the source txo, into the txi that referenced it
			txiPtr.sourceTrans.blockHeight = sourceBlockHeight
			txiPtr.sourceTrans.nthInBlock = sourceNthTrans
			// txiPtr.J_vout is already there of course
		}
	}
	return nil
}
