package weddingcake

import (
	"fmt"
	"testing"
)

/*
func TestExampleTreesNfgs(t *testing.T) {

	sc := StoreConfig{
		HashLength:              32,
		ReassuranceBytesCount:   2,
		NodeFormatSpecsPerLevel: 10,
	}

	count := 27_000
	fmt.Printf("Tree of %d hashes...\n", count)
	presentationArray := make([]ShallowTreeHash, count)
	for i := range count {
		hash := helperRandomHash(sc.HashLength)
		presentationArray[i].Hash = hash
		presentationArray[i].PresentationIndex = int64(i)
	}
	st := sc.GenerateShallowTree(presentationArray)
	ts := st.CountLevelShapes()
	sc.ChooseNodeFormatSpecsForTreeShape(ts)

	count = 27_000 * 256
	fmt.Printf("Tree of %d hashes...\n", count)
	presentationArray = make([]ShallowTreeHash, count)
	for i := range count {
		hash := helperRandomHash(sc.HashLength)
		presentationArray[i].Hash = hash
		presentationArray[i].PresentationIndex = int64(i)
	}
	st = sc.GenerateShallowTree(presentationArray)
	ts = st.CountLevelShapes()
	sc.ChooseNodeFormatSpecsForTreeShape(ts)
}*/

func TestDesignTreeFormat(t *testing.T) {

	sc := StoreConfig{
		HashLength:              32,
		ReassuranceBytesCount:   2,
		NodeFormatSpecsPerLevel: 10,
		NodeIdConfig:            ID16[NodeIdType]{},
		HashIndexIdConfig:       ID16[HashIndexIdType]{},
	}

	count := 27_000
	fmt.Printf("Tree of %d hashes...\n", count)
	presentationArray := make([]ShallowTreeHash, count)
	for i := range count {
		hash := helperRandomHash(sc.HashLength)
		presentationArray[i].Hash = hash
		presentationArray[i].PresentationIndex = int64(i)
	}
	st := GenerateShallowTree(presentationArray, sc.HashLength, sc.ReassuranceBytesCount)
	_ = sc.DesignTreeFormat(st)

	count = 27_000 * 256
	/* This would produce too many nodes for uint16 node ids
	fmt.Printf("Tree of %d hashes...\n", count)
	presentationArray = make([]ShallowTreeHash, count)
	for i := range count {
		hash := helperRandomHash(hashLength)
		presentationArray[i].Hash = hash
		presentationArray[i].PresentationIndex = int64(i)
	}
	st = GenerateShallowTree(presentationArray, hashLength)
	_ = DesignTreeFormat(st)*/
}
