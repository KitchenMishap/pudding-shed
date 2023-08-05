package tinychain

import (
	"github.com/KitchenMishap/pudding-shed/chainreadinterface"
	"github.com/KitchenMishap/pudding-shed/indexedhashes"
)

// A handle implements the Handle interface
// We choose to use heights for handles, held as int64's
type handle struct {
	height int64
}

// Implement Handle
// IsHandle() confirms that h is a handle
func (h handle) IsHandle() bool {
	return true
}

func (h handle) IsInvalid() bool {
	return h.height == -1
}

// Check that implements
var _ chainreadinterface.Handle = (*handle)(nil)

type handles struct {
}

// Implement IHandles

// HBlockFromHeight() returns the HBlock at a given height in the chain
func (h *handles) HBlockFromHeight(hgt int64) chainreadinterface.HBlock {
	return handle{height: hgt}
}
func (h *handles) HeightFromHBlock(han chainreadinterface.HBlock) int64 {
	return han.(handle).height
}
func (h *handles) HashFromHBlock(han chainreadinterface.HBlock) indexedhashes.Sha256 {
	return HashOfInt(uint64(h.HeightFromHBlock(han)))
}
func (h *handles) HashFromHTransaction(han chainreadinterface.HTransaction) indexedhashes.Sha256 {
	return HashOfInt(uint64(h.HeightFromHTransaction(han)))
}
func (h *handles) HTransactionFromHeight(hgt int64) chainreadinterface.HTransaction {
	return handle{height: hgt}
}
func (h *handles) HeightFromHTransaction(han chainreadinterface.HTransaction) int64 {
	return han.(handle).height
}
