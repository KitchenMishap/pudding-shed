package chainreadinterface

import "github.com/KitchenMishap/pudding-shed/indexedhashes"

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
	HBlockFromHeight(blockHeight int64) HBlock
	HeightFromHBlock(hBlock HBlock) int64
	HeightFromHTransaction(hTrans HTransaction) int64
	HashFromHBlock(hBlock HBlock) indexedhashes.Sha256
	HashFromHTransaction(hTransaction HTransaction) indexedhashes.Sha256
}
