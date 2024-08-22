package jsonblock

import (
	"crypto/sha256"
	"github.com/KitchenMishap/pudding-shed/indexedhashes"
)

func HashOfString(in string) indexedhashes.Sha256 {
	h := sha256.New()
	h.Write([]byte(in))

	var outBytes [32]byte
	o := h.Sum(nil)
	for i := 0; i < len(o); i++ {
		outBytes[i] = o[i]
	}
	return outBytes
}
