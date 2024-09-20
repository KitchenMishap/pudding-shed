package concurrency

import (
	"fmt"
	"github.com/KitchenMishap/pudding-shed/jsonblock"
	"sync"
)

// NewSequencerContainer creates a SequenceContainer
// Objects are sent in to SequenceContainer.InChan in a pseudo-sequence.
// Objects come out of outChan in strict sequence order.
// The intention is to place this in a pipeline after a thread pool which may have disrupted the original sequence.
// To compensate for bloating, at a point upstream where things ARE still in sequence, you should
// call SequencerContainer.WaitForNotBloated() before sending objects to your thread pool.
func NewSequencerContainer(firstOut int64, bloatedCount int, outChan *chan *jsonblock.JsonBlockEssential) *SequencerContainer {
	result := SequencerContainer{}
	result.InChan = make(chan *jsonblock.JsonBlockEssential)
	result.outChan = outChan
	result.theMap = make(map[int64]*jsonblock.JsonBlockEssential)
	result.nextOut = firstOut
	result.bloatedCount = bloatedCount
	result.bloatedCondition = sync.NewCond(&sync.Mutex{})
	go result.worker()
	return &result
}

type SequencerContainer struct {
	InChan           chan *jsonblock.JsonBlockEssential
	outChan          *chan *jsonblock.JsonBlockEssential
	theMap           map[int64]*jsonblock.JsonBlockEssential
	nextOut          int64
	bloatedCount     int
	bloatedCondition *sync.Cond
	bloated          bool
}

func (sc *SequencerContainer) worker() {
	for newVal := range sc.InChan {
		fmt.Println("Received sequence ", newVal.J_height)
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
			fmt.Println("Outputing sequence ", newVal.J_height, " directly")
			*sc.outChan <- newVal
			sc.nextOut++
			// May have freed up some others to go
			for sc.sendNextOut() {
			}
		} else {
			// Longing for a different sequence number to come in
			// Put this one aside in the map
			fmt.Println("Storing sequence ", newVal.J_height, " (container size ", len(sc.theMap), ") due to out of sequence")
			sc.theMap[int64(newVal.J_height)] = newVal
			// If we become bloated, lock the input valve upstream
			if len(sc.theMap) == sc.bloatedCount { // Important, == not > or >=
				fmt.Println("BECAME BLOATED")
				sc.setBloatedCondition(true)
			}
		}
	}
}

func (sc *SequencerContainer) sendNextOut() bool {
	next, ok := sc.theMap[sc.nextOut]
	if !ok {
		return false
	} else {
		fmt.Println("Outputing sequence ", sc.nextOut, " from container")
		*sc.outChan <- next
		// If we were exactly bloated, unlock the input valve upstream
		if len(sc.theMap) == sc.bloatedCount {
			fmt.Println("NO LONGER BLOATED")
			sc.setBloatedCondition(false)
		}
		delete(sc.theMap, sc.nextOut)
		sc.nextOut++
		return true
	}
}

func (sc *SequencerContainer) setBloatedCondition(bloated bool) {
	sc.bloated = bloated
	if !bloated {
		// Tell everyone that we're not bloated for the time being
		sc.bloatedCondition.Broadcast()
	}
}

func (sc *SequencerContainer) WaitForNotBloated() {
	fmt.Println("Waiting for not bloated")
	sc.bloatedCondition.L.Lock()
	for sc.bloated {
		sc.bloatedCondition.Wait()
	}
	sc.bloatedCondition.L.Unlock()
	fmt.Println("Not bloated")
}
