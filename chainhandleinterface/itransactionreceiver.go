package chainhandleinterface

// You will pass an ITransactionReceiver to IBlockChain, which will make calls to it giving you info about a transaction

// BitcoinCore transactions are slightly different from PuddingShed transactions.
// PuddingShed coinbase transactions have no txis.
// Here we are talking about BitcoinCore transactions (ie, with coinbase txis)

type IBitcoinCoreTransactionReceiver interface {
	ResetReceiver()
	ReceiveParentBlockHandle(BlockHandle)
	ReceiveTransactionHash([32]byte)
	ReceiveIntField(field string, value int64)
	ReceiveTxiHandleToAppend(TxiHandle)
	ReceiveTxoHandleToAppend(TxoHandle)
}

// You could implement your own, or this example concrete type will do the job just fine

func NewBitcoinCoreTransactionReceiver() *BitcoinCoreTransactionReceiver {
	result := BitcoinCoreTransactionReceiver{}
	result.TxiHandles = make([]TxiHandle, 0)
	result.FieldMap = make(map[string]int64)
	return &result
}

type BitcoinCoreTransactionReceiver struct {
	ParentBlockHandle BlockHandle
	TransactionHash   [32]byte
	TxiHandles        []TxiHandle
	TxoHandles        []TxoHandle
	FieldMap          map[string]int64
}

// TransactionReceiver implements ITransactionReceiver
var _ IBitcoinCoreTransactionReceiver = (*BitcoinCoreTransactionReceiver)(nil) // Check that implements

func (tr *BitcoinCoreTransactionReceiver) ResetReceiver() {
	tr.TxiHandles = tr.TxiHandles[:0]
	tr.TxoHandles = tr.TxoHandles[:0]
}
func (tr *BitcoinCoreTransactionReceiver) ReceiveParentBlockHandle(blockHandle BlockHandle) {
	tr.ParentBlockHandle = blockHandle
}
func (tr *BitcoinCoreTransactionReceiver) ReceiveTransactionHash(hash [32]byte) {
	tr.TransactionHash = hash
}
func (tr *BitcoinCoreTransactionReceiver) ReceiveIntField(field string, value int64) {
	tr.FieldMap[field] = value
}

func (tr *BitcoinCoreTransactionReceiver) ReceiveTxiHandleToAppend(handle TxiHandle) {
	tr.TxiHandles = append(tr.TxiHandles, handle)
}
func (tr *BitcoinCoreTransactionReceiver) ReceiveTxoHandleToAppend(handle TxoHandle) {
	tr.TxoHandles = append(tr.TxoHandles, handle)
}
