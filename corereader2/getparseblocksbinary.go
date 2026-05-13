package corereader2

import (
	"fmt"
	"io"
	"sync"

	"github.com/KitchenMishap/pudding-shed/indexedhashes"
	"github.com/KitchenMishap/pudding-shed/intrinsicobjects"
)

type HeightNumberedBlock struct {
	BlockHeight int64
	BlockHash   indexedhashes.Sha256
	Block       intrinsicobjects.Block
}

// For SequencerContainer
func (hnb *HeightNumberedBlock) SequenceNumber() int64 { return hnb.BlockHeight }

// GetAndParseBlocksBinary takes unnumbered block hashes at its input channel, in height sequence.
// Parsed blocks come out, numbered but out of sequence
func GetAndParseBlocks(inChan chan indexedhashes.Sha256, useJson bool,
	outChan chan *HeightNumberedBlock, threads int) {
	// Adorn the hashes with block heights and squirt them into a channel
	chanNumbered := make(chan *HeightNumberedBlock)
	go func() {
		blockHeight := int64(0)
		for blockHash := range inChan {
			numbered := HeightNumberedBlock{}
			numbered.BlockHeight = blockHeight
			numbered.BlockHash = blockHash
			numbered.Block = intrinsicobjects.Block{}
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

				var blockReq string
				if useJson {
					// Request the block as json
					blockReq = "http://127.0.0.1:8332/rest/block/" + hashString + ".json"
				} else {
					// Request the block as binary
					blockReq = "http://127.0.0.1:8332/rest/block/" + hashString + ".bin"
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
									if useJson {
										err = intrinsicobjects.ParseJsonBlock(bodyOutBlock, &numbered.Block)
										if err != nil {
											success = false
											fmt.Println(err.Error())
											fmt.Println("Error parsing JSON")
										} else {
											success = true
										}
									} else {
										intrinsicobjects.ParseBinaryBlock(bodyOutBlock, &numbered.Block)
										success = true
									}
									if success {
										outChan <- numbered
										// Break out of retry loop
										if retry > 0 {
											fmt.Printf("Retry succeeded\n")
										}
										break
									}
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
