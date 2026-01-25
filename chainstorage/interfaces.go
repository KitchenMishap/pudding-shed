package chainstorage

import (
	"github.com/KitchenMishap/pudding-shed/chainreadinterface"
	"github.com/KitchenMishap/pudding-shed/jsonblock"
	"github.com/KitchenMishap/pudding-shed/transactionindexing"
	"github.com/KitchenMishap/pudding-shed/wordfile"
)

// Just stores hashes (block, transaction, address) associated with each appended block
type IAppendableHashesChain interface {
	AppendHashes(hashes *jsonblock.JsonBlockHashes) error
	Close()
	Sync() error
	CountHashes() (int64, int64, int64, error)
}

type IAppendableChain interface {
	AppendBlock(chainreadinterface.IBlockChain, chainreadinterface.IBlockHandle) error
	Close()
	Sync() error
	GetAsChainReadInterface() chainreadinterface.IBlockChain
	SelfTestTransHashes() error
}

// See also ITransactionIndexer,
// An optional part of an appendable chain that delegates construction of transaction
// hash indexing to an external actor

type IAppendableHashesChainFactory interface {
	Exists() bool
	Create() error
	Open() (IAppendableHashesChain, error)
}

type IAppendableChainFactory interface {
	Exists() bool
	Create() error
	CreateFromHashStores() error
	Open() (IAppendableChain, error)
	OpenReadOnly() (chainreadinterface.IBlockChain, chainreadinterface.IHandleCreator, IParents, IPrivilegedFiles, error)
}

type IAppendableChainFactoryWithIndexer interface {
	IAppendableChainFactory
	OpenWithIndexer() (IAppendableChain, transactionindexing.ITransactionIndexer, error)
}

type IParents interface {
	ParentBlockOfTrans(transactionHeight int64) (int64, error)
	ParentTransOfTxi(transactionHeight int64) (int64, error)
	ParentTransOfTxo(transactionHeight int64) (int64, error)
}

// IPrivilegedFiles give access to certain files by a privileged program (pudding-server)
type IPrivilegedFiles interface {
	TxoSatsFile() wordfile.ReadAtWordCounter
	TxiTxFile() wordfile.ReadAtWordCounter
	TxiVoutFile() wordfile.ReadAtWordCounter
	TransFirstTxiFile() wordfile.ReadAtWordCounter
	TransFirstTxoFile() wordfile.ReadAtWordCounter
}
