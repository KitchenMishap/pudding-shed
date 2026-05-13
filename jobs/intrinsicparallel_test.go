package jobs

import (
	"testing"
)

func TestStreamBlockHashes(t *testing.T) {
	path := "E:\\Data\\TwoYearsBinary"
	years := 2
	gbMem := 16
	threads := 30
	do1 := true
	do2 := true
	do3 := true
	err := RunIntrinsic(path, "delegated", years, threads, gbMem, do1, do2, do3, 0)

	if err != nil {
		t.Error(err)
	}
}
