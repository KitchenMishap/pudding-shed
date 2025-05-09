package chainstorage

import (
	"github.com/KitchenMishap/pudding-shed/chainreadinterface"
	"github.com/KitchenMishap/pudding-shed/indexedhashes"
)

// Address implements IAddress
type Address struct {
	handle     AddressHandle
	data       *concreteReadableChain
	populated  bool
	txoHeights []int64
}

func (adr *Address) populate() error {
	adr.txoHeights = []int64{} // In case we hit an error
	var addressHeight int64
	if adr.handle.heightSpecified {
		addressHeight = adr.handle.height
	} else {
		addressHash := adr.handle.hash
		var err error
		addressHeight, err = adr.data.addrHashes.IndexOfHash(&addressHash)
		if err != nil {
			return err
		}
	}
	if !adr.handle.hashSpecified {
		err := adr.data.addrHashes.GetHashAtIndex(addressHeight, &adr.handle.hash)
		if err != nil {
			return err
		}
		adr.handle.hashSpecified = true
	}
	firstTxoHeight, err := adr.data.addrFirstTxo.ReadWordAt(addressHeight)
	if err != nil {
		return err
	}
	adr.txoHeights = []int64{firstTxoHeight}
	additionalTxos, err := adr.data.addrAdditionalTxos.GetArray(addressHeight)
	if err != nil {
		return err
	}

	adr.txoHeights = append(adr.txoHeights, additionalTxos...)
	adr.populated = true
	return nil
}

func (adr *Address) TxoCount() (int64, error) {
	if !adr.populated {
		err := adr.populate()
		if err != nil {
			return -1, err
		}
	}
	return int64(len(adr.txoHeights)), nil
}

func (adr *Address) NthTxo(n int64) (chainreadinterface.ITxoHandle, error) {
	if !adr.populated {
		err := adr.populate()
		if err != nil {
			return nil, err
		}
	}
	handle := TxoHandle{}
	handle.txxHeightSpecified = true
	handle.txxHeight = adr.txoHeights[n]
	handle.transIndexSpecified = false
	return &handle, nil
}

func (adr *Address) Hash() indexedhashes.Sha256 {
	if !adr.populated {
		adr.populate()
	}
	return adr.handle.hash
}

func (adr *Address) HashSpecified() bool {
	return true // Because Hash() will populate the hash if its not already there
}

func (adr *Address) Height() int64 {
	return adr.handle.height
}

func (adr *Address) HeightSpecified() bool {
	return adr.handle.heightSpecified
}
