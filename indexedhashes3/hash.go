package indexedhashes3

import (
	"encoding/binary"
	"github.com/KitchenMishap/pudding-shed/indexedhashes"
)

type Hash indexedhashes.Sha256

func (h *Hash) toTruncatedHash() truncatedHash {
	result := [24]byte{}
	copy(result[:], (*h)[8:32])
	return result
}

func (h *Hash) toAbbreviatedHash() abbreviatedHash {
	return abbreviatedHash(binary.LittleEndian.Uint64((*h)[0:8]))
}

func NewHashFromTruncatedHashAbbreviatedHash(t *truncatedHash, a abbreviatedHash) *Hash {
	h := Hash{}
	binary.LittleEndian.PutUint64(h[0:8], uint64(a))
	copy((h)[8:32], (*t)[0:24])
	return &h
}
