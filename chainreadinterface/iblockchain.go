package chainreadinterface

type IBlockTree interface {
	GenesisBlock() IBlockHandle
	ParentBlock(block IBlockHandle) IBlockHandle
	GenesisTransaction() ITransHandle
	PreviousTransaction(trans ITransHandle) ITransHandle
	IsBlockTree() bool
	BlockInterface(IBlockHandle) IBlock
	TransInterface(ITransHandle) ITransaction
	TxiInterface(ITxiHandle) ITxi
	TxoInterface(ITxoHandle) ITxo
}

type IBlockChain interface {
	IBlockTree
	LatestBlock() IBlockHandle
	NextBlock(block IBlockHandle) IBlockHandle
	LatestTransaction() ITransHandle
	NextTransaction(trans ITransHandle) ITransHandle
}
