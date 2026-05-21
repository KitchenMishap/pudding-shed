package jobs

import (
	"testing"
)

func TestStreamBlockHashes(t *testing.T) {
	const path = "E:\\Data\\888888"
	const years = 17 // Note that 9 years are needed to test the Segwit era
	const gbMem = 64
	const threads = 40
	const do1 = false
	const do2 = true
	const do3 = false
	const json = false
	const useHandlesInterface = true
	const useAdaptivePassesPhase2 = true
	err := RunIntrinsic(path, json, "delegated", years, threads, gbMem, do1, do2, do3,
		0, true, useHandlesInterface, useAdaptivePassesPhase2)

	if err != nil {
		t.Error(err)
	}
}
