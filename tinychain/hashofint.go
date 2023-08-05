package tinychain

import (
	"crypto/sha256"
	"encoding/binary"
	"github.com/KitchenMishap/pudding-shed/indexedhashes"
)

func HashOfInt(in uint64) indexedhashes.Sha256 {
	var bytes [8]byte
	binary.LittleEndian.PutUint64(bytes[0:8], in)

	h := sha256.New()
	h.Write(bytes[0:8])

	var outBytes [32]byte
	o := h.Sum(nil)
	for i := 0; i < len(o); i++ {
		outBytes[i] = o[i]
	}
	return outBytes
}
