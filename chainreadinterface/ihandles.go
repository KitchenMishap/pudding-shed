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
	HeightSpecified() bool
	HashSpecified() bool
	IsTransHandle()
	IsInvalid() bool
}

type ITxiHandle interface {
	ParentTrans() ITransHandle
	ParentIndex() int64
	TxiHeight() int64
	ParentSpecified() bool
	TxiHeightSpecified() bool
}

type ITxoHandle interface {
	ParentTrans() ITransHandle
	ParentIndex() int64
	TxoHeight() int64
	ParentSpecified() bool
	TxoHeightSpecified() bool
}
