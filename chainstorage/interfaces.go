package chainstorage

import "github.com/KitchenMishap/pudding-shed/chainreadinterface"

type IAppendableChain interface {
	AppendBlock(chainreadinterface.IBlockChain, chainreadinterface.IBlockHandle) error
	Close() error
}

type IAppendableChainFactory interface {
	Exists() bool
	Create() error
	Open() IAppendableChain
	OpenReadOnly() chainreadinterface.IBlockChain
}