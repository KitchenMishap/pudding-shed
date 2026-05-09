package jobs

import "testing"

func TestStreamBlockHashes(t *testing.T) {
	path := "E:\\Data\\TwoYearsBinary"
	years := 2
	blocks := int64(100888)
	gbMem := 64

	err := RunIntrinsic(path, years, blocks, gbMem, true, true, true)

	if err != nil {
		t.Error(err)
	}
}
