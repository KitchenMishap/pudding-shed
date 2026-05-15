package indexedhashes3

import (
	"math"
)

// A bin is conceptually a container (which can grow) of items which are binEntryBytes.
// It is the way a bin is held whilst in memory.
// It is implemented as a contiguous slice of bytes.
// (This design is re-factored after a previous "slice of slices" design caused horrendous gc avalanches)
// Index parameters in this code file are indices of entries (not bytes)
type bin struct {
	bytes []byte
}

type BinsArray struct {
	bins []*bin
}

// Reset reuses existing bins as far as possible
func (a BinsArray) Reuse(expectedEntriesPerBin int64, bytesPerEntry int64, numberOfBins int64) {
	// First empty each existing bin (keeping any capacity)
	for i := 0; i < len(a.bins); i++ {
		a.bins[i].bytes = a.bins[i].bytes[:0]
	}
	// Append extra bins if needed (with new capacity)
	for i := int64(len(a.bins)); i < numberOfBins; i++ {
		a.bins = append(a.bins, &bin{})
		a.bins[i].bytes = make([]byte, 0, expectedEntriesPerBin*bytesPerEntry)
	}
	// Trim down the outer array if necessary (keeping capacity)
	a.bins = a.bins[:numberOfBins]

	// We now have the right number of empty bins, and each has sensible capacity
}

func NewBinsArray(expectedEntriesPerBin int64, bytesPerEntry int64, numberOfBins int64) *BinsArray {
	result := BinsArray{}
	result.bins = make([]*bin, numberOfBins)
	for i := range numberOfBins {
		item := bin{}
		// Empty but with capacity
		item.bytes = make([]byte, 0, expectedEntriesPerBin*bytesPerEntry)
		result.bins[i] = &item
	}
	return &result
}

func (b *bin) getEntry(index int64, p *HashIndexingParams) binEntryBytes {
	bytesPerEntry := p.BytesPerBinEntry()
	return b.bytes[bytesPerEntry*index : bytesPerEntry*(index+1)]
}

func (b *bin) setEntry(index int64, val binEntryBytes) {
	bytesPerEntry := int64(len(val))
	copy(b.bytes[bytesPerEntry*index:], val)
}

func (b *bin) length(bytesPerEntry int64) int64 {
	return int64(len(b.bytes)) / bytesPerEntry
}

func (b *bin) insertBinEntry(sn sortNum, hi hashIndex, th truncatedHash, p *HashIndexingParams) int64 {
	bytesPerEntry := p.BytesPerBinEntry()

	newEntry := newBinEntryBytes(th, hi, sn, p)
	insertionPoint := b.findIndexBasedOnSortNum(sn, p)
	if insertionPoint > b.length(bytesPerEntry) {
		panic("insertionPoint too large")
	}
	// For some reason I couldn't get slices.Insert() to compile
	// Alternative insertion algorithm:
	if len(b.bytes) == 0 {
		// Special case if empty
		b.bytes = append(b.bytes, newEntry...)
	} else {
		// First append a copy of the last item
		count := b.length(bytesPerEntry)
		b.bytes = append(b.bytes, b.bytes[bytesPerEntry*(count-1):bytesPerEntry*count]...)
		count++

		// Then copy stuff forward by one
		copy(b.bytes[bytesPerEntry*(insertionPoint+1):], b.bytes[bytesPerEntry*insertionPoint:bytesPerEntry*(count-1)])

		// Finally drop in the inserted value
		b.setEntry(insertionPoint, newEntry)
	}
	return insertionPoint
}

func (b *bin) lookupByHash(th truncatedHash, sn sortNum, p *HashIndexingParams) hashIndex {
	bytesPerEntry := p.BytesPerBinEntry()
	firstMatchingIndex := b.findIndexBasedOnSortNum(sn, p)
	// Search sequentially from here until everything matches or sn does not match
	for index := firstMatchingIndex; index < b.length(bytesPerEntry); index++ {
		entry := b.getEntry(index, p)
		hiFound, snFound := entry.getHashIndexSortNum(p)
		if snFound != sn {
			return -1 // hash not present in bin
		}
		thFound := entry.getTruncatedHash()
		if thFound.equals(th) {
			return hiFound
		}
	}
	return -1 // hash not present in bin
}

