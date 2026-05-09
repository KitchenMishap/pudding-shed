package intrinsicobjects

import (
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/KitchenMishap/pudding-shed/indexedhashes"
)

// Advised by Gemini AI to avoid "Only one usage of each socket address (...) is normally permitted"
var TheOneAndOnlyTransport = &http.Transport{
	// High numbers to saturate my 36 cores
	MaxIdleConns:        100,
	MaxIdleConnsPerHost: 100,
	// Keeps connections open so they don't go into TIME_WAIT
	IdleConnTimeout: 90 * time.Second,
}

var TheOneAndOnlyClient = &http.Client{
	Transport: TheOneAndOnlyTransport,
}

// Streams 32-byte hashes of blocks to a channel, starting with the genesis block
func StreamBlockHashesFromGenesis(numBlocks int64, channel chan indexedhashes.Sha256) error {
	if numBlocks == 0 {
		close(channel)
		return nil
	}

	sentBlocks := int64(0)

	// We are interested in 1000 blocks at a time (just chosen as a round number less than the 2000 max headers we
	// can get in one go)
	// But we will get the block hashes from the "prev hash" field of each header.
	// So we (at first sight) actually need to get 1001 headers (eg 0..1000) to know the first 1000 hashes.
	// FURTHERMORE (on a second look), we will ALSO need the hash of the first block of the next 1000+ headers
	// that we'll request. This is found in the "prev hash" of index 1001.
	// So in fact, we need to get 1002 headers (eg 0..1001) at a time.
	// (this scheme avoids expensive hashing of headers and the extra code to do so)

	// The first request will be 1002 block headers as binary starting with the genesis block (indices 0..1001)
	// The long hex number within the following is the hash of the genesis block.
	headersReq := "http://127.0.0.1:8332/rest/headers/1002/000000000019d6689c085ae165831e934ff763ae46a2a6c172b3f1b60a8ce26f.bin"

	for {
		var resp *http.Response
		var err error
		resp, err = TheOneAndOnlyClient.Get(headersReq)

		if err != nil {
			fmt.Println(err.Error())
			fmt.Println("StreamBlockHashes(): Could not GET from local bitcoin REST server.")
			fmt.Println("Are you sure Bitcoin Core is running, with the correct parameters?")
			fmt.Println("Recommend: bitcoin-qt.exe -txindex -disablewallet -server -rest")
			close(channel)
			return err
		}
		expectedSize := 80 * 1002
		bodyOutHeaders := make([]byte, expectedSize)
		_, err = io.ReadFull(resp.Body, bodyOutHeaders)
		err = resp.Body.Close()
		if err != nil {
			fmt.Println(err.Error())
			close(channel)
			return err
		}

		// Each header is 80 bytes. header counts from 0
		for header := range len(bodyOutHeaders)/80 - 2 {
			// We will ignore the first header, as the second header contains the hash of the first block (as "prev hash")
			offset := (header + 1) * 80

			// The hash of the next block to request
			hashBytes := bodyOutHeaders[offset+4 : offset+36]
			hash := indexedhashes.Sha256{}
			copy(hash[:], hashBytes)
			channel <- hash

			// The block height
			blockHeight := sentBlocks

			if blockHeight%10_000 == 0 {
				fmt.Printf("Block %d initiated\n", blockHeight)
			}

			sentBlocks++
			if sentBlocks == numBlocks {
				fmt.Println("Sent all blocks")
				close(channel)
				return nil
			}
		}
		// We have output all the block hashes for which we had block hashes given in the headers.
		// (EXCEPT the prevHash in the last header, which we're going to use to get the next batch of headers)
		// But we still want more blocks.
		// If we didn't get the 1002 headers we requested, then we've exhausted the blockchain before
		// the requested number of blocks.
		if len(bodyOutHeaders) < 80*1002 {
			fmt.Println("Blockchain exhausted")
			close(channel)
			return nil // Let's not treat it as an error
		}
		// We need more headers.
		offset := 1001 * 80 // Offset to the hash of the first block to feature in the next batch of 1002 headers
		hashBytes := bodyOutHeaders[offset+4 : offset+36]
		hexHash := ToHexHash(hashBytes)

		// Request the block as binary
		headersReq = "http://127.0.0.1:8332/rest/headers/1002/" + hexHash + ".bin"
	}
}

type HeightNumberedBlock struct {
	BlockHeight int64
	BlockHash   indexedhashes.Sha256
	Block       Block
}

// For SequencerContainer
func (hnb *HeightNumberedBlock) SequenceNumber() int64 { return hnb.BlockHeight }

// Unnumbered block hashes go in, in height sequence
// Parsed blocks come out, numbered but out of sequence
func GetAndParseBlocks(inChan chan indexedhashes.Sha256, outChan chan *HeightNumberedBlock, threads int) {
	// Adorn the hashes with block heights and squirt them into a channel
	chanNumbered := make(chan *HeightNumberedBlock)
	go func() {
		blockHeight := int64(0)
		for blockHash := range inChan {
			numbered := HeightNumberedBlock{}
			numbered.BlockHeight = blockHeight
			numbered.BlockHash = blockHash
			numbered.Block = Block{}
			chanNumbered <- &numbered

			blockHeight++
		}
		close(chanNumbered)
	}()

	// A thread pool to fill in the Blocks of these objects (get from Core and parse)
	var wg sync.WaitGroup
	for range threads {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for numbered := range chanNumbered {
				// Convert binary hash to string hash
				hashString := ToHexHash(numbered.BlockHash[0:32])

				// Request the block as binary
				blockReq := "http://127.0.0.1:8332/rest/block/" + hashString + ".bin"
				if numbered.BlockHeight%10_000 == 0 {
					fmt.Printf("Block %d: %s\n", numbered.BlockHeight, blockReq)
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
									ParseBinaryBlock(bodyOutBlock, &numbered.Block)
									outChan <- numbered
									// Success! Break out of retry loop
									if retry > 0 {
										fmt.Printf("Retry succeeded\n")
									}
									success = true
									break
								} else {
									fmt.Println(err.Error())
									fmt.Printf("Error closing response body\n")
								}
							} else {
								fmt.Println(err.Error())
								fmt.Printf("ReadAll() returned error\n")
							}
						} else {
							fmt.Println(resp.Status)
							fmt.Printf("Response is not 200 OK\n")
						}
					} else {
						fmt.Println(err.Error())
						fmt.Printf("Get() returned error\n")
					}
				}
				if success == false {
					fmt.Println("Retries exhausted getting block from Bitcoin Core, block height: ", numbered.BlockHeight)
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

// ToHexHash converts from a 32-byte slice from a binary header
// to a big-endian hex string for REST/RPC requests.
func ToHexHash(rawHash []byte) string {
	reversed := make([]byte, 32)
	for i := 0; i < 32; i++ {
		reversed[i] = rawHash[31-i]
	}
	return hex.EncodeToString(reversed)
}
