package chainreadinterface

type ITransaction interface {
	ITransHandle
	TxiCount() int64
	NthTxi(n int64) ITxiHandle
	TxoCount() int64
	NthTxo(n int64) ITxoHandle
}

type ITxi interface {
	ITxiHandle
	SourceTxo() ITxoHandle
}

type ITxo interface {
	ITxoHandle
	Satoshis() int64
}
