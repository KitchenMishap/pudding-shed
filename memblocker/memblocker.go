package memblocker

import (
	"runtime"
	"sync"
)

func NewMemBlocker(maxBytes uint64) *MemBlocker {
	result := MemBlocker{}
	result.maxBytes = maxBytes
	result.someBytesFreedCond = sync.NewCond(&sync.Mutex{})
	return &result
}

type MemBlocker struct {
	maxBytes           uint64 // The number of process allocated bytes above which we must wait
	memStats           runtime.MemStats
	lastHeapSize       uint64
	someBytesFreedCond *sync.Cond
}

func (mb *MemBlocker) IsThereFreeMemory() bool {
	allocatedBytes := mb.countAllocatedMemory()
	return allocatedBytes < mb.maxBytes
}

// WaitForSpareMemory waits until this process's allocated bytes on the heap falls below maxBytes
func (mb *MemBlocker) WaitForSpareMemory() {
	mb.someBytesFreedCond.L.Lock()
	allocatedBytes := mb.countAllocatedMemory()
	for allocatedBytes >= mb.maxBytes {
		// Currently too much memory is allocated on the heap of this process.
		// We will have to wait for a routine who's in the habit of calling MemoryWasFreed()...
		mb.someBytesFreedCond.Wait()
		allocatedBytes = mb.countAllocatedMemory()
	}
	// Now allocatedBytes has dropped below the threshold
	mb.someBytesFreedCond.L.Unlock()
}

func (mb *MemBlocker) StartFreeingMem() {
	mb.someBytesFreedCond.L.Lock()
}
func (mb *MemBlocker) MemoryWasFreed() {
	mb.someBytesFreedCond.L.Unlock()
	mb.someBytesFreedCond.Signal()
}

// countAllocatedMemory returns the current process heap allocation.
// Use it as a very rough estimate (lower bound) as to how much memory
// this program is using
func (mb *MemBlocker) countAllocatedMemory() uint64 {
	runtime.ReadMemStats(&mb.memStats)
	mb.lastHeapSize = mb.memStats.HeapAlloc
	return mb.memStats.HeapAlloc
}

func (mb *MemBlocker) LastHeapSize() uint64 {
	return mb.lastHeapSize
}
