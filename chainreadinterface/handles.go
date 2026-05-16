package chainreadinterface

type handleHash [32]byte

type gpHandle struct {
	hash    handleHash
	integer int64
}

type BlockHandle struct {
	gpHandle
}
