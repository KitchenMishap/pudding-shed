package jsonblock

import (
	"errors"
	"github.com/KitchenMishap/pudding-shed/indexedhashes"
	"strings"
)

// HashesBlockChain - Just the hashes (block, transaction, address)
// One block is in memory at any one time
type HashesBlockChain struct {
	blockFetcher       IBlockJsonFetcher
	currentJsonBytes   []byte
	currentBlock       *JsonBlockHashes
	latestBlockVisited int64
	nextBlockChannel   chan nextHashesBlockReport
}

type nextHashesBlockReport struct {
	parsedHeight     int64
	jsonBytes        []byte
	parsedJson       *JsonBlockHashes
	errorEncountered error
}

func CreateHashesBlockChain(fetcher IBlockJsonFetcher) *HashesBlockChain {
	res := HashesBlockChain{
		blockFetcher:       fetcher,
		currentJsonBytes:   nil,
		currentBlock:       nil,
		latestBlockVisited: -1,
		nextBlockChannel:   make(chan nextHashesBlockReport),
	}

	res.startParsingNextBlock(0) // We expect the next block asked for to be the genesis block (0)

	return &res
}

func (hbc *HashesBlockChain) startParsingNextBlock(nextHeight int64) {
	go func() {
		// (1) Fetch bytes
		bytes, err := hbc.blockFetcher.FetchBlockJsonBytes(nextHeight)
		if err != nil {
			// Something happened, send error report to channel
			hbc.nextBlockChannel <- nextHashesBlockReport{nextHeight, nil, nil, err}
		} else {
			// (2) Parse json
			block, err := ParseJsonBlockHashes(bytes)
			if err != nil {
				// Something happened, send error report to channel
				hbc.nextBlockChannel <- nextHashesBlockReport{nextHeight, bytes, nil, err}
			} else {
				// (3) We can safely do the following in this parallel go routine too
				// For each address in each txo, generate a hash as an id
				err = PostJsonEncodeAddressHashes2(block)
				if err != nil {
					// Something happened, send error report to channel
					hbc.nextBlockChannel <- nextHashesBlockReport{nextHeight, bytes, block, err}
				} else {
					// Convert the hash strings to binary
					err = PostJsonEncodeSha256s2(block)
					if err != nil {
						// Something happened, send error report to channel
						hbc.nextBlockChannel <- nextHashesBlockReport{nextHeight, bytes, block, err}
					} else {
						// Success! Send it via channel back to main goroutine
						hbc.nextBlockChannel <- nextHashesBlockReport{nextHeight, bytes, block, nil}
					}
				}
			}
		}
	}()
}

func (hbc *HashesBlockChain) SwitchBlock(blockHeightRequested int64) (*JsonBlockHashes, error) {
	if blockHeightRequested-hbc.latestBlockVisited > 1 {
		return nil, errors.New("blocks visited must first be visited in sequence from genesis block")
	}

	// Arrange for requested block to be represented in obc
	if hbc.currentBlock == nil || int64(hbc.currentBlock.J_height) != blockHeightRequested {
		// It's not there already

		// Wait (if necessary) and see what comes through next from the goroutine channel
		waitingForUs := <-hbc.nextBlockChannel

		// Is it not the one we want?
		if waitingForUs.parsedHeight != blockHeightRequested {
			println("found ", waitingForUs.parsedHeight, " but waiting for ", blockHeightRequested)
			// Ask for it and wait for it
			hbc.startParsingNextBlock(blockHeightRequested)
			waitingForUs = <-hbc.nextBlockChannel
			println("now found ", waitingForUs.parsedHeight, " after requesting ", blockHeightRequested)
		}

		// Is it rubbish?
		if waitingForUs.errorEncountered != nil {
			// Ask for the genesis block, just so there's always something passing through the channel for next time
			hbc.startParsingNextBlock(0)
			return nil, waitingForUs.errorEncountered
		}

		// Not rubbish
		if blockHeightRequested > hbc.latestBlockVisited {
			hbc.latestBlockVisited = blockHeightRequested
		}
		hbc.currentJsonBytes = waitingForUs.jsonBytes
		hbc.currentBlock = waitingForUs.parsedJson

		// Ask for the most likely next block, so there's always something passing through the channel for next time
		hbc.startParsingNextBlock(blockHeightRequested + 1)
	}

	return hbc.currentBlock, nil
}

func PostJsonEncodeSha256s2(block *JsonBlockHashes) error {
	// First the block hash
	err := indexedhashes.HashHexToSha256(block.J_hash, &block.hash)
	if err != nil {
		return err
	}

	// Then the transaction hashes
	for nthTrans := range block.J_tx {
		transPtr := &block.J_tx[nthTrans]
		err = indexedhashes.HashHexToSha256(transPtr.J_txid, &transPtr.txid)
		if err != nil {
			return err
		}
	}
	return nil
}

func PostJsonEncodeAddressHashes2(block *JsonBlockHashes) error {
	for nthTrans := range block.J_tx {
		transPtr := &block.J_tx[nthTrans]
		// The addresses are in the txos
		for nthTxo := range transPtr.J_vout {
			txoPtr := &transPtr.J_vout[nthTxo]
			addrPtr := &txoPtr.J_scriptPubKey
			adornTxoAddressWithPuddingHash2(addrPtr)
		}
	}
	return nil
}

func adornTxoAddressWithPuddingHash2(addrPtr *JsonScriptPubKeyEssential2) {
	address := addrPtr.J_address // Remember some types of address are case sensitive
	hex := strings.ToLower(addrPtr.J_hex)

	// THE FOLLOWING HASH IS PECULIAR TO PUDDING SHED SOFTWARE AND NOT IN GENERAL USE BY BITCOINERS
	// BUT NOTE THAT IT CAN BE GENERATED JUST BY HASHING A USUAL ASCII ADDRESS STRING
	// The hash we use to identify an address is as follows: (this hash is peculiar to pudding-shed)
	// If address is more than 10 characters, we assume it's a human-readable ASCII address, we use the hash of that.
	// Otherwise, we use the hash of hex expressed as ASCII
	hash := indexedhashes.Sha256{}
	if len(address) > 10 { // So addresses of "unknown", "none", "", etc aren't accidentally hashed
		hash = HashOfString(address)
	} else {
		hash = HashOfString(hex)
	}

	addrPtr.puddingHash = hash
}
