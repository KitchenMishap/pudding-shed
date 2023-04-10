package chainreadinterface

type Handle interface {
}

type HBlock interface {
	Handle
}

type HTransaction interface {
	Handle
}

type IHandles interface {
	HBlockFromHeight(BlockHeight int64) HBlock
	HBlockFromHash(BlockHash [32]byte) HBlock
	HeightFromHBlock(hBlock HBlock) int64
	HashFromHBlock(hBlock HBlock) [32]byte
	HTransactionFromHash(TransactionHash [32]byte) HTransaction
	HashFromHTransaction(hTransaction HTransaction) [32]byte
}
