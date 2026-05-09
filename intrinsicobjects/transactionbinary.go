package intrinsicobjects

import (
	"encoding/binary"
	"fmt"
	"hash"

	"github.com/KitchenMishap/pudding-shed/indexedhashes"
)

// Returns number of bytes read
func ParseBinaryTransaction(bin []byte, byteIndex int, h hash.Hash, targetTrans *Transaction) int {
	startByteIndex := byteIndex

	h.Reset()                             // This is for building the txid hash. It does NOT include various segwit bytes!
	h.Write(bin[byteIndex : byteIndex+4]) // First 4 bytes of the hashable thing

	// Read the 4 byte transaction version
	targetTrans.Version = binary.LittleEndian.Uint32(bin[byteIndex : byteIndex+4])
	byteIndex += 4

	// SegWit check
	targetTrans.IsSegWit = false
	//if bb.Height >= 481824 { // We DON'T know the block height, so we can't check for the segwit activation block
	if bin[byteIndex] == 0x00 && bin[byteIndex+1] != 0x00 {
		targetTrans.IsSegWit = true
		if bin[byteIndex+1] != 0x01 { // 0x01 is the only allowable flag at time of writing (2026)
			fmt.Errorf("unknown SegWit flag: 0x%02x", bin[byteIndex+1])
		}
		byteIndex += 2 // (weirdly!) only increment by two if they're 0,1
	}

	txiCountByteOffset := byteIndex // This is used for skipping segwit for txid hash

	// Txi count
	var txiCount uint64
	var bytes int
	txiCount, bytes = ReadCompactSize(bin, byteIndex)
	byteIndex += bytes
	targetTrans.Txis = make([]Txi, txiCount)

	// Loop through the txis
	for i := uint64(0); i < txiCount; i++ {
		// Read TxId of vin
		copy(targetTrans.Txis[i].TxId[:], bin[byteIndex:byteIndex+32])
		byteIndex += 32

		// Read 4 bytes of vout of vin
		(*targetTrans).Txis[i].VOut = int64(binary.LittleEndian.Uint32(bin[byteIndex : byteIndex+4]))
		byteIndex += 4

		// Read ScriptSig length
		var ssLen uint64
		ssLen, bytes = ReadCompactSize(bin, byteIndex)
		byteIndex += bytes

		// Skip that number of bytes
		byteIndex += int(ssLen)

		// Skip 4 bytes of "sequence" value
		byteIndex += 4
	}

	// Txo count
	var txoCount uint64
	txoCount, bytes = ReadCompactSize(bin, byteIndex)
	byteIndex += bytes
	targetTrans.Txos = make([]Txo, txoCount)

	// Loop through the txos
	for i := uint64(0); i < txoCount; i++ {
		// Eight bytes of satoshis value
		targetTrans.Txos[i].Value = int64(binary.LittleEndian.Uint64(bin[byteIndex : byteIndex+8]))
		byteIndex += 8

		// Read ScriptPubKey length
		var spkLen uint64
		spkLen, bytes = ReadCompactSize(bin, byteIndex)
		byteIndex += bytes

		// Read that number of bytes
		scriptPubKey := bin[byteIndex : byteIndex+int(spkLen)]
		byteIndex += int(spkLen)

		// Take the hash of it (this is puddingHash3, unique to pudding-shed, not a proper bitcoin thing)
		targetTrans.Txos[i].AddressPuddingHash3 = indexedhashes.HashOfBytes(scriptPubKey)
	}

	segwitByteOffset := byteIndex // This is used for skipping segwit for txid hash

	if targetTrans.IsSegWit {
		// Loop through the witness data
		// There is a witness stack for each txi
		for i := uint64(0); i < txiCount; i++ {
			// Read how many items are in this input's witness stack
			var itemCount uint64
			itemCount, bytes = ReadCompactSize(bin, byteIndex)
			byteIndex += bytes

			for j := uint64(0); j < itemCount; j++ {
				// Each item in the stack has its own length
				var itemLen uint64
				itemLen, bytes = ReadCompactSize(bin, byteIndex)
				byteIndex += bytes

				// Jump over the actual witness data
				byteIndex += int(itemLen)
			}
		}
	}

	locktimeByteOffset := byteIndex // This is used for skipping segwit for txid hash
	// Skip 4 bytes of locktime
	byteIndex += 4

	// Now we finish calculating the txid, skipping those various segwit sequences if they're there
	// We've already added the version to the hash
	h.Write(bin[txiCountByteOffset:segwitByteOffset])       // Then the txis and txos
	h.Write(bin[locktimeByteOffset : locktimeByteOffset+4]) // Then the locktime
	// Hash it TWICE
	var firstHash [32]byte
	h.Sum(firstHash[:0])
	h.Reset()
	h.Write(firstHash[:])
	var finalHash [32]byte
	h.Sum(finalHash[:0])

	targetTrans.TxId = finalHash

	// targetTrans is now a complete(ish) description of a transaction, with txi's, txo's, and a txid
	return byteIndex - startByteIndex
}
