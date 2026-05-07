package corereaderbin

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"sync"

	"github.com/KitchenMishap/pudding-shed/indexedhashes"
)

type BlockBinary struct {
	Height       int64
	Hash         [32]byte
	HashString   string
	Binary       []byte
	Transactions []TransactionBinary
}

func (bb BlockBinary) SequenceNumber() int64 { return bb.Height }

func (bb *BlockBinary) Parse() error {
	// Skip block header
	byteIndex := 80
	// Read the transaction count
	var txCount uint64
	var bytes int
	txCount, bytes = ReadCompactSize(bb.Binary, byteIndex)
	byteIndex += bytes
	bb.Transactions = make([]TransactionBinary, txCount)

	h := sha256.New() // For txid hash

	// Now read the transactions (txCount of them)
	for tx := range txCount {
		h.Reset()                                   // This is for building the txid hash. It does NOT include various segwit bytes!
		h.Write(bb.Binary[byteIndex : byteIndex+4]) // First 4 bytes of the hashable thing
		trans := TransactionBinary{}
		// Read the 4 byte transaction version
		trans.Version = binary.LittleEndian.Uint32(bb.Binary[byteIndex : byteIndex+4])
		byteIndex += 4

		// SegWit check
		trans.IsSegWit = false
		if bb.Height >= 481824 { // segwit activation block
			if bb.Binary[byteIndex] == 0x00 && bb.Binary[byteIndex+1] == 0x01 {
				trans.IsSegWit = true
				byteIndex += 2 // (weirdly!) only increment by two if they're 0,1
			}
		}

		txiCountByteOffset := byteIndex // This is used for skipping segwit for txid hash

		// Txi count
		var txiCount uint64
		txiCount, bytes = ReadCompactSize(bb.Binary, byteIndex)
		byteIndex += bytes
		trans.Txis = make([]TxiBinary, txiCount)

		// Loop through the txis
		for i := uint64(0); i < txiCount; i++ {
			// Read TxId of vin
			copy(trans.Txis[i].TxId[:], bb.Binary[byteIndex:byteIndex+32])
			byteIndex += 32

			// Read 4 bytes of vout of vin
			trans.Txis[i].Vout = binary.LittleEndian.Uint32(bb.Binary[byteIndex : byteIndex+4])
			byteIndex += 4

			// Read ScriptSig length
			var ssLen uint64
			ssLen, bytes = ReadCompactSize(bb.Binary, byteIndex)
			byteIndex += bytes

			// Skip that number of bytes
			byteIndex += int(ssLen)

			// Skip 4 bytes of "sequence" value
			byteIndex += 4
		}

		// Txo count
		var txoCount uint64
		txoCount, bytes = ReadCompactSize(bb.Binary, byteIndex)
		byteIndex += bytes
		trans.Txos = make([]TxoBinary, txoCount)

		// Loop through the txos
		for i := uint64(0); i < txoCount; i++ {
			// Eight bytes of satoshis value
			trans.Txos[i].Value = binary.LittleEndian.Uint64(bb.Binary[byteIndex : byteIndex+8])
			byteIndex += 8

			// Read ScriptPubKey length
			var spkLen uint64
			spkLen, bytes = ReadCompactSize(bb.Binary, byteIndex)
			byteIndex += bytes

			// Read that number of bytes and convert to hex string
			hexStringScriptPubKey := hex.EncodeToString(bb.Binary[byteIndex : byteIndex+int(spkLen)])
			byteIndex += int(spkLen)

			// Take the hash of it (this is puddingHash2, unique to pudding-shed, not a proper bitcoin thing
			trans.Txos[i].PuddingHash2 = indexedhashes.HashOfString(hexStringScriptPubKey)
		}

		segwitByteOffset := byteIndex // This is used for skipping segwit for txid hash

		if trans.IsSegWit {
			// Loop through the witness data
			// There is a witness stack for each txi
			for i := uint64(0); i < txiCount; i++ {
				// Read how many items are in this input's witness stack
				var itemCount uint64
				itemCount, bytes = ReadCompactSize(bb.Binary, byteIndex)
				byteIndex += bytes

				for j := uint64(0); j < itemCount; j++ {
					// Each item in the stack has its own length
					var itemLen uint64
					itemLen, bytes = ReadCompactSize(bb.Binary, byteIndex)
					byteIndex += bytes

					// Jump over the actual witness data
					byteIndex += int(itemLen)
				}
			}
		}

		locktimeByteOffset := byteIndex // This is used for skipping segwit for txid hash
		byteIndex += 4

		// Now we finish calculating the txid, skipping those various segwit sequences if they're there
		// We've already added the version to the hash
		h.Write(bb.Binary[txiCountByteOffset:segwitByteOffset])       // Then the txis and txos
		h.Write(bb.Binary[locktimeByteOffset : locktimeByteOffset+4]) // Then the locktime
		// Hash it TWICE
		var firstHash [32]byte
		h.Sum(firstHash[:0])
		h.Reset()
		h.Write(firstHash[:])
		var finalHash [32]byte
		h.Sum(finalHash[:0])

		trans.Txid = finalHash

		// trans is now a complete(ish) description of a transaction, with txi's, txo's, and a txid
		bb.Transactions[tx] = trans
	}
	return nil
}

func GetBlocksBinary(inChan chan BlockBinary, outChan chan BlockBinary, threads int) {
	var wg sync.WaitGroup
	for range threads {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for in := range inChan {
				// Convert binary hash to string hash
				in.HashString = ToHexHash(in.Hash[0:32])

				// Request the block as binary
				blockReq := "http://127.0.0.1:8332/rest/block/" + in.HashString + ".bin"
				if in.Height%10_000 == 0 {
					fmt.Printf("Block %d: %s\n", in.Height, blockReq)
				}

				success := false
				for retry := 0; retry < 3; retry++ {
					if retry > 0 {
						fmt.Println("Retrying...")
					}
					resp, err := TheOneAndOnlyClient.Get(blockReq)
					if err == nil {
						if resp.StatusCode == 200 {
							bodyOutBlock, err := io.ReadAll(resp.Body)
							if err == nil {
								err = resp.Body.Close()
								if err == nil {
									out := BlockBinary{}
									out.Height = in.Height
									out.Hash = in.Hash
									out.HashString = in.HashString
									out.Binary = bodyOutBlock
									err = out.Parse()
									if err != nil {
										panic("Parse error")
									}
									outChan <- out
									// Success! Break out of retry loop
									if retry > 0 {
										fmt.Printf("Retry succeeded\n")
									}
									success = true
									break
								} else {
									fmt.Println(err.Error())
									fmt.Printf("Error closing response body, might retry...\n")
								}
							} else {
								fmt.Println(err.Error())
								fmt.Printf("ReadAll() returned error, might retry...\n")
							}
						} else {
							fmt.Println(resp.Status)
							fmt.Printf("Response is not 200 OK, might retry...\n")
						}
					} else {
						fmt.Println(err.Error())
						fmt.Printf("Get() returned error, might retry...\n")
					}
				}
				if success == false {
					fmt.Println("Retries exhausted getting block from Bitcoin Core, block height: ", in.Height)
					panic("Retries failed") // ToDo
				}
			}
		}()
	}
	go func() {
		wg.Wait()
		close(outChan)
	}()
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
