package wordfile

import (
	"log"
	"os"
)
import "testing"

func TestWordFile1(t *testing.T) {
	HelperWordfile(1, 0xFF, t)
}
func TestWordFile2(t *testing.T) {
	HelperWordfile(2, 0xFFFF, t)
}
func TestWordFile4(t *testing.T) {
	HelperWordfile(4, 0xFFFFFFFF, t)
}
func TestWordFile7(t *testing.T) {
	HelperWordfile(7, 0xFFFFFFFFFFFFFF, t)
}
func TestWordFile8(t *testing.T) {
	HelperWordfile(8, 0x7FFFFFFFFFFFFFFF, t)
}

// Tests wordfile with a given word size
// mask is a mask the size of the maximum word
func HelperWordfile(wordsize int64, mask int64, t *testing.T) {
	// Create an empty file for testing
	f, err := os.Create("Temp_Testing\\WordfileTesting.int")
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			t.Error("Closing file should not give error")
		}
	}(f)
	// Treat it as a n-byte wordfile
	wf := NewWordFile(f, wordsize)

	// Check it's empty
	cf, err := wf.CountWords()
	if err != nil {
		t.Error("CountWords should not give error on valid file")
	}
	if cf != 0 {
		t.Error("CountWords on empty file should give zero")
	}
	// Check a read at 0 gives an error
	log.Println("Note: An EOF here is a PASS")
	rr, err := wf.ReadWordAt(0)
	if err == nil {
		t.Error("ReadWordAt(0) on empty file should give error")
	}

	// Write at word zero and read it back
	val := 0x0000000000000000 & mask
	err = wf.WriteWordAt(val, 0)
	if err != nil {
		t.Error("Write at offset zero should not give error")
	}
	rr, err = wf.ReadWordAt(0)
	if err != nil {
		t.Error("Read after write should not give error")
	}
	if rr != val {
		t.Error("Read of written zero should be zero")
	}

	// Write at word two and check that word one reads back as zero even though not written
	val = 0x2222222222222222 & mask
	err = wf.WriteWordAt(val, 2)
	if err != nil {
		t.Error("Should be able to write past end of file without error")
	}
	cf, _ = wf.CountWords()
	if cf != 3 {
		t.Error("CountWords() should give three after writing word 2")
	}
	rr, _ = wf.ReadWordAt(1)
	if rr != 0 {
		t.Error("Writing at word 2 should leave a zero at the skipped word one")
	}

	// Go back and write to word 1
	val = 0x1111111111111111 & mask
	err = wf.WriteWordAt(val, 1)
	if err != nil {
		t.Error("Write to middle of file should not give error")
	}
	// Check all three words
	r0, err := wf.ReadWordAt(0)
	r1, err := wf.ReadWordAt(1)
	r2, err := wf.ReadWordAt(2)
	if r0 != 0x0000000000000000&mask {
		t.Error("Expected to read back 0x00")
	}
	if r1 != 0x1111111111111111&mask {
		t.Error("Expected to read back 0x11")
	}
	if r2 != 0x2222222222222222&mask {
		t.Error("Expected to read back 0x22")
	}
}
