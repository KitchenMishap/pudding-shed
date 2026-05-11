package intrinsicobjectscri

import (
	"errors"
	"fmt"
	"sort"

	"github.com/KitchenMishap/pudding-shed/chainreadinterface"
	"github.com/KitchenMishap/pudding-shed/intrinsicobjects"
	"github.com/KitchenMishap/pudding-shed/transactionindexing"
)

// OneBlockHolder provides an IBlockchain restricted to accessing blocks in sequence.
// GenesisBlock() must be called first, and only once. At this point OneBlockHolder takes an
// *intrinsicobjects.Block from its InChan and presumes it to be the genesis block.
// Subsequent calls to NextBlock must be in sequence, and cause another *intrinsicobjects.Block to be
// taken from InChan. Out of sequence blocks will generate errors, based on block's PrevHash field
type OneBlockHolder struct {
	InChan             chan *intrinsicobjects.Block
	currentBlock       *Block // intrinsicobjects.Block are converted to intrinsicobjectscri.Block on receipt
	currentBlockHeight int64  // inferred from block sequence injected

	timestampsForMedian []uint32

	// We'll sometimes need the following.
	// We send it indexing info as we discover it.
	// We retrieve that info as we need it.
	transactionIndexer       transactionindexing.ITransactionIndexer
	latestTransactionVisited int64
}

func CreateOneBlockHolder(transactionIndexer transactionindexing.ITransactionIndexer) *OneBlockHolder {
	res := OneBlockHolder{
		InChan:             make(chan *intrinsicobjects.Block),
		currentBlock:       nil,
		currentBlockHeight: -1,

		timestampsForMedian: make([]uint32, 0, 12), // Empty, capacity for 11 plus 1 spare

		transactionIndexer:       transactionIndexer,
		latestTransactionVisited: -1,
	}
	return &res
}

// Functions in intrinsicobjectscri.OneBlockChain to implement chainreadinterface.IBlockTree
// as part of chainreadinterface.IBlockChain

func (obh *OneBlockHolder) InvalidBlock() chainreadinterface.IBlockHandle {
	bh := BlockHandle{}
	bh.isInvalid = true
	return &bh
}

func (obh *OneBlockHolder) InvalidTrans() chainreadinterface.ITransHandle {
	th := TransHandle{}
	th.isInvalid = true
	return &th
}

func (obh *OneBlockHolder) GenesisBlock() chainreadinterface.IBlockHandle {
	if obh.currentBlockHeight != -1 {
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
		panic(err)
	} // ToDo
	obh.currentBlockHeight = 0
	err = obh.PostIntrinsicGatherTransHashes(obh.currentBlock)
	if err != nil {
		panic(err)
	} // ToDo
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
	return obh.currentBlock
}

func (obh *OneBlockHolder) ParentBlock(_ chainreadinterface.IBlockHandle) chainreadinterface.IBlockHandle {
	panic("OneBlockHolder: ParentBlock() not supported")
}

func (obh *OneBlockHolder) GenesisTransaction() (chainreadinterface.ITransHandle, error) {
	panic("OneBlockHolder: GenesisTransaction() not supported")
}

func (obh *OneBlockHolder) PreviousTransaction(_ chainreadinterface.ITransHandle) chainreadinterface.ITransHandle {
	panic("OneBlockHolder: PreviousTransaction() not supported")
}

func (obh *OneBlockHolder) IsBlockTree() bool { return false } // This is a BlockChain not a full tree

func (obh *OneBlockHolder) BlockInterface(handle chainreadinterface.IBlockHandle) (chainreadinterface.IBlock, error) {
	if !handle.HashSpecified() {
		panic("OneBlockHolder: only supports BlockInterface() by hash")
	}
	handleHash, err := handle.Hash()
	if err != nil {
		return nil, err
	}
	currentHash, err := obh.currentBlock.Hash()
	if err != nil {
		return nil, err
	}
	if handleHash != currentHash {
		panic("OneBlockHolder: block with this hash not loaded")
	}
	return obh.currentBlock, nil
}

func (obh *OneBlockHolder) TransInterface(handle chainreadinterface.ITransHandle) (chainreadinterface.ITransaction, error) {
	if !handle.HashSpecified() {
		return nil, errors.New("this function assumes hash is specified in ITransHandle")
	}
	hash, err := handle.Hash()
	if err != nil {
		return nil, err
	}
	index, ok := obh.currentBlock.txidMap[hash]
	if !ok {
		return nil, errors.New("transaction not found in current block based on hash")
	}

	trans := obh.currentBlock.transactions[index]
	return trans, nil
}

