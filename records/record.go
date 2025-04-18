package records

import (
	"encoding/binary"
	"errors"
)

type Record []byte

type fieldDescriptor struct {
	byteOffset int
	wordSize   int
}

type Descriptor struct {
	m map[string]fieldDescriptor
}

// Compiler check that implements
var _ RecordDescriptor = (*Descriptor)(nil)

func NewRecordDescriptor() *Descriptor {
	result := Descriptor{}
	result.m = make(map[string]fieldDescriptor)
	return &result
}

func (rd *Descriptor) RecordSize() int {
	result := 0
	for _, v := range rd.m {
		result += v.wordSize
	}
	return result
}

func (rd *Descriptor) AppendWordDescription(name string, byteSize int) {
	recordSizeBefore := rd.RecordSize()
	rd.m[name] = fieldDescriptor{
		byteOffset: recordSizeBefore,
		wordSize:   byteSize,
	}
}

func (rd *Descriptor) FieldDescriptor(fieldName string) (offset int, byteSize int, err error) {
	fd, success := rd.m[fieldName]
	if !success {
		return -1, -1, errors.New("Field not found: " + fieldName)
	}
	return fd.byteOffset, fd.wordSize, nil
}

func (r *Record) GetWord(d RecordDescriptor, fieldName string) (word uint64, err error) {
	byteOffset, byteSize, err := d.FieldDescriptor(fieldName)
	if err != nil {
		return 0, err
	}

	eightBytes := [8]byte{}
	for i := 0; i < byteSize; i++ {
		eightBytes[i] = (*r)[byteOffset+i]
	}
	return binary.LittleEndian.Uint64(eightBytes[:]), nil
}

func (r *Record) PutWord(d RecordDescriptor, fieldName string, value uint64) error {
	byteOffset, byteSize, err := d.FieldDescriptor(fieldName)
	if err != nil {
		return err
	}
	eightBytes := [8]byte{}
	binary.LittleEndian.PutUint64(eightBytes[:], value)
	for i := 0; i < byteSize; i++ {
		(*r)[byteOffset+i] = eightBytes[i]
	}
	return nil
}
