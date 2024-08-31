package chainstorage

import (
	"errors"
	"github.com/KitchenMishap/pudding-shed/chainreadinterface"
	"github.com/KitchenMishap/pudding-shed/indexedhashes"
	"github.com/KitchenMishap/pudding-shed/intarrayarray"
	"github.com/KitchenMishap/pudding-shed/jsonblock"
	"github.com/KitchenMishap/pudding-shed/wordfile"
)

type concreteReadableChain struct {
	blkFirstTrans       wordfile.ReadAtWordCounter
	blkHashes           indexedhashes.HashReader
	trnHashes           indexedhashes.HashReader
	addrHashes          indexedhashes.HashReader
	trnFirstTxi         wordfile.ReadAtWordCounter
	trnFirstTxo         wordfile.ReadAtWordCounter
	txiTx               wordfile.ReadAtWordCounter
	txiVout             wordfile.ReadAtWordCounter
	txoSats             wordfile.ReadAtWordCounter
	txoAddress          wordfile.ReadAtWordCounter
	txoSpentTxi         wordfile.ReadAtWordCounter
	addrFirstTxo        wordfile.ReadAtWordCounter
	parentBlockOfTrans  wordfile.ReadAtWordCounter
	parentTransOfTxi    wordfile.ReadAtWordCounter
	parentTransOfTxo    wordfile.ReadAtWordCounter
	blkNonEssentialInts map[string]wordfile.ReadAtWordCounter
	trnNonEssentialInts map[string]wordfile.ReadAtWordCounter
	addrAdditionalTxos  intarrayarray.IntArrayMapStoreReadOnly
}

// Check that implements
var _ chainreadinterface.IBlockChain = (*concreteReadableChain)(nil)
var _ chainreadinterface.IHandleCreator = (*concreteReadableChain)(nil)
var _ IParents = (*concreteReadableChain)(nil)

// Functions to implement IBlockTree as part of IBlockChain

func (crc *concreteReadableChain) GenesisBlock() chainreadinterface.IBlockHandle {
	return &BlockHandle{HashHeight{height: 0, hashSpecified: false, heightSpecified: true}, crc}
}
func (crc *concreteReadableChain) ParentBlock(block chainreadinterface.IBlockHandle) chainreadinterface.IBlockHandle {
	if block.IsInvalid() {
		return crc.InvalidBlock()
	}
	if !block.HeightSpecified() {
		panic("this implementation of ParentBlock assumes block height is specified")
	}
	parentHeight := block.Height() - 1
	return &BlockHandle{HashHeight{height: parentHeight, hashSpecified: false, heightSpecified: true}, crc}
}
func (crc *concreteReadableChain) GenesisTransaction() (chainreadinterface.ITransHandle, error) {
	return &TransHandle{HashHeight{height: 0, hashSpecified: false, heightSpecified: true}, crc}, nil
}
func (crc *concreteReadableChain) PreviousTransaction(trans chainreadinterface.ITransHandle) chainreadinterface.ITransHandle {
	if trans.IsInvalid() {
		return crc.InvalidTrans()
	}
	if !trans.HeightSpecified() {
		panic("this implementation of PreviousTransaction assumes trans height is specified")
	}
	prevHeight := trans.Height() - 1
	return &TransHandle{HashHeight{height: prevHeight, hashSpecified: false, heightSpecified: true}, crc}
}
func (crc *concreteReadableChain) IsBlockTree() bool {
	return false
}
func (crc *concreteReadableChain) BlockInterface(hBlock chainreadinterface.IBlockHandle) (chainreadinterface.IBlock, error) {
	if hBlock.HeightSpecified() {
		blockHeight := hBlock.Height()
		return &Block{height: blockHeight, data: crc}, nil
	} else if hBlock.HashSpecified() {
		hash, err := hBlock.Hash()
		if err != nil {
			return &Block{}, err
		}
		blockHeight, err := crc.blkHashes.IndexOfHash(&hash)
		if err != nil {
			return &Block{}, err
		}
		return &Block{height: blockHeight, data: crc}, nil
	} else {
		panic("neither height nor hash was specified in BlockHandle")
	}
}
func (crc *concreteReadableChain) TransInterface(hTrans chainreadinterface.ITransHandle) (chainreadinterface.ITransaction, error) {
	if hTrans.HeightSpecified() {
		transHeight := hTrans.Height()
		return &Transaction{height: transHeight, data: crc}, nil
	} else if hTrans.HashSpecified() {
		hash, err := hTrans.Hash()
		if err != nil {
			return &Transaction{}, err
		}
		transHeight, err := crc.trnHashes.IndexOfHash(&hash)
		if err != nil {
			return &Transaction{}, err
		}
		return &Transaction{height: transHeight, data: crc}, nil
	} else {
		panic("neither height nor hash was specified in TransHandle")
	}
}

