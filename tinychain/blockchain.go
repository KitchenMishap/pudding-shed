package tinychain

import "github.com/KitchenMishap/pudding-shed/chainreadinterface"

// Blockchain implements IBlockchain (and so implicitly IBlockTree)
type Blockchain struct {
	blocks       []Block
	transactions []Transaction
}

// Compiler check that implements
var _ chainreadinterface.IBlockChain = (*Blockchain)(nil)

// Functions to implement IBlockTree as part of IBlockChain

func (bc *Blockchain) InvalidBlock() chainreadinterface.IBlockHandle {
	return &BlockHandle{HashHeight{-1}}
}
func (bc *Blockchain) InvalidTrans() chainreadinterface.ITransHandle {
	return &TransHandle{HashHeight{-1}}
}

func (bc *Blockchain) GenesisBlock() chainreadinterface.IBlockHandle {
	return &bc.blocks[0]
}

func (bc *Blockchain) ParentBlock(b chainreadinterface.IBlockHandle) chainreadinterface.IBlockHandle {
	if bc.IsBlockTree() {
		panic("BlockTree not supported by this code. Only Blockchains (ie longest chain) work here")
	}
	height := b.Height() - 1
	if height >= 0 {
		return &bc.blocks[height]
	}
	return bc.InvalidBlock()
}

func (bc *Blockchain) GenesisTransaction() (chainreadinterface.ITransHandle, error) {
	block := bc.GenesisBlock()
	blockInt, err := bc.BlockInterface(block)
	if err != nil {
		return &TransHandle{}, err
	}
	hGenesisTrans, err := blockInt.NthTransaction(0)
	if err != nil {
		return &TransHandle{}, err
	}
	return hGenesisTrans, nil
}

func (bc *Blockchain) PreviousTransaction(t chainreadinterface.ITransHandle) chainreadinterface.ITransHandle {
	if bc.IsBlockTree() {
		panic("BlockTree not supported by this function. Only Blockchains (ie longest chain) work here")
	}
	// We are in the tinychain package, so we know t to support transaction Height
	if !t.HeightSpecified() {
		panic("Only transaction handles indexed by height are supported by this function")
	}
	transHeight := t.Height() - 1
	transHandle := TransHandle{}
	transHandle.height = transHeight
	return &transHandle
}

func (bc *Blockchain) IsBlockTree() bool {
	// A block tree can have multiple blocks as a block's parent block
	// Not the case here (even though we derive from IBlockTree and implement its functions)
	return false
}

func (bc *Blockchain) BlockInterface(b chainreadinterface.IBlockHandle) (chainreadinterface.IBlock, error) {
	// We are in the tinychain package, so we know Heights are specified in IBlockHandles
	if !b.HeightSpecified() {
		panic("This function requires blocks to be specified by height")
	}
	blockHeight := b.Height()
	return &bc.blocks[blockHeight], nil
}

func (bc *Blockchain) TransInterface(t chainreadinterface.ITransHandle) (chainreadinterface.ITransaction, error) {
	// We are in the tinychain package, so we know Heights are specified in ITransHandles
	if !t.HeightSpecified() {
		panic("This function requires transactions to be specified by height")
	}
	transHeight := t.Height()
	return &bc.transactions[transHeight], nil
}

func (bc *Blockchain) TxoInterface(txo chainreadinterface.ITxoHandle) (chainreadinterface.ITxo, error) {
	// This is a bit of a fiddle for tinychain package. We have stored the Txos in the Transactions.
	if !txo.ParentSpecified() {
		panic("This function depends upon the txo handle having the parent specified")
	}
	parentTrans := txo.ParentTrans()
	if !parentTrans.HeightSpecified() {
		panic("This function depends on transaction handles having height specified")
	}
	parentTransHeight := parentTrans.Height()
	parentTransObject := bc.transactions[parentTransHeight]
	parentIndex := txo.ParentIndex()
	txoObject := parentTransObject.txos[parentIndex]
	return &txoObject, nil
}

func (bc *Blockchain) TxiInterface(txi chainreadinterface.ITxiHandle) (chainreadinterface.ITxi, error) {
	// This is a bit of a fiddle for tinychain package. We have stored the Txis in the Transactions.
	if !txi.ParentSpecified() {
		panic("This function depends upon the txi handle having the parent specified")
	}
	parentTrans := txi.ParentTrans()
	if !parentTrans.HeightSpecified() {
		panic("This function depends on transaction handles having height specified")
	}
	parentTransHeight := parentTrans.Height()
	parentTransObject := bc.transactions[parentTransHeight]
	parentIndex := txi.ParentIndex()
	txiObject := parentTransObject.txis[parentIndex]
	return &txiObject, nil
}

// Implement the rest of IBlockChain
func (bc *Blockchain) LatestBlock() (chainreadinterface.IBlockHandle, error) {
	blocks := len(bc.blocks)
	return &bc.blocks[blocks-1], nil
}

func (bc *Blockchain) NextBlock(handle chainreadinterface.IBlockHandle) (chainreadinterface.IBlockHandle, error) {
	blocks := len(bc.blocks)
	if !handle.HeightSpecified() {
		panic("This function depends on block handles specifying block height")
	}
	nextHeight := handle.Height() + 1
	next := BlockHandle{}
	if nextHeight == int64(blocks) {
		next.height = -1
	} else {
		next.height = nextHeight
	}
	return &next, nil
}

func (bc *Blockchain) LatestTransaction() (chainreadinterface.ITransHandle, error) {
	transactions := len(bc.transactions)
	latest := TransHandle{}
	latest.height = int64(transactions - 1)
	return &latest, nil
}

func (bc *Blockchain) NextTransaction(t chainreadinterface.ITransHandle) (chainreadinterface.ITransHandle, error) {
	// This returns a height of -1 for next after last transaction
	if !t.HeightSpecified() {
		panic("This function depends on transaction handles which specify transaction height")
	}
	nextHeight := t.Height() + 1
	transactions := len(bc.transactions)
	if nextHeight == int64(transactions) {
		nextHeight = -1
	}
	next := TransHandle{}
	next.height = nextHeight
	return &next, nil
}
