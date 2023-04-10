package tinychain

import "github.com/KitchenMishap/pudding-shed/chainreadinterface"

// A handle implements the Handle interface
// We choose to use heights for handles, held as int64's
type handle struct {
	height int64
}

// Implement Handle
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
func (h *handles) HBlockFromHeight(hgt int64) chainreadinterface.HBlock {
	return handle{height: hgt}
}
func (h *handles) HeightFromHBlock(han chainreadinterface.HBlock) int64 {
	return han.(handle).height
}
func (h *handles) hTransactionFromHeight(hgt int64) chainreadinterface.HTransaction {
	return handle{height: hgt}
}
func (h *handles) heightFromHTransaction(han chainreadinterface.HTransaction) int64 {
	return han.(handle).height
}
