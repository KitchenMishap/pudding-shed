package chainhandleinterface

type IBlockChain interface {
	IsBlockHandleInvalid(BlockHandle) bool
	GenesisBlock() (BlockHandle, error)
	NextBlock(BlockHandle) (BlockHandle, error)
	GetBlockInfo(BlockHandle, IBlockReceiver) error
	GetTransactionInfo(TransactionHandle, IBitcoinCoreTransactionReceiver) error
	GetTxiInfo(TxiHandle, ITxiReceiver) error
	GetTxoInfo(TxoHandle, ITxoReceiver) error
}
