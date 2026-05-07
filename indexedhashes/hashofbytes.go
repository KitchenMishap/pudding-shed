package indexedhashes

import "crypto/sha256"

func HashOfBytes(in []byte) Sha256 {
	h := sha256.New()
	h.Write(in)

	var outBytes [32]byte
	o := h.Sum(nil)
	for i := 0; i < len(o); i++ {
		outBytes[i] = o[i]
	}
	return outBytes
}
