package chainhandleinterface

// The usage and meaning of these "Bytes" fields is completely up to any particular implementation of IBlockChain.
// A client of IBlockChain should never try to interpret them, nor store them between sessions of the software.

// Handles only make sense in the context of the longest chain.
// SubHandles can however be used within blocks and transactions that are outside the longest chain.

type BlockHandle struct {
	Bytes [4]byte // Just four, until someday we may need more
}
type TransactionSubHandle struct {
	Bytes [4]byte // Just four, until someday we may need more
}
type TxiSubHandle struct {
	Bytes [4]byte
}
type TxoSubHandle struct {
	Bytes [4]byte
}

// A client of IBlockChain is permitted to construct the following from their component handles and subhandles
type TransactionHandle struct {
	BH  BlockHandle
	TSH TransactionSubHandle
}
type TxiHandle struct {
	TH    TransactionHandle
	TXISH TxiSubHandle
}
type TxoHandle struct {
	TH    TransactionHandle
	TXOSH TxoSubHandle
}

type AddressHandle struct {
	Bytes [8]byte
}
