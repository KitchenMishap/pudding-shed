package jobs

import "testing"

func TestStreamBlockHashes(t *testing.T) {
	path := "E:\\Data\\TenYearsBinary"
	years := 10
	gbMem := 64
	threads := 30
	err := RunIntrinsic(path, "delegated", years, threads, gbMem, true, true, true)

	if err != nil {
		t.Error(err)
	}
}
