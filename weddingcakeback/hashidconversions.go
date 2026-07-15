package weddingcakeback

// This file contains conversions between various ids that refer to presented hashes.
// Do not use casts between these types elsewhere! Except in tests.

// GlobalPiType uses -1 to mean "no match". Zero corresponds to the first ever presented hash.
// TierTopIndex has no special value for "no match". Zero corresponds to the supplied offset.
// SingleTreePi uses 0 to mean "no match". One corresponds to the supplied offset
// HashIndexId uses 0 to mean "no match". One corresponds to the supplied offset.

func GlobalPiFromSingleTreePi(singleTreePi SingleTreePiType) GlobalPiType {
	return GlobalPiType(singleTreePi) - 1
}

//	func GlobalPiFromSingleTreePi(singleTreePi SingleTreePiType, offset GlobalPiType) GlobalPiType {
//		return GlobalPiType(singleTreePi) - 1 + offset
//	}
func GlobalPiWithinRange(globalPi GlobalPiType, offset GlobalPiType, count int) bool {
	return globalPi >= offset && globalPi < offset+GlobalPiType(count)
}
func GlobalPiFromUint64(globalPi uint64) GlobalPiType {
	return GlobalPiType(globalPi)
}
func GlobalPiFromTierTopIndex(ttIndex TierTopIndex, offset GlobalPiType) GlobalPiType {
	return offset + GlobalPiType(ttIndex)
}
func TierTopIndexFromGlobalPi(globalPi GlobalPiType, offset GlobalPiType) TierTopIndex {
	return TierTopIndex(globalPi - offset)
}
func SingleTreePiFromGlobalPi(globalPi GlobalPiType) SingleTreePiType {
	return SingleTreePiType(globalPi + 1)
}

//	func SingleTreePiFromGlobalPi(globalPi GlobalPiType, offset GlobalPiType) SingleTreePiType {
//		return SingleTreePiType(globalPi - offset + 1)
//	}
func HashIndexIdFromGlobalPi(global GlobalPiType, offset GlobalPiType) HashIndexIdType {
	// The value reserved for "no match"
	if global == GlobalPiNoMatch {
		return HashIndexIdNoMatch
	}
	if global < offset {
		panic("Global presentation index lower than first presentation index")
	}
	// firstGlobalPi maps to hashIndexId 1
	return HashIndexIdType(global - offset + 1)
}

func GlobalPiFromHashIndexId(hashIndexId HashIndexIdType, offset GlobalPiType) GlobalPiType {
	// The value reserved for "no match"
	if hashIndexId == HashIndexIdNoMatch {
		return GlobalPiNoMatch
	}
	// HashIndexIdType 1 maps to firstGlobalPi
	return GlobalPiType(hashIndexId) - 1 + offset
}
