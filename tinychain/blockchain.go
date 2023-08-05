package tinychain

import "github.com/KitchenMishap/pudding-shed/chainreadinterface"

type blockchain struct {
	blocks       []block
	transactions []transaction
}

// Implement IBlockTree
func (bc *blockchain) GenesisBlock() chainreadinterface.HBlock {
	return theHandles.HBlockFromHeight(0)
}

func (bc *blockchain) ParentBlock(b chainreadinterface.HBlock) chainreadinterface.HBlock {
	// This will return a height of -1 for parent of genesis block
	height := theHandles.HeightFromHBlock(b)
	return theHandles.HBlockFromHeight(height - 1)
}

func (bc *blockchain) GenesisTransaction() chainreadinterface.HTransaction {
	hgb := bc.GenesisBlock()
	igb := bc.BlockInterface(hgb)
	return igb.NthTransactionHandle(0)
}

func (bc *blockchain) PreviousTransaction(t chainreadinterface.HTransaction) chainreadinterface.HTransaction {
	// This will return a height of -1 for previous to genesis transaction
	transHeight := theHandles.HeightFromHTransaction(t)
	return theHandles.HTransactionFromHeight(transHeight - 1)
}

// Implement the rest of IBlockChain
func (bc *blockchain) LatestBlock() chainreadinterface.HBlock {
	return theHandles.HBlockFromHeight(int64(len(bc.blocks) - 1))
}

func (bc *blockchain) NextBlock(hb chainreadinterface.HBlock) chainreadinterface.HBlock {
	// This will return a height of -1 for next after latest block
	if hb == bc.LatestBlock() {
		return theHandles.HBlockFromHeight(-1)
	}
	blockHeight := theHandles.HeightFromHBlock(hb)
	return theHandles.HBlockFromHeight(blockHeight + 1)
}

func (bc *blockchain) BlockInterface(hBlock chainreadinterface.HBlock) chainreadinterface.IBlock {
	blockHeight := theHandles.HeightFromHBlock(hBlock)
	b := bc.blocks[blockHeight]
	return &b
}

func (bc *blockchain) LatestTransaction() chainreadinterface.HTransaction {
	transHeight := int64(len(bc.transactions) - 1)
	return theHandles.HTransactionFromHeight(transHeight)
}

func (bc *blockchain) NextTransaction(t chainreadinterface.HTransaction) chainreadinterface.HTransaction {
	// This returns a height of -1 for next after last transaction
	transHeight := theHandles.HeightFromHTransaction(t) + 1
	if transHeight == int64(len(bc.transactions)) {
		transHeight = -1
	}
	return theHandles.HTransactionFromHeight(transHeight)
}

func (bc *blockchain) TransactionInterface(hTransaction chainreadinterface.HTransaction) chainreadinterface.ITransaction {
	th := theHandles.HeightFromHTransaction(hTransaction)
	return &bc.transactions[th]
}

// Compiler check that implements
var _ chainreadinterface.IBlockChain = (*blockchain)(nil)
