package chainhandleinterface

// You will pass an ITxiReceiver to IBlockChain, which will make calls to it giving you info about a transaction input

type ITxiReceiver interface {
	ReceiveParentTransactionHandle(TransactionHandle)
	ReceiveIncomingTxid([32]byte)
	ReceiveIncomingVout(int64)
}

// You could implement your own, or this example concrete type will do the job just fine

func NewTxiReceiver() *TxiReceiver {
	result := TxiReceiver{}
	return &result
}

type TxiReceiver struct {
	ParentTransactionHandle TransactionHandle
	IncomingTxid            [32]byte
	IncomingVout            int64
}

// TxiReceiver implements ITxiReceiver
var _ ITxiReceiver = (*TxiReceiver)(nil) // Check that implements

func (tir *TxiReceiver) ReceiveParentTransactionHandle(handle TransactionHandle) {
	tir.ParentTransactionHandle = handle
}
func (tir *TxiReceiver) ReceiveIncomingTxid(txid [32]byte) { tir.IncomingTxid = txid }
func (tir *TxiReceiver) ReceiveIncomingVout(vout int64)    { tir.IncomingVout = vout }
