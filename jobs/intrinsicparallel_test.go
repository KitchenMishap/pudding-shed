package jobs

import (
	"testing"
)

func TestStreamBlockHashes(t *testing.T) {
	path := "E:\\Data\\NineYearsNewJson"
	years := 9 // Note that 9 years are needed to test the Segwit era
	gbMem := 16
	threads := 30
	do1 := true
	do2 := true
	do3 := true
	json := true
	err := RunIntrinsic(path, json, "delegated", years, threads, gbMem, do1, do2, do3, 0)

	if err != nil {
		t.Error(err)
	}
}
