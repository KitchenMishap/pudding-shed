package intrinsicobjectschi

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"sort"

	"github.com/KitchenMishap/pudding-shed/chainhandleinterface"
	"github.com/KitchenMishap/pudding-shed/intrinsicobjects"
	"github.com/KitchenMishap/pudding-shed/transactionindexing"
)

type OneBlockHolder struct {
	InChan       chan *intrinsicobjects.Block
	currentBlock *Block // intrinsicobjects.Block are converted to intrinsicobjectschi.Block on receipt

	timestampsForMedian []uint32

	transactionIndexer       transactionindexing.ITransactionIndexer
	latestTransactionVisited int64
}

// intrinsicobjectschi.OneBlockHolder implements chainhandleinterface.IBlockChain
var _ chainhandleinterface.IBlockChain = (*OneBlockHolder)(nil) // Check that implements

func CreateOneBlockHolder(transactionIndexer transactionindexing.ITransactionIndexer) *OneBlockHolder {
	result := OneBlockHolder{}
	result.InChan = make(chan *intrinsicobjects.Block)
	result.currentBlock = nil

	result.timestampsForMedian = make([]uint32, 0, 12) // Empty, capacity for 11 plus 1 spare

	result.transactionIndexer = transactionIndexer
	result.latestTransactionVisited = -1

	return &result
}

func (obh *OneBlockHolder) GenesisBlock() (chainhandleinterface.BlockHandle, error) {
	if obh.currentBlock != nil {
		panic("OneBlockHolder: Can only visit Genesis block once")
	}
	fmt.Println("Attempting to receive genesis block...")
	// Convert an intrinsicobjects.Block to an intrinsicobjectscri.Block on receipt
	var err error
	intrinsicBlock := <-obh.InChan
	timestamp := intrinsicBlock.Time
	mediantime := obh.PostIntrinsicCalculateMedianTime(timestamp)
	obh.currentBlock, err = NewBlock(intrinsicBlock, 0, mediantime)
	if err != nil {
		return makeInvalidBlockHandle(), err
	}
	err = obh.PostIntrinsicGatherTransHashes(obh.currentBlock)
	if err != nil {
		return makeInvalidBlockHandle(), err
	}
	fmt.Println("...Received genesis block")

	if obh.currentBlock == nil {
		panic("OneBlockHolder: First block was nil")
	}
	// Sample four bytes of the genesis hash
	if obh.currentBlock.intrinsic.BlockHash[0] != 0x6f {
		panic("OneBlockHolder: First block was not Genesis block")
	}
	if obh.currentBlock.intrinsic.BlockHash[1] != 0xe2 {
		panic("OneBlockHolder: First block was not Genesis block")
	}
	if obh.currentBlock.intrinsic.BlockHash[2] != 0x8c {
		panic("OneBlockHolder: First block was not Genesis block")
	}
	if obh.currentBlock.intrinsic.BlockHash[3] != 0x0a {
		panic("OneBlockHolder: First block was not Genesis block")
	}
	blockHandle := chainhandleinterface.BlockHandle{}
	blockHandle.Bytes = [4]byte{} // Block handle is block height (zero for this genesis block)
	return blockHandle, nil
}

func (obh *OneBlockHolder) NextBlock(handle chainhandleinterface.BlockHandle) (chainhandleinterface.BlockHandle, error) {
	providedBlockHeight := int64(binary.LittleEndian.Uint32(handle.Bytes[:]))
	if providedBlockHeight == obh.currentBlock.blockHeight {
		newIntrinsicBlock := <-obh.InChan
		if newIntrinsicBlock == nil {
			// No more blocks
			return obh.InvalidBlockHandle(), nil
		} // No next block
		timestamp := newIntrinsicBlock.Time
		mediantime := obh.PostIntrinsicCalculateMedianTime(timestamp)
		// Convert incoming intrinsicobjects.Block to intrinisicobjectschi.Block
		currentBlockHeight := providedBlockHeight + 1
		var err error
		existingBlockHash := obh.currentBlock.intrinsic.BlockHash
		obh.currentBlock, err = NewBlock(newIntrinsicBlock, currentBlockHeight, mediantime)
		if err != nil {
			return obh.InvalidBlockHandle(), err
		}
		if obh.currentBlock.intrinsic.PrevHash != existingBlockHash {
			panic("blocks supplied out of sequence")
		}
		err = obh.PostIntrinsicGatherTransHashes(obh.currentBlock)
		if err != nil {
			return makeInvalidBlockHandle(), err
		}
		return makeBlockHandle(obh.currentBlock.blockHeight), nil
	}
	panic("Specified block handle must match that already held")
}

