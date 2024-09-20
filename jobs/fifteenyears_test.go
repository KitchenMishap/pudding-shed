package jobs

import "testing"

func TestOneYearDelegated(t *testing.T) {
	SeveralYearsParallel(4, "delegated")
}

func TestOneYearUndelegated(t *testing.T) {
	SeveralYearsPrimaries(1, "separate files")
}
