package intarrayarray

import "testing"

func TestCreator(t *testing.T) {
	creator := NewConcreteMapStoreCreator("name", "folder", 3, 3, 3)
	creator.CreateMap()

	writable, _ := creator.OpenMap()

	writable.AppendToArray(1234, 12345)

	writable.AppendToArray(5678, 12345)
	writable.AppendToArray(5678, 67890)

	writable.AppendToArray(12345678, 123)
	writable.AppendToArray(12345678, 234)
	writable.AppendToArray(12345678, 345)

	writable.FlushFile()

	readable, _ := creator.OpenMapReadOnly()
	arr, _ := readable.GetArray(1234)
	if len(arr) != 1 {
		t.Fail()
	}
	arr, _ = readable.GetArray(5678)
	if len(arr) != 2 {
		t.Fail()
	}
	arr, _ = readable.GetArray(12345678)
	if len(arr) != 3 {
		t.Fail()
	}
	arr, _ = readable.GetArray(12)
	if len(arr) != 0 {
		t.Fail()
	}
}
