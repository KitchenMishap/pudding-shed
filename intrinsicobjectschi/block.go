package intrinsicobjectschi

import (
	//"math"

	"github.com/KitchenMishap/pudding-shed/intrinsicobjects"
)

// A Block object with an intrinsicobjects.Block, with adornments for
// some info not available intrinsically

type Block struct {
	intrinsic *intrinsicobjects.Block

	blockHeight int64
	medianTime  uint32
}

func NewBlock(intrinsic *intrinsicobjects.Block, blockHeight int64, mediantime uint32) (*Block, error) {
	result := Block{}
	result.intrinsic = intrinsic
	result.blockHeight = blockHeight
	result.medianTime = mediantime

	return &result, nil
}
