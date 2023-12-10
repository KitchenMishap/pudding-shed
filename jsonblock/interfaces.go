package jsonblock

import "github.com/KitchenMishap/pudding-shed/indexedhashes"

// IBlockJsonFetcher describes an object that serves each block as json bytes
type IBlockJsonFetcher interface {
	CountBlocks() (int64, error)
	FetchBlockJsonBytes(height int64) ([]byte, error)
}

// ITransLocatorByHash describes an object that obtains a transaction's Indices Path (block/trans) based on hash
type ITransLocatorByHash interface {
	GetTransIndicesPathByHash(sha256 indexedhashes.Sha256) (ITransIndicesPath, error)
}

// ITransLocatorStore describes an object that stores and retrieves transactions' Indices Path (block/trans) based on hash
type ITransLocatorStore interface {
	ITransLocatorByHash
	StoreIndicesPathForHash(sha256 indexedhashes.Sha256, blockHeight int64, nthTransInBlock int64) error
}

type ITransIndicesPath interface {
	BlockHeight() int64
	NthTransInBlock() int64
}