func (crc *concreteReadableChain) TxiInterface(hTxi chainreadinterface.ITxiHandle) (chainreadinterface.ITxi, error) {
	if hTxi.TxiHeightSpecified() {
		return &Txi{height: hTxi.TxiHeight(), data: crc}, nil
	} else if hTxi.ParentSpecified() {
		hTrans := hTxi.ParentTrans()
		trans, err := crc.TransInterface(hTrans)
		if err != nil {
			return &Txi{}, err
		}
		transFirstTxiHeight, err := crc.trnFirstTxi.ReadWordAt(trans.Height())
		if err != nil {
			return &Txi{}, err
		}
		index := hTxi.ParentIndex()
		txiHeight := transFirstTxiHeight + index
		return &Txi{height: txiHeight, data: crc}, nil
	} else {
		panic("hTxi specifies neither txiHeight nor parent")
	}
}

func (crc *concreteReadableChain) TxoInterface(hTxo chainreadinterface.ITxoHandle) (chainreadinterface.ITxo, error) {
	if hTxo.TxoHeightSpecified() {
		return &Txo{height: hTxo.TxoHeight(), data: crc}, nil
	} else if hTxo.ParentSpecified() {
		hTrans := hTxo.ParentTrans()
		trans, err := crc.TransInterface(hTrans)
		if err != nil {
			return &Txo{}, err
		}
		transFirstTxoHeight, err := crc.trnFirstTxo.ReadWordAt(trans.Height())
		if err != nil {
			return &Txo{}, err
		}
		index := hTxo.ParentIndex()
		txoHeight := transFirstTxoHeight + index
		return &Txo{height: txoHeight, data: crc}, nil
	} else {
		panic("hTxo specifies neither txoHeight nor parent")
	}
}

func (crc *concreteReadableChain) AddressInterface(hAddress chainreadinterface.IAddressHandle) (chainreadinterface.IAddress, error) {
	result := Address{}

	result.hashSpecified = hAddress.HashSpecified()
	if result.hashSpecified {
		result.hash = hAddress.Hash()
	}
	result.heightSpecified = hAddress.HeightSpecified()
	if result.heightSpecified {
		result.height = hAddress.Height()
	}

	result.data = crc
	result.populated = false

	return &result, nil
}

// Functions to implement IBlockChain

func (crc *concreteReadableChain) InvalidBlock() chainreadinterface.IBlockHandle {
	return &BlockHandle{HashHeight{height: -1, hashSpecified: false, heightSpecified: true}, crc}
}
func (crc *concreteReadableChain) InvalidTrans() chainreadinterface.ITransHandle {
	return &TransHandle{HashHeight{height: -1, hashSpecified: false, heightSpecified: true}, crc}
}

func (crc *concreteReadableChain) LatestBlock() (chainreadinterface.IBlockHandle, error) {
	blocksInChain, err := crc.blkHashes.CountHashes()
	if err != nil || blocksInChain == 0 {
		return crc.InvalidBlock(), err
	}
	return &BlockHandle{HashHeight{
		height:          blocksInChain - 1,
		hash:            indexedhashes.Sha256{},
		heightSpecified: true,
		hashSpecified:   false,
	}, crc}, nil
}

func (crc *concreteReadableChain) NextBlock(hBlock chainreadinterface.IBlockHandle) (chainreadinterface.IBlockHandle, error) {
	givenBlockNum := hBlock.Height()
	nextBlockNum := givenBlockNum + 1
	blocksInChain, err := crc.blkHashes.CountHashes()
	if err != nil || nextBlockNum >= blocksInChain {
		return crc.InvalidBlock(), err
	}
	return &BlockHandle{HashHeight{
		height:          nextBlockNum,
		hash:            indexedhashes.Sha256{},
		heightSpecified: true,
		hashSpecified:   false,
	}, crc}, nil
}