// ToDo Sequencing through all txis/txos will take an unnecessary amount of time looking up hashes of transactions?

func (obh *OneBlockHolder) TxiInterface(handle chainreadinterface.ITxiHandle) (chainreadinterface.ITxi, error) {
	if !handle.ParentSpecified() {
		return nil, errors.New("this function assumes Parent is specified in ITxiHandle")
	}
	parentHandle := handle.ParentTrans()
	transHash, err := parentHandle.Hash()
	if err != nil {
		return nil, err
	}
	transIndex, ok := obh.currentBlock.txidMap[transHash]
	if !ok {
		return nil, errors.New("transaction not found in current block based on hash")
	}
	return &obh.currentBlock.transactions[transIndex].puddingShedTxis[handle.ParentIndex()], nil
}

func (obh *OneBlockHolder) TxoInterface(handle chainreadinterface.ITxoHandle) (chainreadinterface.ITxo, error) {
	if !handle.ParentSpecified() {
		panic("this function assumes Parent is specified in ITxoHandle")
	}
	parentHandle := handle.ParentTrans()
	transHash, err := parentHandle.Hash()
	if err != nil {
		return nil, err
	}
	transIndex, ok := obh.currentBlock.txidMap[transHash]
	if !ok {
		return nil, errors.New("transaction not found in current block based on hash")
	}
	return &obh.currentBlock.transactions[transIndex].txos[handle.ParentIndex()], nil
}

func (obh *OneBlockHolder) AddressInterface(handle chainreadinterface.IAddressHandle) (chainreadinterface.IAddress, error) {
	// intrinsicchain.AddressHandle sneakily supports chainreadinterface.IAddress with limited functionality, so
	// we use one of those
	if handle.HashSpecified() {
		result := AddressHandle{}
		result.puddingHash3 = handle.Hash()
		return &result, nil
	}
	return nil, errors.New("intrinsicobjectscri.OneBlockChain.AddressInterface(): This code depends on the address handle specifying a hash")
}

// Functions in intrinsicchain.OneBlockChain to implement chainreadinterface.IBlockChain

func (obh *OneBlockHolder) LatestBlock() (chainreadinterface.IBlockHandle, error) {
	panic("OneBlockHolder: LatestBlock() not supported")
}

func (obh *OneBlockHolder) NextBlock(bh chainreadinterface.IBlockHandle) (chainreadinterface.IBlockHandle, error) {
	if bh.HashSpecified() {
		hash, err := bh.Hash()
		if err != nil {
			return nil, err
		}
		if hash == obh.currentBlock.intrinsic.BlockHash {
			originalBlockHash := hash
			newIntrinsicBlock := <-obh.InChan
			if newIntrinsicBlock == nil {
				// No more blocks
				return obh.InvalidBlock(), nil
			}
			timestamp := newIntrinsicBlock.Time
			mediantime := obh.PostIntrinsicCalculateMedianTime(timestamp)
			// Convert incoming intrinsicobjects.Block to intrinisicobjectscri.Block
			obh.currentBlockHeight++
			obh.currentBlock, err = NewBlock(newIntrinsicBlock, obh.currentBlockHeight, mediantime)
			if err != nil {
				return nil, err
			}
			if obh.currentBlock.intrinsic.PrevHash != originalBlockHash {
				panic("blocks supplied out of sequence")
			}
			err = obh.PostIntrinsicGatherTransHashes(obh.currentBlock)
			if err != nil {
				return nil, err
			}
			return obh.currentBlock, nil
		}
		panic("Wrong block seems to be loaded")
	}
	panic("this fn assumes that block hash is specified")
}

func (obh *OneBlockHolder) LatestTransaction() (chainreadinterface.ITransHandle, error) {
	panic("OneBlockHolder: LatestTransaction() not supported")
}

func (obh *OneBlockHolder) NextTransaction(_ chainreadinterface.ITransHandle) (chainreadinterface.ITransHandle, error) {
	panic("OneBlockHolder: NextTransaction() not supported (but probably could be)")
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
	for nthTrans := range block.transactions {
		transHeight++
		err = obh.transactionIndexer.StoreTransHeightToParentBlock(transHeight, blockHeight)
		if err != nil {
			return err
		}
		transPtr := &(block.transactions[nthTrans])
		err = obh.transactionIndexer.StoreTransHashToHeight(&((*transPtr).intrinsic.TxId), transHeight)
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
