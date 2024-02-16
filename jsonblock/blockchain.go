package jsonblock

import (
	"errors"
	"github.com/KitchenMishap/pudding-shed/chainreadinterface"
	"github.com/KitchenMishap/pudding-shed/indexedhashes"
	"github.com/KitchenMishap/pudding-shed/transactionindexing"
)

// OneBlockChain - We call it a one block chain, as one block is in memory at any point in time
type OneBlockChain struct {
	blockFetcher             IBlockJsonFetcher
	transactionIndexer       transactionindexing.ITransactionIndexer
	currentJsonBytes         []byte
	currentBlock             *JsonBlockEssential
	latestBlockVisited       int64
	latestTransactionVisited int64
	nextBlockChannel         chan nextBlockReport
}

type nextBlockReport struct {
	parsedHeight     int64
	jsonBytes        []byte
	parsedJson       *JsonBlockEssential
	errorEncountered error
}

func CreateOneBlockChain(
	fetcher IBlockJsonFetcher,
	indexer transactionindexing.ITransactionIndexer,
) *OneBlockChain {
	res := OneBlockChain{
		blockFetcher:             fetcher,
		transactionIndexer:       indexer,
		currentJsonBytes:         nil,
		currentBlock:             nil,
		latestBlockVisited:       -1,
		latestTransactionVisited: -1,
		nextBlockChannel:         make(chan nextBlockReport),
	}

	res.startParsingNextBlock(0) // We expect the next block asked for to be the genesis block (0)

	return &res
}

func (obc *OneBlockChain) startParsingNextBlock(nextHeight int64) {
	go func() {
		// (1) Fetch bytes
		bytes, err := obc.blockFetcher.FetchBlockJsonBytes(nextHeight)
		if err != nil {
			// Something happened, send error report to channel
			obc.nextBlockChannel <- nextBlockReport{nextHeight, nil, nil, err}
		} else {
			// (2) Parse json
			block, err := parseJsonBlock(bytes)
			if err != nil {
				// Something happened, send error report to channel
				obc.nextBlockChannel <- nextBlockReport{nextHeight, bytes, nil, err}
			} else {
				// (3) We can safely do the following in this parallel go routine too
				postJsonRemoveCoinbaseTxis(block)
				// Convert bitcoin floats to satoshi ints
				postJsonCalculateSatoshis(block)
				// Convert the hash strings to binary
				err = postJsonEncodeSha256s(block)
				if err != nil {
					// Something happened, send error report to channel
					obc.nextBlockChannel <- nextBlockReport{nextHeight, bytes, block, err}
				} else {
					// Success! Send it via channel back to main goroutine
					// (Note there's some further processing needed, but we can only do that in the main goroutine)
					obc.nextBlockChannel <- nextBlockReport{nextHeight, bytes, block, nil}
				}
			}
		}
	}()
}

func (obc *OneBlockChain) switchBlock(blockHeightRequested int64) (*JsonBlockEssential, error) {
	if blockHeightRequested-obc.latestBlockVisited > 1 {
		return nil, errors.New("blocks visited must first be visited in sequence from genesis block")
	}

	// Arrange for requested block to be represented in obc
	if obc.currentBlock == nil || int64(obc.currentBlock.J_height) != blockHeightRequested {
		// It's not there already

		// Wait (if necessary) and see what comes through next from the goroutine channel
		waitingForUs := <-obc.nextBlockChannel

		// Is it not the one we want?
		if waitingForUs.parsedHeight != blockHeightRequested {
			println("found ", waitingForUs.parsedHeight, " but waiting for ", blockHeightRequested)
			// Ask for it and wait for it
			obc.startParsingNextBlock(blockHeightRequested)
			waitingForUs = <-obc.nextBlockChannel
			println("now found ", waitingForUs.parsedHeight, " after requesting ", blockHeightRequested)
		}

		// Is it rubbish?
		if waitingForUs.errorEncountered != nil {
			// Ask for the genesis block, just so there's always something passing through the channel for next time
			obc.startParsingNextBlock(0)
			return nil, waitingForUs.errorEncountered
		}

		// Not rubbish
		if blockHeightRequested > obc.latestBlockVisited {
			obc.latestBlockVisited = blockHeightRequested
		}
		obc.currentJsonBytes = waitingForUs.jsonBytes
		obc.currentBlock = waitingForUs.parsedJson

		// Ask for the most likely next block, so there's always something passing through the channel for next time
		obc.startParsingNextBlock(blockHeightRequested + 1)

		// Gather the transaction hashes, as we'll need to look them up, from now and forevermore
		err := obc.postJsonGatherTransHashes(obc.currentBlock)
		if err != nil {
			return nil, err
		}

		// Store in each array element (tx, txi, txo), the element's index into that array
		postJsonArrayIndicesIntoElements(obc.currentBlock)

		err = obc.postJsonUpdateTransReferences(obc.currentBlock)
		if err != nil {
			return nil, err
		}

		// Gather a map of the non-essential ints
		postJsonGatherNonEssentialInts(obc.currentBlock)
	}

	return obc.currentBlock, nil
}

