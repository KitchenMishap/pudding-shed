package jobs

import "testing"

func TestFourYearsDelegated(t *testing.T) {
	SeveralYearsParallel(7, "delegated")
}

func TestOneYearUndelegated(t *testing.T) {
	SeveralYearsPrimaries(1, "separate files")
}
