package intrinsicobjectscri

import (
	"errors"
	"fmt"

	"github.com/KitchenMishap/pudding-shed/chainreadinterface"
	"github.com/KitchenMishap/pudding-shed/intrinsicobjects"
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
}

func CreateOneBlockHolder() *OneBlockHolder {
	res := OneBlockHolder{
		InChan:             make(chan *intrinsicobjects.Block),
		currentBlock:       nil,
		currentBlockHeight: -1,
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
	th.vOut = -1
	return &th
}

func (obh *OneBlockHolder) GenesisBlock() chainreadinterface.IBlockHandle {
	if obh.currentBlockHeight != -1 {
		panic("OneBlockHolder: Can only visit Genesis block once")
	}
	fmt.Println("Attempting to receive genesis block...")
	// Convert an intrinsicobjects.Block to an intrinsicobjectscri.Block on receipt
	obh.currentBlock = NewBlock(<-obh.InChan)
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
	obh.currentBlockHeight = 0
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
	if handle.Hash() != obh.currentBlock.Hash() {
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

	trans := &obh.currentBlock.transactions[index]
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
	return &obh.currentBlock.transactions[transIndex].txis[handle.ParentIndex()], nil
}

func (obh *OneBlockHolder) TxoInterface(handle chainreadinterface.ITxoHandle) (chainreadinterface.ITxo, error) {
	if !handle.ParentSpecified() {
		return nil, errors.New("this function assumes Parent is specified in ITxoHandle")
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
			obh.currentBlock = NewBlock(<-obh.InChan) // Convert incoming intrinsicobjects.Block to intrinisicobjectscri.Block
			if obh.currentBlock.intrinsic.PrevHash != originalBlockHash {
				panic("blocks supplied out of sequence")
			}
			obh.currentBlockHeight++
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
