package chainreadinterface

type IAddress interface {
	IAddressHandle
	TxoCount() (int64, error)           // May return a "not supported" error
	NthTxo(n int64) (ITxoHandle, error) // May return a "not supported" error
}
