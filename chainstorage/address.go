package chainstorage

import "github.com/KitchenMishap/pudding-shed/chainreadinterface"

// Address implements IAddress
type Address struct {
	AddressHandle
	data       *concreteReadableChain
	populated  bool
	txoHeights []int64
}

func (adr *Address) populate() error {
	adr.txoHeights = []int64{} // In case we hit an error
	var addressHeight int64
	if adr.AddressHandle.heightSpecified {
		addressHeight = adr.AddressHandle.height
	} else {
		addressHash := adr.AddressHandle.hash
		var err error
		addressHeight, err = adr.data.addrHashes.IndexOfHash(&addressHash)
		if err != nil {
			return err
		}
	}
	firstTxoHeight, err := adr.data.addrFirstTxo.ReadWordAt(addressHeight)
	if err != nil {
		return err
	}
	adr.txoHeights = []int64{firstTxoHeight}
	additionalTxos := adr.data.addrAdditionalTxos.GetArray(addressHeight)
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
