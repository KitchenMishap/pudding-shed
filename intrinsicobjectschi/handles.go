package intrinsicobjectschi

import (
	"encoding/binary"

	"github.com/KitchenMishap/pudding-shed/chainhandleinterface"
)

// Functions in here must be local to intrinsicobjectschi (ie, lower case) so that they are not exposed to
// clients of this particular chainhandlesinterface.IBlockChain implementation

// In here are policies of the internals of handles for intrinsicobjectschi

func makeBlockHandle(blockHeight int64) chainhandleinterface.BlockHandle {
	if blockHeight < 0 || blockHeight > 0xFFFFFFFF {
		panic("Cannot represent this block height")
	}
	result := chainhandleinterface.BlockHandle{}
	binary.LittleEndian.PutUint32(result.Bytes[:], uint32(blockHeight))
	return result
}

func makeInvalidBlockHandle() chainhandleinterface.BlockHandle {
	result := chainhandleinterface.BlockHandle{}
	binary.LittleEndian.PutUint32(result.Bytes[:], 0xFFFFFFFF)
	return result
}

func isInvalidBlockHandle(handle chainhandleinterface.BlockHandle) bool {
	return binary.LittleEndian.Uint32(handle.Bytes[:]) == 0xFFFFFFFF
}

func makeTransactionSubHandle(indexWithinBlock int64) chainhandleinterface.TransactionSubHandle {
	if indexWithinBlock < 0 || indexWithinBlock > 0xFFFFFFFF {
		panic("Cannot represent this transaction index")
	}
	result := chainhandleinterface.TransactionSubHandle{}
	binary.LittleEndian.PutUint32(result.Bytes[:], uint32(indexWithinBlock))
	return result
}

func indexInBlockFromTransactionSubHandle(handle chainhandleinterface.TransactionSubHandle) int64 {
	return int64(binary.LittleEndian.Uint32(handle.Bytes[:]))
}

func makeTxiSubHandle(indexWithinTrans int64) chainhandleinterface.TxiSubHandle {
	if indexWithinTrans < 0 || indexWithinTrans > 0xFFFFFFFF {
		panic("Cannot represent this txi index")
	}
	result := chainhandleinterface.TxiSubHandle{}
	binary.LittleEndian.PutUint32(result.Bytes[:], uint32(indexWithinTrans))
	return result
}

func makeTxoSubHandle(indexWithinTrans int64) chainhandleinterface.TxoSubHandle {
	if indexWithinTrans < 0 || indexWithinTrans > 0xFFFFFFFF {
		panic("Cannot represent this txi index")
	}
	result := chainhandleinterface.TxoSubHandle{}
	binary.LittleEndian.PutUint32(result.Bytes[:], uint32(indexWithinTrans))
	return result
}

func txiIndexFromTxiSubHandle(handle chainhandleinterface.TxiSubHandle) int64 {
	return int64(binary.LittleEndian.Uint32(handle.Bytes[:]))
}
func txoIndexFromTxiSubHandle(handle chainhandleinterface.TxoSubHandle) int64 {
	return int64(binary.LittleEndian.Uint32(handle.Bytes[:]))
}
