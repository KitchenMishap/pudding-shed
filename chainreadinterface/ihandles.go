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
