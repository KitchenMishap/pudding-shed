package wordfile

import (
	"github.com/KitchenMishap/pudding-shed/memfile"
	"log"
	"os"
)
import "testing"

func TestWordFile1(t *testing.T) {
	HelperWordfile(1, 0xFF, false, t)
	HelperWordfile(1, 0xFF, true, t)
}
func TestWordFile2(t *testing.T) {
	HelperWordfile(2, 0xFFFF, false, t)
	HelperWordfile(2, 0xFFFF, true, t)
}
func TestWordFile4(t *testing.T) {
	HelperWordfile(4, 0xFFFFFFFF, false, t)
	HelperWordfile(4, 0xFFFFFFFF, true, t)
}
func TestWordFile7(t *testing.T) {
	HelperWordfile(7, 0xFFFFFFFFFFFFFF, false, t)
	HelperWordfile(7, 0xFFFFFFFFFFFFFF, true, t)
}
func TestWordFile8(t *testing.T) {
	HelperWordfile(8, 0x7FFFFFFFFFFFFFFF, false, t)
	HelperWordfile(8, 0x7FFFFFFFFFFFFFFF, true, t)
}

// Tests wordfile with a given word size
// mask is a mask the size of the maximum word
func HelperWordfile(wordsize int64, mask int64, appendOptimized bool, t *testing.T) {
	// Create an empty file for testing
	f, err := os.Create("Temp_Testing\\WordfileTesting.int")
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			t.Error("Closing file should not give error")
		}
	}(f)

	var wf *WordFile
	if appendOptimized {
		ao, err := memfile.NewAppendOptimizedFile(f)
		if err != nil {
			t.Error("Should be able to create append optimized file from file")
		}
		// Treat it as a n-byte wordfile
		wf = NewWordFile(ao, wordsize, 0)
	} else {
		// Treat it as a n-byte wordfile
		wf = NewWordFile(f, wordsize, 0)
	}

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

	// Write a whole bunch of words and then read back the last one
	for i := 0; i < 16384; i++ {
		err := wf.WriteWordAt(123, int64(i))
		if err != nil {
			t.Error("Write should not give an error")
		}
	}
	val, err = wf.ReadWordAt(16383)
	if err != nil {
		t.Error("Should be able to read back last val")
	}
	if val != 123 {
		t.Error("Read back something else after writing 123")
	}
}
