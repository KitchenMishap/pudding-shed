package concurrency

// Sequenceable is a type that can be passed through a SequenceContainer
type Sequenceable interface {
	SequenceNumber() int64
}

// NewSequencerContainer creates a SequenceContainer
// Objects are sent in to SequenceContainer.InChan in a pseudo-sequence.
// Objects come out of outChan in strict sequence order.
// The intention is to place this in a pipeline after a thread pool which may have disrupted the original sequence.
func NewSequencerContainer(firstOut int64, bloatedCount int64, outChan* chan Sequenceable) *SequencerContainer {
	result := SequencerContainer{}
	result.InChan = make( chan Sequenceable )
	result.outChan = outChan
	result.theMap = make map[int64]Sequenceable
	result.nextOut = firstOut
	result.bloatedCount = bloatedCount
	result.currentlyBloated = false;
	go result.worker()
	return &result
}

type SequencerContainer struct {
	InChan chan Sequenceable
	outChan* chan Sequenceable
	theMap map[int64]Sequenceable
	nextOut int64
	bloatedCount int64
	currentlyBloated bool
}

func (sc *SequencerContainer) worker()
{
	newVal := <-container.InChan
	if newVal == nil {
		// input channel closed, should be able to empty the map
		for sendNextOut() {
		}
		if len(sc.theMap) > 0 {
			panic( "The sequencerContainer became jammed")
		}
		sc.outChan.Close()
		return
	} else if newVal.SequenceNumber() == sc.nextOut {
		sc.outChan <- newVal
		sc.nextOut++
		// May have freed up some others to go
		for sendNextOut() {
		}
	} else {
		// Waiting for a different sequence number to come in
		// Put this one aside in the map
		sc.theMap[newVal.SequenceNumber()] = newVal
		sc.currentlyBloated = len(sc.theMap) >= sc.bloatedCount
	}
}

func (sc *SequencerContainer) sendNextOut() bool {
	next, ok := sc.theMap[sc.nextOut]
	if !ok {
		return false
	} else {
		sc.outChan<- next
		sc.nextOut++
		delete(sc.theMap, sc.nextOut)
		sc.currentlyBloated = int64(len(sc.theMap)) >= sc.bloatedCount
		return true;
	}
}
