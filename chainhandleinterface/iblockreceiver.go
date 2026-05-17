package chainhandleinterface

// You will pass an IBlockReceiver to IBlockChain, which will make calls to it giving you info about a block

type IBlockReceiver interface {
	ResetReceiver()
	ReceiveBlockHeight(int64)
	ReceiveBlockHash([32]byte)
	ReceiveIntField(field string, value int64)
	ReceiveTransactionHandleToAppend(TransactionHandle)
}

// You could implement your own, or this example concrete type will do the job just fine

func NewBlockReceiver() *BlockReceiver {
	result := BlockReceiver{}
	result.TransactionHandles = make([]TransactionHandle, 0)
	result.FieldMap = make(map[string]int64)
	return &result
}

type BlockReceiver struct {
	BlockHeight        int64
	BlockHash          [32]byte
	TransactionHandles []TransactionHandle
	FieldMap           map[string]int64
}

// BlockReceiver implements IBlockReceiver
var _ IBlockReceiver = (*BlockReceiver)(nil) // Check that implements

func (br *BlockReceiver) ResetReceiver()                            { br.TransactionHandles = br.TransactionHandles[:0] }
func (br *BlockReceiver) ReceiveBlockHeight(height int64)           { br.BlockHeight = height }
func (br *BlockReceiver) ReceiveBlockHash(hash [32]byte)            { br.BlockHash = hash }
func (br *BlockReceiver) ReceiveIntField(field string, value int64) { br.FieldMap[field] = value }
func (br *BlockReceiver) ReceiveTransactionHandleToAppend(handle TransactionHandle) {
	br.TransactionHandles = append(br.TransactionHandles, handle)
}
