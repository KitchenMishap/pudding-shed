package intarrayarray

import "testing"

func TestCreator(t *testing.T) {
	creator := NewConcreteMapStoreCreator("name", "folder", 3, 3, 3)
	if !creator.MapExists() {
		creator.CreateMap()
	}
	writable := creator.OpenMap()

	writable.AppendToArray(1234, 12345)

	writable.AppendToArray(5678, 12345)
	writable.AppendToArray(5678, 67890)

	writable.AppendToArray(12345678, 123)
	writable.AppendToArray(12345678, 234)
	writable.AppendToArray(12345678, 345)

	writable.FlushFile()

	readable := creator.OpenMapReadOnly()
	if len(readable.GetArray(1234)) != 1 {
		t.Fail()
	}
	if len(readable.GetArray(5678)) != 2 {
		t.Fail()
	}
	if len(readable.GetArray(12345678)) != 3 {
		t.Fail()
	}
	if len(readable.GetArray(12)) != 0 {
		t.Fail()
	}
}
