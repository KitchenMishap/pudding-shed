package concurrency

import (
	"github.com/KitchenMishap/pudding-shed/jsonblock"
	"sync"
)

// ToDo: parameterize SequencerContainer / SequencerContainerHashes by type as they duplicate code

// NewSequencerContainer creates a SequenceContainer
// Objects are sent in to SequenceContainer.InChan in a pseudo-sequence.
// Objects come out of outChan in strict sequence order.
// The intention is to place this in a pipeline after a thread pool which may have disrupted the original sequence.
// To compensate for bloating, at a point upstream where things ARE still in sequence, you should
// call SequencerContainer.WaitForNotBloated() before sending objects to your thread pool.
func NewSequencerContainerHashes(firstOut int64, bloatedCount int, outChan *chan *jsonblock.JsonBlockHashes) *SequencerContainerHashes {
	result := SequencerContainerHashes{}
	result.InChan = make(chan *jsonblock.JsonBlockHashes)
	result.outChan = outChan
	result.theMap = make(map[int64]*jsonblock.JsonBlockHashes)
	result.nextOut = firstOut
	result.bloatedCount = bloatedCount
	result.bloatedCondition = sync.NewCond(&sync.Mutex{})
	go result.worker()
	return &result
}

type SequencerContainerHashes struct {
	InChan           chan *jsonblock.JsonBlockHashes
	outChan          *chan *jsonblock.JsonBlockHashes
	theMap           map[int64]*jsonblock.JsonBlockHashes
	nextOut          int64
	bloatedCount     int
	bloatedCondition *sync.Cond
	bloated          bool
}

func (sc *SequencerContainerHashes) worker() {
	for newVal := range sc.InChan {
		if newVal == nil {
			// input channel closed, should be able to empty the map
			for sc.sendNextOut() {
			}
			if len(sc.theMap) > 0 {
				panic("The sequencerContainer became jammed")
			}
			close(*sc.outChan)
			return
		} else if int64(newVal.J_height) == sc.nextOut {
			*sc.outChan <- newVal
			sc.nextOut++
			// May have freed up some others to go
			for sc.sendNextOut() {
			}
		} else {
			// Longing for a different sequence number to come in
			// Put this one aside in the map
			sc.theMap[int64(newVal.J_height)] = newVal
			// If we become bloated, lock the input valve upstream
			if len(sc.theMap) == sc.bloatedCount { // Important, == not > or >=
				sc.setBloatedCondition(true)
			}
		}
	}
}

func (sc *SequencerContainerHashes) sendNextOut() bool {
	next, ok := sc.theMap[sc.nextOut]
	if !ok {
		return false
	} else {
		*sc.outChan <- next
		// If we were exactly bloated, unlock the input valve upstream
		if len(sc.theMap) == sc.bloatedCount {
			sc.setBloatedCondition(false)
		}
		delete(sc.theMap, sc.nextOut)
		sc.nextOut++
		return true
	}
}

func (sc *SequencerContainerHashes) setBloatedCondition(bloated bool) {
	sc.bloated = bloated
	if !bloated {
		// Tell everyone that we're not bloated for the time being
		sc.bloatedCondition.Broadcast()
	}
}

func (sc *SequencerContainerHashes) WaitForNotBloated() {
	sc.bloatedCondition.L.Lock()
	for sc.bloated {
		sc.bloatedCondition.Wait()
	}
	sc.bloatedCondition.L.Unlock()
}
