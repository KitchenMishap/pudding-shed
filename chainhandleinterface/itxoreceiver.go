package chainhandleinterface

// You will pass an ITxoReceiver to IBlockChain, which will make calls to it giving you info about a transaction output

type ITxoReceiver interface {
	ResetReceiver()
	ReceiveParentTransactionHandle(TransactionHandle)
	ReceiveSatoshisValue(int64)
	ReceiveScriptPubByteToAppend(byte)
}

// You could implement your own, or this example concrete type will do the job just fine

func NewTxoReceiver() *TxoReceiver {
	result := TxoReceiver{}
	return &result
}

type TxoReceiver struct {
	ParentTransactionHandle TransactionHandle
	SatoshisValue           int64
	ScriptPubBytes          []byte
}

// TxoReceiver implements ITxoReceiver
var _ ITxoReceiver = (*TxoReceiver)(nil) // Check that implements

func (tor *TxoReceiver) ResetReceiver() { tor.ScriptPubBytes = tor.ScriptPubBytes[:0] }
func (tor *TxoReceiver) ReceiveParentTransactionHandle(handle TransactionHandle) {
	tor.ParentTransactionHandle = handle
}
func (tor *TxoReceiver) ReceiveSatoshisValue(value int64) { tor.SatoshisValue = value }
func (tor *TxoReceiver) ReceiveScriptPubByteToAppend(b byte) {
	tor.ScriptPubBytes = append(tor.ScriptPubBytes, b)
}
