package jobs

import "testing"

func TestStreamBlockHashes(t *testing.T) {
	path := "E:\\Data\\TwoYearsBinary"
	years := 2
	gbMem := 16

	err := RunIntrinsic(path, years, gbMem, true, true, true)

	if err != nil {
		t.Error(err)
	}
}
