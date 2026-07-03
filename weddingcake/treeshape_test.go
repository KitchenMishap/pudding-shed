package weddingcake

import (
	"fmt"
	"testing"
)

func TestTreeShapeFiveHundredThousand(t *testing.T) {
	const hashLength = 32
	const reassuranceBytesCount = 2
	const count = 500_000
	presentationArray := make([]ShallowTreeHash, count)
	for i := range count {
		hash := helperRandomHash(hashLength)
		presentationArray[i].Hash = hash
		presentationArray[i].PresentationIndex = int64(i)
	}

	st := GenerateShallowTree(presentationArray, hashLength, reassuranceBytesCount)
	ts := st.CountLevelShapes()
	fmt.Printf("For %d hashes...\n", count)
	ts.Print()
}
