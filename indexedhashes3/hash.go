package indexedhashes3

import (
	"encoding/binary"
	"github.com/KitchenMishap/pudding-shed/indexedhashes"
)

type Hash indexedhashes.Sha256

func (h *Hash) toTruncatedHash(p *Params) truncatedHash {
	byteStart := p.wholeBytesInBinNum()
	return (*h)[byteStart:32]
}

func (h *Hash) toBinNum(p *Params) binNum {
	LSBs64 := binary.LittleEndian.Uint64((*h)[0:8])
	mask := uint64(0xFFFFFFFFFFFFFFFF) >> (64 - p.bitsPerBinNum)
	return binNum(int64(LSBs64 & mask))
}
