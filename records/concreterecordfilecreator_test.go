package records

import "testing"

// Test Fails!
func Test_ConcreteRecordFileCreator(t *testing.T) {
	// This record type has a byte and an int64
	rd := NewRecordDescriptor()
	rd.AppendWordDescription("myField1", 1)
	rd.AppendWordDescription("myField2", 8)
	if rd.RecordSize() != 9 {
		t.Error("Recordsize should be 9")
		return
	}

	// RecordFileCreator
	rfc := NewConcreteRecordFileCreator("recordfile", "Temp_Testing", rd)
	// Create one record of zeroes
	rfc.CreateRecordFileFilledZeros(1)

	// Is it zeroes?
	rf, err := rfc.OpenRecordFile()
	rec, err := rf.ReadRecordAt(0)
	if err != nil {
		t.Error(err)
		return
	}
	val, err := rec.GetWord(rd, "myField1")
	if err != nil {
		t.Error(err)
		return
	}
	if val != 0 {
		t.Error("field should be zero")
		return
	}
	val, err = rec.GetWord(rd, "myField2")
	if err != nil {
		t.Error(err)
		return
	}
	if val != 0 {
		t.Error("field should be zero")
		return
	}

	// Put together one record, with 255 and 65536
	err = rec.PutWord(rd, "myField1", 255)
	if err != nil {
		t.Error(err)
		return
	}
	err = rec.PutWord(rd, "myField2", 65536)
	if err != nil {
		t.Error(err)
		return
	}

	// Write the 255,36635 to record 0
	err = rf.WriteRecordAt(rec, 0)
	if err != nil {
		t.Error(err)
		return
	}
	// Write the 255,36635 to record 1
	err = rf.WriteRecordAt(rec, 1)
	if err != nil {
		t.Error(err)
		return
	}

	// Read back from record 1
	rec2, err := rf.ReadRecordAt(1)
	if err != nil {
		t.Error(err)
		return
	}
	val, err = rec2.GetWord(rd, "myField1")
	if err != nil {
		t.Error(err)
		return
	}
	if val != 255 {
		t.Error("field should be 255")
		return
	}
	val, err = rec2.GetWord(rd, "myField2")
	if err != nil {
		t.Error(err)
		return
	}
	if val != 65536 {
		t.Error("field should be 65536")
		return
	}

	rf.Close()
}
