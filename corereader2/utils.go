package corereader2

import "encoding/hex"

// ToHexHash converts from a 32-byte slice from a binary header
// to a big-endian hex string for REST/RPC requests.
func ToHexHash(rawHash []byte) string {
	reversed := make([]byte, 32)
	for i := 0; i < 32; i++ {
		reversed[i] = rawHash[31-i]
	}
	return hex.EncodeToString(reversed)
}