func (crc *concreteReadableChain) LatestTransaction() (chainreadinterface.ITransHandle, error) {
	transInChain, err := crc.trnHashes.CountHashes()
	if err != nil || transInChain == 0 {
		return crc.InvalidTrans(), err
	}
	return &TransHandle{HashHeight{
		height:          transInChain - 1,
		hash:            indexedhashes.Sha256{},
		heightSpecified: true,
		hashSpecified:   false,
	}, crc}, nil
}

func (crc *concreteReadableChain) NextTransaction(hTrans chainreadinterface.ITransHandle) (chainreadinterface.ITransHandle, error) {
	givenTransNum := hTrans.Height()
	nextTransNum := givenTransNum + 1
	transInChain, err := crc.trnHashes.CountHashes()
	if err != nil || nextTransNum >= transInChain {
		return crc.InvalidTrans(), err
	}
	return &TransHandle{HashHeight{
		height:          nextTransNum,
		hash:            indexedhashes.Sha256{},
		heightSpecified: true,
		hashSpecified:   false,
	}, crc}, nil
}

func (crc *concreteReadableChain) BlockHandleByHeight(blockHeight int64) (chainreadinterface.IBlockHandle, error) {
	result := BlockHandle{}
	result.heightSpecified = true
	result.height = blockHeight
	result.hashSpecified = false
	result.data = crc
	return &result, nil
}
func (crc *concreteReadableChain) TransactionHandleByHeight(transactionHeight int64) (chainreadinterface.ITransHandle, error) {
	result := TransHandle{}
	result.heightSpecified = true
	result.height = transactionHeight
	result.hashSpecified = false
	result.data = crc
	return &result, nil
}
func (crc *concreteReadableChain) TxiHandleByHeight(txiHeight int64) (chainreadinterface.ITxiHandle, error) {
	result := TxiHandle{}
	result.txxHeightSpecified = true
	result.txxHeight = txiHeight
	result.transIndexSpecified = false
	result.data = crc
	return &result, nil
}
func (crc *concreteReadableChain) TxoHandleByHeight(txoHeight int64) (chainreadinterface.ITxoHandle, error) {
	result := TxoHandle{}
	result.txxHeightSpecified = true
	result.txxHeight = txoHeight
	result.transIndexSpecified = false
	result.data = crc
	return &result, nil
}
func (crc *concreteReadableChain) AddressHandleByHeight(addressHeight int64) (chainreadinterface.IAddressHandle, error) {
	result := AddressHandle{}
	result.heightSpecified = true
	result.height = addressHeight
	return &result, nil
}
func (crc *concreteReadableChain) AddressHandleByHash(sha256 indexedhashes.Sha256) (chainreadinterface.IAddressHandle, error) {
	result := AddressHandle{}
	index, _ := crc.addrHashes.IndexOfHash(&sha256)
	if index == -1 {
		return nil, errors.New("address hash not found")
	}
	result.hashSpecified = false
	result.heightSpecified = true
	result.height = index
	return &result, nil
}
func (crc *concreteReadableChain) AddressHandleByString(address string) (chainreadinterface.IAddressHandle, error) {
	hash := jsonblock.HashOfString(address)
	return crc.AddressHandleByHash(hash)
}

func (crc *concreteReadableChain) TransactionHandleByHash(sha256 indexedhashes.Sha256) (chainreadinterface.ITransHandle, error) {
	result := TransHandle{}
	index, _ := crc.trnHashes.IndexOfHash(&sha256)
	if index == -1 {
		return nil, errors.New("transaction hash not found")
	}
	result.hashSpecified = false
	result.heightSpecified = true
	result.height = index
	return &result, nil
}

func (crc *concreteReadableChain) BlockHandleByHash(sha256 indexedhashes.Sha256) (chainreadinterface.IBlockHandle, error) {
	result := BlockHandle{}
	index, _ := crc.blkHashes.IndexOfHash(&sha256)
	if index == -1 {
		return nil, errors.New("block hash not found")
	}
	result.hashSpecified = false
	result.heightSpecified = true
	result.height = index
	return &result, nil
}

func (crc *concreteReadableChain) ParentBlockOfTrans(transactionHeight int64) (int64, error) {
	return crc.parentBlockOfTrans.ReadWordAt(transactionHeight)
}
func (crc *concreteReadableChain) ParentTransOfTxi(transactionHeight int64) (int64, error) {
	return crc.parentTransOfTxi.ReadWordAt(transactionHeight)
}
func (crc *concreteReadableChain) ParentTransOfTxo(transactionHeight int64) (int64, error) {
	return crc.parentTransOfTxo.ReadWordAt(transactionHeight)
}
