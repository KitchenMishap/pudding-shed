package chainreadinterface

type Handle interface {
	// IsHandle is merely here to stop a consumer handing around pointers to arbitrary objects as Handles
	IsHandle() bool
	IsInvalid() bool
}

type HBlock interface {
	Handle
}

type HTransaction interface {
	Handle
}

type IHandles interface {
	HBlockFromHeight(BlockHeight int64) HBlock
	HeightFromHBlock(hBlock HBlock) int64
}
