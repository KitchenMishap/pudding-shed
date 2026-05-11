package intrinsicobjects

import (
	"crypto/sha256"
	"encoding/binary"
	"math/big"
)

func ParseBinaryBlock(bin []byte, targetBlock *Block) {
	//-------
	// Header
	//-------

	// Hash is found by sha256 (twice) of block header (first 80 bytes of block)
	h := sha256.New()
	h.Write(bin[0:80])
	// Hash it TWICE
	var firstHash [32]byte
	h.Sum(firstHash[:0])
	h.Reset()
	h.Write(firstHash[:])
	var blockHash [32]byte
	h.Sum(blockHash[:0])
	targetBlock.BlockHash = blockHash

	// First 4 bytes in the header is the  block version
	targetBlock.Version = binary.LittleEndian.Uint32(bin[0:4])
	// 32 bytes prev block hash
	copy(targetBlock.PrevHash[0:32], bin[4:36])
	// 32 bytes Merkle Root
	copy(targetBlock.MerkleRoot[0:32], bin[36:68])
	// 4 bytes timestamp
	targetBlock.Time = binary.LittleEndian.Uint32(bin[68:72])
	// 4 bytes that are used regarding difficulty
	targetBlock.NBits = binary.LittleEndian.Uint32(bin[72:76])
	targetBlock.Difficulty = CalculateDifficulty(targetBlock.NBits)
	// 4 bytes counter that was used to solve the proof of work
	targetBlock.Nonce = binary.LittleEndian.Uint32(bin[76:80])

	//---------------------------
	// Rest of block after header
	//---------------------------
	// Skip block header
	byteIndex := 80
	targetBlock.StrippedSize = 80 // plus more later...
	targetBlock.Weight = 80 * 4   // plus more later...
	// Read the transaction count
	var txCount uint64
	var bytes int
	txCount, bytes = ReadCompactSize(bin, byteIndex)
	byteIndex += bytes
	targetBlock.StrippedSize += bytes
	targetBlock.Weight += 4 * bytes

	targetBlock.Transactions = make([]Transaction, txCount)

	// Now read the transactions (txCount of them)
	for tx := range txCount {
		bytes = ParseBinaryTransaction(bin, byteIndex, h, &(targetBlock.Transactions[tx]))
		byteIndex += bytes
		targetBlock.Weight += targetBlock.Transactions[tx].Weight
		targetBlock.StrippedSize += targetBlock.Transactions[tx].StrippedSize
	}

	targetBlock.Size = len(bin)
}

func CalculateDifficulty(nBits uint32) float64 {
	exponent := uint(nBits >> 24)
	mantissa := uint64(nBits & 0x00ffffff)

	target := big.NewInt(0).SetUint64(mantissa)
	if exponent > 3 {
		target.Lsh(target, 8*(exponent-3))
	} else if exponent < 3 {
		target.Rsh(target, 8*(3-exponent))
	}

	// Genesis Target (Difficulty 1.0)
	genTarget, _ := big.NewInt(0).SetString("00000000ffff0000000000000000000000000000000000000000000000000000", 16)

	fGen, _ := big.NewFloat(0).SetInt(genTarget).Float64()
	fCur, _ := big.NewFloat(0).SetInt(target).Float64()

	// Mimic the JSON float-to-int64 truncation
	return fGen / fCur
}

// ReadCompactSize returns the value and the number of bytes consumed.
func ReadCompactSize(data []byte, offset int) (uint64, int) {
	first := data[offset]
	switch {
	case first < 253:
		return uint64(first), 1
	case first == 253:
		return uint64(binary.LittleEndian.Uint16(data[offset+1:])), 3
	case first == 254:
		return uint64(binary.LittleEndian.Uint32(data[offset+1:])), 5
	default: // 255
		return binary.LittleEndian.Uint64(data[offset+1:]), 9
	}
}
