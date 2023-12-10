package chainreadinterface

type ITransaction interface {
	ITransHandle
	TxiCount() (int64, error)
	NthTxi(n int64) (ITxiHandle, error)
	TxoCount() (int64, error)
	NthTxo(n int64) (ITxoHandle, error)
}

// Note that the Bitcoin Core "standard" json for a coinbase transaction DOES have an entry in vin for "coinbase".
// We instead define a coinbase transaction as not having ANY vins (not even coinbase)

type ITxi interface {
	ITxiHandle
	SourceTxo() (ITxoHandle, error)
}

type ITxo interface {
	ITxoHandle
	Satoshis() (int64, error)
}
