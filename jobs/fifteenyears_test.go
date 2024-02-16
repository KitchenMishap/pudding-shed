package jobs

import "testing"

func TestOneYearDelegated(t *testing.T) {
	SeveralYearsPrimaries(1, "delegated")
}

func TestOneYearUndelegated(t *testing.T) {
	SeveralYearsPrimaries(1, "separate files")
}
