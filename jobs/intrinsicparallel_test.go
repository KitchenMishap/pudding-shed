package jobs

import "testing"

func TestStreamBlockHashes(t *testing.T) {
	path := "E:\\Data\\TwoYearsBinary"
	years := 2
	gbMem := 16
	threads := 10 // ToDo low for now whilst long task is running
	err := RunIntrinsic(path, "delegated", years, threads, gbMem, true, true, true)

	if err != nil {
		t.Error(err)
	}
}
