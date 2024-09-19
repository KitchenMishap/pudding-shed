package concurrency

import "github.com/KitchenMishap/pudding-shed/jsonblock"

// NewSequencerContainer creates a SequenceContainer
// Objects are sent in to SequenceContainer.InChan in a pseudo-sequence.
// Objects come out of outChan in strict sequence order.
// The intention is to place this in a pipeline after a thread pool which may have disrupted the original sequence.
func NewSequencerContainer(firstOut int64, bloatedCount int64, outChan chan *jsonblock.JsonBlockEssential) *SequencerContainer {
	result := SequencerContainer{}
	result.InChan = make(chan *jsonblock.JsonBlockEssential)
	result.outChan = outChan
	result.theMap = make(map[int64]*jsonblock.JsonBlockEssential)
	result.nextOut = firstOut
	result.bloatedCount = bloatedCount
	result.currentlyBloated = false
	go result.worker()
	return &result
}

type SequencerContainer struct {
	InChan           chan *jsonblock.JsonBlockEssential
	outChan          chan *jsonblock.JsonBlockEssential
	theMap           map[int64]*jsonblock.JsonBlockEssential
	nextOut          int64
	bloatedCount     int64
	currentlyBloated bool
}

func (sc *SequencerContainer) worker() {
	newVal := <-sc.InChan
	if newVal == nil {
		// input channel closed, should be able to empty the map
		for sc.sendNextOut() {
		}
		if len(sc.theMap) > 0 {
			panic("The sequencerContainer became jammed")
		}
		close(sc.outChan)
		return
	} else if int64(newVal.J_height) == sc.nextOut {
		sc.outChan <- newVal
		sc.nextOut++
		// May have freed up some others to go
		for sc.sendNextOut() {
		}
	} else {
		// Waiting for a different sequence number to come in
		// Put this one aside in the map
		sc.theMap[int64(newVal.J_height)] = newVal
		sc.currentlyBloated = int64(len(sc.theMap)) >= sc.bloatedCount
	}
}

func (sc *SequencerContainer) sendNextOut() bool {
	next, ok := sc.theMap[sc.nextOut]
	if !ok {
		return false
	} else {
		sc.outChan <- next
		sc.nextOut++
		delete(sc.theMap, sc.nextOut)
		sc.currentlyBloated = int64(len(sc.theMap)) >= sc.bloatedCount
		return true
	}
}
