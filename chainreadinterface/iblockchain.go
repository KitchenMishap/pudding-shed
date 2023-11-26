package chainreadinterface

type IBlockTree interface {
	InvalidBlock() IBlockHandle
	InvalidTrans() ITransHandle
	GenesisBlock() IBlockHandle
	ParentBlock(block IBlockHandle) IBlockHandle
	GenesisTransaction() (ITransHandle, error)
	PreviousTransaction(trans ITransHandle) ITransHandle // Todo [  ] Cannot be implemented in some cases
	IsBlockTree() bool
	BlockInterface(IBlockHandle) (IBlock, error)
	TransInterface(ITransHandle) (ITransaction, error)
	TxiInterface(ITxiHandle) (ITxi, error)
	TxoInterface(ITxoHandle) (ITxo, error)
}

type IBlockChain interface {
	IBlockTree
	LatestBlock() (IBlockHandle, error)
	NextBlock(block IBlockHandle) (IBlockHandle, error)
	LatestTransaction() (ITransHandle, error)
	NextTransaction(trans ITransHandle) (ITransHandle, error)
}