func (obh *OneBlockHolder) InvalidBlockHandle() chainhandleinterface.BlockHandle {
	return makeInvalidBlockHandle()
}

func (obh *OneBlockHolder) IsBlockHandleInvalid(handle chainhandleinterface.BlockHandle) bool {
	return isInvalidBlockHandle(handle)
}

func (obh *OneBlockHolder) GetBlockInfo(blockHandle chainhandleinterface.BlockHandle,
	receiver chainhandleinterface.IBlockReceiver) error {
	receiver.ResetReceiver()
	// Check that we are being asked for the block we are holding
	if blockHandle != makeBlockHandle(obh.currentBlock.blockHeight) {
		return errors.New("not holding this block")
	}
	// Send info about the current block to the receiver
	receiver.ReceiveBlockHeight(obh.currentBlock.blockHeight)
	receiver.ReceiveBlockHash(obh.currentBlock.intrinsic.BlockHash)
	receiver.ReceiveIntField("time", int64(obh.currentBlock.intrinsic.Time))
	receiver.ReceiveIntField("mediantime", int64(obh.currentBlock.medianTime))
	receiver.ReceiveIntField("size", int64(obh.currentBlock.intrinsic.Size))
	receiver.ReceiveIntField("strippedsize", int64(obh.currentBlock.intrinsic.StrippedSize))
	receiver.ReceiveIntField("weight", int64(obh.currentBlock.intrinsic.Weight))
	receiver.ReceiveIntField("difficulty", int64(math.Round(obh.currentBlock.intrinsic.Difficulty)))
	// ToDo could send other info here, eg merkle root
	for i := range obh.currentBlock.intrinsic.Transactions {
		transHandle := chainhandleinterface.TransactionHandle{}
		transHandle.BH = blockHandle
		transHandle.TSH = makeTransactionSubHandle(int64(i))
		receiver.ReceiveTransactionHandleToAppend(transHandle)
	}
	return nil
}

func (obh *OneBlockHolder) GetTransactionInfo(transHandle chainhandleinterface.TransactionHandle,
	receiver chainhandleinterface.IBitcoinCoreTransactionReceiver) error {
	receiver.ResetReceiver()
	// Check that we are being asked for the block we are holding
	if transHandle.BH != makeBlockHandle(obh.currentBlock.blockHeight) {
		return errors.New("transaction not in this held block")
	}
	txIndex := indexInBlockFromTransactionSubHandle(transHandle.TSH)
	// Send info about the current block to the receiver
	receiver.ReceiveParentBlockHandle(transHandle.BH)
	trans := &obh.currentBlock.intrinsic.Transactions[txIndex]
	receiver.ReceiveTransactionHash(trans.TxId)
	receiver.ReceiveIntField("size", int64(trans.Size))
	receiver.ReceiveIntField("vsize", int64(trans.VSize))
	receiver.ReceiveIntField("weight", int64(trans.Weight))
	isCoinbaseTrans := txIndex == 0

	txiHandle := chainhandleinterface.TxiHandle{}
	txiHandle.TH = transHandle
	for txiIndex := range trans.BitcoinCoreTxis {
		txiHandle.TXISH = makeTxiSubHandle(int64(txiIndex))
		// pudding shed (this software) treats coinbase transactions differently from Bitcoin Core
		// in pudding shed, coinbase transactions have NO txis
		if !isCoinbaseTrans {
			receiver.ReceiveTxiHandleToAppend(txiHandle)
		}
	}

	txoHandle := chainhandleinterface.TxoHandle{}
	txoHandle.TH = transHandle
	for txoIndex := range trans.Txos {
		txoHandle.TXOSH = makeTxoSubHandle(int64(txoIndex))
		receiver.ReceiveTxoHandleToAppend(txoHandle)
	}
	return nil
}

