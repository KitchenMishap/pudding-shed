package chainreadinterface

type ITransaction interface {
	ITransHandle
	TxiCount() (int64, error)
	NthTxi(n int64) (ITxiHandle, error)
	TxoCount() (int64, error)
	NthTxo(n int64) (ITxoHandle, error)
}

type ITxi interface {
	ITxiHandle
	SourceTxo() (ITxoHandle, error)
}

type ITxo interface {
	ITxoHandle
	Satoshis() (int64, error)
}
