package jsonblock

import (
	"errors"
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
		InChan:                   make(chan *JsonBlockEssential, 1),
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
	obh.currentBlock = <-obh.InChan
	if obh.currentBlock == nil {
		panic("OneBlockHolder: First block was nil")
	}
	if obh.currentBlock.J_height != 0 {
		panic("OneBlockHolder: First block was not height zero")
	}
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
			obh.currentBlock = <-obh.InChan
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
