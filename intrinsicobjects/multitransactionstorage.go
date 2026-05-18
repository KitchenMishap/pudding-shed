package intrinsicobjects

// Previously, we called make() for txis and txos within each transaction.
// This caused huge gc pressure.
// Additionally, we took ScrptPubKey's (byte slices) as slices of the block binary,
// which meant the whole block binary was kept as long as the transaction object was kept!
// We solve both of these problems with a "storage" object held per block.

// If you want to process a transaction outside of a block, just supply one of these.

type MultiTransactionStorage struct {
	Txis    []Txi
	Txos    []Txo
	Scripts []byte
}

func NewMultiTransactionStorage() *MultiTransactionStorage {
	result := MultiTransactionStorage{}
	result.Txis = make([]Txi, 0, 5000)        // A suitable head-start for txis in a block
	result.Txos = make([]Txo, 0, 5000)        // A suitable head-start for txos in a block
	result.Scripts = make([]byte, 0, 100_000) // A suitable head-start for scripts in txos in a block
	return &result
}

func (s *MultiTransactionStorage) Reset() {
	s.Txis = s.Txis[:0]
	s.Txos = s.Txos[:0]
	s.Scripts = s.Scripts[:0]
}
