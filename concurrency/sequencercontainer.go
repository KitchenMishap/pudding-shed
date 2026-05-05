package concurrency

import (
	"sync"
)

// NewSequencerContainer creates a SequenceContainer
// Objects are sent in to SequenceContainer.InChan in a pseudo-sequence.
// Objects come out of outChan in strict sequence order.
// The intention is to place this in a pipeline after a thread pool which may have disrupted the original sequence.
// To compensate for bloating, at a point upstream where things ARE still in sequence, you should
// call SequencerContainer.WaitForNotBloated() before sending objects to your thread pool.
func NewSequencerContainer[T Sequencable](firstOut int64, bloatedCount int, outChan chan T) *SequencerContainer[T] {
	result := SequencerContainer[T]{}
	result.InChan = make(chan T)
	result.outChan = outChan
	result.theMap = make(map[int64]T)
	result.nextOut = firstOut
	result.bloatedCount = bloatedCount
	result.bloatedCondition = sync.NewCond(&sync.Mutex{})
	go result.worker()
	return &result
}

type SequencerContainer[T Sequencable] struct {
	InChan           chan T
	outChan          chan T
	theMap           map[int64]T
	nextOut          int64
	bloatedCount     int
	bloatedCondition *sync.Cond
	bloated          bool
}

func (sc *SequencerContainer[T]) worker() {
	for newVal := range sc.InChan {
		seq := newVal.SequenceNumber()
		if seq == sc.nextOut {
			sc.outChan <- newVal
			sc.nextOut++
			// May have freed up some others to go
			for sc.sendNextOut() {
			}
		} else {
			// Longing for a different sequence number to come in
			// Put this one aside in the map
			sc.theMap[seq] = newVal
			// If we become bloated, lock the input valve upstream
			if len(sc.theMap) == sc.bloatedCount { // Important, == not > or >=
				sc.setBloatedCondition(true)
			}
		}
	}
	// input channel closed, should be able to empty the map
	for sc.sendNextOut() {
	}
	if len(sc.theMap) > 0 {
		panic("Sequencer jammed: missing sequence numbers in stream")
	}
	close(sc.outChan)
	return

}

func (sc *SequencerContainer[T]) sendNextOut() bool {
	next, ok := sc.theMap[sc.nextOut]
	if !ok {
		return false
	} else {
		sc.outChan <- next
		// If we were exactly bloated, unlock the input valve upstream
		if len(sc.theMap) == sc.bloatedCount {
			sc.setBloatedCondition(false)
		}
		delete(sc.theMap, sc.nextOut)
		sc.nextOut++
		return true
	}
}

func (sc *SequencerContainer[T]) setBloatedCondition(bloated bool) {
	sc.bloatedCondition.L.Lock()
	sc.bloated = bloated
	if !bloated {
		// Tell everyone that we're not bloated for the time being
		sc.bloatedCondition.Broadcast()
	}
	sc.bloatedCondition.L.Unlock()
}

func (sc *SequencerContainer[T]) WaitForNotBloated() {
	sc.bloatedCondition.L.Lock()
	for sc.bloated {
		sc.bloatedCondition.Wait()
	}
	sc.bloatedCondition.L.Unlock()
}
