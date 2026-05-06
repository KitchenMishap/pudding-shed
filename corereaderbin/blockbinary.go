package corereaderbin

import (
	"fmt"
	"io"
	"sync"
)

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

				resp, err := TheOneAndOnlyClient.Get(blockReq)
				if err != nil {
					fmt.Println(err.Error())
					panic(err) // ToDo
				}
				if resp.StatusCode != 200 {
					panic("Response not OK: " + resp.Status)
				}
				bodyOutBlock, _ := io.ReadAll(resp.Body)
				err = resp.Body.Close()
				if err != nil {
					fmt.Println(err.Error())
					panic(err) // ToDo
				}

				// Send the binary block to the channel
				out := BlockBinary{}
				out.Height = in.Height
				out.Hash = in.Hash
				out.HashString = in.HashString
				out.Binary = bodyOutBlock

				outChan <- out
			}
		}()
	}
	go func() {
		wg.Wait()
		close(outChan)
	}()
}