// Because there are a (very) few repeated transaction hashes on the blockchain, our method
// currently cannot find some of those repeated hashes by index. In that circumstance,
// we return a hash of all zeroes!
func (b *bin) lookupByIndex(hi hashIndex, bn binNum, p *HashIndexingParams) *Hash {
	bytesPerEntry := p.BytesPerBinEntry()
	// No shortcut. Sequential search
	for index := int64(0); index < b.length(bytesPerEntry); index++ {
		entry := b.getEntry(index, p)
		hiFound, snFound := entry.getHashIndexSortNum(p)
		if hiFound == hi {
			th := entry.getTruncatedHash()
			ah := newAbbreviatedHashFromBinNumSortNum(bn, snFound, p)
			h := NewHashFromTruncatedHashAbbreviatedHash(th, ah)
			return h
		}
	}
	// Looks like we've got one of those repeated hash situations!
	exceptionalZeroHash := Hash{}
	return &exceptionalZeroHash
}

// Finds the first index for which sn(index-1) < sn, and sn(index) >= sn.
// This might equal the length of the bin, ie one past the end, if sortNum is greater than any sortNums in the bin.
// This COULD be the first entry having that sortNum,
// or it COULD be the index after the interval where sortNum is crossed.
// If you need to know which of these is the case, you'll need to do a further check.
func (b *bin) findIndexBasedOnSortNum(sn sortNum, p *HashIndexingParams) int64 {
	if len(b.bytes) == 0 {
		return 0
	}
	startIndex := int64(0)
	endIndex := b.length(p.BytesPerBinEntry()) - 1

	// Special case: if just one in sequence, but sn > sn in sequence...
	if endIndex == startIndex {
		entry := b.getEntry(startIndex, p)
		_, snOnly := entry.getHashIndexSortNum(p)
		if sn > snOnly {
			return startIndex + 1
		}
	}

	for endIndex != startIndex {
		startIndex, endIndex = b.homeInOnSortNum(startIndex, endIndex, sn, p)
	}
	return startIndex
}

// This is the core of the "truncated secant" (I think) method for quickly searching a sorted sequence
// We are ultimately looking for the first index where sn(index-1) < sn, and sn(index) >= sn.
// If newStart==newEnd then we have found that point and you should no longer iterate.
// Note that if (endIndex+1, endIndex+1) is returned, then the index you want is beyond the searched sequence.
func (b *bin) homeInOnSortNum(startIndex int64, endIndex int64, sn sortNum, p *HashIndexingParams) (newStart int64, newEnd int64) {
	if startIndex == endIndex {
		panic("homeInOnSortNum(): you should have already finished the iteration")
	}
	entryStart := b.getEntry(startIndex, p)
	_, snStart := entryStart.getHashIndexSortNum(p)
	// First check for an easy out
	if sn <= snStart {
		return startIndex, startIndex // END
	}
	entryEnd := b.getEntry(endIndex, p)
	_, snEnd := entryEnd.getHashIndexSortNum(p)
	// Check for another easy out
	if sn > snEnd {
		return endIndex + 1, endIndex + 1 // END
	}
	if snEnd == snStart {
		panic("homeInOnSortNum(): this flat case should have been covered by one of the above easy outs")
	}
	if snStart > snEnd {
		panic("this bin should be sorted but it is not")
	}
	// Do most common case first for efficiency
	// We already know sn > snStart
	// We know snEnd >= sn
	if /* snStart < sn && */ snEnd > sn {
		// Easy out if indices are adjacent
		if endIndex-startIndex == 1 {
			return endIndex, endIndex // END
		}
		// Draw a straight line between (startIndex, snStart) and (endIndex, snEnd)
		slope := float64(snEnd-snStart) / float64(endIndex-startIndex)
		// Find the index where the line intersects the horizontal line at height sn
		floatIndex := float64(startIndex) + float64(sn-snStart)/slope
		midIndex := int64(math.Round(floatIndex))
		// Nudge if equal to limits
		if midIndex == startIndex {
			midIndex++
		} else if midIndex == endIndex {
			midIndex--
		}
		// Find actual sn on the "curve" at intIndex
		entryMid := b.getEntry(midIndex, p)
		_, snMid := entryMid.getHashIndexSortNum(p)
		if snMid < sn {
			return midIndex, endIndex
		} else if snMid > sn {
			return startIndex, midIndex
		} else {
			// Equal
			return midIndex, midIndex // END
		}
	} else {
		// sn must be equal to snEnd
		// We know snEnd > snStart
		// Keep stepping back one to find the first occurance of sn in this sequence of sn's (could of course
		// be a sequence of just one!)
		endSequenceIndex := endIndex
		snEndSequence := sn
		for snEndSequence == sn {
			endSequenceIndex--
			endSequenceEntry := b.getEntry(endSequenceIndex, p)
			_, snEndSequence = endSequenceEntry.getHashIndexSortNum(p)
		}
		return endSequenceIndex + 1, endSequenceIndex + 1 // END
	}
}
