package memblocker

import (
	"fmt"
	"runtime"
	"testing"
	"time"
)

type bigObject struct {
	chunk [100000000]byte // 100Mb
}

func TestMemBlocker(t *testing.T) {
	// We limit memory use to 8GB
	mb := NewMemBlocker(8000000000)

	// Start with nothing allocated
	mapOfBig := make(map[int]*bigObject)

	// When they end
	chan1 := make(chan bool)
	chan2 := make(chan bool)

	// The first goroutine sleeps 30 seconds then starts freeing
	go func() {
		time.Sleep(30 * time.Second)
		fmt.Println("Done sleeping")
		mapSize2 := len(mapOfBig)
		for mapSize2 > 0 {
			fmt.Println("Freeing", time.Now())
			mb.StartFreeingMem()
			for k, _ := range mapOfBig {
				delete(mapOfBig, k)
				break
			}
			runtime.GC()
			fmt.Println("Allocated: ", mb.countAllocatedMemory())
			mb.MemoryWasFreed()
			time.Sleep(2 * time.Second)
			mapSize2 = len(mapOfBig)
		}
		fmt.Println("First goroutine completed")
		chan1 <- true
	}()

	// The second goroutine allocates 0.1Gb every second 160 times over
	// (So would allocate 16GB if left unchecked)
	go func() {
		for i := 0; i < 160; i++ {
			mb.WaitForSpareMemory()
			// Allocate 100MB
			fmt.Println("Allocating")
			mapOfBig[i] = &bigObject{}
			fmt.Println("Allocated: ", mb.countAllocatedMemory())
			time.Sleep(1 * time.Second)
		}
		fmt.Println("Second goroutine completed")
		chan2 <- true
	}()

	_ = <-chan1
	_ = <-chan2
	fmt.Println("Finished")
}
