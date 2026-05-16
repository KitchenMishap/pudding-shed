package intrinsicobjectschi

import (
	"github.com/KitchenMishap/pudding-shed/chainhandleinterface"
	"github.com/KitchenMishap/pudding-shed/intrinsicobjects"
)

type OneBlockHolder struct {
	InChan chan *intrinsicobjects.Block
}

// intrinsicobjectschi.OneBlockHolder implements chainhandleinterface.IBlockChain
var _ chainhandleinterface.IBlockChain = (*OneBlockHolder)(nil) // Check that implements

func CreateOneBlockHolder() *OneBlockHolder {
	panic("Not Implemented")
}

func (obh *OneBlockHolder) GenesisBlock() (chainhandleinterface.BlockHandle, error) {
	panic("Not Implemented")
}

func (obh *OneBlockHolder) NextBlock(handle chainhandleinterface.BlockHandle) (chainhandleinterface.BlockHandle, error) {
	panic("Not Implemented")
}

func (obh *OneBlockHolder) IsBlockHandleInvalid(handle chainhandleinterface.BlockHandle) bool {
	panic("Not Implemented")
}

func (obh *OneBlockHolder) BlockInterface(handle chainhandleinterface.BlockHandle) (chainhandleinterface.IBlock, error) {
	panic("Not Implemented")
}
