package chainhandleinterface

type IBlockChain interface {
	IsBlockHandleInvalid(BlockHandle) bool
	GenesisBlock() (BlockHandle, error)
	BlockInterface(BlockHandle) (IBlock, error)
}
