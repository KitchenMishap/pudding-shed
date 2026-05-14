package jobs

import (
	"testing"
)

func TestStreamBlockHashes(t *testing.T) {
	path := "E:\\Data\\NineYearsBinary"
	years := 9 // Note that 9 years are needed to test the Segwit era
	gbMem := 64
	threads := 30
	do1 := false
	do2 := true
	do3 := false
	json := false
	err := RunIntrinsic(path, json, "delegated", years, threads, gbMem, do1, do2, do3, 0, true)

	if err != nil {
		t.Error(err)
	}
}
