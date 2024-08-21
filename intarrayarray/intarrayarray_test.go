package intarrayarray

import "testing"

func TestIntArrayArray(t *testing.T) {
	helper(t, 10, 1, 0x100)
	helper(t, 1000, 5, 0x10000000000)
}

func helper(t *testing.T, arrayCount int64, elementByteSize int64, valueModulus int64) {
	// Construct
	iaa1 := NewIntArrayArray(arrayCount, elementByteSize)
	for i := int64(0); i < arrayCount; i += 9 {
		for j := int64(0); j < i; j++ {
			val := j % valueModulus
			iaa1.AppendToArray(i, val)
		}
	}

	// Save
	iaa1.SaveFile("Testing.iaa")

	// Load
	iaa2 := NewIntArrayArray(arrayCount, elementByteSize)
	iaa2.LoadFile("Testing.iaa")

	// Verify
	for i := int64(0); i < arrayCount; i += 9 {
		arr := iaa2.GetArray(i)
		for j := int64(0); j < i; j++ {
			expectedVal := j % valueModulus
			if arr[j] != expectedVal {
				t.Fail()
			}
		}
	}
}
