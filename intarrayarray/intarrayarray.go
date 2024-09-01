package intarrayarray

import (
	"encoding/binary"
	"os"
)

type IntArrayArray struct {
	elementByteSize int64
	arrayCount      int64
	elementTally    int64
	intToIntArray   map[int64][]int64
}

func NewIntArrayArray(arrayCount int64, elementByteSize int64) IntArrayArray {
	result := IntArrayArray{}
	result.arrayCount = arrayCount
	result.elementByteSize = elementByteSize
	result.elementTally = 0
	result.intToIntArray = make(map[int64][]int64)
	return result
}

func (iaa *IntArrayArray) AppendToArray(key int64, val int64) {
	originalArray, exists := iaa.intToIntArray[key]
	if exists {
		iaa.intToIntArray[key] = append(originalArray, val)
	} else {
		iaa.intToIntArray[key] = []int64{val}
	}
	iaa.elementTally++
}

func (iaa *IntArrayArray) GetArray(key int64) []int64 {
	originalArray, exists := iaa.intToIntArray[key]
	if exists {
		return originalArray
	} else {
		return []int64{} // Empty slice
	}
}

func (iaa *IntArrayArray) SaveFile(filepath string) error {
	// First write to a byte array
	bufferSize := int64(0)
	// Room for the elements
	bufferSize += iaa.elementTally * int64(iaa.elementByteSize)
	// Room for the arrayCounts (assume for now they take 9 bytes each, the biggest possible varint)
	bufferSize += int64(iaa.arrayCount) * int64(9)

	buffer := make([]byte, 0, bufferSize) // Zero sized slice of bytes with capacity for more

	converter := VarInt{}
	for i := int64(0); i < iaa.arrayCount; i++ {
		// What is the array's size if it exists?
		arrayLen := int64(0)
		originalArray, exists := iaa.intToIntArray[i]
		if exists {
			arrayLen = int64(len(originalArray))
		}

		// Size of one array as a varint
		converter.FromInt64(arrayLen)
		buffer = append(buffer, converter.ToBytes()...) // Concatenation of slices

		for j := int64(0); j < arrayLen; j++ {
			byt8 := [8]byte{}
			binary.LittleEndian.PutUint64(byt8[:], uint64(originalArray[j]))
			// Only want the first few LS bytes
			buffer = append(buffer, byt8[0:iaa.elementByteSize]...) // Concatenation of slices
		}
	}
	// Second write the file as a whole
	err := os.WriteFile(filepath, buffer, 0755)
	if err != nil {
		return err
	}
	return err
}

func (iaa *IntArrayArray) LoadFile(filepath string) error {
	iaa.elementTally = 0
	iaa.intToIntArray = make(map[int64][]int64)

	// First read the file as a whole
	buffer, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}

	// Then interpret it
	ptr := int64(0)
	converter := VarInt{}
	for i := int64(0); i < iaa.arrayCount; i++ {
		// read the array size as a varint (might be size zero)
		ptr += int64(converter.FromBytes(buffer[ptr:]))
		arraySize := converter.ToInt64()
		arr := make([]int64, arraySize)

		// Read the elements of the array
		byteSize := iaa.elementByteSize
		for j := int64(0); j < arraySize; j++ {
			byts := [8]byte{0, 0, 0, 0, 0, 0, 0, 0}
			// We just need the first few LS bytes
			copy(byts[0:byteSize], buffer[ptr:ptr+byteSize])
			element := int64(binary.LittleEndian.Uint64(byts[0:8]))
			arr[j] = element
			ptr += byteSize
		}
		iaa.elementTally += arraySize
		iaa.intToIntArray[i] = arr
	}
	return nil
}
