package chainstorage

import "github.com/KitchenMishap/pudding-shed/chainreadinterface"

type IAppendableChain interface {
	AppendBlock(chainreadinterface.IHandles, chainreadinterface.IBlockChain, chainreadinterface.IBlock) error
	Close() error
}

type IAppendableChainFactory interface {
	Exists() bool
	Create() error
	Open() IAppendableChain
	OpenReadOnly() chainreadinterface.IBlockChain
}
