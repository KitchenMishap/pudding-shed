package records

import (
	"errors"
	"github.com/KitchenMishap/pudding-shed/memfile"
	"log"
)

// A RecordFile is very similar to a WordFile
type RecordFile struct {
	file        memfile.AppendableLookupFile
	recordSize  int
	recordCount int64
}

// Compiler check that implements
var _ ReadWriteAtRecordCounter = (*RecordFile)(nil)

func NewRecordFile(file memfile.AppendableLookupFile, rd RecordDescriptor, recordCount int64) *RecordFile {
	p := new(RecordFile)
	p.file = file
	p.recordSize = rd.RecordSize()
	p.recordCount = recordCount
	return p
}

func (rf *RecordFile) ReadRecordAt(off int64) (Record, error) {
	rf.Sync()
	bytes := make([]byte, rf.recordSize)
	_, err := rf.file.ReadAt(bytes[:], off*int64(rf.recordSize))
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func (rf *RecordFile) WriteRecordAt(val Record, off int64) error {
	if len(val) != int(rf.recordSize) {
		return errors.New("val length does not equal recordSize")
	}
	_, err := rf.file.WriteAt(val, off*int64(rf.recordSize))
	if err != nil {
		return err
	}
	if off+1 > rf.recordCount {
		rf.recordCount = off + 1
	}
	return nil
}

func (rf *RecordFile) CountRecords() (records int64, err error) {
	return rf.recordCount, nil
}

func (rf *RecordFile) Close() error {
	err := rf.file.Close()
	if err != nil {
		log.Println(err)
		log.Println("Close(): Could not call Close()")
	}
	return err
}

func (rf *RecordFile) Sync() error {
	return rf.file.Sync()
}

func (rf *RecordFile) RecordSize() int {
	return rf.recordSize
}
