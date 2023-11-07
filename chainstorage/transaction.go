package chainstorage

import (
	"github.com/KitchenMishap/pudding-shed/chainreadinterface"
	"github.com/KitchenMishap/pudding-shed/indexedhashes"
)

// Transaction implements ITransaction
type Transaction struct {
	height int64
	data   *concreteReadableChain
}

// Functions that implement ITransHandle as part of ITransaction

func (tr *Transaction) Height() int64 {
	return tr.height
}
func (tr *Transaction) Hash() (indexedhashes.Sha256, error) {
	hash := indexedhashes.Sha256{}
	err := tr.data.trnHashes.GetHashAtIndex(tr.height, &hash)
	if err != nil {
		return indexedhashes.Sha256{}, err
	}
	return hash, nil
}
func (tr *Transaction) HeightSpecified() bool {
	return true
}
func (tr *Transaction) HashSpecified() bool {
	return true
}
func (tr *Transaction) IsTransHandle() {}
func (tr *Transaction) IsInvalid() bool {
	return tr.height == -1
}

// Functions that implement ITransaction

func (tr *Transaction) TxiCount() (int64, error) {
	transInChain, err := tr.data.trnHashes.CountHashes()
	if err != nil {
		return -1, err
	}
	txisInChain, err := tr.data.txiTx.CountWords()
	if err != nil {
		return -1, err
	}

	tranFirstTxiHeight, err := tr.data.trnFirstTxi.ReadWordAt(tr.height)
	if err != nil {
		return -1, err
	}
	nextTranHeight := tr.height + 1
	if nextTranHeight < transInChain {
		nextTransFirstTxiHeight, err := tr.data.trnFirstTxi.ReadWordAt(nextTranHeight)
		if err != nil {
			return -1, err
		}
		return nextTransFirstTxiHeight - tranFirstTxiHeight, nil
	} else {
		// There might not be a next transaction
		return txisInChain - tranFirstTxiHeight, nil
	}
}
func (tr *Transaction) NthTxi(n int64) (chainreadinterface.ITxiHandle, error) {
	transFirstTxiHeight, err := tr.data.trnFirstTxi.ReadWordAt(tr.height)
	if err != nil {
		return &TxiHandle{}, err
	}
	txiHeight := transFirstTxiHeight + n
	return &TxiHandle{TxxHandle{
		TransIndex:          TransIndex{},
		txxHeight:           txiHeight,
		transIndexSpecified: false,
		txxHeightSpecified:  true,
	}}, nil
}

func (tr *Transaction) TxoCount() (int64, error) {
	transInChain, err := tr.data.trnHashes.CountHashes()
	if err != nil {
		return -1, err
	}
	txosInChain, err := tr.data.txoSats.CountWords()
	if err != nil {
		return -1, err
	}

	tranFirstTxoHeight, err := tr.data.trnFirstTxo.ReadWordAt(tr.height)
	if err != nil {
		return -1, err
	}
	nextTranHeight := tr.height + 1
	if nextTranHeight < transInChain {
		nextTransFirstTxoHeight, err := tr.data.trnFirstTxo.ReadWordAt(nextTranHeight)
		if err != nil {
			return -1, err
		}
		return nextTransFirstTxoHeight - tranFirstTxoHeight, nil
	} else {
		// There might not be a next transaction
		return txosInChain - tranFirstTxoHeight, nil
	}
}
func (tr *Transaction) NthTxo(n int64) (chainreadinterface.ITxoHandle, error) {
	transFirstTxoHeight, err := tr.data.trnFirstTxo.ReadWordAt(tr.height)
	if err != nil {
		return &TxoHandle{}, err
	}
	txoHeight := transFirstTxoHeight + n
	return &TxoHandle{TxxHandle{
		TransIndex:          TransIndex{},
		txxHeight:           txoHeight,
		transIndexSpecified: false,
		txxHeightSpecified:  true,
	}}, nil
}

// Compiler check that implements
var _ chainreadinterface.ITransaction = (*Transaction)(nil)

// Txi implements ITxi
type Txi struct {
	height int64
	data   *concreteReadableChain
}

// Compiler check that implements
var _ chainreadinterface.ITxi = (*Txi)(nil)

// Functions to implement ITxiHandle as part of ITxi

func (txi *Txi) ParentTrans() chainreadinterface.ITransHandle {
	return &TransHandle{}
}
func (txi *Txi) ParentIndex() int64 {
	return -1
}
func (txi *Txi) TxiHeight() int64 {
	return txi.height
}
func (txi *Txi) ParentSpecified() bool {
	return false
}
func (txi *Txi) TxiHeightSpecified() bool {
	return true
}

// Functions to implement ITxi

func (txi *Txi) SourceTxo() (chainreadinterface.ITxoHandle, error) {
	sourceTransHeight, err := txi.data.txiTx.ReadWordAt(txi.height)
	if err != nil {
		return &TxoHandle{}, err
	}
	sourceTransVout, err := txi.data.txiVout.ReadWordAt(txi.height)
	if err != nil {
		return &TxoHandle{}, err
	}
	return &TxoHandle{TxxHandle{
		TransIndex: TransIndex{
			TransHandle: TransHandle{HashHeight{
				height:          sourceTransHeight,
				hash:            indexedhashes.Sha256{},
				heightSpecified: true,
				hashSpecified:   false,
			}, txi.data},
			index: sourceTransVout,
		},
		txxHeight:           -1,
		transIndexSpecified: true,
		txxHeightSpecified:  false,
	}}, nil
}

// Txo implements ITxo
type Txo struct {
	height int64
	data   *concreteReadableChain
}

// Compiler check that implements
var _ chainreadinterface.ITxo = (*Txo)(nil)

// Functions to implement ITxoHandle as part of ITxo

func (txo *Txo) ParentTrans() chainreadinterface.ITransHandle {
	return &TransHandle{}
}
func (txo *Txo) ParentIndex() int64 {
	return -1
}
func (txo *Txo) TxoHeight() int64 {
	return txo.height
}
func (txo *Txo) ParentSpecified() bool {
	return false
}
func (txo *Txo) TxoHeightSpecified() bool {
	return true
}

// Functions to implement ITxo

func (txo *Txo) Satoshis() (int64, error) {
	sats, err := txo.data.txoSats.ReadWordAt(txo.height)
	if err != nil {
		return -1, err
	}
	return sats, nil
}
