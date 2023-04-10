package chainreadinterface

type IBlockTree interface {
	GenesisBlock() HBlock
	ParentBlock(hBlock HBlock) HBlock
	GenesisTransaction() HTransaction
	PreviousTransaction(hTransaction HTransaction) HTransaction
}

type IBlockChain interface {
	IBlockTree
	LatestBlock() HBlock
	NextBlock(hBlock HBlock) HBlock
	BlockInterface(hBlock HBlock) IBlock
	LatestTransaction() HTransaction
	NextTransaction(hTransaction HTransaction) HTransaction
	TransactionInterface(hTransaction HTransaction) ITransaction
}
