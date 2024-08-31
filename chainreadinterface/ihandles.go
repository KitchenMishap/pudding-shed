package chainreadinterface

import "github.com/KitchenMishap/pudding-shed/indexedhashes"

type IBlockHandle interface {
	Height() int64
	Hash() (indexedhashes.Sha256, error)
	HeightSpecified() bool
	HashSpecified() bool
	IsBlockHandle()
	IsInvalid() bool
}

type ITransHandle interface {
	Height() int64
	Hash() (indexedhashes.Sha256, error)
	IndicesPath() (int64, int64) // Block, trans
	HeightSpecified() bool
	HashSpecified() bool
	IndicesPathSpecified() bool
	IsTransHandle()
	IsInvalid() bool
}

type ITxiHandle interface {
	ParentTrans() ITransHandle
	ParentIndex() int64
	TxiHeight() int64
	IndicesPath() (int64, int64, int64) // Block, Trans, Vin
	ParentSpecified() bool
	TxiHeightSpecified() bool
	IndicesPathSpecified() bool
}

type ITxoHandle interface {
	ParentTrans() ITransHandle
	ParentIndex() int64
	TxoHeight() int64
	IndicesPath() (int64, int64, int64) // Block, Trans, Vout
	ParentSpecified() bool
	TxoHeightSpecified() bool
	IndicesPathSpecified() bool
}

type IAddressHandle interface {
	Height() int64
	Hash() indexedhashes.Sha256
	HeightSpecified() bool
	HashSpecified() bool
}

// IHandleCreator may supply handles, but this doesn't imply existence of the underlying object
type IHandleCreator interface {
	BlockHandleByHeight(blockHeight int64) (IBlockHandle, error)
	TransactionHandleByHeight(transactionHeight int64) (ITransHandle, error)
	TxiHandleByHeight(txiHeight int64) (ITxiHandle, error)
	TxoHandleByHeight(txoHeight int64) (ITxoHandle, error)
}
