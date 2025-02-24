package indexedhashes3

import (
	"math"
)

// A bin is a slice of binEntryBytes.
// It is the way a bin is held whilst in memory.
// Note that it is not directly a whole slice of bytes... you'd have to do some rearranging to get such a thing.

type bin []binEntryBytes

func newEmptyBin() bin {
	result := bin(make([]binEntryBytes, 0))
	return result
}

func (b *bin) insertBinEntry(sn sortNum, hi hashIndex, th *truncatedHash, p *HashIndexingParams) int64 {
	newEntry := newBinEntryBytes(th, hi, sn, p)
	insertionPoint := b.findIndexBasedOnSortNum(sn, p)

	// For some reason I couldn't get slices.Insert() to compile
	// Alternative insertion algorithm:
	if len(*b) == 0 {
		// Special case if empty
		*b = append(*b, newEntry)
	} else {
		// First append a copy of the last item
		*b = append(*b, (*b)[len(*b)-1])

		// Then copy everything forward by one
		copy((*b)[insertionPoint+1:], (*b)[insertionPoint:len(*b)-1])

		// Finally drop in the inserted value
		(*b)[insertionPoint] = newEntry
	}
	return insertionPoint
}

func (b *bin) lookupByHash(th *truncatedHash, sn sortNum, p *HashIndexingParams) hashIndex {
	firstMatchingIndex := b.findIndexBasedOnSortNum(sn, p)
	// Search sequentially from here until everything matches or sn does not match
	for index := firstMatchingIndex; index < int64(len(*b)); index++ {
		hiFound, snFound := (*b)[index].getHashIndexSortNum(p)
		if snFound != sn {
			return -1 // hash not present in bin
		}
		thFound := (*b)[index].getTruncatedHash()
		if thFound.equals(th) {
			//fmt.Println("Found in ", index, "-th bin entry")
			return hiFound
		}
	}
	return -1 // hash not present in bin
}

// Because there are a (very) few repeated transaction hashes on the blockchain, our method
// currently cannot find some of those repeated hashes by index. In that circumstance,
// we return a hash of all zeroes!
func (b *bin) lookupByIndex(hi hashIndex, bn binNum, p *HashIndexingParams) *Hash {
	// No shortcut. Sequential search
	for index := 0; index < len(*b); index++ {
		hiFound, snFound := (*b)[index].getHashIndexSortNum(p)
		if hiFound == hi {
			th := (*b)[index].getTruncatedHash()
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
	if len(*b) == 0 {
		return 0
	}
	startIndex := int64(0)
	endIndex := int64(len(*b) - 1)

	// Special case: if just one in sequence, but sn > sn in sequence...
	if endIndex == startIndex {
		_, snOnly := (*b)[startIndex].getHashIndexSortNum(p)
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
	_, snStart := (*b)[startIndex].getHashIndexSortNum(p)
	// First check for an easy out
	if sn <= snStart {
		return startIndex, startIndex // END
	}
	_, snEnd := (*b)[endIndex].getHashIndexSortNum(p)
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
		_, snMid := (*b)[midIndex].getHashIndexSortNum(p)
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
			_, snEndSequence = (*b)[endSequenceIndex].getHashIndexSortNum(p)
		}
		return endSequenceIndex + 1, endSequenceIndex + 1 // END
	}
}