func postJsonRemoveCoinbaseTxis(block *JsonBlockEssential) {
	// Coinbase Txi's are a transaction's entry in vin[] which represent the concept of coinbase in Bitcoin Core's JSON
	// We detect them by way of absence of hash string

	// Go through transactions
	for nthTrans := range block.J_tx {
		transPtr := &block.J_tx[nthTrans]
		// If there's a coinbase Txi, it will be the only one in vin
		if len(transPtr.J_vin) == 1 {
			if transPtr.J_vin[0].J_txid == "" { // This is how we identify a "coinbase txi"
				transPtr.J_vin = transPtr.J_vin[:0] // Slice to an empty array. Better than nil as encodes to empty array in JSON
			}
		}
	}
}

func postJsonEncodeSha256s(block *JsonBlockEssential) error {
	// First the block hash
	err := indexedhashes.HashHexToSha256(block.J_hash, &block.hash)
	if err != nil {
		return err
	}

	// Then the trans hashes
	for nthTrans := range block.J_tx {
		transPtr := &block.J_tx[nthTrans]
		err = indexedhashes.HashHexToSha256(transPtr.J_txid, &transPtr.txid)
		if err != nil {
			return err
		}
		// Then the references to trans hashes in the txis
		for nthTxi := range transPtr.J_vin {
			txiPtr := &transPtr.J_vin[nthTxi]
			err = indexedhashes.HashHexToSha256(txiPtr.J_txid, &txiPtr.txid)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func postJsonCalculateSatoshis(block *JsonBlockEssential) {
	const satoshisPerBitcoin = float64(100_000_000)
	for nthTrans := range block.J_tx {
		transPtr := &block.J_tx[nthTrans]
		for nthTxo := range transPtr.J_vout {
			txoPtr := &transPtr.J_vout[nthTxo]
			txoPtr.satoshis = int64(satoshisPerBitcoin * txoPtr.J_value)
		}
	}
}

func (obc *OneBlockChain) postJsonGatherTransHashes(block *JsonBlockEssential) error {
	blockHeight := int64(block.J_height)
	firstTransHeight := obc.latestTransactionVisited + 1
	err := obc.transactionIndexer.StoreBlockHeightToFirstTrans(blockHeight, firstTransHeight)
	if err != nil {
		return err
	}
	transHeight := obc.latestTransactionVisited
	for nthTrans := range block.J_tx {
		transHeight++
		err = obc.transactionIndexer.StoreTransHeightToParentBlock(transHeight, blockHeight)
		if err != nil {
			return err
		}
		transPtr := &block.J_tx[nthTrans]
		err = obc.transactionIndexer.StoreTransHashToHeight(&transPtr.txid, transHeight)
		if err != nil {
			return err
		}
	}
	obc.latestTransactionVisited = transHeight
	return nil
}

func postJsonArrayIndicesIntoElements(block *JsonBlockEssential) {
	// We copy Height into each element of the Tx array,
	// And place each element's index into each element
	for nth := range block.J_tx {
		txPtr := &block.J_tx[nth]
		txPtr.handle.blockHeight = int64(block.J_height)
		txPtr.handle.nthInBlock = int64(nth)
		// Then we go deeper, into the Tx[nth]'s vin[] and vout[]
		for vinIndex := range txPtr.J_vin {
			txPtr.J_vin[vinIndex].parentTrans.blockHeight = int64(block.J_height)
			txPtr.J_vin[vinIndex].parentTrans.nthInBlock = int64(nth)
			txPtr.J_vin[vinIndex].parentVIndex = int64(vinIndex)
		}
		for voutIndex := range txPtr.J_vout {
			txPtr.J_vout[voutIndex].parentTrans.blockHeight = int64(block.J_height)
			txPtr.J_vout[voutIndex].parentTrans.nthInBlock = int64(nth)
			txPtr.J_vout[voutIndex].parentVIndex = int64(voutIndex)
		}
	}
}

// postJsonUpdateTransReferences() uses the accrued transaction hash map to
// locate the txos (by way of block/nthTrans/vindex path indices) corresponding to each txi.
func (obc *OneBlockChain) postJsonUpdateTransReferences(block *JsonBlockEssential) error {
	// Use the map to locate the txos (by indices path) referenced by trans hashes in the txis in this block
	for nthTrans := range block.J_tx {
		transPtr := &block.J_tx[nthTrans]
		for nthTxi := range transPtr.J_vin {
			txiPtr := &transPtr.J_vin[nthTxi]

			// Look up the path indices by source transaction hash
			sourceTransHeight, err := obc.transactionIndexer.RetrieveTransHashToHeight(&txiPtr.txid)
			if err != nil {
				return err
			}
			sourceBlockHeight, err := obc.transactionIndexer.RetrieveTransHeightToParentBlock(sourceTransHeight)
			if err != nil {
				return err
			}
			sourceBlockFirstTrans, err := obc.transactionIndexer.RetrieveBlockHeightToFirstTrans(sourceBlockHeight)
			if err != nil {
				return err
			}
			sourceNthTrans := sourceTransHeight - sourceBlockFirstTrans

			// Store the path indices of the source txo, into the txi that referenced it
			txiPtr.sourceTrans.blockHeight = sourceBlockHeight
			txiPtr.sourceTrans.nthInBlock = sourceNthTrans
			// txiPtr.J_vout is already there of course
		}
	}
	return nil
}

func postJsonGatherNonEssentialInts(block *JsonBlockEssential) {
	block.nonEssentialInts = make(map[string]int64)
	block.nonEssentialInts["version"] = block.J_version
	block.nonEssentialInts["time"] = block.J_time
	block.nonEssentialInts["mediantime"] = block.J_mediantime
	block.nonEssentialInts["nonce"] = block.J_nonce
	block.nonEssentialInts["difficulty"] = int64(block.J_difficulty)
	block.nonEssentialInts["strippedsize"] = block.J_strippedsize
	block.nonEssentialInts["size"] = block.J_size
	block.nonEssentialInts["weight"] = block.J_weight
	for nthTrans := range block.J_tx {
		transPtr := &block.J_tx[nthTrans]
		transPtr.nonEssentialInts = make(map[string]int64)
		transPtr.nonEssentialInts["version"] = transPtr.J_version
		transPtr.nonEssentialInts["size"] = transPtr.J_size
		transPtr.nonEssentialInts["vsize"] = transPtr.J_vsize
		transPtr.nonEssentialInts["weight"] = transPtr.J_weight
		transPtr.nonEssentialInts["locktime"] = transPtr.J_locktime
	}
}

// Functions in jsonblock.OneBlockChain to implement chainreadinterface.IBlockTree as part of chainreadinterface.IBlockChain

func (obc *OneBlockChain) InvalidBlock() chainreadinterface.IBlockHandle {
	bh := BlockHandle{}
	bh.height = -1
	return &bh
}

func (obc *OneBlockChain) InvalidTrans() chainreadinterface.ITransHandle {
	th := TransHandle{}
	th.nthInBlock = -1
	return &th
}

func (obc *OneBlockChain) GenesisBlock() chainreadinterface.IBlockHandle {
	block, err := obc.switchBlock(0)
	if err != nil {
		panic("couldn't switch to genesis block")
	}
	return block
}

func (obc *OneBlockChain) ParentBlock(block chainreadinterface.IBlockHandle) chainreadinterface.IBlockHandle {
	if block.IsInvalid() {
		return obc.InvalidBlock()
	}
	if block.HeightSpecified() {
		if block.Height() == 0 {
			return obc.InvalidBlock()
		}
		prevBlock, err := obc.switchBlock(block.Height() - 1)
		if err != nil {
			panic(err)
		}
		return prevBlock
	} else {
		panic("this function expects block to have height specified")
	}
}

func (obc *OneBlockChain) GenesisTransaction() (chainreadinterface.ITransHandle, error) {
	genesisBlock, err := obc.switchBlock(0)
	if err != nil {
		return nil, err
	}
	trans, err := genesisBlock.NthTransaction(0)
	return trans, err
}

func (obc *OneBlockChain) PreviousTransaction(trans chainreadinterface.ITransHandle) chainreadinterface.ITransHandle {
	// Todo [  ] Replace panics with error returns
	if !trans.IndicesPathSpecified() {
		panic("this function assumes indices path specified for trans")
	}
	blockHeight, nthInBlock := trans.IndicesPath()
	if nthInBlock == 0 {
		if blockHeight == 0 {
			return obc.InvalidTrans()
		}
		// Load previous block
		block, err := obc.switchBlock(blockHeight - 1)
		if err != nil {
			panic(err)
		}
		transCount, err := block.TransactionCount()
		if err != nil {
			panic(err)
		}
		trans, err := block.NthTransaction(transCount - 1)
		if err != nil {
			panic(err)
		}
		return trans
	}
	nthInBlock -= 1
	// Load block in case not loaded
	block, err := obc.switchBlock(blockHeight)
	if err != nil {
		panic(err)
	}
	nthTrans, err := block.NthTransaction(nthInBlock)
	if err != nil {
		panic(err)
	}
	return nthTrans
}

func (obc *OneBlockChain) IsBlockTree() bool { return false } // This is a BlockChain not a full tree

func (obc *OneBlockChain) BlockInterface(handle chainreadinterface.IBlockHandle) (chainreadinterface.IBlock, error) {
	if !handle.HeightSpecified() {
		panic("this function assumes block handle specifies height")
	}
	res, err := obc.switchBlock(handle.Height())
	return res, err
}

func (obc *OneBlockChain) TransInterface(handle chainreadinterface.ITransHandle) (chainreadinterface.ITransaction, error) {
	if !handle.IndicesPathSpecified() {
		return nil, errors.New("this function assumes indices path is specified in ITransHandle")
	}
	blockHeight, nthInBlock := handle.IndicesPath()
	block, err := obc.switchBlock(blockHeight)
	if err != nil {
		return nil, err
	}

	trans := &block.J_tx[nthInBlock]
	return trans, nil
}

func (obc *OneBlockChain) TxiInterface(handle chainreadinterface.ITxiHandle) (chainreadinterface.ITxi, error) {
	if !handle.IndicesPathSpecified() {
		return nil, errors.New("this function assumes indices path is specified in ITxiHandle")
	}
	blockHeight, nthInBlock, vIndex := handle.IndicesPath()
	block, err := obc.switchBlock(blockHeight)
	if err != nil {
		return nil, err
	}

	txi := &block.J_tx[nthInBlock].J_vin[vIndex]
	return txi, nil
}

func (obc *OneBlockChain) TxoInterface(handle chainreadinterface.ITxoHandle) (chainreadinterface.ITxo, error) {
	if !handle.IndicesPathSpecified() {
		return nil, errors.New("this function assumes indices path is specified in ITxoHandle")
	}
	blockHeight, nthInBlock, vIndex := handle.IndicesPath()
	block, err := obc.switchBlock(blockHeight)
	if err != nil {
		return nil, err
	}

	txo := &block.J_tx[nthInBlock].J_vout[vIndex]
	return txo, nil
}

// Functions in jsonblock.OneBlockChain to implement chainreadinterface.IBlockChain

func (obc *OneBlockChain) LatestBlock() (chainreadinterface.IBlockHandle, error) {
	count, err := obc.blockFetcher.CountBlocks()
	if err != nil {
		return nil, err
	}
	block, err := obc.switchBlock(count - 1)
	if err != nil {
		return nil, err
	}
	return block, nil
}

func (obc *OneBlockChain) NextBlock(bh chainreadinterface.IBlockHandle) (chainreadinterface.IBlockHandle, error) {
	blockHeight := bh.Height()
	count, err := obc.blockFetcher.CountBlocks()
	if err != nil {
		return nil, err
	}
	if blockHeight+1 >= count {
		return obc.InvalidBlock(), nil
	}

	block, err := obc.switchBlock(blockHeight + 1)
	if err != nil {
		return nil, err
	}
	return block, nil
}

func (obc *OneBlockChain) LatestTransaction() (chainreadinterface.ITransHandle, error) {
	count, err := obc.blockFetcher.CountBlocks()
	if err != nil {
		return nil, err
	}
	block, err := obc.switchBlock(count - 1)
	if err != nil {
		return nil, err
	}

	transCount, err := block.TransactionCount()
	if err != nil {
		return nil, err
	}
	handle := TransHandle{
		blockHeight: int64(block.J_height),
		nthInBlock:  transCount - 1,
	}
	return &handle, nil
}

func (obc *OneBlockChain) NextTransaction(transHandle chainreadinterface.ITransHandle) (chainreadinterface.ITransHandle, error) {
	if !transHandle.IndicesPathSpecified() {
		return nil, errors.New("this function assumes ITransHandle specifies indices path")
	}
	blockHeight, nthInBlock := transHandle.IndicesPath()
	block, err := obc.switchBlock(blockHeight)
	if err != nil {
		return nil, err
	}

	blockTransCount, err := block.TransactionCount()
	if err != nil {
		return nil, err
	}
	if nthInBlock+1 >= blockTransCount {
		// Next block
		nextBlock, err := obc.NextBlock(block)
		if err != nil {
			return nil, err
		}
		if nextBlock.IsInvalid() {
			return obc.InvalidTrans(), nil
		}
		// First trans in next block
		_, err = obc.switchBlock(nextBlock.Height())
		if err != nil {
			return nil, err
		}
		return &obc.currentBlock.J_tx[0], nil
	}
	return &obc.currentBlock.J_tx[nthInBlock+1], nil
}