func (obh *OneBlockHolder) GetTxiInfo(txiHandle chainhandleinterface.TxiHandle,
	receiver chainhandleinterface.ITxiReceiver) error {
	// Check that we are being asked for the block we are holding
	if txiHandle.TH.BH != makeBlockHandle(obh.currentBlock.blockHeight) {
		return errors.New("txi not in this held block")
	}
	txIndex := indexInBlockFromTransactionSubHandle(txiHandle.TH.TSH)
	trans := &obh.currentBlock.intrinsic.Transactions[txIndex]
	txiIndex := txiIndexFromTxiSubHandle(txiHandle.TXISH)
	txi := &trans.BitcoinCoreTxis[txiIndex]
	// Send info about the txi to the receiver
	receiver.ReceiveParentTransactionHandle(txiHandle.TH)
	receiver.ReceiveIncomingTxid(txi.TxId)
	receiver.ReceiveIncomingVout(txi.VOut)
	return nil
}

func (obh *OneBlockHolder) GetTxoInfo(txoHandle chainhandleinterface.TxoHandle,
	receiver chainhandleinterface.ITxoReceiver) error {
	receiver.ResetReceiver()
	// Check that we are being asked for the block we are holding
	if txoHandle.TH.BH != makeBlockHandle(obh.currentBlock.blockHeight) {
		return errors.New("txo not in this held block")
	}
	txIndex := indexInBlockFromTransactionSubHandle(txoHandle.TH.TSH)
	trans := &obh.currentBlock.intrinsic.Transactions[txIndex]
	txoIndex := txoIndexFromTxiSubHandle(txoHandle.TXOSH)
	txo := &trans.Txos[txoIndex]
	// Send info about the txo to the receiver
	receiver.ReceiveParentTransactionHandle(txoHandle.TH)
	receiver.ReceiveSatoshisValue(txo.Value)
	for byteIndex := range txo.ScriptPubKey {
		receiver.ReceiveScriptPubByteToAppend(txo.ScriptPubKey[byteIndex])
	}
	return nil
}

func (obh *OneBlockHolder) PostIntrinsicGatherTransHashes(block *Block) error {
	// (Only needs doing if we have a transaction indexer)
	if obh.transactionIndexer == nil {
		return nil
	}
	// Some things that need doing for new blocks encountered, after the intrinsic parsing is done,
	// and in the context of a growing chain of blocks
	blockHeight := int64(block.blockHeight)
	firstTransHeight := obh.latestTransactionVisited + 1
	err := obh.transactionIndexer.StoreBlockHeightToFirstTrans(blockHeight, firstTransHeight)
	if err != nil {
		return err
	}
	transHeight := obh.latestTransactionVisited
	for nthTrans := range block.intrinsic.Transactions {
		transHeight++
		err = obh.transactionIndexer.StoreTransHeightToParentBlock(transHeight, blockHeight)
		if err != nil {
			return err
		}
		transPtr := &(block.intrinsic.Transactions[nthTrans])
		err = obh.transactionIndexer.StoreTransHashToHeight(&((*transPtr).TxId), transHeight)
		if err != nil {
			return err
		}
	}
	obh.latestTransactionVisited = transHeight
	return nil
}

func (obh *OneBlockHolder) PostIntrinsicCalculateMedianTime(timestamp uint32) uint32 {
	// 1. Maintain the sliding window (max 11)
	obh.timestampsForMedian = append(obh.timestampsForMedian, timestamp)
	if len(obh.timestampsForMedian) > 11 {
		obh.timestampsForMedian = obh.timestampsForMedian[1:]
	}

	// 2. Create a temporary copy to avoid scrambling your sliding window
	n := len(obh.timestampsForMedian)
	temp := make([]uint32, n)
	copy(temp, obh.timestampsForMedian)

	// 3. Sort the timestamps numerically
	sort.Slice(temp, func(i, j int) bool {
		return temp[i] < temp[j]
	})

	// 4. The Core Rule: Pick the element at Index (Size / 2)
	// For N=1, Index 0
	// For N=2, Index 1 (The later one)
	// For N=11, Index 5 (The middle one)
	return temp[n/2]
}
