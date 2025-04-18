package records

import "testing"

// Test Fails!
func TestNewRecordFile(t *testing.T) {
	rd := NewRecordDescriptor()
	rd.AppendWordDescription("myField1", 1)
	rd.AppendWordDescription("myField2", 8)
	if rd.RecordSize() != 9 {
		t.Error("Ended up with wrong RecordSize()")
	}
}
