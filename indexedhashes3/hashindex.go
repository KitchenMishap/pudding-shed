package indexedhashes3

import (
	"encoding/binary"
)

type hashIndex int64

func (hi *hashIndex) toBytes(p *Params) []byte {
	buf := [8]byte{}
	binary.LittleEndian.PutUint64(buf[:], uint64(*hi))
	return buf[0:p.bytesPerHashIndex]
}

func (hi *hashIndex) fromBytes(buf []byte, p *Params) {
	b := [8]byte{}
	copy(b[:len(buf)], buf[:])
	*hi = hashIndex(binary.LittleEndian.Uint64(buf))
}
